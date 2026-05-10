# Todoist CLI 🤖

A powerful and user-friendly command-line interface for Todoist, designed specifically to play well with Google Calendar integrations and provide a fluid experience for AI agents and power users.

## ✨ Features

- **"Calendar Magic" Task Creation**: Automatically formats task titles with start and end times (e.g., `Meeting (15:00 - 16:00)`) making it perfectly compatible with two-way calendar sync systems.
- **Human-Friendly Dates**: Forget strict ISO formats. Use simple `YYYY-MM-DD HH:MM` date strings.
- **Flexible Fetching**: Pull your tasks using built-in or custom presets, OR execute any Raw Todoist filter query directly from the terminal.
- **Custom Presets**: Add, edit, delete fetch presets via the `presets` command. Override built-in ones or create new ones.
- **Local Caching**: Fast project and section ID resolution without repeatedly hitting the Todoist API.

## 🚀 Installation

1. Clone the repository
2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build the CLI:

   ```bash
   go build -o todoist-cli .
   ```

4. Set your Todoist API token via environment variable or `.env` file:

   ```bash
   export TODOIST_API_TOKEN=your_todoist_token_here
   ```

   Or create a `.env` file in the root directory:

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
./todoist-cli create -name "Learn Go" -start "2026-03-25 17:00" -duration 120 -project "Learning" -section "Active" -labels "coding,focus" -priority 1
```

**Required flags:**

- `-name`: Task name (e.g., "Review PRs")
- `-start`: Start date. Accepts `YYYY-MM-DD` for all-day tasks or `YYYY-MM-DD HH:MM` to include a time block (e.g., `2026-03-25` or `2026-03-25 17:00`).

**Optional flags:**

- `-duration`: Duration in minutes. Optional, no default. If specified, it will be added to the task on Todoist.
- `-project`: Target project name (it will be fuzzy-matched locally). If omitted, goes to Inbox.
- `-section`: Target section name within the project. Use with `-project` for unambiguous matching.
- `-labels`: Comma-separated labels (e.g., `important,coding`).
- `-desc`: Task description text.
- `-priority`: Priority level (1-4). `1` is Urgent/Red, `4` is Normal.

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

**Custom Presets:**
Define your own presets or override built-in ones. Presets are stored in `.cache/presets.json`.

```bash
# Generate a template from built-in defaults
./todoist-cli presets init

# List all active presets
./todoist-cli presets

# Add, edit, or delete presets
./todoist-cli presets add sprint "(today | overdue) & #Sprint"
./todoist-cli presets edit foco "today & p1"
./todoist-cli presets delete sprint
```

Then use them with `fetch`:

```bash
./todoist-cli fetch sprint
```

### 3. Managing Presets (`presets`)

```bash
./todoist-cli presets                  # List all active presets
./todoist-cli presets init             # Generate .cache/presets.json from defaults
./todoist-cli presets add <name> <q>   # Add a preset
./todoist-cli presets edit <name> <q>  # Edit or override a preset
./todoist-cli presets delete <name>    # Delete a preset
./todoist-cli presets help             # Show usage
```

Built-in presets (`foco`, `radar`) cannot be deleted, only overridden with `edit`.

## 📂 Project Structure

- `main.go`: Main entry point and CLI flag parsing
- `internal/task/`: Core logic for task creation (Calendar magic) and dynamic fetching
- `internal/client/`: Unified Todoist REST API client
- `internal/cache/`: Local filesystem caching for quick Project and Section ID lookups
- `internal/models/`: Go structs mapping Todoist JSON responses

## 🛠️ Built With

- Go 1.26.1+
- [godotenv](https://github.com/joho/godotenv) (for loading `.env` securely)
