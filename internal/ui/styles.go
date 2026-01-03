package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Header and Footer
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	// Status bar
	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	// Loading and empty states
	LoadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	EmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	// Application rows
	AppStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	SelectedAppStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Bold(true)

	// Connection rows
	ConnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

	SelectedConnStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#DDDDDD")).
				Background(lipgloss.Color("#4A3B7C"))

	// Error display
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	// Table view styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	TableSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4A3B7C"))
)
