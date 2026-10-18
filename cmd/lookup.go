package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/suer/first-task-cli/internal/firestore"
	"github.com/suer/first-task-cli/internal/model"
)

func loadTags(ctx context.Context, c *firestore.Client) ([]model.Tag, error) {
	docs, err := c.List(ctx, model.TagsCollection, "")
	if err != nil {
		return nil, err
	}
	tags := make([]model.Tag, 0, len(docs))
	for _, d := range docs {
		tags = append(tags, model.TagFromDocument(d))
	}
	return tags, nil
}

func loadProjects(ctx context.Context, c *firestore.Client, done bool) ([]model.Project, error) {
	collection := model.ProjectsCollection
	if done {
		collection = model.CompletedProjectsCollection
	}
	docs, err := c.List(ctx, collection, "")
	if err != nil {
		return nil, err
	}
	projects := make([]model.Project, 0, len(docs))
	for _, d := range docs {
		projects = append(projects, model.ProjectFromDocument(d))
	}
	return projects, nil
}

func loadTasks(ctx context.Context, c *firestore.Client, done bool) ([]model.Task, error) {
	collection := model.TasksCollection
	if done {
		collection = model.CompletedTasksCollection
	}
	docs, err := c.List(ctx, collection, "displayOrder")
	if err != nil {
		return nil, err
	}
	tasks := make([]model.Task, 0, len(docs))
	for _, d := range docs {
		tasks = append(tasks, model.TaskFromDocument(d))
	}
	return tasks, nil
}

func resolveTagIDs(tags []model.Tag, names []string) ([]string, error) {
	ids := make([]string, 0, len(names))
	for _, name := range names {
		var id string
		for _, t := range tags {
			if strings.EqualFold(t.Name, name) {
				id = t.ID
				break
			}
		}
		if id == "" {
			return nil, fmt.Errorf("tag %q not found", name)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// resolveProjectID accepts an id (or unique prefix) or an exact title; an empty ref means no project.
func resolveProjectID(projects []model.Project, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	for _, p := range projects {
		if strings.EqualFold(p.Title, ref) {
			return p.ID, nil
		}
	}
	ids := make([]string, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	i, err := matchID(ids, ref)
	if err != nil {
		return "", fmt.Errorf("project %q not found", ref)
	}
	return projects[i].ID, nil
}
