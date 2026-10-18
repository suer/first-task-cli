package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/suer/first-task-cli/internal/firestore"
	"github.com/suer/first-task-cli/internal/model"
)

func newTaskCmd() *cobra.Command {
	task := &cobra.Command{Use: "task", Short: "Manage tasks"}
	task.AddCommand(newTaskListCmd(), newTaskShowCmd(), newTaskAddCmd(), newTaskEditCmd(), newTaskDoneCmd(), newTaskUndoneCmd(), newTaskRmCmd())
	return task
}

// findTask looks the task up in both the active and completed collections.
func findTask(ctx context.Context, c *firestore.Client, ref string) (model.Task, string, error) {
	var all []model.Task
	var collections []string
	for _, done := range []bool{false, true} {
		tasks, err := loadTasks(ctx, c, done)
		if err != nil {
			return model.Task{}, "", err
		}
		collection := model.TasksCollection
		if done {
			collection = model.CompletedTasksCollection
		}
		for _, t := range tasks {
			all = append(all, t)
			collections = append(collections, collection)
		}
	}
	ids := make([]string, len(all))
	for i, t := range all {
		ids[i] = t.ID
	}
	i, err := matchID(ids, ref)
	if err != nil {
		return model.Task{}, "", err
	}
	return all[i], collections[i], nil
}

func newTaskListCmd() *cobra.Command {
	var (
		done      bool
		tagName   string
		project   string
		query     string
		dueBefore string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List and search tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := newClient(ctx)
			if err != nil {
				return err
			}
			tasks, err := loadTasks(ctx, c, done)
			if err != nil {
				return err
			}
			tags, err := loadTags(ctx, c)
			if err != nil {
				return err
			}
			projects, err := loadProjects(ctx, c, false)
			if err != nil {
				return err
			}

			if tagName != "" {
				ids, err := resolveTagIDs(tags, []string{tagName})
				if err != nil {
					return err
				}
				tasks = filterTasks(tasks, func(t model.Task) bool { return contains(t.TagIDs, ids[0]) })
			}
			if project != "" {
				id, err := resolveProjectID(projects, project)
				if err != nil {
					return err
				}
				tasks = filterTasks(tasks, func(t model.Task) bool { return t.ProjectID == id })
			}
			if query != "" {
				q := strings.ToLower(query)
				tasks = filterTasks(tasks, func(t model.Task) bool {
					return strings.Contains(strings.ToLower(t.Title), q) || strings.Contains(strings.ToLower(t.Memo), q)
				})
			}
			if dueBefore != "" {
				limit, err := parseDate(dueBefore)
				if err != nil {
					return err
				}
				tasks = filterTasks(tasks, func(t model.Task) bool { return t.DueDate != nil && !t.DueDate.After(limit.AddDate(0, 0, 1)) })
			}

			if jsonOutput {
				return printJSON(tasks)
			}
			return printTasks(tasks, tags, projects)
		},
	}
	cmd.Flags().BoolVar(&done, "done", false, "list completed tasks")
	cmd.Flags().StringVar(&tagName, "tag", "", "filter by tag name")
	cmd.Flags().StringVar(&project, "project", "", "filter by project title or id")
	cmd.Flags().StringVarP(&query, "query", "q", "", "search title and memo")
	cmd.Flags().StringVar(&dueBefore, "due-before", "", "tasks due on or before this date")
	return cmd
}

func newTaskShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ID",
		Short: "Show a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := newClient(ctx)
			if err != nil {
				return err
			}
			t, _, err := findTask(ctx, c, args[0])
			if err != nil {
				return err
			}
			if jsonOutput {
				return printJSON(t)
			}
			tags, err := loadTags(ctx, c)
			if err != nil {
				return err
			}
			projects, err := loadProjects(ctx, c, false)
			if err != nil {
				return err
			}
			fmt.Printf("ID:       %s\n", t.ID)
			fmt.Printf("Title:    %s\n", t.Title)
			fmt.Printf("Memo:     %s\n", t.Memo)
			fmt.Printf("Project:  %s\n", projectTitle(projects, t.ProjectID))
			fmt.Printf("Tags:     %s\n", strings.Join(tagNames(tags, t.TagIDs), ", "))
			fmt.Printf("Start:    %s\n", formatDate(t.StartDate))
			fmt.Printf("Due:      %s\n", formatDate(t.DueDate))
			fmt.Printf("Done:     %s\n", formatDate(t.CompletedAt))
			fmt.Printf("Created:  %s\n", formatDate(t.CreatedAt))
			fmt.Printf("Updated:  %s\n", formatDate(t.UpdatedAt))
			return nil
		},
	}
}

