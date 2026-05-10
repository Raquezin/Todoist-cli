package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"todoist-cli/internal/client"
	"todoist-cli/internal/sanitize"
)

var CacheFile = ".cache/proyectos_cache.json"

var SectionCacheFile = ".cache/secciones_cache.json"

var (
	memProjectMu   sync.RWMutex
	memProject     map[string]nameID
	memProjectFile string

	memSectionMu   sync.RWMutex
	memSection     []nameID
	memSectionFile string
)

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

func lookupMemProject(nameLower string) (string, bool) {
	memProjectMu.RLock()
	defer memProjectMu.RUnlock()
	if memProjectFile != CacheFile || memProject == nil {
		return "", false
	}
	e, ok := memProject[nameLower]
	if !ok {
		return "", false
	}
	return e.ID, true
}

func storeMemProject(data map[string]nameID) {
	memProjectMu.Lock()
	memProject = data
	memProjectFile = CacheFile
	memProjectMu.Unlock()
}

func lookupMemSection(nameLower, projectID, displayName string) (string, bool) {
	memSectionMu.RLock()
	defer memSectionMu.RUnlock()
	if memSectionFile != SectionCacheFile || len(memSection) == 0 {
		return "", false
	}
	return matchSection(memSection, nameLower, projectID, displayName, true)
}

func storeMemSection(data []nameID) {
	memSectionMu.Lock()
	memSection = data
	memSectionFile = SectionCacheFile
	memSectionMu.Unlock()
}

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
	if name == "" {
		return ""
	}
	nameLower := strings.ToLower(name)

	if id, ok := lookupMemProject(nameLower); ok {
		fmt.Printf("⚡ [Cache] Project '%s' found in memory.\n", sanitize.TerminalLimit(name, 120))
		return id
	}

	id := cachedLookup(CacheFile, "Project", name, func() (map[string]nameID, error) {
		projects, err := apiClient.GetProjects()
		if err != nil {
			return nil, err
		}
		data := make(map[string]nameID, len(projects))
		for _, p := range projects {
			data[strings.ToLower(p.Name)] = nameID{Name: p.Name, ID: p.ID}
		}
		storeMemProject(data)
		return data, nil
	})
	if id != "" {
		memProjectMu.Lock()
		if memProjectFile == CacheFile && memProject != nil {
			memProject[nameLower] = nameID{Name: name, ID: id}
		}
		memProjectMu.Unlock()
	}
	return id
}

func GetSectionID(apiClient *client.TodoistClient, name, projectID string) string {
	if name == "" {
		return ""
	}
	nameLower := strings.ToLower(name)

	if id, ok := lookupMemSection(nameLower, projectID, name); ok {
		return id
	}

	var cacheData []nameID
	if fileData, err := os.ReadFile(SectionCacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err != nil {
			fmt.Println("⚠️ Warning: Section cache file is corrupted. Regenerating...")
		} else if id, ok := matchSection(cacheData, nameLower, projectID, name, true); ok {
			storeMemSection(cacheData)
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

	storeMemSection(freshData)

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
	memProjectMu.RLock()
	if memProjectFile == CacheFile && len(memProject) > 0 {
		result := make(map[string]string, len(memProject))
		for _, entry := range memProject {
			result[entry.ID] = entry.Name
		}
		memProjectMu.RUnlock()
		return result
	}
	memProjectMu.RUnlock()

	cacheData := make(map[string]nameID)
	result := make(map[string]string)
	if fileData, err := os.ReadFile(CacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err != nil {
			fmt.Printf("⚠️ Warning: Failed to unmarshal cache data: %s\n", sanitize.Terminal(err.Error()))
		} else {
			storeMemProject(cacheData)
			for _, entry := range cacheData {
				result[entry.ID] = entry.Name
			}
		}
	}
	return result
}

func GetSectionMap(apiClient *client.TodoistClient) (map[string]string, error) {
	memSectionMu.RLock()
	if memSectionFile == SectionCacheFile && len(memSection) > 0 {
		result := make(map[string]string, len(memSection))
		for _, entry := range memSection {
			result[entry.ID] = entry.Name
		}
		memSectionMu.RUnlock()
		return result, nil
	}
	memSectionMu.RUnlock()

	cacheData := make([]nameID, 0)
	result := make(map[string]string)
	if fileData, err := os.ReadFile(SectionCacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err == nil {
			storeMemSection(cacheData)
			for _, entry := range cacheData {
				result[entry.ID] = entry.Name
			}
			if len(result) > 0 {
				return result, nil
			}
		}
	}

	sections, err := apiClient.GetSections()
	if err != nil {
		return result, err
	}

	freshData := make([]nameID, 0, len(sections))
	result = make(map[string]string, len(sections))
	for _, s := range sections {
		freshData = append(freshData, nameID{Name: s.Name, ID: s.ID, ProjectID: s.ProjectID})
		result[s.ID] = s.Name
	}

	if cacheBytes, err := json.MarshalIndent(freshData, "", "    "); err == nil {
		_ = atomicWrite(SectionCacheFile, cacheBytes)
	}

	storeMemSection(freshData)

	return result, nil
}

func GetCachedProjectID(name string) string {
	nameLower := strings.ToLower(name)

	if id, ok := lookupMemProject(nameLower); ok {
		return id
	}

	cacheData := make(map[string]nameID)
	if fileData, err := os.ReadFile(CacheFile); err == nil {
		_ = json.Unmarshal(fileData, &cacheData)
		storeMemProject(cacheData)
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

	storeMemProject(cacheData)
	return nil
}
