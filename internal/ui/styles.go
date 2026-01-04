package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kostyay/netmon/internal/config"
)

// Theme-aware style getters

// HeaderStyle returns the style for the main header title.
func HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Header.TitleFg))
}

// FooterStyle returns the style for footer text.
func FooterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Footer.FgColor))
}

// FooterKeyStyle returns the style for keyboard shortcut keys in footer.
func FooterKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Footer.KeyFgColor))
}

// FooterDescStyle returns the style for key descriptions in footer.
func FooterDescStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Footer.DescFgColor))
}

// StatusStyle returns the style for status bar text.
func StatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Status.FgColor))
}

// LoadingStyle returns the style for loading indicators.
func LoadingStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Status.FgColor)).
		Italic(true)
}

// EmptyStyle returns the style for empty state messages.
func EmptyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Status.FgColor)).
		Italic(true)
}

// AppStyle returns the style for application rows in grouped view.
func AppStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.FgColor))
}

// SelectedAppStyle returns the style for selected application in grouped view.
func SelectedAppStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.CursorFgColor)).
		Background(lipgloss.Color(config.CurrentTheme.Styles.Table.CursorBgColor)).
		Bold(true)
}

// ConnStyle returns the style for connection rows.
func ConnStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.FgColor))
}

// SelectedConnStyle returns the style for the selected row in table view.
func SelectedConnStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.CursorFgColor)).
		Background(lipgloss.Color(config.CurrentTheme.Styles.Table.CursorBgColor))
}

// ErrorStyle returns the style for error messages.
func ErrorStyle() lipgloss.Style {
	// Keep error as red for visibility
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555")).
		Bold(true)
}

// TableHeaderStyle returns the style for table column headers.
func TableHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.HeaderFgColor)).
		Bold(true)
}

// TableHeaderSelectedStyle returns the style for the selected column header.
func TableHeaderSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.SelectedColumn)).
		Bold(true)
}

// SortIndicatorStyle returns the style for sort direction indicators.
func SortIndicatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.SortIndicator))
}

// FrameStyle returns the style for the main content frame with border.
func FrameStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.CurrentTheme.Styles.Table.HeaderFgColor)).
		Width(width - 2).   // Account for border
		Height(height - 2). // Account for border
		Padding(0, 1)
}
