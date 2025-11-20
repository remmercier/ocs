package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func parseFlags() (string, bool) {
	dir := flag.String("dir", os.Getenv("HOME")+"/.local/share/opencode/storage/session", "directory to scan for session JSON files")
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debug output")
	flag.Parse()
	return *dir, debug
}

func checkDirectory(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", dir)
	}
}

func loadSessions(dir string) (Sessions, error) {
	var sessions Sessions
	err := sessions.Scan(dir)
	if err != nil {
		return nil, err
	}
	sessions.Sort()
	return sessions, nil
}

func logSessions(sessions Sessions, debug bool) {
	if debug {
		log.Printf("Found %d sessions", len(sessions))
		if len(sessions) > 0 {
			log.Printf("First session: ID=%s Title=%s Dir=%s Created=%v", sessions[0].ID, sessions[0].Title, sessions[0].Directory, sessions[0].Time.Created)
		}
	}
}

func runProgram(sessions Sessions, lastCursor int) model {
	m := newModel(sessions, lastCursor)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v", err)
	}
	return finalModel.(model)
}

func handleCommand(selectedCommand string, selectedIndex int) int {
	fmt.Println(selectedCommand)
	cmd := exec.Command("/bin/bash", "-c", selectedCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Printf("Command failed: %v", err)
	}
	return selectedIndex
}

func main() {
	dir, debug := parseFlags()
	if debug {
		log.Printf("Scanning dir: %s", dir)
	}
	checkDirectory(dir)
	sessions, err := loadSessions(dir)
	if err != nil {
		log.Fatalf("Error loading sessions: %v", err)
	}
	logSessions(sessions, debug)

	lastCursor := -1
	for {
		if debug {
			log.Printf("Setting cursor to %d", lastCursor)
		}
		finalModel := runProgram(sessions, lastCursor)
		if finalModel.shouldRefresh {
			sessions, err = loadSessions(dir)
			if err != nil {
				log.Fatalf("Error reloading sessions: %v", err)
			}
			lastCursor = -1
		} else if finalModel.selectedCommand != "" {
			lastCursor = handleCommand(finalModel.selectedCommand, finalModel.selectedIndex)
		} else {
			break
		}
	}
}
