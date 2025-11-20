package main

import (
	"github.com/charmbracelet/bubbles/table"
)

type model struct {
	table           table.Model
	sessions        Sessions
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

func newModel(sessions Sessions, cursor int) model {
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

	m := model{
		sessions:      sessions,
		width:         width,
		height:        height,
		selectedIndex: -1,
		shouldRefresh: false,
		showHelp:      false,
	}
	m.rows = m.buildRows()
	m.columns = columns
	m.table = m.createTable(cursor)
	m.baseWidths = baseWidths
	m.visibleColumns = visibleColumns
	m.rebuildTable()
	return m
}
