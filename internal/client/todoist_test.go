package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProjects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			t.Errorf("Expected Authorization header")
		}
		w.Write([]byte(`{"results":[{"id":"1","name":"Test Project"},{"id":"2","name":"Another"}],"next_cursor":""}`))
	}))
	defer ts.Close()

	oldBase := BaseURL
	BaseURL = ts.URL
	defer func() { BaseURL = oldBase }()

	client := New("fake-token")
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

func TestGetProjectsAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
	}))
	defer ts.Close()

	oldBase := BaseURL
	BaseURL = ts.URL
	defer func() { BaseURL = oldBase }()

	client := New("invalid-token")
	_, err := client.GetProjects()
	if err == nil {
		t.Fatal("Expected error for unauthorized request")
	}
}
