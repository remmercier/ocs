package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func getTime(created interface{}) time.Time {
	switch v := created.(type) {
	case float64:
		return time.UnixMilli(int64(v))
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.UnixMilli(i)
		}
	}
	return time.Time{}
}

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

func loadAndSortSessions(dir string) []Session {
	sessions, err := scanSessions(dir)
	if err != nil {
		log.Fatalf("Error scanning sessions: %v", err)
	}
	sort.Slice(sessions, func(i, j int) bool {
		ti := getTime(sessions[i].Time.Created)
		tj := getTime(sessions[j].Time.Created)
		return ti.After(tj)
	})
	return sessions
}

func logSessions(sessions []Session, debug bool) {
	if debug {
		log.Printf("Found %d sessions", len(sessions))
		if len(sessions) > 0 {
			log.Printf("First session: ID=%s Title=%s Dir=%s Created=%v", sessions[0].ID, sessions[0].Title, sessions[0].Directory, sessions[0].Time.Created)
		}
	}
}

func runProgram(sessions []Session, lastCursor int) model {
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
	sessions := loadAndSortSessions(dir)
	logSessions(sessions, debug)

	lastCursor := -1
	for {
		if debug {
			log.Printf("Setting cursor to %d", lastCursor)
		}
		finalModel := runProgram(sessions, lastCursor)
		if finalModel.shouldRefresh {
			sessions = loadAndSortSessions(dir)
			lastCursor = -1
		} else if finalModel.selectedCommand != "" {
			lastCursor = handleCommand(finalModel.selectedCommand, finalModel.selectedIndex)
		} else {
			break
		}
	}
}
