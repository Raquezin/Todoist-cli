package task

import (
	"fmt"
	"os"
	"strings"
	"time"

	"todoist-cli/internal/cache"
	"todoist-cli/internal/client"
	"todoist-cli/internal/models"
)

type Creator struct {
	Client *client.TodoistClient
	Loc    *time.Location
}

func NewCreator(apiClient *client.TodoistClient) *Creator {
	tz := os.Getenv("TZ")
	var loc *time.Location
	var err error

	if tz != "" {
		loc, err = time.LoadLocation(tz)
	}

	if loc == nil || err != nil {
		loc = time.Local
	}

	return &Creator{Client: apiClient, Loc: loc}
}

func (c *Creator) Create(name, startStr string, duration int, projectName string, labels []string, description string, priority int) error {
	startStr = strings.TrimSpace(startStr)
	var startDt time.Time

	dateOnlyFormats := []string{
		"2006-01-02",
		"2006/01/02",
	}

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

	if duration > 0 && !hasTime {
		return fmt.Errorf("duration cannot be used with an all-day task (specify a time like 'YYYY-MM-DD HH:MM')")
	}

	if priority < 1 || priority > 4 {
		return fmt.Errorf("priority must be between 1 (Urgent) and 4 (Normal)")
	}

	if duration < 0 {
		return fmt.Errorf("duration must be a positive integer")
	}

	projectID := cache.GetProjectID(c.Client, projectName)
	if projectID == "" && projectName != "" {
		fmt.Printf("⚠️ Warning: Project '%s' not found. Using Inbox.\n", projectName)
	}

	taskReq := models.TaskRequest{
		Content:     name,
		ProjectID:   projectID,
		Labels:      labels,
		Description: description,
		Priority:    models.ToAPIPriority(priority),
	}

	if hasTime {
		taskReq.DueDatetime = startDt.Format(time.RFC3339)
		if duration > 0 {
			taskReq.Duration = duration
			taskReq.DurationUnit = "minute"
		}
	} else {
		taskReq.DueDate = startDt.Format("2006-01-02")
	}

	taskRes, err := c.Client.CreateTask(taskReq)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Println("✅ Task created successfully!")
	fmt.Printf("   Title: %s\n", taskRes.Content)
	return nil
}