func newTaskAddCmd() *cobra.Command {
	var (
		memo    string
		project string
		tagList []string
		start   string
		due     string
	)
	cmd := &cobra.Command{
		Use:   "add TITLE",
		Short: "Create a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := newClient(ctx)
			if err != nil {
				return err
			}

			now := time.Now()
			t := model.Task{ID: newID(), Title: args[0], Memo: memo, CreatedAt: &now, UpdatedAt: &now}

			if project != "" {
				projects, err := loadProjects(ctx, c, false)
				if err != nil {
					return err
				}
				if t.ProjectID, err = resolveProjectID(projects, project); err != nil {
					return err
				}
			}
			if len(tagList) > 0 {
				tags, err := loadTags(ctx, c)
				if err != nil {
					return err
				}
				if t.TagIDs, err = resolveTagIDs(tags, tagList); err != nil {
					return err
				}
			}
			if start != "" {
				d, err := parseDate(start)
				if err != nil {
					return err
				}
				t.StartDate = &d
			}
			if due != "" {
				d, err := parseDate(due)
				if err != nil {
					return err
				}
				t.DueDate = &d
			}

			existing, err := loadTasks(ctx, c, false)
			if err != nil {
				return err
			}
			for _, e := range existing {
				if e.DisplayOrder >= t.DisplayOrder {
					t.DisplayOrder = e.DisplayOrder + 1
				}
			}

			if err := c.Create(ctx, model.TasksCollection, t.Document()); err != nil {
				return err
			}
			fmt.Println(t.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&memo, "memo", "", "memo")
	cmd.Flags().StringVar(&project, "project", "", "project title or id")
	cmd.Flags().StringSliceVar(&tagList, "tag", nil, "tag name (repeatable)")
	cmd.Flags().StringVar(&start, "start", "", "start date (YYYY-MM-DD, today, tomorrow)")
	cmd.Flags().StringVar(&due, "due", "", "due date (YYYY-MM-DD, today, tomorrow)")
	return cmd
}

func newTaskEditCmd() *cobra.Command {
	var (
		title   string
		memo    string
		project string
		tagList []string
		start   string
		due     string
	)
	cmd := &cobra.Command{
		Use:   "edit ID",
		Short: "Update a task (pass an empty string to clear --project, --start or --due)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := newClient(ctx)
			if err != nil {
				return err
			}
			t, collection, err := findTask(ctx, c, args[0])
			if err != nil {
				return err
			}

			fields := cmd.Flags()
			if fields.Changed("title") {
				t.Title = title
			}
			if fields.Changed("memo") {
				t.Memo = memo
			}
			if fields.Changed("project") {
				projects, err := loadProjects(ctx, c, false)
				if err != nil {
					return err
				}
				if t.ProjectID, err = resolveProjectID(projects, project); err != nil {
					return err
				}
			}
			if fields.Changed("tag") {
				tags, err := loadTags(ctx, c)
				if err != nil {
					return err
				}
				if t.TagIDs, err = resolveTagIDs(tags, tagList); err != nil {
					return err
				}
			}
			if fields.Changed("start") {
				if t.StartDate, err = optDate(start); err != nil {
					return err
				}
			}
			if fields.Changed("due") {
				if t.DueDate, err = optDate(due); err != nil {
					return err
				}
			}
			now := time.Now()
			t.UpdatedAt = &now

			return c.Patch(ctx, collection, t.Document())
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "title")
	cmd.Flags().StringVar(&memo, "memo", "", "memo")
	cmd.Flags().StringVar(&project, "project", "", "project title or id")
	cmd.Flags().StringSliceVar(&tagList, "tag", nil, "replace tags with these names (repeatable)")
	cmd.Flags().StringVar(&start, "start", "", "start date")
	cmd.Flags().StringVar(&due, "due", "", "due date")
	return cmd
}

func newTaskDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done ID",
		Short: "Mark a task as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return toggleTask(cmd.Context(), args[0], true)
		},
	}
}

func newTaskUndoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "undone ID",
		Short: "Move a completed task back to active",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return toggleTask(cmd.Context(), args[0], false)
		},
	}
}

func toggleTask(ctx context.Context, ref string, done bool) error {
	c, err := newClient(ctx)
	if err != nil {
		return err
	}
	t, collection, err := findTask(ctx, c, ref)
	if err != nil {
		return err
	}
	target := model.TasksCollection
	if done {
		target = model.CompletedTasksCollection
	}
	if collection == target {
		return fmt.Errorf("task %s is already %s", shortID(t.ID), map[bool]string{true: "completed", false: "active"}[done])
	}

	now := time.Now()
	t.UpdatedAt = &now
	if done {
		t.CompletedAt = &now
	} else {
		t.CompletedAt = nil
	}
	return c.Move(ctx, collection, target, t.Document())
}

func newTaskRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm ID",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			c, err := newClient(ctx)
			if err != nil {
				return err
			}
			t, collection, err := findTask(ctx, c, args[0])
			if err != nil {
				return err
			}
			if err := c.Delete(ctx, collection, t.ID); err != nil {
				return err
			}
			fmt.Printf("Deleted %s %q\n", shortID(t.ID), t.Title)
			return nil
		},
	}
}

func optDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	d, err := parseDate(s)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func filterTasks(tasks []model.Task, keep func(model.Task) bool) []model.Task {
	out := tasks[:0:0]
	for _, t := range tasks {
		if keep(t) {
			out = append(out, t)
		}
	}
	return out
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func projectTitle(projects []model.Project, id string) string {
	for _, p := range projects {
		if p.ID == id {
			return p.Title
		}
	}
	return "-"
}

func tagNames(tags []model.Tag, ids []string) []string {
	var names []string
	for _, t := range tags {
		if contains(ids, t.ID) {
			names = append(names, t.Name)
		}
	}
	sort.Strings(names)
	return names
}

func printTasks(tasks []model.Task, tags []model.Tag, projects []model.Project) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tTITLE\tDUE\tPROJECT\tTAGS")
	for _, t := range tasks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", shortID(t.ID), t.Title, formatDate(t.DueDate), projectTitle(projects, t.ProjectID), strings.Join(tagNames(tags, t.TagIDs), ","))
	}
	return w.Flush()
}
