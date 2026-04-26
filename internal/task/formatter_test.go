package task

import (
	"testing"
	"time"

	"todoist-cli/internal/models"
)

func TestFormatTask(t *testing.T) {
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	projectMap := map[string]string{
		"proj1": "Work",
		"proj2": "Personal",
	}

	tests := []struct {
		name     string
		task     models.FilteredTask
		expected string
	}{
		{
			name: "No date no project",
			task: models.FilteredTask{
				Content:  "Do something",
				Priority: 4,
			},
			expected: "Do something | - | P1 | Inbox",
		},
		{
			name: "Date only same year",
			task: models.FilteredTask{
				Content:  "Do something",
				Priority: 3,
				Due: &models.Due{
					Date: "2026-03-27",
				},
				ProjectID: "proj1",
			},
			expected: "Do something | 27 Mar | P2 | Work",
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
			expected: "Do something | 27 Mar 2027 | P3 | Inbox",
		},
		{
			name: "Datetime same year",
			task: models.FilteredTask{
				Content:  "Meeting",
				Priority: 4,
				Due: &models.Due{
					Datetime: "2026-03-27T14:30:00Z",
				},
				ProjectID: "proj1",
			},
			expected: "Meeting | 27 Mar 14:30 | P1 | Work",
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
			expected: "Meeting next year | 27 Mar 2027 14:30 | P1 | Inbox",
		},
		{
			name: "With duration minutes",
			task: models.FilteredTask{
				Content:  "Meeting",
				Priority: 4,
				Due: &models.Due{
					Datetime: "2026-03-27T14:30:00Z",
				},
				ProjectID: "proj1",
				Duration: &models.Duration{
					Amount: 30,
					Unit:   "minute",
				},
			},
			expected: "Meeting | 27 Mar 14:30 | P1 | Work | ⏱ 30m",
		},
		{
			name: "With duration hours",
			task: models.FilteredTask{
				Content:  "Deep Work",
				Priority: 4,
				Due: &models.Due{
					Date: "2026-03-27",
				},
				ProjectID: "proj1",
				Duration: &models.Duration{
					Amount: 2,
					Unit:   "hour",
				},
			},
			expected: "Deep Work | 27 Mar | P1 | Work | ⏱ 2h",
		},
		{
			name: "With labels",
			task: models.FilteredTask{
				Content:  "Task with labels",
				Priority: 3,
				Due: &models.Due{
					Date: "2026-03-27",
				},
				ProjectID: "proj2",
				Labels:    []string{"urgent", "home"},
			},
			expected: "Task with labels | 27 Mar | P2 | Personal | @urgent @home",
		},
		{
			name: "Full: labels + duration",
			task: models.FilteredTask{
				Content:  "Full task",
				Priority: 2,
				Due: &models.Due{
					Datetime: "2026-03-27T10:00:00Z",
				},
				ProjectID: "proj1",
				Labels:    []string{"coding"},
				Duration: &models.Duration{
					Amount: 90,
					Unit:   "minute",
				},
			},
			expected: "Full task | 27 Mar 10:00 | P3 | Work | @coding | ⏱ 90m",
		},
		{
			name: "Due string fallback",
			task: models.FilteredTask{
				Content:  "Recurring",
				Priority: 4,
				Due: &models.Due{
					String: "every day at 10:00",
				},
				ProjectID: "proj2",
			},
			expected: "Recurring | every day at 10:00 | P1 | Personal",
		},
		{
			name: "Long due string truncated",
			task: models.FilteredTask{
				Content:  "CIN",
				Priority: 4,
				Due: &models.Due{
					String: "Every 2 weeks Mon @ 17:40 ending 2026-06-01",
				},
			},
			expected: "CIN | Every 2 weeks Mon @ 17:40 ending... | P1 | Inbox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTask(tt.task, now, projectMap)
			if result != tt.expected {
				t.Errorf("FormatTask() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatDueEdgeCases(t *testing.T) {
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)

	// Test unparseable date
	due := &models.Due{
		Datetime: "invalid-date-format",
	}
	res := formatDue(due, now)
	if res != "invalid-date-format" {
		t.Errorf("Expected fallback to raw datetime string, got %s", res)
	}

	// Test unparseable date but valid date fallback
	due2 := &models.Due{
		Date: "invalid-date-format",
	}
	res2 := formatDue(due2, now)
	if res2 != "invalid-date-format" {
		t.Errorf("Expected fallback to raw date string, got %s", res2)
	}

	// Test nil due
	res3 := formatDue(nil, now)
	if res3 != "-" {
		t.Errorf("Expected -, got %s", res3)
	}
}
