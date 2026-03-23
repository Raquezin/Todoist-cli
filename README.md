# Todoist CLI 🤖

A powerful and user-friendly command-line interface for Todoist, designed specifically to play well with Google Calendar integrations and provide a fluid experience for AI agents and power users.

## ✨ Features

- **"Calendar Magic" Task Creation**: Automatically formats task titles with start and end times (e.g., `Meeting (15:00 - 16:00)`) making it perfectly compatible with two-way calendar sync systems.
- **Human-Friendly Dates**: Forget strict ISO formats. Use simple `YYYY-MM-DD HH:MM` date strings.
- **Flexible Fetching**: Pull your tasks using built-in presets (`foco`, `radar`) OR execute any Raw Todoist filter query directly from the terminal.
- **Local Project Caching**: Fast project ID resolution without hitting the Todoist API for every command.

## 🚀 Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Build the CLI:
   ```bash
   go build -o todoist-cli cmd/main.go
   ```
4. Create a `.env` file in the root directory and add your Todoist API token:
   ```env
   TODOIST_API_TOKEN=your_todoist_token_here
   ```

## 📖 Usage & Commands

The CLI offers an intuitive set of subcommands. You can always view the quick guide in your terminal by running:
```bash
./todoist-cli help
```

### 1. Creating Tasks (`create`)

Use the `create` subcommand to add new tasks. The title will automatically be appended with the duration block for Calendar integrations.

```bash
./todoist-cli create -name "Learn Go" -start "2026-03-25 17:00" -duration 120 -project "Learning" -labels "coding,focus" -priority 4
```

**Required flags:**
- `-name`: Task name (e.g., "Review PRs")
- `-start`: Start date and time. It accepts multiple formats, but `YYYY-MM-DD HH:MM` is recommended (e.g., `2026-03-25 17:00`)

**Optional flags:**
- `-duration`: Duration in minutes. Default is `60`. This is used to calculate the end time in the title.
- `-project`: Target project name (it will be fuzzy-matched locally). If omitted, goes to Inbox.
- `-labels`: Comma-separated labels (e.g., `important,coding`).
- `-desc`: Task description text.
- `-priority`: Priority level (1-4). `1` is Normal, `4` is Urgent/Red.

### 2. Fetching Tasks (`fetch`)

Use the `fetch` command to list your active tasks. You can use predefined presets or write your own Todoist filter logic.

**Using Presets:**
```bash
# Shows priority 1 tasks for today (excluding meetings/studying by default)
./todoist-cli fetch foco   

# Shows important tasks for the next 7 days
./todoist-cli fetch radar  
```

**Using Raw Todoist Filters:**
If the keyword is not a preset, the CLI treats it as a raw Todoist query string!
```bash
# Fetch tasks matching exact Todoist syntax
./todoist-cli fetch "today & #Work"
./todoist-cli fetch "p1 & overdue"
```

## 📂 Project Structure

- `cmd/main.go`: Main entry point and CLI flag parsing
- `internal/task/`: Core logic for task creation (Calendar magic) and dynamic fetching
- `internal/client/`: Unified Todoist REST API client
- `internal/cache/`: Local filesystem caching for quick Project ID lookups
- `internal/models/`: Go structs mapping Todoist JSON responses

## 🛠️ Built With

- Go 1.21+
- [godotenv](https://github.com/joho/godotenv) (for loading `.env` securely)
