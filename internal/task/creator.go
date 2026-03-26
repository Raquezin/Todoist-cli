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

	// Formatos de solo fecha
	dateOnlyFormats := []string{
		"2006-01-02",
		"2006/01/02",
	}

	// Lista de formatos con hora soportados
	timeFormats := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006/01/02 15:04",
	}

	parsed := false
	hasTime := false

	for _, format := range timeFormats {
		if t, errParse := time.ParseInLocation(format, startStr, c.Loc); errParse == nil {
			startDt = t
			parsed = true
			hasTime = true
			break
		}
	}

	if !parsed {
		for _, format := range dateOnlyFormats {
			if t, errParse := time.ParseInLocation(format, startStr, c.Loc); errParse == nil {
				startDt = t
				parsed = true
				hasTime = false
				break
			}
		}
	}

	if !parsed {
		return fmt.Errorf("invalid date format: '%s'. Use 'YYYY-MM-DD' or 'YYYY-MM-DD HH:MM'", startStr)
	}

	projectID := cache.GetProjectID(c.Token, projectName)
	if projectID == "" && projectName != "" {
		fmt.Printf("⚠️ Warning: Project '%s' not found. Using Inbox.\n", projectName)
	}

	title := name

	// Mapear prioridad (UI de Todoist: 1 es urgente, API: 4 es urgente)
	apiPriority := 5 - priority
	if apiPriority < 1 {
		apiPriority = 1
	} else if apiPriority > 4 {
		apiPriority = 4
	}

	taskReq := models.TaskRequest{
		ProjectID:   projectID,
		Labels:      labels,
		Description: description,
		Priority:    apiPriority,
	}

	if hasTime {
		endDt := startDt.Add(time.Duration(duration) * time.Minute)
		// Recuperamos la "magia" del calendario
		title = fmt.Sprintf("%s (%s - %s)", name, startDt.Format("15:04"), endDt.Format("15:04"))
		taskReq.DueDatetime = startDt.Format(time.RFC3339)
		taskReq.Duration = duration
		taskReq.DurationUnit = "minute"
	} else {
		taskReq.DueDate = startDt.Format("2006-01-02")
	}

	taskReq.Content = title

	todoistClient := client.New(c.Token)
	taskRes, err := todoistClient.CreateTask(taskReq)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Println("✅ Task created successfully!")
	fmt.Printf("   Title: %s\n", taskRes.Content)
	return nil
}
