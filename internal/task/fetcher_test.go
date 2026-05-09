package task

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"todoist-cli/internal/client"
)

func TestFetcherFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" {
			_, _ = w.Write([]byte(`{"results":[{"id":"proj1","name":"Work"}]}`))
			return
		}

		if r.URL.Path == "/tasks/filter" {
			cursor := r.URL.Query().Get("cursor")
			if cursor == "" {
				_, _ = w.Write([]byte(`{
					"results": [
						{"id":"t1", "content":"Task 1", "project_id":"proj1", "priority":4}
					],
					"next_cursor": "page2"
				}`))
			} else {
				_, _ = w.Write([]byte(`{
					"results": [
						{"id":"t2", "content":"Task 2", "project_id":"proj1", "priority":4}
					],
					"next_cursor": ""
				}`))
			}
			return
		}
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL
	f := NewFetcher(apiClient)

	// Will fetch preset "foco"
	err := f.Fetch("foco")
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Fetch empty results test
	tsEmpty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tasks/filter" {
			_, _ = w.Write([]byte(`{"results": []}`))
		}
	}))
	defer tsEmpty.Close()

	apiClientEmpty := client.New("fake-token")
	apiClientEmpty.BaseURL = tsEmpty.URL
	fEmpty := NewFetcher(apiClientEmpty)

	err = fEmpty.Fetch("custom query")
	if err != nil {
		t.Fatalf("Fetch empty failed: %v", err)
	}
}

func TestFetcherPaginationLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tasks/filter" {
			// Always return next_cursor to force infinite loop
			_, _ = w.Write([]byte(`{
				"results": [{"id":"t3", "content":"Infinite Task", "project_id":"proj1", "priority":4}],
				"next_cursor": "neverending"
			}`))
			return
		}
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL
	f := NewFetcher(apiClient)

	// Should stop after maxPages (20)
	err := f.Fetch("custom")
	if err != nil {
		t.Fatalf("Fetch failed on pagination limit: %v", err)
	}
}

func TestFetcherAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL
	f := NewFetcher(apiClient)

	err := f.Fetch("custom")
	if err == nil {
		t.Fatal("Expected error from Fetch due to API failure")
	}
}
