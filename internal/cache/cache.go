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

var SectionCacheFile = ".cache/secciones_cache.json"

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

type nameID struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	ProjectID string `json:"project_id,omitempty"`
}

type ProjectCache = nameID

type SectionCache = nameID

func cachedLookup(cacheFile, entityType, name string, fetchFunc func() (map[string]nameID, error)) string {
	if name == "" {
		return ""
	}
	nameLower := strings.ToLower(name)

	cacheData := make(map[string]nameID)
	if fileData, err := os.ReadFile(cacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err == nil {
			if entry, exists := cacheData[nameLower]; exists {
				fmt.Printf("⚡ [Cache] %s '%s' found locally.\n", entityType, sanitize.TerminalLimit(name, 120))
				return entry.ID
			}
		} else {
			fmt.Printf("⚠️ Warning: %s cache file is corrupted. Regenerating...\n", entityType)
		}
	}

	fmt.Printf("🔄 [API] Fetching %ss from Todoist...\n", entityType)

	freshData, err := fetchFunc()
	if err != nil {
		fmt.Printf("❌ Error fetching %ss: %s\n", entityType, sanitize.Terminal(err.Error()))
		return ""
	}

	if cacheBytes, err := json.MarshalIndent(freshData, "", "    "); err == nil {
		if err := atomicWrite(cacheFile, cacheBytes); err != nil {
			fmt.Printf("⚠️ Warning: Failed to write %s cache file: %s\n", entityType, sanitize.Terminal(err.Error()))
		} else {
			fmt.Printf("💾 %s cache updated.\n", entityType)
		}
	}

	if entry, exists := freshData[nameLower]; exists {
		return entry.ID
	}
	return ""
}

func GetProjectID(apiClient *client.TodoistClient, name string) string {
	return cachedLookup(CacheFile, "Project", name, func() (map[string]nameID, error) {
		projects, err := apiClient.GetProjects()
		if err != nil {
			return nil, err
		}
		data := make(map[string]nameID, len(projects))
		for _, p := range projects {
			data[strings.ToLower(p.Name)] = nameID{Name: p.Name, ID: p.ID}
		}
		return data, nil
	})
}

func GetSectionID(apiClient *client.TodoistClient, name, projectID string) string {
	if name == "" {
		return ""
	}
	nameLower := strings.ToLower(name)

	var cacheData []nameID
	if fileData, err := os.ReadFile(SectionCacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err != nil {
			fmt.Println("⚠️ Warning: Section cache file is corrupted. Regenerating...")
		} else if id, ok := matchSection(cacheData, nameLower, projectID, name, true); ok {
			return id
		}
	}

	fmt.Println("🔄 [API] Fetching sections from Todoist...")

	sections, err := apiClient.GetSections()
	if err != nil {
		fmt.Printf("❌ Error fetching sections: %s\n", sanitize.Terminal(err.Error()))
		return ""
	}

	freshData := make([]nameID, 0, len(sections))
	for _, s := range sections {
		freshData = append(freshData, nameID{Name: s.Name, ID: s.ID, ProjectID: s.ProjectID})
	}

	if cacheBytes, err := json.MarshalIndent(freshData, "", "    "); err == nil {
		if err := atomicWrite(SectionCacheFile, cacheBytes); err != nil {
			fmt.Printf("⚠️ Warning: Failed to write section cache file: %s\n", sanitize.Terminal(err.Error()))
		} else {
			fmt.Println("💾 Section cache updated.")
		}
	}

	id, _ := matchSection(freshData, nameLower, projectID, name, false)
	return id
}

func matchSection(entries []nameID, nameLower, projectID, displayName string, fromCache bool) (id string, matched bool) {
	if projectID != "" {
		for _, e := range entries {
			if strings.ToLower(e.Name) == nameLower && e.ProjectID == projectID {
				if fromCache {
					fmt.Printf("⚡ [Cache] Section '%s' found locally.\n", sanitize.TerminalLimit(displayName, 120))
				}
				return e.ID, true
			}
		}
		return "", false
	}

	var matches []nameID
	for _, e := range entries {
		if strings.ToLower(e.Name) == nameLower {
			matches = append(matches, e)
		}
	}

	if len(matches) == 0 {
		return "", false
	}
	if len(matches) == 1 {
		if fromCache {
			fmt.Printf("⚡ [Cache] Section '%s' found locally.\n", sanitize.TerminalLimit(displayName, 120))
		}
		return matches[0].ID, true
	}

	fmt.Printf("⚠️ Warning: Multiple sections named '%s' found. Specify -project to disambiguate.\n", sanitize.TerminalLimit(displayName, 120))
	return "", false
}

func GetAllCachedProjects() map[string]string {
	cacheData := make(map[string]nameID)
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
	cacheData := make(map[string]nameID)
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

	cacheData := make(map[string]nameID, len(projects))
	for _, p := range projects {
		cacheData[strings.ToLower(p.Name)] = nameID{Name: p.Name, ID: p.ID}
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
