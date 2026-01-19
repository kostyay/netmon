package ui

import (
	"strings"

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

// AddedConnStyle returns the style for newly added connections.
func AddedConnStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.AddedFgColor))
}

// RemovedConnStyle returns the style for removed connections.
func RemovedConnStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Table.RemovedFgColor))
}

// FrameStyle returns the style for the main content frame with border.
func FrameStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.CurrentTheme.Styles.Table.HeaderFgColor)).
		Width(width-2).   // Account for border
		Height(height-2). // Account for border
		Padding(0, 1)
}

// RenderFrameWithTitle renders content in a frame with a centered title on the top border.
func RenderFrameWithTitle(content string, title string, width, height int) string {
	borderColor := lipgloss.Color(config.CurrentTheme.Styles.Table.HeaderFgColor)
	titleColor := lipgloss.Color(config.CurrentTheme.Styles.Header.TitleFg)

	// Border characters for rounded border
	topLeft := "╭"
	topRight := "╮"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"
	vertical := "│"

	// Style for border characters
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	// Calculate inner width (content area without borders)
	innerWidth := width - 2

	// Build top border with centered title
	titleWithPadding := " " + title + " "
	titleLen := len(titleWithPadding)

	// Calculate padding on each side of the title
	remainingWidth := innerWidth - titleLen
	if remainingWidth < 0 {
		remainingWidth = 0
		titleWithPadding = titleWithPadding[:innerWidth]
	}
	leftPad := remainingWidth / 2
	rightPad := remainingWidth - leftPad

	topBorder := borderStyle.Render(topLeft)
	topBorder += borderStyle.Render(strings.Repeat(horizontal, leftPad))
	topBorder += titleStyle.Render(titleWithPadding)
	topBorder += borderStyle.Render(strings.Repeat(horizontal, rightPad))
	topBorder += borderStyle.Render(topRight)

	// Build bottom border
	bottomBorder := borderStyle.Render(bottomLeft)
	bottomBorder += borderStyle.Render(strings.Repeat(horizontal, innerWidth))
	bottomBorder += borderStyle.Render(bottomRight)

	// Style for content area with padding
	contentStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(height-2).
		Padding(0, 1)

	styledContent := contentStyle.Render(content)

	// Build complete frame
	var result string
	result = topBorder + "\n"

	// Add content lines with vertical borders
	lines := splitLines(styledContent)
	for _, line := range lines {
		// Ensure line is padded to inner width
		result += borderStyle.Render(vertical) + padRight(line, innerWidth) + borderStyle.Render(vertical) + "\n"
	}

	result += bottomBorder

	return result
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "\n")
}

// padRight pads a string to the specified width.
func padRight(s string, width int) string {
	// Use lipgloss to measure visible width (handles ANSI escape codes)
	visibleWidth := lipgloss.Width(s)
	if visibleWidth >= width {
		return s
	}
	padding := width - visibleWidth
	return s + strings.Repeat(" ", padding)
}

// DimmedStyle returns a style for dimmed background content when modal is visible.
func DimmedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Modal.DimmedFgColor)).
		Faint(true)
}
