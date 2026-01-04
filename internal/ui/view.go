package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kostyay/netmon/internal/model"
)

// Layout constants for fixed header/footer with scrollable content.
const (
	headerHeight = 2 // title + blank line (crumbs moved to footer)
	footerHeight = 3 // blank line + crumbs + keybindings
)

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Wait for viewport to be initialized
	if !m.ready {
		return LoadingStyle().Render("Initializing...")
	}

	view := m.CurrentView()
	if view == nil {
		return LoadingStyle().Render("Initializing...")
	}

	var b strings.Builder

	// === HEADER (fixed at top) ===
	header := HeaderStyle().Render("netmon - Network Monitor")
	b.WriteString(header)
	b.WriteString("\n")

	// Error display (replaces blank line if error present)
	if m.lastError != nil {
		b.WriteString(ErrorStyle().Render(fmt.Sprintf("Error: %s", m.lastError.Error())))
		b.WriteString("\n")
	}

	// === CONTENT (scrollable via viewport) ===
	var content string
	if m.snapshot == nil {
		content = LoadingStyle().Render("Loading...")
	} else if len(m.snapshot.Applications) == 0 {
		content = EmptyStyle().Render("No network connections found")
	} else {
		switch view.Level {
		case LevelProcessList:
			content = m.renderProcessList()
		case LevelConnections:
			content = m.renderConnectionsList()
		}
	}

	// Set content, sync scroll position, and render viewport
	m.viewport.SetContent(content)
	m.ensureCursorVisible()
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// === FOOTER (fixed at bottom) ===
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderBreadcrumbs renders the navigation breadcrumbs.
func (m Model) renderBreadcrumbs() string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}

	var parts []string
	parts = append(parts, "Processes")

	if view.Level == LevelConnections {
		parts = append(parts, view.ProcessName)
	}

	crumbs := strings.Join(parts, " > ")
	return StatusStyle().Render(fmt.Sprintf("📍 %s  |  Refresh: %.1fs", crumbs, m.refreshInterval.Seconds()))
}

// renderFooter renders the two-row footer with crumbs and keybindings.
func (m Model) renderFooter() string {
	var b strings.Builder

	// Row 1: Breadcrumbs
	b.WriteString(m.renderBreadcrumbs())
	b.WriteString("\n")

	// Row 2: Keybindings
	b.WriteString(m.renderKeybindings())

	return b.String()
}

