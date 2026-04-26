package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"todoist-cli/internal/models"
)

func TestGetProjects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			t.Errorf("Expected Authorization header")
		}
		if r.Method != "GET" || r.URL.Path != "/projects" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"1","name":"Test Project"},{"id":"2","name":"Another"}],"next_cursor":""}`))
	}))
	defer ts.Close()

	client := New("fake-token")
	client.BaseURL = ts.URL

	projects, err := client.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects failed: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}

	if projects[0].ID != "1" || projects[0].Name != "Test Project" {
		t.Errorf("Unexpected first project: %+v", projects[0])
	}
}

func TestCreateTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header")
		}
		if r.Method != "POST" || r.URL.Path != "/tasks" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var req models.TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode body: %v", err)
		}

		if req.Content != "Buy milk" {
			t.Errorf("Unexpected content: %s", req.Content)
		}

		_, _ = w.Write([]byte(`{"content":"Buy milk","priority":1,"project_id":"123"}`))
	}))
	defer ts.Close()

	client := New("fake-token")
	client.BaseURL = ts.URL

	taskReq := models.TaskRequest{Content: "Buy milk", Priority: 1}
	taskRes, err := client.CreateTask(taskReq)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if taskRes.Content != "Buy milk" || taskRes.ProjectID != "123" {
		t.Errorf("Unexpected response: %+v", taskRes)
	}
}

func TestFilterTasks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/tasks/filter" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}

		query := r.URL.Query().Get("query")
		cursor := r.URL.Query().Get("cursor")

		if query != "today" {
			t.Errorf("Unexpected query: %s", query)
		}
		if cursor != "cursor123" {
			t.Errorf("Unexpected cursor: %s", cursor)
		}

		_, _ = w.Write([]byte(`{"results":[{"id":"1","content":"Task 1","project_id":"p1","priority":1}],"next_cursor":"cursor456"}`))
	}))
	defer ts.Close()

	client := New("fake-token")
	client.BaseURL = ts.URL

	res, err := client.FilterTasks("today", "cursor123")
	if err != nil {
		t.Fatalf("FilterTasks failed: %v", err)
	}

	if len(res.Results) != 1 || res.Results[0].Content != "Task 1" {
		t.Errorf("Unexpected results: %+v", res.Results)
	}
	if res.NextCursor != "cursor456" {
		t.Errorf("Unexpected next cursor: %s", res.NextCursor)
	}
}

func TestDoRequestErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
			return
		}
		if r.URL.Path == "/tasks/filter" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Forbidden"}`))
			return
		}
		if r.URL.Path == "/badjson" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"invalid-json`))
			return
		}
	}))
	defer ts.Close()

	client := New("invalid-token")
	client.BaseURL = ts.URL

	// Test HTTP Error directly on FilterTasks
	_, err := client.FilterTasks("test", "")
	if err == nil || err.Error() != "API error (403): {\"error\":\"Forbidden\"}" {
		t.Errorf("Expected 403 error, got %v", err)
	}

	// Test specific error path
	err = client.doRequest("GET", "/error", nil, nil)
	if err == nil || err.Error() != "API error (401): {\"error\":\"Unauthorized\"}" {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test bad json response
	var dummy struct{}
	err = client.doRequest("GET", "/badjson", nil, &dummy)
	if err == nil {
		t.Fatal("Expected JSON decode error")
	}

	// Test bad URL
	client.BaseURL = "http://invalid-url-\x00"
	err = client.doRequest("GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("Expected request build error")
	}
}

func TestGetProjectsNetworkError(t *testing.T) {
	client := New("fake-token")
	client.BaseURL = "http://invalid-url-\x00"
	_, err := client.GetProjects()
	if err == nil {
		t.Fatal("Expected error for invalid URL in GetProjects")
	}
}

func TestCreateTaskNetworkError(t *testing.T) {
	client := New("fake-token")
	client.BaseURL = "http://invalid-url-\x00"
	_, err := client.CreateTask(models.TaskRequest{})
	if err == nil {
		t.Fatal("Expected error for invalid URL in CreateTask")
	}
}

func TestRateLimitRetry(t *testing.T) {
	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"Too Many Requests"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer ts.Close()

	client := New("fake-token")
	client.BaseURL = ts.URL

	_, err := client.GetProjects()
	if err != nil {
		t.Fatalf("Expected success after retries, got %v", err)
	}
	if requests != 3 {
		t.Errorf("Expected 3 requests, got %d", requests)
	}
}

func TestRateLimitRetryExhausted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"Too Many Requests"}`))
	}))
	defer ts.Close()

	client := New("fake-token")
	client.BaseURL = ts.URL

	_, err := client.GetProjects()
	if err == nil {
		t.Fatal("Expected error after exhausting retries")
	}
}
