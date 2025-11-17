package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var debug bool
	dir := flag.String("dir", os.Getenv("HOME")+"/.local/share/opencode/storage/session", "directory to scan for session JSON files")
	flag.BoolVar(&debug, "debug", false, "enable debug output")
	flag.Parse()

	if debug {
		log.Printf("Scanning dir: %s", *dir)
	}
	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", *dir)
	}
	sessions, err := scanSessions(*dir)
	if err != nil {
		log.Fatalf("Error scanning sessions: %v", err)
	}
	if debug {
		log.Printf("Found %d sessions", len(sessions))
		if len(sessions) > 0 {
			log.Printf("First session: ID=%s Title=%s Dir=%s Created=%v", sessions[0].ID, sessions[0].Title, sessions[0].Directory, sessions[0].Time.Created)
		}
	}

	lastCursor := -1
	for {
		if debug {
			log.Printf("Setting cursor to %d", lastCursor)
		}
		m := newModel(sessions, lastCursor)
		p := tea.NewProgram(m, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			log.Fatalf("Error running program: %v", err)
		}
		if finalModel.(model).selectedCommand != "" {
			lastCursor = finalModel.(model).selectedIndex
			fmt.Println(finalModel.(model).selectedCommand)
			cmd := exec.Command("/bin/bash", "-c", finalModel.(model).selectedCommand)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				log.Printf("Command failed: %v", err)
			}
			// continue loop
		} else {
			// quit
			break
		}
	}
}
