package ui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	table    table.Model
	quitting bool
}

func NewModel() Model {
	columns := []table.Column{
		{Title: "Application", Width: 20},
		{Title: "PID", Width: 8},
		{Title: "Port", Width: 8},
		{Title: "Protocol", Width: 10},
	}

	rows := []table.Row{
		{"No data", "-", "-", "-"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(false)
	t.SetStyles(s)

	return Model{
		table: t,
	}
}

var _ tea.Model = Model{}
