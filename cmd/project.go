package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/suer/first-task-cli/internal/model"
)

func newProjectCmd() *cobra.Command {
	project := &cobra.Command{Use: "project", Short: "Manage projects"}

	var done bool
	list := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			projects, err := loadProjects(cmd.Context(), c, done)
			if err != nil {
				return err
			}
			if jsonOutput {
				return printJSON(projects)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTITLE\tSTART\tDUE")
			for _, p := range projects {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", shortID(p.ID), p.Title, formatDate(p.StartDate), formatDate(p.DueDate))
			}
			return w.Flush()
		},
	}
	list.Flags().BoolVar(&done, "done", false, "list completed projects")
	project.AddCommand(list)

	project.AddCommand(&cobra.Command{
		Use:   "add TITLE",
		Short: "Create a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			now := time.Now()
			p := model.Project{ID: newID(), Title: args[0], CreatedAt: &now, UpdatedAt: &now}
			if err := c.Create(cmd.Context(), model.ProjectsCollection, p.Document()); err != nil {
				return err
			}
			fmt.Println(p.ID)
			return nil
		},
	})

	project.AddCommand(&cobra.Command{
		Use:   "rm ID",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			projects, err := loadProjects(cmd.Context(), c, false)
			if err != nil {
				return err
			}
			ids := make([]string, len(projects))
			for i, p := range projects {
				ids[i] = p.ID
			}
			i, err := matchID(ids, args[0])
			if err != nil {
				return err
			}
			return c.Delete(cmd.Context(), model.ProjectsCollection, projects[i].ID)
		},
	})

	return project
}
