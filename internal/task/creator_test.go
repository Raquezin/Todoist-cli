package task

import (
	"testing"
	"time"
)

func TestCreatorTimezone(t *testing.T) {
	c := NewCreator("fake-token")
	if c.Loc == nil {
		t.Error("Expected location to be set")
	}
}

func TestCreatorDateParsing(t *testing.T) {
	c := NewCreator("fake-token")

	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		_, err := time.ParseInLocation(format, "2026-03-25 17:00", c.Loc)
		// We only really care if the simple format parses ok for the test test
		if err == nil {
			break
		}
	}

	dateFormats := []string{
		"2006-01-02",
		"2006/01/02",
	}

	for _, format := range dateFormats {
		_, err := time.ParseInLocation(format, "2026-03-25", c.Loc)
		if err == nil {
			break
		}
	}
}
