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

func truncateWithEllipsis(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen < 5 {
		return text[:maxLen]
	}
	startLen := (maxLen - 3) / 2
	endLen := maxLen - 3 - startLen
	return text[:startLen] + "..." + text[len(text)-endLen:]
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table           table.Model
	sessions        []Session
	rows            []table.Row
	columns         []table.Column
	baseWidths      []int
	visibleColumns  []bool
	width           int
	height          int
	selectedCommand string
	selectedIndex   int
	shouldRefresh   bool
	showHelp        bool
}

func (m model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width != m.width || msg.Height-3 != m.height {
			oldCursor := m.table.Cursor()
			m.width = msg.Width
			m.height = msg.Height - 3
			m.rebuildTable()
			m.table.SetCursor(oldCursor)
		}
	case tea.KeyMsg:
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		switch msg.String() {
		case "?":
			m.showHelp = true
			return m, nil
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.shouldRefresh = true
			return m, tea.Quit
		case "n":
			m.selectedCommand = "opencode"
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
		case "i":
			m.visibleColumns[0] = !m.visibleColumns[0]
			m.rebuildTable()
		case "t":
			m.visibleColumns[1] = !m.visibleColumns[1]
			m.rebuildTable()
		case "c":
			m.visibleColumns[3] = !m.visibleColumns[3]
			m.rebuildTable()
		case "d":
			m.visibleColumns[2] = !m.visibleColumns[2]
			m.rebuildTable()
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.showHelp {
		return `Help:
↑/↓: Navigate sessions
Enter: Open selected session
r: Refresh sessions
n: New session
i: Toggle ID column
t: Toggle Title column
c: Toggle Created column
d: Toggle Directory column
?: Show this help
q/Ctrl+C: Quit

Press any key to close help`
	}
	left := " OpenCode session viewer"
	right := "press '?' for help  "
	header := lipgloss.JoinHorizontal(lipgloss.Top, left, lipgloss.PlaceHorizontal(m.width-lipgloss.Width(left), lipgloss.Right, right))
	styledHeader := lipgloss.NewStyle().Bold(true).Render(header)
	return styledHeader + "\n\n" + m.table.View()
}

func (m *model) rebuildTable() {
	// calculate widths
	totalPercent := 0
	for i, vis := range m.visibleColumns {
		if vis {
			totalPercent += m.baseWidths[i]
		}
	}
	newColumns := []table.Column{}
	colIndexMap := []int{} // to map visible index to original
	for i, vis := range m.visibleColumns {
		if vis {
			w := m.baseWidths[i] * m.width / totalPercent
			newColumns = append(newColumns, table.Column{Title: m.columns[i].Title, Width: w})
			colIndexMap = append(colIndexMap, i)
		}
	}
	// rebuild rows
	newRows := make([]table.Row, len(m.rows))
	for i, row := range m.rows {
		newRow := table.Row{}
		for _, origIdx := range colIndexMap {
			newRow = append(newRow, row[origIdx])
		}
		newRows[i] = newRow
	}
	// apply truncate for directory if visible
	dirVisibleIndex := -1
	for j, origIdx := range colIndexMap {
		if origIdx == 2 {
			dirVisibleIndex = j
			break
		}
	}
	if dirVisibleIndex >= 0 {
		dirW := newColumns[dirVisibleIndex].Width
		home, _ := os.UserHomeDir()
		for i := range newRows {
			dir := m.sessions[i].Directory
			if strings.HasPrefix(dir, home) {
				dir = "~" + dir[len(home):]
			}
			dir = truncateWithEllipsis(dir, dirW-2)
			newRows[i][dirVisibleIndex] = dir
		}
	}
	// apply styling for non-existent directories
	for i, row := range newRows {
		if _, err := os.Stat(m.sessions[i].Directory); os.IsNotExist(err) {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
			for j := range row {
				row[j] = style.Render(row[j])
			}
		}
	}
	// set table
	m.table = table.New(
		table.WithColumns(newColumns),
		table.WithRows(newRows),
		table.WithFocused(true),
		table.WithHeight(m.height),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("15"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.table.SetStyles(s)
	sumW := 0
	for _, col := range newColumns {
		sumW += col.Width
	}
	m.table.SetWidth(sumW)
}

func newModel(sessions []Session, cursor int) model {
	width := 80
	height := 20
	baseWidths := []int{15, 35, 30, 20}
	visibleColumns := []bool{false, true, true, true}

	columns := []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Title", Width: 0},
		{Title: "Directory", Width: 0},
		{Title: "Created", Width: 0},
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
		row := table.Row{
			s.ID,
			s.Title,
			dir,
			createdTime,
		}
		rows[i] = row
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
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("15"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	t.SetWidth(width)

	m := model{
		table:         t,
		sessions:      sessions,
		rows:          rows,
		columns:       columns,
		width:         width,
		height:        height,
		selectedIndex: -1,
		shouldRefresh: false,
		showHelp:      false,
	}
	m.baseWidths = baseWidths
	m.visibleColumns = visibleColumns
	m.rebuildTable()
	return m
}
