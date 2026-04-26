package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test help command
	os.Args = []string{"todoist-cli", "help"}
	if err := run(); err != nil {
		t.Errorf("Expected nil error for help, got %v", err)
	}

	// Test -h flag
	os.Args = []string{"todoist-cli", "-h"}
	if err := run(); err != nil {
		t.Errorf("Expected nil error for -h, got %v", err)
	}
}

func TestRunNoCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Missing env variable
	os.Setenv("TODOIST_API_TOKEN", "")
	os.Args = []string{"todoist-cli", "create"}
	err := run()
	if err == nil || err.Error() != "TODOIST_API_TOKEN not found in environment or .env file" {
		t.Errorf("Expected token error, got %v", err)
	}

	os.Setenv("TODOIST_API_TOKEN", "fake-token")

	// No command
	os.Args = []string{"todoist-cli"}
	err = run()
	if err == nil || err.Error() != "no command provided" {
		t.Errorf("Expected no command error, got %v", err)
	}

	// Unknown command
	os.Args = []string{"todoist-cli", "unknown"}
	err = run()
	if err == nil || err.Error() != "command 'unknown' not recognized" {
		t.Errorf("Expected unknown command error, got %v", err)
	}
}

func TestRunCreateValidations(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Setenv("TODOIST_API_TOKEN", "fake-token")

	// Missing required flags
	os.Args = []string{"todoist-cli", "create"}
	err := run()
	if err == nil || err.Error() != "-name and -start flags are required.\nUsage: ./todoist-cli create -name \"Task\" -start \"YYYY-MM-DD\" or \"YYYY-MM-DD HH:MM\" (also accepts / instead of -)" {
		t.Errorf("Expected missing flags error, got %v", err)
	}

	// Invalid priority
	os.Args = []string{"todoist-cli", "create", "-name", "A", "-start", "2026-01-01", "-priority", "5"}
	err = run()
	if err == nil || err.Error() != "-priority must be between 1 (Urgent) and 4 (Normal)" {
		t.Errorf("Expected invalid priority error, got %v", err)
	}
}

func TestRunFetchValidations(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Setenv("TODOIST_API_TOKEN", "fake-token")

	// Missing query
	os.Args = []string{"todoist-cli", "fetch"}
	err := run()
	if err == nil || err.Error() != "a command or filter is required for fetch.\nExample: ./todoist-cli fetch foco\nExample: ./todoist-cli fetch \"today & #Work\"" {
		t.Errorf("Expected missing query error, got %v", err)
	}
}

func TestRunCreateSuccess(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Setenv("TODOIST_API_TOKEN", "fake-token")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects" {
			_, _ = w.Write([]byte(`{"results":[{"id":"proj1","name":"Work"}]}`))
			return
		}
		if r.URL.Path == "/tasks" {
			_, _ = w.Write([]byte(`{"content":"Test Task","id":"task1"}`))
			return
		}
	}))
	defer ts.Close()
	os.Setenv("TODOIST_API_URL", ts.URL)
	defer os.Unsetenv("TODOIST_API_URL")

	os.Args = []string{
		"todoist-cli", "create",
		"-name", "Test",
		"-start", "2026-01-01 10:00",
		"-labels", "tag1, tag2",
		"-desc", "A description",
		"-project", "Work",
		"-duration", "30",
	}

	err := run()
	if err != nil {
		t.Errorf("Expected nil, got error: %v", err)
	}
}

func TestRunCreateParseError(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Setenv("TODOIST_API_TOKEN", "fake-token")

	// Force parse error with invalid integer for duration
	os.Args = []string{"todoist-cli", "create", "-duration", "not-an-int"}
	err := run()
	if err == nil {
		t.Errorf("Expected parse error, got nil")
	}
}

func TestRunSubcommandHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Setenv("TODOIST_API_TOKEN", "fake-token")

	// create --help
	os.Args = []string{"todoist-cli", "create", "--help"}
	if err := run(); err != nil {
		t.Errorf("Expected nil error for create --help, got %v", err)
	}

	// fetch --help
	os.Args = []string{"todoist-cli", "fetch", "--help"}
	if err := run(); err != nil {
		t.Errorf("Expected nil error for fetch --help, got %v", err)
	}
}

func TestRunCreateValidationsAdvanced(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Setenv("TODOIST_API_TOKEN", "fake-token")

	// Missing start flag but name present
	os.Args = []string{"todoist-cli", "create", "-name", "A"}
	err := run()
	if err == nil || !strings.Contains(err.Error(), "-name and -start flags are required") {
		t.Errorf("Expected missing start flag error, got %v", err)
	}
}
