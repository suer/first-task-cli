package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/suer/first-task-cli/internal/firestore"
)

// Documents are always read back from JSON, so round trip through it.
func viaJSON(t *testing.T, d firestore.Document) firestore.Document {
	t.Helper()
	data, err := json.Marshal(d.Fields)
	if err != nil {
		t.Fatal(err)
	}
	var fields firestore.Fields
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	return firestore.Document{ID: d.ID, Fields: fields}
}

func TestTaskDocumentRoundTrip(t *testing.T) {
	now := time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)
	in := Task{
		ID: "ABC", Title: "write", Memo: "m", CreatedAt: &now, UpdatedAt: &now,
		DueDate: &now, DisplayOrder: 3, TagIDs: []string{"t1", "t2"}, ProjectID: "p1",
	}
	out := TaskFromDocument(viaJSON(t, in.Document()))

	if out.Title != in.Title || out.DisplayOrder != 3 || out.ProjectID != "p1" {
		t.Fatalf("unexpected task: %+v", out)
	}
	if len(out.TagIDs) != 2 || out.TagIDs[1] != "t2" {
		t.Fatalf("unexpected tags: %v", out.TagIDs)
	}
	if out.CompletedAt != nil || out.StartDate != nil {
		t.Fatalf("nil timestamps must stay nil: %+v", out)
	}
	if out.DueDate == nil || !out.DueDate.Equal(now) {
		t.Fatalf("unexpected due date: %v", out.DueDate)
	}
}

func TestProjectUsesLegacyCompletedKey(t *testing.T) {
	now := time.Now()
	doc := Project{ID: "P", Title: "x", CompletedAt: &now}.Document()
	if _, ok := doc.Fields["complatedAt"]; !ok {
		t.Fatal("expected complatedAt key")
	}
	if !ProjectFromDocument(viaJSON(t, doc)).Done() {
		t.Fatal("expected project to be done")
	}
}
