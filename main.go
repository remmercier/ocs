package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
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

	m.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		if width == 0 {
			return false
		}
		m.currentWidth = width
		m.table.Clear()
		populateTable(m.table, m.sessions, m.visible, width)
		// Update header
		left := "OpenCode Session browser"
		right := "Press ? for help"
		padLen := width - len(left) - len(right)
		if padLen > 0 {
			m.header.SetText(left + strings.Repeat(" ", padLen) + right)
		} else {
			m.header.SetText(left + " " + right)
		}
		return false
	})

	// Set up input capture
	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			row, _ := m.table.GetSelection()
			if row > 1 && row-2 < len(m.sessions) {
				selectedSession := m.sessions[row-2]
				m.selectedCommand = fmt.Sprintf("cd '%s' ; opencode -s %s", selectedSession.Directory, selectedSession.ID)
				m.selectedIndex = row - 2
				m.app.Stop()
			}
		case tcell.KeyEscape:
			m.app.Stop()
		case tcell.KeyCtrlD:
			m.app.Stop()
		}
		switch event.Rune() {
		case 'q', 'Q':
			m.app.Stop()
		case 'r', 'R':
			m.shouldRefresh = true
			m.app.Stop()
		case 'n', 'N':
			m.selectedCommand = "opencode"
			m.app.Stop()
		case 'i', 'I':
			m.visible[0] = !m.visible[0]
			m.table.Clear()
			populateTable(m.table, m.sessions, m.visible, m.currentWidth)
		case 't', 'T':
			m.visible[1] = !m.visible[1]
			m.table.Clear()
			populateTable(m.table, m.sessions, m.visible, m.currentWidth)
		case 'd', 'D':
			m.visible[2] = !m.visible[2]
			m.table.Clear()
			populateTable(m.table, m.sessions, m.visible, m.currentWidth)
		case 'c', 'C':
			m.visible[3] = !m.visible[3]
			m.table.Clear()
			populateTable(m.table, m.sessions, m.visible, m.currentWidth)
		case '?':
			m.app.SetRoot(m.helpModal, false).SetFocus(m.helpModal)
		case 'v', 'V':
			row, _ := m.table.GetSelection()
			if row > 1 && row-2 < len(m.sessions) {
				selectedSession := m.sessions[row-2]
				if _, err := exec.LookPath("python3"); err == nil {
					if _, err := exec.LookPath("ocs_messages.py"); err == nil {
						pager := "more"
						if _, err := exec.LookPath("glow"); err == nil {
							pager = "glow -p"
						} else if _, err := exec.LookPath("less"); err == nil {
							pager = "less"
						}
						m.selectedCommand = fmt.Sprintf("ocs_messages.py %s | %s", selectedSession.ID, pager)
						m.selectedIndex = row - 2
						m.app.Stop()
					}
				}
			}
		}
		return event
	})

	if err := m.app.SetRoot(m.flex, true).Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
	return m
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