// renderKeybindings renders the keybindings row.
func (m Model) renderKeybindings() string {
	keyStyle := FooterKeyStyle()
	descStyle := FooterDescStyle()

	view := m.CurrentView()
	var parts []string

	if view != nil && view.Level == LevelProcessList {
		parts = []string{
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("←→") + descStyle.Render(" Column"),
			keyStyle.Render("Enter") + descStyle.Render(" Drill-in"),
			keyStyle.Render("+/-") + descStyle.Render(" Refresh"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	} else {
		parts = []string{
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("←→") + descStyle.Render(" Column"),
			keyStyle.Render("Enter") + descStyle.Render(" Sort"),
			keyStyle.Render("Esc") + descStyle.Render(" Back"),
			keyStyle.Render("+/-") + descStyle.Render(" Refresh"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	}

	footerText := strings.Join(parts, "  ")
	return FooterStyle().Render(footerText)
}

// renderProcessList renders the process list table (Level 0).
func (m Model) renderProcessList() string {
	if m.snapshot == nil {
		return LoadingStyle().Render("Loading...")
	}

	view := m.CurrentView()
	if view == nil {
		return ""
	}

	var b strings.Builder

	// Render summary
	totalConns := m.snapshot.TotalConnections()
	if m.snapshot.SkippedCount > 0 {
		summary := StatusStyle().Render(fmt.Sprintf("Showing %d connections (%d hidden)",
			totalConns, m.snapshot.SkippedCount))
		b.WriteString(summary)
	} else {
		summary := StatusStyle().Render(fmt.Sprintf("Showing %d connections", totalConns))
		b.WriteString(summary)
	}
	b.WriteString("\n\n")

	// Render header
	b.WriteString(m.renderProcessListHeader())
	b.WriteString("\n")

	// Render each process row
	for i, app := range m.snapshot.Applications {
		isSelected := i == view.Cursor

		// Aggregate TX/RX stats for all PIDs of this app
		txStr, rxStr := m.getAggregatedNetIO(app.PIDs)

		// Build row content
		row := fmt.Sprintf("%-20s %5d %5d %5d %8s %8s",
			truncateString(app.Name, 20),
			len(app.Connections),
			app.EstablishedCount,
			app.ListenCount,
			txStr,
			rxStr,
		)

		// Add selection marker
		if isSelected {
			row = "▶ " + row
			b.WriteString(SelectedConnStyle().Render(row))
		} else {
			row = "  " + row
			b.WriteString(ConnStyle().Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderProcessListHeader renders the header for process list table.
func (m Model) renderProcessListHeader() string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}

	var b strings.Builder

	columns := []struct {
		label  string
		column int
		width  int
	}{
		{"Process", 0, 22},
		{"Conns", 1, 6},
		{"ESTAB", 2, 6},
		{"LISTEN", 3, 6},
		{"TX", 4, 9},
		{"RX", 5, 9},
	}

	headerStyle := TableHeaderStyle()
	selectedStyle := TableHeaderSelectedStyle()

	for i, col := range columns {
		if i > 0 {
			b.WriteString(" ")
		}

		isSelected := int(view.SelectedColumn) == col.column

		header := col.label
		padWidth := col.width - len(header)
		if padWidth < 0 {
			padWidth = 0
		}
		paddedHeader := header + strings.Repeat(" ", padWidth)

		if isSelected {
			b.WriteString(selectedStyle.Render(paddedHeader))
		} else {
			b.WriteString(headerStyle.Render(paddedHeader))
		}
	}

	return b.String()
}

// renderConnectionsList renders connections for a specific process (Level 1).
func (m Model) renderConnectionsList() string {
	if m.snapshot == nil {
		return LoadingStyle().Render("Loading...")
	}

	view := m.CurrentView()
	if view == nil {
		return ""
	}

	// Find the selected process
	var selectedApp *model.Application
	for i := range m.snapshot.Applications {
		if m.snapshot.Applications[i].Name == view.ProcessName {
			selectedApp = &m.snapshot.Applications[i]
			break
		}
	}

	if selectedApp == nil {
		return EmptyStyle().Render("Process not found")
	}

	var b strings.Builder

	// === HEADER SECTION ===
	// Process name (bold)
	b.WriteString(HeaderStyle().Render(selectedApp.Name))
	b.WriteString("\n")

	// PIDs and TX/RX stats
	txStr, rxStr := m.getAggregatedNetIO(selectedApp.PIDs)
	statsLine := fmt.Sprintf("PIDs: %s  |  TX: %s  RX: %s  |  %d connections",
		formatPIDList(selectedApp.PIDs),
		txStr, rxStr,
		len(selectedApp.Connections))
	b.WriteString(StatusStyle().Render(statsLine))
	b.WriteString("\n\n")

	// === CONNECTIONS TABLE ===
	// Header
	b.WriteString(m.renderConnectionsHeader())
	b.WriteString("\n")

	// Sort connections
	conns := m.sortConnectionsForView(selectedApp.Connections)

	// Render each connection (no PID column - redundant at this level)
	for i, conn := range conns {
		isSelected := i == view.Cursor

		row := fmt.Sprintf("%-5s %-21s %-21s %-11s",
			conn.Protocol,
			truncateAddr(conn.LocalAddr, 21),
			truncateAddr(conn.RemoteAddr, 21),
			conn.State,
		)

		if isSelected {
			row = "▶ " + row
			b.WriteString(SelectedConnStyle().Render(row))
		} else {
			row = "  " + row
			b.WriteString(ConnStyle().Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatPIDList formats a slice of PIDs for display.
func formatPIDList(pids []int32) string {
	if len(pids) == 0 {
		return "-"
	}
	if len(pids) == 1 {
		return fmt.Sprintf("%d", pids[0])
	}
	if len(pids) <= 3 {
		strs := make([]string, len(pids))
		for i, p := range pids {
			strs[i] = fmt.Sprintf("%d", p)
		}
		return strings.Join(strs, ", ")
	}
	return fmt.Sprintf("%d, %d +%d more", pids[0], pids[1], len(pids)-2)
}

// renderConnectionsHeader renders the header for connections table.
func (m Model) renderConnectionsHeader() string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}

	var b strings.Builder

	columns := []struct {
		label      string
		sortColumn SortColumn
		width      int
	}{
		{"Proto", SortProtocol, 6},
		{"Local", SortLocal, 22},
		{"Remote", SortRemote, 22},
		{"State", SortState, 12},
	}

	headerStyle := TableHeaderStyle()
	selectedStyle := TableHeaderSelectedStyle()
	sortStyle := SortIndicatorStyle()

	for i, col := range columns {
		if i > 0 {
			b.WriteString(" ")
		}

		isSelected := view.SelectedColumn == col.sortColumn
		isSorted := view.SortColumn == col.sortColumn

		header := col.label

		var sortIndicator string
		if isSorted {
			if view.SortAscending {
				sortIndicator = "↑"
			} else {
				sortIndicator = "↓"
			}
		}

		padWidth := col.width - len(header)
		if isSorted {
			padWidth -= 1
		}
		if padWidth < 0 {
			padWidth = 0
		}
		paddedHeader := header + strings.Repeat(" ", padWidth)

		if isSelected {
			b.WriteString(selectedStyle.Render(paddedHeader))
		} else {
			b.WriteString(headerStyle.Render(paddedHeader))
		}

		if isSorted {
			b.WriteString(sortStyle.Render(sortIndicator))
		}
	}

	return b.String()
}

// sortConnectionsForView sorts connections based on current view state.
func (m Model) sortConnectionsForView(conns []model.Connection) []model.Connection {
	view := m.CurrentView()
	if view == nil {
		return conns
	}

	sorted := make([]model.Connection, len(conns))
	copy(sorted, conns)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool
		switch view.SortColumn {
		case SortPID:
			less = sorted[i].PID < sorted[j].PID
		case SortProtocol:
			less = sorted[i].Protocol < sorted[j].Protocol
		case SortLocal:
			less = sorted[i].LocalAddr < sorted[j].LocalAddr
		case SortRemote:
			less = sorted[i].RemoteAddr < sorted[j].RemoteAddr
		case SortState:
			less = sorted[i].State < sorted[j].State
		default:
			less = sorted[i].LocalAddr < sorted[j].LocalAddr
		}

		if view.SortAscending {
			return less
		}
		return !less
	})

	return sorted
}

func formatPIDs(pids []int32) string {
	if len(pids) == 0 {
		return ""
	}

	if len(pids) == 1 {
		return fmt.Sprintf("PID: %d", pids[0])
	}

	// Multiple PIDs: show first few + count
	if len(pids) <= 3 {
		strs := make([]string, len(pids))
		for i, p := range pids {
			strs[i] = fmt.Sprintf("%d", p)
		}
		return fmt.Sprintf("PIDs: %s", strings.Join(strs, ", "))
	}

	return fmt.Sprintf("PIDs: %d, %d +%d more", pids[0], pids[1], len(pids)-2)
}

func truncateAddr(addr string, maxLen int) string {
	if maxLen < 4 {
		if len(addr) <= maxLen {
			return addr
		}
		return addr[:maxLen] // Can't fit ellipsis, just truncate
	}
	if len(addr) <= maxLen {
		return addr
	}
	return addr[:maxLen-3] + "..."
}

// truncateString truncates a string to maxLen with ellipsis if needed.
func truncateString(s string, maxLen int) string {
	if maxLen < 4 {
		if len(s) <= maxLen {
			return s
		}
		return s[:maxLen] // Can't fit ellipsis, just truncate
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatBytes formats bytes into human-readable units.
// Returns '--' for nil stats, otherwise formats as '1.2 MB', '256 KB', '89 B'.
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// formatBytesOrDash formats bytes or returns '--' if nil.
func formatBytesOrDash(stats *model.NetIOStats, isSent bool) string {
	if stats == nil {
		return "--"
	}
	if isSent {
		return formatBytes(stats.BytesSent)
	}
	return formatBytes(stats.BytesRecv)
}

// getAggregatedNetIO returns formatted TX and RX strings aggregated across all PIDs.
// Returns "--" for each if no stats are available.
func (m Model) getAggregatedNetIO(pids []int32) (tx, rx string) {
	var totalSent, totalRecv uint64
	hasStats := false

	for _, pid := range pids {
		if stats, ok := m.netIOCache[pid]; ok {
			totalSent += stats.BytesSent
			totalRecv += stats.BytesRecv
			hasStats = true
		}
	}

	if !hasStats {
		return "--", "--"
	}

	return formatBytes(totalSent), formatBytes(totalRecv)
}

// ensureCursorVisible adjusts the viewport scroll position to keep the cursor visible.
// Must be called after SetContent so the viewport knows its total height.
func (m *Model) ensureCursorVisible() {
	lineNumber := m.cursorLinePosition()

	// Scroll up if cursor is above visible area
	if lineNumber < m.viewport.YOffset {
		m.viewport.SetYOffset(lineNumber)
		return
	}

	// Scroll down if cursor is below visible area
	visibleEnd := m.viewport.YOffset + m.viewport.Height
	if lineNumber >= visibleEnd {
		m.viewport.SetYOffset(lineNumber - m.viewport.Height + 1)
	}
}

// cursorLinePosition calculates which line the currently selected item is on.
// This is used to ensure the cursor stays visible when scrolling.
func (m Model) cursorLinePosition() int {
	view := m.CurrentView()
	if view == nil || m.snapshot == nil {
		return 0
	}

	// Account for summary line, blank line, header row
	const tableHeaderLines = 3
	return view.Cursor + tableHeaderLines
}

