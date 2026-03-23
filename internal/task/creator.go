package task

import (
	"fmt"
	"strings"
	"time"

	"todoist-cli/internal/cache"
	"todoist-cli/internal/client"
	"todoist-cli/internal/models"
)

type Creator struct {
	Token string
	Loc   *time.Location
}

func NewCreator(token string) *Creator {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		fmt.Printf("❌ Error loading timezone: %v\n", err)
		loc = time.Local
	}
	return &Creator{Token: token, Loc: loc}
}

func (c *Creator) Create(name, startStr string, duration int, projectName string, labels []string, description string, priority int) error {
	// Intentar parsear con un formato mucho más sencillo y amigable
	startStr = strings.TrimSpace(startStr)
	var startDt time.Time
	var err error

	// Lista de formatos soportados, desde el más fácil al más estricto ISO
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006/01/02 15:04",
	}

	parsed := false
	for _, format := range formats {
		if t, errParse := time.ParseInLocation(format, startStr, c.Loc); errParse == nil {
			startDt = t
			parsed = true
			break
		}
	}

	if !parsed {
		return fmt.Errorf("invalid date format: '%s'. Use 'YYYY-MM-DD HH:MM' (e.g. 2026-03-25 17:00)", startStr)
	}

	endDt := startDt.Add(time.Duration(duration) * time.Minute)

	// Recuperamos la "magia" del calendario
	title := fmt.Sprintf("%s (%s - %s)", name, startDt.Format("15:04"), endDt.Format("15:04"))

	projectID := cache.GetProjectID(c.Token, projectName)
	if projectID == "" && projectName != "" {
		fmt.Printf("⚠️ Warning: Project '%s' not found. Using Inbox.\n", projectName)
	}

	taskReq := models.TaskRequest{
		Content:     title,
		DueDatetime: startDt.Format(time.RFC3339),
		ProjectID:   projectID,
		Labels:      labels,
		Description: description,
		Priority:    priority,
	}

	todoistClient := client.New(c.Token)
	taskRes, err := todoistClient.CreateTask(taskReq)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Println("✅ Task created successfully!")
	fmt.Printf("   Title: %s\n", taskRes.Content)
	return nil
}
