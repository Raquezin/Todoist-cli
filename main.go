package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"todoist-cli/internal/client"
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

1. Create a task:
   ./todoist-cli create -name "Meeting" -start "2026-03-25 17:00" -duration 60 -project "Work" -priority 1
   ./todoist-cli create -name "Buy milk" -start "2026-03-25"

2. Fetch tasks (Presets):
   ./todoist-cli fetch foco    # Today's priority 1 tasks, excluding meetings
   ./todoist-cli fetch radar   # Next 7 days, important tasks

3. Fetch tasks (Raw Todoist Filter):
   ./todoist-cli fetch "today & #Work"
   ./todoist-cli fetch "p1 & overdue"

For more details, check the README.md file.`)
}

func run() error {
	_ = godotenv.Load()
	token := strings.TrimSpace(os.Getenv("TODOIST_API_TOKEN"))

	if len(os.Args) < 2 {
		printHelp()
		return fmt.Errorf("no command provided")
	}

	if os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		return nil
	}

	if token == "" {
		return fmt.Errorf("TODOIST_API_TOKEN not found in environment or .env file")
	}

	command := os.Args[1]
	todoistClient := client.New(token)

	switch command {
	case "create":
		createCmd := flag.NewFlagSet("create", flag.ContinueOnError)
		name := createCmd.String("name", "", "Task name (Required)")
		start := createCmd.String("start", "", "Start date, e.g. '2026-03-25' or '2026-03-25 17:00' (Required)")
		duration := createCmd.Int("duration", 0, "Duration in minutes (optional)")
		project := createCmd.String("project", "", "Project name")
		labelsIn := createCmd.String("labels", "", "Comma-separated labels (e.g. important,coding)")
		desc := createCmd.String("desc", "", "Task description")
		priority := createCmd.Int("priority", 4, "Priority 1 (Urgent) to 4 (Normal)")

		if len(os.Args) > 2 && (os.Args[2] == "help" || os.Args[2] == "--help" || os.Args[2] == "-h" || os.Args[2] == "-help") {
			fmt.Println("USAGE:\n  ./todoist-cli create [options]\n\nOPTIONS:")
			createCmd.PrintDefaults()
			return nil
		}

		if err := createCmd.Parse(os.Args[2:]); err != nil {
			return err
		}

		if *name == "" || *start == "" {
			return fmt.Errorf("-name and -start flags are required.\nUsage: ./todoist-cli create -name \"Task\" -start \"YYYY-MM-DD\" or \"YYYY-MM-DD HH:MM\" (also accepts / instead of -)")
		}

		if *duration < 0 {
			return fmt.Errorf("-duration must be a positive integer")
		}

		if *priority < 1 || *priority > 4 {
			return fmt.Errorf("-priority must be between 1 (Urgent) and 4 (Normal)")
		}

		var labels []string
		if *labelsIn != "" {
			for _, e := range strings.Split(*labelsIn, ",") {
				trimmed := strings.TrimSpace(e)
				if trimmed != "" {
					labels = append(labels, trimmed)
				}
			}
		}

		creator := task.NewCreator(todoistClient)
		if err := creator.Create(*name, *start, *duration, *project, labels, *desc, *priority); err != nil {
			return err
		}

	case "fetch":
		if len(os.Args) > 2 && (os.Args[2] == "help" || os.Args[2] == "--help" || os.Args[2] == "-h" || os.Args[2] == "-help") {
			fmt.Println("USAGE:\n  ./todoist-cli fetch <preset|filter>\n\nEXAMPLES:\n  ./todoist-cli fetch foco\n  ./todoist-cli fetch \"today & #Work\"")
			return nil
		}
		if len(os.Args) < 3 {
			return fmt.Errorf("a command or filter is required for fetch.\nExample: ./todoist-cli fetch foco\nExample: ./todoist-cli fetch \"today & #Work\"")
		}
		queryName := os.Args[2]
		fetcher := task.NewFetcher(todoistClient)
		if err := fetcher.Fetch(queryName); err != nil {
			return err
		}

	case "help":
		printHelp()

	default:
		printHelp()
		return fmt.Errorf("command '%s' not recognized", command)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
}
