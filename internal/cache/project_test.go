package cache

import (
	"encoding/json"
	"os"
	"testing"
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

	cacheData := map[string]string{"test project": "proj123"}
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

	cacheData := map[string]string{"proj1": "id1", "proj2": "id2"}
	jsonData, _ := json.Marshal(cacheData)
	if err := os.WriteFile(CacheFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test cache: %v", err)
	}

	all := GetAllCachedProjects()
	if len(all) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(all))
	}
	if all["proj1"] != "id1" {
		t.Errorf("Expected id1, got %s", all["proj1"])
	}
}
