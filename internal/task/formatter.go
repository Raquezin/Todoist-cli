package task

import (
	"fmt"
	"strings"
	"time"

	"todoist-cli/internal/models"
	"todoist-cli/internal/sanitize"
)

const maxDateLen = 35
const maxContentLen = 500
const maxLabelLen = 100
const maxProjectLen = 120

func formatDue(due *models.Due, now time.Time) string {
	if due == nil {
		return "-"
	}

	var parsed time.Time
	var err error
	hasTime := false

	if due.Datetime != "" {
		parsed, err = time.Parse(time.RFC3339, due.Datetime)
		if err == nil {
			hasTime = true
		}
	}

	if (err != nil || due.Datetime == "") && due.Date != "" {
		parsed, err = time.Parse("2006-01-02", due.Date)
		if err == nil {
			hasTime = false
		}
	}

	if err == nil && !parsed.IsZero() {
		var fmtStr string
		if parsed.Year() != now.Year() {
			if hasTime {
				fmtStr = "02 Jan 2006 15:04"
			} else {
				fmtStr = "02 Jan 2006"
			}
		} else {
			if hasTime {
				fmtStr = "02 Jan 15:04"
			} else {
				fmtStr = "02 Jan"
			}
		}
		return parsed.Format(fmtStr)
	}

	if due.String != "" {
		return sanitize.TerminalLimit(due.String, maxDateLen)
	}

	dateVal := due.Datetime
	if dateVal == "" {
		dateVal = due.Date
	}
	if dateVal != "" {
		return sanitize.TerminalLimit(dateVal, maxDateLen)
	}
	return "-"
}

func formatLabels(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, len(labels))
	for i, l := range labels {
		parts[i] = "@" + sanitize.TerminalLimit(l, maxLabelLen)
	}
	return strings.Join(parts, " ")
}

func formatDuration(dur *models.Duration) string {
	if dur == nil || dur.Amount <= 0 {
		return ""
	}
	unit := dur.Unit
	switch unit {
	case "minute":
		unit = "m"
	case "hour":
		unit = "h"
	}
	return fmt.Sprintf("%d%s", dur.Amount, unit)
}

// FormatTask formats a FilteredTask into a readable string.
// projectMap maps project IDs to project names.
func FormatTask(t models.FilteredTask, now time.Time, projectMap map[string]string) string {
	content := sanitize.TerminalLimit(t.Content, maxContentLen)
	content = strings.ReplaceAll(content, "|", "¦")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", "")
	date := formatDue(t.Due, now)
	priority := fmt.Sprintf("P%d", models.ToUIPriority(t.Priority))

	project := projectMap[t.ProjectID]
	if project == "" {
		project = "Inbox"
	}
	project = sanitize.TerminalLimit(project, maxProjectLen)

	labels := formatLabels(t.Labels)
	duration := formatDuration(t.Duration)

	var b strings.Builder
	b.WriteString(content)
	b.WriteString(" | ")
	b.WriteString(date)
	b.WriteString(" | ")
	b.WriteString(priority)
	b.WriteString(" | ")
	b.WriteString(project)

	if labels != "" {
		b.WriteString(" | ")
		b.WriteString(labels)
	}

	if duration != "" {
		b.WriteString(" | ⏱ ")
		b.WriteString(duration)
	}

	return b.String()
}
