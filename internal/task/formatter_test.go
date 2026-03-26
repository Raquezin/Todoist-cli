package task

import (
	"testing"
	"time"

	"todoist-cli/internal/models"
)

func TestFormatTask(t *testing.T) {
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     models.FilteredTask
		expected string
	}{
		{
			name: "No date",
			task: models.FilteredTask{
				Content:  "Do something",
				Priority: 4,
			},
			expected: "👉 [no date] (P1) Do something",
		},
		{
			name: "Date only same year",
			task: models.FilteredTask{
				Content:  "Do something",
				Priority: 3,
				Due: &models.Due{
					Date: "2026-03-27",
				},
			},
			expected: "👉 [27 mar] (P2) Do something",
		},
		{
			name: "Date only different year",
			task: models.FilteredTask{
				Content:  "Do something",
				Priority: 2,
				Due: &models.Due{
					Date: "2027-03-27",
				},
			},
			expected: "👉 [27 mar 2027] (P3) Do something",
		},
		{
			name: "Datetime same year",
			task: models.FilteredTask{
				Content:  "Meeting",
				Priority: 4,
				Due: &models.Due{
					Datetime: "2026-03-27T14:30:00Z",
				},
			},
			expected: "👉 [27 mar 14:30] (P1) Meeting",
		},
		{
			name: "Datetime different year",
			task: models.FilteredTask{
				Content:  "Meeting next year",
				Priority: 4,
				Due: &models.Due{
					Datetime: "2027-03-27T14:30:00Z",
				},
			},
			expected: "👉 [27 mar 2027 14:30] (P1) Meeting next year",
		},
		{
			name: "With duration minutes",
			task: models.FilteredTask{
				Content:  "Meeting",
				Priority: 4,
				Due: &models.Due{
					Datetime: "2026-03-27T14:30:00Z",
				},
				Duration: &models.Duration{
					Amount: 30,
					Unit:   "minute",
				},
			},
			expected: "👉 [27 mar 14:30] (P1) Meeting [⏱️ 30m]",
		},
		{
			name: "With duration hours",
			task: models.FilteredTask{
				Content:  "Deep Work",
				Priority: 4,
				Due: &models.Due{
					Date: "2026-03-27",
				},
				Duration: &models.Duration{
					Amount: 2,
					Unit:   "hour",
				},
			},
			expected: "👉 [27 mar] (P1) Deep Work [⏱️ 2h]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTask(tt.task, now)
			if result != tt.expected {
				t.Errorf("FormatTask() = %q, want %q", result, tt.expected)
			}
		})
	}
}
