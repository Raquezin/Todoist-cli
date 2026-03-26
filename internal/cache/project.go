package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"todoist-cli/internal/client"
)

var CacheFile = ".cache/proyectos_cache.json"

func GetProjectID(token, name string) string {
	if name == "" {
		return ""
	}
	nameLower := strings.ToLower(name)
	cacheData := make(map[string]string)

	if fileData, err := os.ReadFile(CacheFile); err == nil {
		if err := json.Unmarshal(fileData, &cacheData); err == nil {
			if id, exists := cacheData[nameLower]; exists {
				fmt.Printf("⚡ [Cache] Project '%s' found locally.\n", name)
				return id
			}
		} else {
			fmt.Println("⚠️ Warning: Cache file is corrupted. Regenerating...")
		}
	}

	fmt.Println("🔄 [API] Fetching projects from Todoist...")

	todoistClient := client.New(token)
	projects, err := todoistClient.GetProjects()
	if err != nil {
		fmt.Printf("❌ Error fetching projects: %v\n", err)
		return ""
	}

	for _, p := range projects {
		cacheData[strings.ToLower(p.Name)] = p.ID
	}

	if cacheBytes, err := json.MarshalIndent(cacheData, "", "    "); err == nil {
		os.MkdirAll(filepath.Dir(CacheFile), 0755)
		os.WriteFile(CacheFile, cacheBytes, 0644)
		fmt.Println("💾 Project cache updated.")
	}

	return cacheData[nameLower]
}

func GetAllCachedProjects() map[string]string {
	cacheData := make(map[string]string)
	if fileData, err := os.ReadFile(CacheFile); err == nil {
		json.Unmarshal(fileData, &cacheData)
	}
	return cacheData
}

func GetCachedProjectID(name string) string {
	nameLower := strings.ToLower(name)
	cacheData := GetAllCachedProjects()
	return cacheData[nameLower]
}

func RefreshCache(token string) error {
	todoistClient := client.New(token)
	projects, err := todoistClient.GetProjects()
	if err != nil {
		return err
	}

	cacheData := make(map[string]string)
	for _, p := range projects {
		cacheData[strings.ToLower(p.Name)] = p.ID
	}

	cacheBytes, err := json.MarshalIndent(cacheData, "", "    ")
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(CacheFile), 0755)
	return os.WriteFile(CacheFile, cacheBytes, 0644)
}
