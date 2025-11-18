# ocs - Opencode Session Browser

## Description
`ocs` is a terminal user interface (TUI) application for browsing and launching opencode sessions. It scans the opencode session storage directory for JSON files, extracts metadata (session ID, title, working directory, and creation timestamp), and displays them in an interactive table. You can navigate the table, select a session with Enter to launch it in opencode, and return to the browser after closing opencode.

![Screenshot](docs/screenshot.png)

## Features
- **Interactive Table**: Browse sessions with columns for ID, Title, Directory (abbreviated with ~), and Created timestamp.
- **Full-Screen Mode**: Uses alternate screen for a clean, distraction-free interface.
- **Cursor Persistence**: Remembers the last selected row when returning from opencode.
- **Session Launching**: Runs `opencode -s <session_id>` in the session's working directory.
- **Configurable Directory**: Specify the session storage directory with `-dir` (default: `~/.local/share/opencode/storage/session`).
- **Debug Mode**: Enable debug output with `-debug`.

## Installation
1. Ensure you have Go installed.
2. Clone or download the project.
3. Run `make install` to build and install `ocs` to `~/.local/bin`.
4. Run `ocs` to start the application (ensure `~/.local/bin` is in your PATH).

## Usage
- **Navigation**: Use arrow keys or vim-style keys (h/j/k/l) to move in the table.
- **Select Session**: Press Enter on a session to launch it in opencode.
- **Quit**: Press 'q' or Ctrl+C to exit.
- **Resize**: The table automatically adjusts to terminal size.

## Flags
- `-dir string`: Directory to scan for session JSON files (default `~/.local/share/opencode/storage/session`)
- `-debug`: Enable debug output

## Dependencies
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) for table component
- [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling
