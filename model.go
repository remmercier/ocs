package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table           table.Model
	sessions        []Session
	rows            []table.Row
	columns         []table.Column
	width           int
	height          int
	selectedCommand string
	selectedIndex   int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width != m.width || msg.Height-1 != m.height {
			oldCursor := m.table.Cursor()
			m.width = msg.Width
			m.height = msg.Height - 1
			m.columns[0].Width = m.width * 15 / 100
			m.columns[1].Width = m.width * 35 / 100
			m.columns[2].Width = m.width * 30 / 100
			m.columns[3].Width = m.width * 20 / 100
			m.table = table.New(
				table.WithColumns(m.columns),
				table.WithRows(m.rows),
				table.WithFocused(true),
				table.WithHeight(m.height),
			)
			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true).
				Bold(false)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(false)
			m.table.SetStyles(s)
			m.table.SetWidth(m.width)
			m.table.SetCursor(oldCursor)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedIndex := m.table.Cursor()
			if selectedIndex >= 0 && selectedIndex < len(m.sessions) {
				selectedSession := m.sessions[selectedIndex]
				m.selectedCommand = fmt.Sprintf("cd '%s' ; opencode -s %s", selectedSession.Directory, selectedSession.ID)
				m.selectedIndex = selectedIndex
				return m, tea.Quit
			}
			return m, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.table.View() + "\n"
}

func newModel(sessions []Session, cursor int) model {
	width := 80
	height := 20
	idW := width * 15 / 100
	titleW := width * 35 / 100
	dirW := width * 30 / 100
	createdW := width * 20 / 100

	columns := []table.Column{
		{Title: "ID", Width: idW},
		{Title: "Title", Width: titleW},
		{Title: "Directory", Width: dirW},
		{Title: "Created", Width: createdW},
	}

	home, _ := os.UserHomeDir()
	rows := make([]table.Row, len(sessions))
	for i, s := range sessions {
		createdTime := ""
		if s.Time.Created != nil {
			switch v := s.Time.Created.(type) {
			case float64:
				createdTime = time.UnixMilli(int64(v)).Format("2006-01-02 15:04")
			case string:
				if i, err := strconv.ParseInt(v, 10, 64); err == nil {
					createdTime = time.UnixMilli(i).Format("2006-01-02 15:04")
				}
			}
		}
		dir := s.Directory
		if strings.HasPrefix(dir, home) {
			dir = "~" + dir[len(home):]
		}
		rows[i] = table.Row{
			s.ID,
			s.Title,
			dir,
			createdTime,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	if cursor >= 0 && cursor < len(rows) {
		t.SetCursor(cursor)
	}

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	t.SetWidth(width)

	return model{
		table:         t,
		sessions:      sessions,
		rows:          rows,
		columns:       columns,
		width:         width,
		height:        height,
		selectedIndex: -1,
	}
}
