package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"todoist-cli/internal/client"
	"todoist-cli/internal/sanitize"
)

var CacheFile = ".cache/proyectos_cache.json"

func atomicWrite(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, "cache-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return err
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, filename)
}

type ProjectCache struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetProjectID(apiClient *client.TodoistClient, name string) string {
	if name == "" {
		return ""
	}
	nameLower := strings.ToLower(name)
	cacheData := make(map[string]ProjectCache)

	if fileData, err := os.ReadFile(CacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err == nil {
			if entry, exists := cacheData[nameLower]; exists {
				fmt.Printf("⚡ [Cache] Project '%s' found locally.\n", sanitize.TerminalLimit(name, 120))
				return entry.ID
			}
		} else {
			fmt.Println("⚠️ Warning: Cache file is corrupted. Regenerating...")
		}
	}

	fmt.Println("🔄 [API] Fetching projects from Todoist...")

	projects, err := apiClient.GetProjects()
	if err != nil {
		fmt.Printf("❌ Error fetching projects: %s\n", sanitize.Terminal(err.Error()))
		return ""
	}

	cacheData = make(map[string]ProjectCache)
	for _, p := range projects {
		cacheData[strings.ToLower(p.Name)] = ProjectCache{ID: p.ID, Name: p.Name}
	}

	if cacheBytes, err := json.MarshalIndent(cacheData, "", "    "); err == nil {
		if err := atomicWrite(CacheFile, cacheBytes); err != nil {
			fmt.Printf("⚠️ Warning: Failed to write cache file: %s\n", sanitize.Terminal(err.Error()))
		} else {
			fmt.Println("💾 Project cache updated.")
		}
	}

	return cacheData[nameLower].ID
}

func GetAllCachedProjects() map[string]string {
	cacheData := make(map[string]ProjectCache)
	result := make(map[string]string)
	if fileData, err := os.ReadFile(CacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err != nil {
			fmt.Printf("⚠️ Warning: Failed to unmarshal cache data: %s\n", sanitize.Terminal(err.Error()))
		} else {
			for _, entry := range cacheData {
				result[entry.ID] = entry.Name
			}
		}
	}
	return result
}

func GetCachedProjectID(name string) string {
	nameLower := strings.ToLower(name)
	cacheData := make(map[string]ProjectCache)
	if fileData, err := os.ReadFile(CacheFile); err == nil {
		_ = json.Unmarshal(fileData, &cacheData)
	}
	return cacheData[nameLower].ID
}

func RefreshCache(apiClient *client.TodoistClient) error {
	projects, err := apiClient.GetProjects()
	if err != nil {
		return err
	}

	cacheData := make(map[string]ProjectCache)
	for _, p := range projects {
		cacheData[strings.ToLower(p.Name)] = ProjectCache{ID: p.ID, Name: p.Name}
	}

	cacheBytes, err := json.MarshalIndent(cacheData, "", "    ")
	if err != nil {
		return err
	}

	if err := atomicWrite(CacheFile, cacheBytes); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}
