package model

import (
	"time"

	"github.com/suer/first-task-cli/internal/firestore"
)

const (
	TasksCollection             = "tasks"
	CompletedTasksCollection    = "completed-tasks"
	ProjectsCollection          = "projects"
	CompletedProjectsCollection = "completed-projects"
	TagsCollection              = "tags"

	TodayTagKind = "today"
)

type Task struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Memo         string     `json:"memo"`
	CompletedAt  *time.Time `json:"completedAt"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
	StartDate    *time.Time `json:"startDate"`
	DueDate      *time.Time `json:"dueDate"`
	DisplayOrder int        `json:"displayOrder"`
	TagIDs       []string   `json:"tagIds"`
	ProjectID    string     `json:"projectId"`
}

func (t Task) Done() bool { return t.CompletedAt != nil }

func TaskFromDocument(d firestore.Document) Task {
	return Task{
		ID:           d.ID,
		Title:        d.Fields.String("title"),
		Memo:         d.Fields.String("memo"),
		CompletedAt:  d.Fields.Time("completedAt"),
		CreatedAt:    d.Fields.Time("createdAt"),
		UpdatedAt:    d.Fields.Time("updatedAt"),
		StartDate:    d.Fields.Time("startDate"),
		DueDate:      d.Fields.Time("dueDate"),
		DisplayOrder: d.Fields.Int("displayOrder"),
		TagIDs:       d.Fields.Strings("tagIds"),
		ProjectID:    d.Fields.String("projectId"),
	}
}

func (t Task) Document() firestore.Document {
	return firestore.Document{ID: t.ID, Fields: firestore.Fields{
		"id":           firestore.String(t.ID),
		"title":        firestore.String(t.Title),
		"memo":         firestore.String(t.Memo),
		"completedAt":  firestore.OptTime(t.CompletedAt),
		"createdAt":    firestore.OptTime(t.CreatedAt),
		"updatedAt":    firestore.OptTime(t.UpdatedAt),
		"startDate":    firestore.OptTime(t.StartDate),
		"dueDate":      firestore.OptTime(t.DueDate),
		"displayOrder": firestore.Int(t.DisplayOrder),
		"tagIds":       firestore.Strings(t.TagIDs),
		"projectId":    firestore.String(t.ProjectID),
	}}
}

type Project struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	StartDate   *time.Time `json:"startDate"`
	DueDate     *time.Time `json:"dueDate"`
}

func (p Project) Done() bool { return p.CompletedAt != nil }

// The app persists the completion time under the misspelled key "complatedAt".
func ProjectFromDocument(d firestore.Document) Project {
	return Project{
		ID:          d.ID,
		Title:       d.Fields.String("title"),
		CompletedAt: d.Fields.Time("complatedAt"),
		CreatedAt:   d.Fields.Time("createdAt"),
		UpdatedAt:   d.Fields.Time("updatedAt"),
		StartDate:   d.Fields.Time("startDate"),
		DueDate:     d.Fields.Time("dueDate"),
	}
}

func (p Project) Document() firestore.Document {
	return firestore.Document{ID: p.ID, Fields: firestore.Fields{
		"id":          firestore.String(p.ID),
		"title":       firestore.String(p.Title),
		"complatedAt": firestore.OptTime(p.CompletedAt),
		"createdAt":   firestore.OptTime(p.CreatedAt),
		"updatedAt":   firestore.OptTime(p.UpdatedAt),
		"startDate":   firestore.OptTime(p.StartDate),
		"dueDate":     firestore.OptTime(p.DueDate),
	}}
}

type Tag struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

func TagFromDocument(d firestore.Document) Tag {
	return Tag{ID: d.ID, Kind: d.Fields.String("kind"), Name: d.Fields.String("name")}
}

func (t Tag) Document() firestore.Document {
	return firestore.Document{ID: t.ID, Fields: firestore.Fields{
		"id":   firestore.String(t.ID),
		"kind": firestore.String(t.Kind),
		"name": firestore.String(t.Name),
	}}
}
