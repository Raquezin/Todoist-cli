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

func TestPriorityConversions(t *testing.T) {
	// Test UI -> API
	uiToAPI := map[int]int{
		1:  4, // P1 -> API 4
		2:  3, // P2 -> API 3
		3:  2, // P3 -> API 2
		4:  1, // P4 -> API 1
		0:  4, // Out of bounds UI -> Cap at 4
		5:  1, // Out of bounds UI -> Cap at 1
		-1: 4, // Out of bounds UI -> Cap at 4
		10: 1, // Out of bounds UI -> Cap at 1
	}

	for ui, expectedAPI := range uiToAPI {
		gotAPI := ToAPIPriority(ui)
		if gotAPI != expectedAPI {
			t.Errorf("ToAPIPriority(%d) = %d; want %d", ui, gotAPI, expectedAPI)
		}
	}

	// Test API -> UI
	apiToUI := map[int]int{
		4:  1, // API 4 -> P1
		3:  2, // API 3 -> P2
		2:  3, // API 2 -> P3
		1:  4, // API 1 -> P4
		0:  4, // Out of bounds API -> Cap at 4
		5:  1, // Out of bounds API -> Cap at 1
		-1: 4, // Out of bounds API -> Cap at 4
		10: 1, // Out of bounds API -> Cap at 1
	}

	for api, expectedUI := range apiToUI {
		gotUI := ToUIPriority(api)
		if gotUI != expectedUI {
			t.Errorf("ToUIPriority(%d) = %d; want %d", api, gotUI, expectedUI)
		}
	}
}
