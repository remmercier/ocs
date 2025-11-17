# Design Document for ocs - Opencode Session Lister

## Overview
`ocs` is a proof-of-concept terminal user interface (TUI) application written in Go using the Bubble Tea framework. It fulfills the description in README.md by scanning the opencode session storage directory for JSON files, extracting key metadata (session ID, working directory, title, and creation timestamp), and displaying the information in an interactive table within the terminal.

## Requirements
- **Directory Scanning**: Scan a configurable directory (default to opencode session storage) for JSON files.
- **Data Extraction**: Parse each JSON file to extract session metadata.
- **TUI Display**: Use Bubble Tea to render an interactive table showing the sessions.
- **Navigation**: Basic keyboard navigation (up/down arrows, enter to select, q to quit).
- **Error Handling**: Gracefully handle missing directories, invalid JSON, etc.
- **Proof-of-Concept Scope**: Focus on displaying the table; future iterations can add features like filtering, sorting, or actions on selected sessions.

## Architecture
- **Main Entry Point**: `main.go` initializes the Bubble Tea program and starts the TUI.
- **Data Layer**: Functions to scan directories and parse JSON files.
- **Model Layer**: Go structs to represent session data.
- **UI Layer**: Bubble Tea model, view, and update functions to handle TUI rendering and interactions.

## Dependencies
- `github.com/charmbracelet/bubbletea`: For TUI framework.
- `github.com/charmbracelet/bubbles/table`: For table rendering.
- Standard library: `os`, `path/filepath`, `encoding/json`, `time`, `log`.

## Data Structures
```go
type Session struct {
    ID         string    `json:"id"`
    WorkingDir string    `json:"working_directory"`
    Title      string    `json:"title"`
    Created    time.Time `json:"created"`
}
```

## UI Design
- **Table Columns**: ID, Title, Working Directory, Created (formatted as YYYY-MM-DD HH:MM).
- **Layout**: Full-screen table with header.
- **Interactions**:
  - Arrow keys: Navigate rows.
  - Enter: Select row (for future actions).
  - q: Quit application.
- **Styling**: Use Bubble Tea styles for borders, colors.

## Implementation Steps
1. Initialize Go module: `go mod init ocs`.
2. Install dependencies: `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/bubbles`.
3. Define the `Session` struct in a separate file (e.g., `session.go`).
4. Implement directory scanning function: `scanSessions(dir string) ([]Session, error)`.
5. Implement JSON parsing: Helper function to parse a single JSON file into `Session`.
6. Create Bubble Tea model struct with table and sessions data.
7. Implement `Init`, `Update`, and `View` methods for the model.
8. In `main.go`, load sessions, initialize model, and start the program.
9. Add error handling and logging.
10. Test with sample JSON files in the session directory.

## Future Improvements
- Add configuration file for session directory path.
- Implement sorting and filtering.
- Add actions on selected sessions (e.g., open in editor).
- Improve error messages and user feedback.
- Add unit tests for parsing and scanning functions.
