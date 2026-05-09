package cache

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"todoist-cli/internal/client"
)

func TestGetCachedProjectID(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "cache_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	oldCache := CacheFile
	CacheFile = tmpfile.Name()
	defer func() { CacheFile = oldCache }()

	cacheData := map[string]ProjectCache{
		"test project": {ID: "proj123", Name: "Test Project"},
	}
	jsonData, _ := json.Marshal(cacheData)
	if err := os.WriteFile(CacheFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test cache: %v", err)
	}

	id := GetCachedProjectID("Test Project")
	if id != "proj123" {
		t.Errorf("Expected 'proj123', got '%s'", id)
	}

	id = GetCachedProjectID("Nonexistent")
	if id != "" {
		t.Errorf("Expected empty string, got '%s'", id)
	}
}

func TestGetAllCachedProjects(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "cache_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	oldCache := CacheFile
	CacheFile = tmpfile.Name()
	defer func() { CacheFile = oldCache }()

	cacheData := map[string]ProjectCache{
		"proj1": {ID: "id1", Name: "Proj1"},
		"proj2": {ID: "id2", Name: "Proj2"},
	}
	jsonData, _ := json.Marshal(cacheData)
	if err := os.WriteFile(CacheFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test cache: %v", err)
	}

	all := GetAllCachedProjects()
	if len(all) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(all))
	}
	if all["id1"] != "Proj1" {
		t.Errorf("Expected Proj1, got %s", all["id1"])
	}
}

func TestGetProjectID(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "cache_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	os.Remove(tmpfile.Name()) // Ensure it does not exist to trigger fetch

	oldCache := CacheFile
	CacheFile = tmpfile.Name()
	defer func() {
		os.Remove(CacheFile)
		CacheFile = oldCache
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" {
			_, _ = w.Write([]byte(`{"results":[{"id":"api123","name":"API Project"}]}`))
		}
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL

	// It should miss cache and fetch from API
	id := GetProjectID(apiClient, "API Project")
	if id != "api123" {
		t.Errorf("Expected api123, got %s", id)
	}

	// Next call should hit cache
	id2 := GetProjectID(apiClient, "API Project")
	if id2 != "api123" {
		t.Errorf("Expected api123 from cache, got %s", id2)
	}
}

func TestRefreshCache(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "cache_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	oldCache := CacheFile
	CacheFile = tmpfile.Name()
	defer func() { CacheFile = oldCache }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" {
			_, _ = w.Write([]byte(`{"results":[{"id":"refreshed1","name":"Refreshed"}]}`))
		}
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL

	err = RefreshCache(apiClient)
	if err != nil {
		t.Fatalf("RefreshCache failed: %v", err)
	}

	id := GetCachedProjectID("Refreshed")
	if id != "refreshed1" {
		t.Errorf("Expected refreshed1, got %s", id)
	}
}

func TestAtomicWriteUsesPrivateFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not portable on Windows")
	}

	tmpfile, err := os.CreateTemp("", "cache_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if err := atomicWrite(tmpfile.Name(), []byte(`{}`)); err != nil {
		t.Fatalf("atomicWrite failed: %v", err)
	}

	info, err := os.Stat(tmpfile.Name())
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("Expected 0600 cache permissions, got %o", got)
	}
}

func TestAtomicWriteErrorPaths(t *testing.T) {
	// Trying to write to a path where we can't create a directory
	err := atomicWrite("/root/restricted/cache.json", []byte{})
	if err == nil {
		t.Error("Expected error when directory cannot be created")
	}
}

func TestGetAllCachedProjectsBadData(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "cache_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	oldCache := CacheFile
	CacheFile = tmpfile.Name()
	defer func() { CacheFile = oldCache }()

	os.WriteFile(CacheFile, []byte("{invalid-json"), 0644)

	// Should not panic, but return an empty map since unmarshal fails
	res := GetAllCachedProjects()
	if len(res) != 0 {
		t.Errorf("Expected empty map for bad JSON, got %v", res)
	}
}

func TestRefreshCacheFailures(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	apiClient := client.New("fake-token")
	apiClient.BaseURL = ts.URL

	err := RefreshCache(apiClient)
	if err == nil {
		t.Error("Expected error from RefreshCache when API fails")
	}
}
