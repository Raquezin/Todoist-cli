package task

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"todoist-cli/internal/cache"
	"todoist-cli/internal/client"
	"todoist-cli/internal/models"
	"todoist-cli/internal/sanitize"
)

const exclusionGlobal = "& !(#Study & /Horario)"
const maxPages = 20
const presetsFile = ".cache/presets.json"

var builtinQueries = map[string]string{
	"foco":  "today & p1 & !@reuniones",
	"radar": "7 days & @importante",
}

func loadUserPresets() (map[string]string, error) {
	data, err := os.ReadFile(presetsFile)
	if err != nil {
		return nil, err
	}
	var user map[string]string
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", presetsFile, err)
	}
	return user, nil
}

func saveUserPresets(presets map[string]string) error {
	data, err := json.MarshalIndent(presets, "", "    ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(".cache", 0755); err != nil {
		return err
	}
	return os.WriteFile(presetsFile, data, 0644)
}

func LoadPresets() map[string]string {
	merged := make(map[string]string, len(builtinQueries))
	for k, v := range builtinQueries {
		merged[k] = v
	}

	user, err := loadUserPresets()
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("⚠️ Warning: %s\n", sanitize.Terminal(err.Error()))
		}
		return merged
	}

	for k, v := range user {
		merged[k] = v
	}
	return merged
}

func ListPresets() {
	presets := LoadPresets()
	fmt.Println("📋 Active presets:")
	hasOverrides := false
	for _, k := range sortedKeys(presets) {
		v := presets[k]
		origin := "  "
		if builtin, isBuiltin := builtinQueries[k]; isBuiltin {
			origin = "  "
			if builtin != v {
				origin = "(o)"
				hasOverrides = true
			}
		}
		fmt.Printf("   %s %-10s %s\n", origin, k+":", v)
	}
	if hasOverrides {
		fmt.Println("\n(o) = overridden built-in preset.")
	}
	fmt.Println("\nCommands: add | edit | delete | init | help")
}

func InitPresets() error {
	if _, err := os.Stat(presetsFile); err == nil {
		return fmt.Errorf("%s already exists. Delete it first or use 'edit' to modify.", presetsFile)
	}
	if err := saveUserPresets(builtinQueries); err != nil {
		return err
	}
	fmt.Println("✅ presets.json created with built-in defaults:")
	for _, k := range sortedKeys(builtinQueries) {
		fmt.Printf("   %-10s %s\n", k+":", builtinQueries[k])
	}
	fmt.Println("\n  Add presets with: ./todoist-cli presets add <name> <query>")
	return nil
}

func AddPreset(name, query string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("preset name is required")
	}
	if _, isBuiltin := builtinQueries[name]; isBuiltin {
		return fmt.Errorf("'%s' is a built-in preset. Use 'edit' to override it.", name)
	}

	user, _ := loadUserPresets()
	if user == nil {
		user = make(map[string]string)
	}
	if _, exists := user[name]; exists {
		return fmt.Errorf("preset '%s' already exists. Use 'edit' to change it.", name)
	}

	user[name] = query
	if err := saveUserPresets(user); err != nil {
		return err
	}
	fmt.Printf("✅ Preset '%s' added.\n", name)
	return nil
}

func EditPreset(name, query string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("preset name is required")
	}

	user, err := loadUserPresets()
	if os.IsNotExist(err) {
		user = make(map[string]string)
	} else if err != nil {
		return err
	}

	old, existed := user[name]
	user[name] = query

	if err := saveUserPresets(user); err != nil {
		return err
	}

	if _, isBuiltin := builtinQueries[name]; isBuiltin {
		fmt.Printf("✅ Built-in preset '%s' overridden.\n", name)
	} else if existed {
		fmt.Printf("✅ Preset '%s' updated.\n", name)
		fmt.Printf("   Old: %s\n", old)
		fmt.Printf("   New: %s\n", query)
	} else {
		fmt.Printf("✅ Preset '%s' added.\n", name)
	}
	return nil
}

func DeletePreset(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("preset name is required")
	}
	if _, isBuiltin := builtinQueries[name]; isBuiltin {
		return fmt.Errorf("'%s' is a built-in preset and cannot be deleted. Use 'edit' to override it.", name)
	}

	user, err := loadUserPresets()
	if os.IsNotExist(err) {
		return fmt.Errorf("preset '%s' not found.", name)
	}
	if err != nil {
		return err
	}

	if _, exists := user[name]; !exists {
		return fmt.Errorf("preset '%s' not found.", name)
	}

	delete(user, name)
	if err := saveUserPresets(user); err != nil {
		return err
	}
	fmt.Printf("✅ Preset '%s' deleted.\n", name)
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type Fetcher struct {
	Client *client.TodoistClient
}

func NewFetcher(apiClient *client.TodoistClient) *Fetcher {
	return &Fetcher{Client: apiClient}
}

func (f *Fetcher) Fetch(queryName string) error {
	presets := LoadPresets()
	queryBase, exists := presets[queryName]
	if !exists {
		queryBase = queryName
	}

	var queryFinal string
	if exists {
		queryFinal = fmt.Sprintf("(%s) %s", queryBase, exclusionGlobal)
		fmt.Printf("\n🔍 Executing preset: [%s]\n", sanitize.Terminal(queryName))
	} else {
		queryFinal = queryBase
		fmt.Printf("\n🔍 Executing custom filter\n")
	}
	fmt.Printf("💻 Sent query: %s\n", sanitize.Terminal(queryFinal))

	var allTasks []models.FilteredTask
	cursor := ""
	pageCount := 0

	for {
		if pageCount >= maxPages {
			fmt.Printf("⚠️ Warning: Reached maximum pagination limit (%d pages). Some tasks might be missing.\n", maxPages)
			break
		}

		apiResp, err := f.Client.FilterTasks(queryFinal, cursor)
		if err != nil {
			return err
		}

		allTasks = append(allTasks, apiResp.Results...)

		if apiResp.NextCursor == "" {
			break
		}
		cursor = apiResp.NextCursor
		pageCount++
	}

	if len(allTasks) == 0 {
		fmt.Println("   🤷‍♂️ Inbox zero. No tasks found for this filter.")
		return nil
	}

	fmt.Printf("   🎯 Found %d tasks:\n", len(allTasks))

	idToName := cache.GetAllCachedProjects()

	now := time.Now()
	for _, t := range allTasks {
		fmt.Printf("      %s\n", FormatTask(t, now, idToName))
	}

	return nil
}
