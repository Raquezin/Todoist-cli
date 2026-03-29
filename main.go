package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"todoist-cli/internal/task"
)

func printHelp() {
	fmt.Println(`🤖 Todoist CLI - Your terminal Todoist assistant

USAGE:
  ./todoist-cli <command> [options]

AVAILABLE COMMANDS:
  create    Create a new task in Todoist with Calendar-friendly formatting
  fetch     Fetch tasks using presets or raw Todoist filters
  help      Show this help message

EXAMPLES:

1. Create a task (Calendar Magic):
   The command will automatically append the time block to the title, e.g., "Meeting (17:00 - 18:00)"
   ./todoist-cli create -name "Meeting" -start "2026-03-25 17:00" -duration 60 -project "Work" -priority 1

2. Fetch tasks (Presets):
   ./todoist-cli fetch foco    # Today's priority 1 tasks, excluding meetings
   ./todoist-cli fetch radar   # Next 7 days, important tasks

3. Fetch tasks (Raw Todoist Filter):
   ./todoist-cli fetch "today & #Work"
   ./todoist-cli fetch "p1 & overdue"

For more details, check the README.md file.`)
}

func main() {
	_ = godotenv.Load()
	token := os.Getenv("TODOIST_API_TOKEN")

	if len(os.Args) >= 2 && (os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
	}

	if token == "" {
		fmt.Println("❌ Error: TODOIST_API_TOKEN not found in environment or .env file")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		createCmd := flag.NewFlagSet("create", flag.ExitOnError)
		name := createCmd.String("name", "", "Task name (Required)")
		start := createCmd.String("start", "", "Start date, e.g. '2026-03-25' or '2026-03-25 17:00' (Required)")
		duration := createCmd.Int("duration", 0, "Duration in minutes (optional)")
		project := createCmd.String("project", "", "Project name")
		labelsIn := createCmd.String("labels", "", "Comma-separated labels (e.g. important,coding)")
		desc := createCmd.String("desc", "", "Task description")
		priority := createCmd.Int("priority", 4, "Priority 1 (Urgent) to 4 (Normal)")

		if err := createCmd.Parse(os.Args[2:]); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		if *name == "" || *start == "" {
			fmt.Println("❌ Error: -name and -start flags are required.")
			fmt.Println("Usage: ./todoist-cli create -name \"Task\" -start \"YYYY-MM-DD\" or \"YYYY-MM-DD HH:MM\"")
			os.Exit(1)
		}

		if *priority < 1 || *priority > 4 {
			fmt.Println("❌ Error: -priority must be between 1 (Urgent) and 4 (Normal).")
			os.Exit(1)
		}

		var labels []string
		if *labelsIn != "" {
			for _, e := range strings.Split(*labelsIn, ",") {
				labels = append(labels, strings.TrimSpace(e))
			}
		}

		creator := task.NewCreator(token)
		if err := creator.Create(*name, *start, *duration, *project, labels, *desc, *priority); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "fetch":
		if len(os.Args) < 3 {
			fmt.Println("❌ Error: A command or filter is required for fetch.")
			fmt.Println("Example: ./todoist-cli fetch foco")
			fmt.Println("Example: ./todoist-cli fetch \"today & #Work\"")
			os.Exit(1)
		}
		queryName := os.Args[2]
		fetcher := task.NewFetcher(token)
		if err := fetcher.Fetch(queryName); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

	case "help":
		printHelp()

	default:
		fmt.Printf("❌ Command '%s' not recognized.\n\n", command)
		printHelp()
		os.Exit(1)
	}
}
