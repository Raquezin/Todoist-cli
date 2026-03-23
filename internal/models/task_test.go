package models

import (
	"encoding/json"
	"testing"
)

func TestProjectMarshalUnmarshal(t *testing.T) {
	p := Project{ID: "123", Name: "Test Project"}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var p2 Project
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if p2.ID != p.ID || p2.Name != p.Name {
		t.Errorf("Expected %+v, got %+v", p, p2)
	}
}

func TestTaskRequestMarshal(t *testing.T) {
	req := TaskRequest{
		Content:     "Test Task",
		DueDatetime: "2026-03-25T17:00:00Z",
		ProjectID:   "123",
		Labels:      []string{"tag1", "tag2"},
		Description: "Test description",
		Priority:    4,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{"content":"Test Task","due_datetime":"2026-03-25T17:00:00Z","project_id":"123","labels":["tag1","tag2"],"description":"Test description","priority":4}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}
