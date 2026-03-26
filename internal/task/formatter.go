package task

import (
	"fmt"
	"strings"
	"time"

	"todoist-cli/internal/models"
)

// FormatTask formats a FilteredTask into a readable string representation
// including its date, priority, content, and duration.
func FormatTask(t models.FilteredTask, now time.Time) string {
	dateStr := "no date"
	if t.Due != nil {
		var parsed time.Time
		var err error

		dateVal := t.Due.Datetime
		if dateVal == "" {
			dateVal = t.Due.Date
		}

		// Try to parse as RFC3339 first
		parsed, err = time.Parse(time.RFC3339, dateVal)
		if err != nil {
			// Fallback to simple date
			parsed, err = time.Parse("2006-01-02", dateVal)
		}

		if err == nil {
			if parsed.Year() != now.Year() {
				if parsed.Hour() == 0 && parsed.Minute() == 0 && parsed.Second() == 0 && t.Due.Datetime == "" {
					dateStr = strings.ToLower(parsed.Format("02 Jan 2006"))
				} else {
					dateStr = strings.ToLower(parsed.Format("02 Jan 2006 15:04"))
				}
			} else {
				if parsed.Hour() == 0 && parsed.Minute() == 0 && parsed.Second() == 0 && t.Due.Datetime == "" {
					dateStr = strings.ToLower(parsed.Format("02 Jan"))
				} else {
					dateStr = strings.ToLower(parsed.Format("02 Jan 15:04"))
				}
			}
		} else {
			dateStr = dateVal
		}
	}

	durStr := ""
	if t.Duration != nil && t.Duration.Amount > 0 {
		unitStr := t.Duration.Unit
		if unitStr == "minute" {
			unitStr = "m"
		} else if unitStr == "hour" {
			unitStr = "h"
		}
		durStr = fmt.Sprintf(" [⏱️ %d%s]", t.Duration.Amount, unitStr)
	}

	uiPriority := 5 - t.Priority
	if uiPriority < 1 {
		uiPriority = 1
	} else if uiPriority > 4 {
		uiPriority = 4
	}

	return fmt.Sprintf("👉 [%s] (P%d) %s%s", dateStr, uiPriority, t.Content, durStr)
}
