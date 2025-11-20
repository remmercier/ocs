package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) toggleColumn(index int) {
	m.visibleColumns[index] = !m.visibleColumns[index]
	m.rebuildTable()
}

func (m model) handleEnter() (model, tea.Cmd) {
	selectedIndex := m.table.Cursor()
	if selectedIndex >= 0 && selectedIndex < len(m.sessions) {
		selectedSession := m.sessions[selectedIndex]
		m.selectedCommand = fmt.Sprintf("cd '%s' ; opencode -s %s", selectedSession.Directory, selectedSession.ID)
		m.selectedIndex = selectedIndex
		return m, tea.Quit
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m *model) handleWindowSizeMsg(msg tea.WindowSizeMsg) {
	if msg.Width != m.width || msg.Height-3 != m.height {
		oldCursor := m.table.Cursor()
		m.width = msg.Width
		m.height = msg.Height - 3
		m.rebuildTable()
		m.table.SetCursor(oldCursor)
	}
}

func (m model) processKey(msg tea.KeyMsg) (model, tea.Cmd) {
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
		return m.handleEnter()
	case "i":
		m.toggleColumn(0)
	case "t":
		m.toggleColumn(1)
	case "c":
		m.toggleColumn(3)
	case "d":
		m.toggleColumn(2)
	}
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleWindowSizeMsg(msg)
	case tea.KeyMsg:
		m, quitCmd := m.processKey(msg)
		if quitCmd != nil {
			return m, quitCmd
		}
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) rebuildTable() {
	newColumns, colIndexMap := m.calculateWidths()
	newRows := m.rebuildRows(colIndexMap)
	m.applyStyling(newRows)
	m.applyTruncation(newRows, newColumns, colIndexMap)
	m.setTable(newColumns, newRows)
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
	tableView := m.table.View()
	lines := strings.Split(tableView, "\n")
	// Reconstruct the table view with a horizontal line after the header
	if len(lines) > 0 {
		headerLine := lines[0]
		rest := ""
		if len(lines) > 1 {
			rest = "\n" + strings.Join(lines[1:], "\n")
		}
		horizontalLine := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("252")).Render(strings.Repeat("─", m.width))
		return styledHeader + "\n\n" + headerLine + "\n" + horizontalLine + rest
	} else {
		return styledHeader + "\n\n" + tableView
	}
}

func sanitize(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError || r < 32 || r > 126 {
			i += size
			continue
		}
		buf.WriteRune(r)
		i += size
	}
	return buf.String()
}

func getCreatedTime(created interface{}) string {
	if created == nil {
		return ""
	}
	switch v := created.(type) {
	case float64:
		return time.UnixMilli(int64(v)).Format("2006-01-02 15:04")
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.UnixMilli(i).Format("2006-01-02 15:04")
		}
	}
	return ""
}

func setTableStyles(t *table.Model) {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(false).
		Bold(true).
		Foreground(lipgloss.Color("15"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
}

func (m *model) calculateWidths() ([]table.Column, []int) {
	totalPercent := 0
	for i, vis := range m.visibleColumns {
		if vis {
			totalPercent += m.baseWidths[i]
		}
	}
	newColumns := []table.Column{}
	colIndexMap := []int{}
	for i, vis := range m.visibleColumns {
		if vis {
			w := m.baseWidths[i] * m.width / totalPercent
			newColumns = append(newColumns, table.Column{Title: m.columns[i].Title, Width: w})
			colIndexMap = append(colIndexMap, i)
		}
	}
	return newColumns, colIndexMap
}

func (m *model) rebuildRows(colIndexMap []int) []table.Row {
	newRows := make([]table.Row, len(m.rows))
	for i, row := range m.rows {
		newRow := table.Row{}
		for _, origIdx := range colIndexMap {
			newRow = append(newRow, row[origIdx])
		}
		newRows[i] = newRow
	}
	return newRows
}

func truncateWithEllipsis(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return string(runes)
	}
	if maxLen < 2 {
		return string(runes[:maxLen])
	}
	startLen := (maxLen - 1) / 2
	endLen := maxLen - 1 - startLen
	return string(runes[:startLen]) + "…" + string(runes[len(runes)-endLen:])
}

func truncateAtEnd(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return string(runes)
	}
	if maxLen < 2 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}

func stripANSI(s string) string {
	var buf strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			for j := i + 2; j < len(s); j++ {
				if s[j] == 'm' {
					i = j + 1
					break
				}
			}
		} else {
			buf.WriteByte(s[i])
			i++
		}
	}
	return buf.String()
}

func extractANSI(s string) string {
	if len(s) > 2 && s[0] == '\x1b' && s[1] == '[' {
		for i := 2; i < len(s); i++ {
			if s[i] == 'm' {
				return s[:i+1]
			}
		}
	}
	return ""
}

func truncateDirectory(dir string, dirW int) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(dir, home) {
		dir = "~" + dir[len(home):]
	}
	return truncateWithEllipsis(dir, dirW-2)
}

func (m *model) applyTruncation(newRows []table.Row, newColumns []table.Column, colIndexMap []int) {
	for j, col := range newColumns {
		w := col.Width
		origIdx := colIndexMap[j]
		if origIdx == 2 {
			w -= 2
		}
		for i := range newRows {
			cell := newRows[i][j]
			ansi := extractANSI(cell)
			plain := stripANSI(cell)
			var truncated string
			if origIdx == 1 { // Title
				truncated = truncateAtEnd(plain, w)
			} else {
				truncated = truncateWithEllipsis(plain, w)
			}
			newRows[i][j] = ansi + truncated
		}
	}
}

func (m *model) applyStyling(newRows []table.Row) {
	for i, row := range newRows {
		if _, err := os.Stat(m.sessions[i].Directory); os.IsNotExist(err) {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
			for j := range row {
				row[j] = style.Render(row[j])
			}
		}
	}
}

func (m *model) setTable(newColumns []table.Column, newRows []table.Row) {
	m.table = table.New(
		table.WithColumns(newColumns),
		table.WithRows(newRows),
		table.WithFocused(true),
		table.WithHeight(m.height),
	)
	setTableStyles(&m.table)
	sumW := 0
	for _, col := range newColumns {
		sumW += col.Width
	}
	m.table.SetWidth(sumW)
}

func (m model) buildRows() []table.Row {
	home, _ := os.UserHomeDir()
	rows := make([]table.Row, len(m.sessions))
	for i, s := range m.sessions {
		createdTime := getCreatedTime(s.Time.Created)
		dir := s.Directory
		if strings.HasPrefix(dir, home) {
			dir = "~" + dir[len(home):]
		}
		row := table.Row{
			sanitize(s.ID),
			sanitize(s.Title),
			sanitize(dir),
			sanitize(createdTime),
		}
		rows[i] = row
	}
	return rows
}

func (m model) createTable(cursor int) table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Title", Width: 0},
		{Title: "Directory", Width: 0},
		{Title: "Created", Width: 0},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(m.rows),
		table.WithFocused(true),
		table.WithHeight(m.height),
	)
	if cursor >= 0 && cursor < len(m.rows) {
		t.SetCursor(cursor)
	}
	setTableStyles(&t)
	t.SetWidth(80)
	return t
}
