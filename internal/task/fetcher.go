package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"todoist-cli/internal/cache"
	"todoist-cli/internal/client"
	"todoist-cli/internal/models"
	"todoist-cli/internal/sanitize"
)

const exclusionGlobalDefault = "& !(#Study & /Horario)"
const maxPages = 20
const fetchTimeout = 120 * time.Second
const presetsFile = ".cache/presets.json"

func exclusionFilter() string {
	if env := os.Getenv("TODOIST_EXCLUDE_FILTER"); env != "" {
		return env
	}
	return exclusionGlobalDefault
}

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
		return fmt.Errorf("%s already exists — delete it first or use 'edit' to modify", presetsFile)
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
		return fmt.Errorf("'%s' is a built-in preset — use 'edit' to override it", name)
	}

	user, _ := loadUserPresets()
	if user == nil {
		user = make(map[string]string)
	}
	if _, exists := user[name]; exists {
		return fmt.Errorf("preset '%s' already exists — use 'edit' to change it", name)
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
		return fmt.Errorf("'%s' is a built-in preset and cannot be deleted — use 'edit' to override it", name)
	}

	user, err := loadUserPresets()
	if os.IsNotExist(err) {
		return fmt.Errorf("preset '%s' not found", name)
	}
	if err != nil {
		return err
	}

	if _, exists := user[name]; !exists {
		return fmt.Errorf("preset '%s' not found", name)
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

func outputJSON(w io.Writer, tasks []models.FilteredTask, projectMap, sectionMap map[string]string) error {
	type jsonTask struct {
		Content   string   `json:"content"`
		Project   string   `json:"project"`
		Section   string   `json:"section,omitempty"`
		Priority  int      `json:"priority"`
		Labels    []string `json:"labels,omitempty"`
		DueDate   string   `json:"due_date,omitempty"`
		DueString string   `json:"due_string,omitempty"`
	}

	out := make([]jsonTask, 0, len(tasks))
	for _, t := range tasks {
		jt := jsonTask{
			Content:  t.Content,
			Project:  projectMap[t.ProjectID],
			Section:  sectionMap[t.SectionID],
			Priority: models.ToUIPriority(t.Priority),
			Labels:   t.Labels,
		}
		if t.Due != nil {
			jt.DueDate = t.Due.Date
			jt.DueString = t.Due.String
		}
		if jt.Project == "" {
			jt.Project = "Inbox"
		}
		out = append(out, jt)
	}

	bytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}
	fmt.Fprintln(w, string(bytes))
	return nil
}

type Fetcher struct {
	Client *client.TodoistClient
	Out    io.Writer
}

func NewFetcher(apiClient *client.TodoistClient) *Fetcher {
	return &Fetcher{Client: apiClient, Out: os.Stdout}
}

func (f *Fetcher) Fetch(queryName string, jsonOut bool) error {
	presets := LoadPresets()
	queryBase, exists := presets[queryName]
	if !exists {
		queryBase = queryName
	}

	var queryFinal string
	if exists {
		queryFinal = fmt.Sprintf("(%s) %s", queryBase, exclusionFilter())
		fmt.Fprintf(f.Out, "\n🔍 Executing preset: [%s]\n", sanitize.Terminal(queryName))
	} else {
		queryFinal = queryBase
		fmt.Fprintf(f.Out, "\n🔍 Executing custom filter\n")
	}
	fmt.Fprintf(f.Out, "💻 Sent query: %s\n", sanitize.Terminal(queryFinal))

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	var allTasks []models.FilteredTask
	cursor := ""
	pageCount := 0

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("fetch timed out or was cancelled after %v: %w", fetchTimeout, err)
		}
		if pageCount >= maxPages {
			fmt.Fprintf(f.Out, "⚠️ Warning: Reached maximum pagination limit (%d pages). Some tasks might be missing.\n", maxPages)
			break
		}

		apiResp, err := f.Client.FilterTasks(ctx, queryFinal, cursor)
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
		fmt.Fprintln(f.Out, "   🤷‍♂️ Inbox zero. No tasks found for this filter.")
		return nil
	}

	fmt.Fprintf(f.Out, "   🎯 Found %d tasks:\n", len(allTasks))

	idToName := cache.GetAllCachedProjects()
	sectionIDToName, err := cache.GetSectionMap(f.Client)
	if err != nil {
		fmt.Fprintf(f.Out, "⚠️ Warning: Failed to fetch section map: %s\n", sanitize.Terminal(err.Error()))
	}

	if jsonOut {
		if err := outputJSON(f.Out, allTasks, idToName, sectionIDToName); err != nil {
			return err
		}
		return nil
	}

	now := time.Now()
	for _, t := range allTasks {
		fmt.Fprintf(f.Out, "      %s\n", FormatTask(t, now, idToName, sectionIDToName))
	}

	return nil
}
