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

// RenderFrameWithTitle renders content in a frame with a centered title on the top border.
// Uses heavy box drawing for modal prominence.
func RenderFrameWithTitle(content string, title string, width, height int) string {
	borderColor := lipgloss.Color(config.CurrentTheme.Styles.Modal.BorderFgColor)
	titleColor := lipgloss.Color(config.CurrentTheme.Styles.Modal.AccentFgColor)
	return renderFrameWithColors(content, title, width, height, borderColor, titleColor)
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

// DangerBorderColor returns the red color for danger modals.
func DangerBorderColor() lipgloss.Color {
	return lipgloss.Color("#FF5555")
}

// RenderDangerFrameWithTitle renders content in a frame with danger/red styling.
// Used for destructive confirmations like kill process.
func RenderDangerFrameWithTitle(content string, title string, width, height int) string {
	dangerColor := DangerBorderColor()
	return renderFrameWithColors(content, title, width, height, dangerColor, dangerColor)
}

// renderFrameWithColors renders a frame with specified border and title colors.
func renderFrameWithColors(content, title string, width, height int, borderColor, titleColor lipgloss.Color) string {
	// Heavy box drawing characters for modal prominence
	topLeft := "┏"
	topRight := "┓"
	bottomLeft := "┗"
	bottomRight := "┛"
	horizontal := "━"
	vertical := "┃"

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	innerWidth := width - 2

	// Build top border with centered title
	titleWithPadding := " " + title + " "
	titleLen := len(titleWithPadding)

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

	bottomBorder := borderStyle.Render(bottomLeft)
	bottomBorder += borderStyle.Render(strings.Repeat(horizontal, innerWidth))
	bottomBorder += borderStyle.Render(bottomRight)

	contentStyle := lipgloss.NewStyle().
		Width(innerWidth).
		Height(height-2).
		Padding(0, 1)

	styledContent := contentStyle.Render(content)

	var result strings.Builder
	result.WriteString(topBorder)
	result.WriteString("\n")

	for _, line := range splitLines(styledContent) {
		result.WriteString(borderStyle.Render(vertical))
		result.WriteString(padRight(line, innerWidth))
		result.WriteString(borderStyle.Render(vertical))
		result.WriteString("\n")
	}

	result.WriteString(bottomBorder)

	return result.String()
}

// LiveIndicatorStyle returns the style for the LIVE indicator (green).
func LiveIndicatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Header.LiveFg)).
		Bold(true)
}

// WarnStyle returns the style for warning/attention text (amber).
func WarnStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Header.WarnFg))
}

// StatsStyle returns the style for muted stats text.
func StatsStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Header.StatsFg))
}

// FooterGroupStyle returns the style for footer group labels (NAV, ACTION).
func FooterGroupStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Footer.GroupFgColor)).
		Bold(true)
}

// BorderStyle returns the style for borders.
func BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.CurrentTheme.Styles.Border.FgColor))
}
