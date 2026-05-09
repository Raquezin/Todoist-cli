package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"todoist-cli/internal/client"
	"todoist-cli/internal/models"
)

func TestCreatorTimezone(t *testing.T) {
	apiClient := client.New("fake-token")
	c := NewCreator(apiClient)
	if c.Loc == nil {
		t.Error("Expected location to be set")
	}
}

func TestCreatorDateParsing(t *testing.T) {
	apiClient := client.New("fake-token")
	c := NewCreator(apiClient)

	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006/01/02 15:04",
	}

	for _, format := range formats {
		_, err := time.ParseInLocation(format, "2026-03-25 17:00", c.Loc)
		if err == nil {
			break
		}
	}

	dateFormats := []string{
		"2006-01-02",
		"2006/01/02",
	}

	for _, format := range dateFormats {
		_, err := time.ParseInLocation(format, "2026-03-25", c.Loc)
		if err == nil {
			break
		}
	}
}

func TestCreatorCreate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" {
			_, _ = w.Write([]byte(`{"results":[{"id":"proj1","name":"Work"}]}`))
			return
		}

		if r.URL.Path == "/tasks" {
			var req models.TaskRequest
			_ = json.NewDecoder(r.Body).Decode(&req)

			if req.Content != "Test Task" {
				t.Errorf("Unexpected content: %s", req.Content)
			}
			if req.ProjectID != "proj1" {
				t.Errorf("Unexpected project ID: %s", req.ProjectID)
			}

			_, _ = w.Write([]byte(`{"content":"Test Task","id":"task1"}`))
			return
		}
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL
	c := NewCreator(apiClient)

	err := c.Create("Test Task", "2026-03-25 17:00", 60, "Work", []string{"tag1"}, "desc", 1)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test invalid date
	err = c.Create("Test Task", "invalid-date", 0, "", nil, "", 4)
	if err == nil {
		t.Fatal("Expected error for invalid date")
	}
}

func TestCreatorCreateErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" {
			_, _ = w.Write([]byte(`{"results":[]}`))
			return
		}
		if r.URL.Path == "/tasks" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Internal Server Error"}`))
			return
		}
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL
	c := NewCreator(apiClient)

	// Valid date, but network will fail
	err := c.Create("Test Task", "2026-03-25", 0, "Nonexistent", nil, "", 4)
	if err == nil {
		t.Fatal("Expected network error from API client")
	}
}

func TestNewCreatorTZFallback(t *testing.T) {
	// Need to test fallback when TZ is invalid

	oldTz := os.Getenv("TZ")
	defer func() {
		if err := os.Setenv("TZ", oldTz); err != nil {
			t.Fatalf("Failed to restore TZ: %v", err)
		}
	}()

	if err := os.Setenv("TZ", "Invalid/Timezone"); err != nil {
		t.Fatalf("Failed to set TZ: %v", err)
	}
	c := NewCreator(client.New("fake"))
	if c.Loc == nil {
		t.Error("Expected fallback to time.Local, got nil")
	}
}
