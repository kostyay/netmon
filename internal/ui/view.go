package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kostyay/netmon/internal/model"
)

// Layout constants for fixed header/footer with scrollable content.
const (
	headerHeight = 4 // title + status + potential error + blank line
	footerHeight = 2 // margin + controls line
)

// FlatConnection represents a single connection with its parent process name.
type FlatConnection struct {
	ProcessName string
	Connection  model.Connection
}

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Wait for viewport to be initialized
	if !m.ready {
		return LoadingStyle.Render("Initializing...")
	}

	var b strings.Builder

	// === HEADER (fixed at top) ===
	header := HeaderStyle.Render("netmon - Network Monitor")
	b.WriteString(header)
	b.WriteString("\n")

	// Status bar (refresh rate)
	status := StatusStyle.Render(fmt.Sprintf("Refresh: %.1fs  [+/- to adjust]", m.refreshInterval.Seconds()))
	b.WriteString(status)
	b.WriteString("\n")

	// Error display
	if m.lastError != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %s", m.lastError.Error())))
	}
	b.WriteString("\n")

	// === CONTENT (scrollable via viewport) ===
	var content string
	if m.snapshot == nil {
		content = LoadingStyle.Render("Loading...")
	} else if len(m.snapshot.Applications) == 0 {
		content = EmptyStyle.Render("No network connections found")
	} else {
		if m.viewMode == ViewGrouped {
			content = m.renderApplications()
		} else {
			content = m.renderTable()
		}
	}

	// Set content and render viewport
	m.viewport.SetContent(content)
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// === FOOTER (fixed at bottom) ===
	var footerText string
	if m.viewMode == ViewGrouped {
		footerText = "↑↓/jk Navigate  ←→/hl Collapse/Expand  v: Table  +/- Refresh  q Quit"
	} else {
		footerText = "↑↓/jk Navigate  1-5 Sort  v: Grouped  +/- Refresh  q Quit"
	}
	footer := FooterStyle.Render(footerText)
	b.WriteString(footer)

	return b.String()
}

func (m Model) renderApplications() string {
	var b strings.Builder

	// Show summary with skipped count if any
	totalConns := 0
	for _, app := range m.snapshot.Applications {
		totalConns += len(app.Connections)
	}
	if m.snapshot.SkippedCount > 0 {
		summary := StatusStyle.Render(fmt.Sprintf("Showing %d connections (%d hidden)", totalConns, m.snapshot.SkippedCount))
		b.WriteString(summary)
		b.WriteString("\n\n")
	}

	for i, app := range m.snapshot.Applications {
		isSelected := i == m.cursor

		// Render app header line
		line := m.renderAppHeader(app, isSelected)
		b.WriteString(line)
		b.WriteString("\n")

		// Render connections if expanded
		if m.expandedApps[app.Name] {
			for _, conn := range app.Connections {
				connLine := m.renderConnection(conn, isSelected)
				b.WriteString(connLine)
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

func (m Model) renderAppHeader(app model.Application, isSelected bool) string {
	// Format: ▼ Chrome (3 connections)                PIDs: 1234, 5678
	expandIcon := "▶"
	if m.expandedApps[app.Name] {
		expandIcon = "▼"
	}

	connCount := fmt.Sprintf("(%d connections)", len(app.Connections))
	pidsStr := formatPIDs(app.PIDs)

	// Build the line with spacing
	leftPart := fmt.Sprintf("%s %s %s", expandIcon, app.Name, connCount)

	// Use actual width with fallback
	availableWidth := m.width
	if availableWidth == 0 {
		availableWidth = 80 // fallback for default terminal width
	}

	// Calculate padding to align PIDs on the right
	padding := availableWidth - len(leftPart) - len(pidsStr) - 4 // 4 for margins
	if padding < 2 {
		padding = 2
	}

	line := fmt.Sprintf("%s%s%s", leftPart, strings.Repeat(" ", padding), pidsStr)

	if isSelected {
		return SelectedAppStyle.Render(line)
	}
	return AppStyle.Render(line)
}

func (m Model) renderConnection(conn model.Connection, isSelected bool) string {
	// Format: │ TCP  127.0.0.1:52341    →  142.250.80.46:443   ESTABLISHED
	line := fmt.Sprintf("  │ %-4s %-21s → %-21s %s",
		conn.Protocol,
		truncateAddr(conn.LocalAddr, 21),
		truncateAddr(conn.RemoteAddr, 21),
		conn.State,
	)

	if isSelected {
		return SelectedConnStyle.Render(line)
	}
	return ConnStyle.Render(line)
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
	if len(addr) <= maxLen {
		return addr
	}
	return addr[:maxLen-3] + "..."
}

// cursorLinePosition calculates which line the currently selected item is on.
// This is used to ensure the cursor stays visible when scrolling.
func (m Model) cursorLinePosition() int {
	if m.snapshot == nil {
		return 0
	}

	// In table view, tableCursor directly corresponds to line position
	if m.viewMode == ViewTable {
		return m.tableCursor
	}

	// In grouped view, calculate position based on expanded apps
	lineNumber := 0
	for i := 0; i < m.cursor && i < len(m.snapshot.Applications); i++ {
		app := m.snapshot.Applications[i]
		lineNumber++ // app header line
		if m.expandedApps[app.Name] {
			lineNumber += len(app.Connections) // connection lines
		}
	}
	return lineNumber
}

// flattenConnections creates a flat slice of all connections with their process names.
func (m Model) flattenConnections() []FlatConnection {
	if m.snapshot == nil {
		return nil
	}

	var flat []FlatConnection
	for _, app := range m.snapshot.Applications {
		for _, conn := range app.Connections {
			flat = append(flat, FlatConnection{
				ProcessName: app.Name,
				Connection:  conn,
			})
		}
	}
	return flat
}

// sortConnections sorts a slice of flat connections based on current sort settings.
func (m Model) sortConnections(conns []FlatConnection) []FlatConnection {
	// Create a copy to avoid mutating the input
	sorted := make([]FlatConnection, len(conns))
	copy(sorted, conns)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool
		switch m.sortColumn {
		case SortProcess:
			less = sorted[i].ProcessName < sorted[j].ProcessName
		case SortProtocol:
			less = sorted[i].Connection.Protocol < sorted[j].Connection.Protocol
		case SortLocal:
			less = sorted[i].Connection.LocalAddr < sorted[j].Connection.LocalAddr
		case SortRemote:
			less = sorted[i].Connection.RemoteAddr < sorted[j].Connection.RemoteAddr
		case SortState:
			less = sorted[i].Connection.State < sorted[j].Connection.State
		default:
			less = sorted[i].ProcessName < sorted[j].ProcessName
		}

		if m.sortAscending {
			return less
		}
		return !less
	})

	return sorted
}

// renderTable renders the full table view with headers and sorted connections.
func (m Model) renderTable() string {
	var b strings.Builder

	// Get and sort connections
	conns := m.flattenConnections()
	conns = m.sortConnections(conns)

	// Render summary with sort indicator
	totalConns := len(conns)
	sortColumnName := m.sortColumnName()
	sortIndicator := "▲"
	if !m.sortAscending {
		sortIndicator = "▼"
	}
	if m.snapshot.SkippedCount > 0 {
		summary := StatusStyle.Render(fmt.Sprintf("Showing %d connections (%d hidden) | Sorted by %s %s",
			totalConns, m.snapshot.SkippedCount, sortColumnName, sortIndicator))
		b.WriteString(summary)
	} else {
		summary := StatusStyle.Render(fmt.Sprintf("Showing %d connections | Sorted by %s %s",
			totalConns, sortColumnName, sortIndicator))
		b.WriteString(summary)
	}
	b.WriteString("\n\n")

	// Render header
	b.WriteString(m.renderTableHeader())
	b.WriteString("\n")

	// Render each row
	for i, fc := range conns {
		isSelected := i == m.tableCursor

		// Build row content
		row := fmt.Sprintf("%-14s %-5s %-21s %-21s %-11s",
			truncateString(fc.ProcessName, 14),
			fc.Connection.Protocol,
			truncateAddr(fc.Connection.LocalAddr, 21),
			truncateAddr(fc.Connection.RemoteAddr, 21),
			fc.Connection.State,
		)

		// Add selection marker
		if isSelected {
			row = "▶ " + row
			b.WriteString(SelectedConnStyle.Render(row))
		} else {
			row = "  " + row
			b.WriteString(ConnStyle.Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// truncateString truncates a string to maxLen with ellipsis if needed.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// sortColumnName returns the display name for the current sort column.
func (m Model) sortColumnName() string {
	switch m.sortColumn {
	case SortProcess:
		return "Process"
	case SortProtocol:
		return "Protocol"
	case SortLocal:
		return "Local"
	case SortRemote:
		return "Remote"
	case SortState:
		return "State"
	default:
		return "Process"
	}
}

// renderTableHeader renders the column headers for table view with sort indicators.
func (m Model) renderTableHeader() string {
	var b strings.Builder

	// Column headers with keyboard shortcuts
	columns := []struct {
		key    string
		label  string
		column SortColumn
	}{
		{"1", "Process", SortProcess},
		{"2", "Proto", SortProtocol},
		{"3", "Local", SortLocal},
		{"4", "Remote", SortRemote},
		{"5", "State", SortState},
	}

	for i, col := range columns {
		if i > 0 {
			b.WriteString("  ")
		}
		header := fmt.Sprintf("[%s]%s", col.key, col.label)

		// Add sort indicator if this is the active sort column
		if m.sortColumn == col.column {
			if m.sortAscending {
				header += "▲"
			} else {
				header += "▼"
			}
		}

		// Apply width padding
		switch col.column {
		case SortProcess:
			header = fmt.Sprintf("%-15s", header)
		case SortProtocol:
			header = fmt.Sprintf("%-8s", header)
		case SortLocal, SortRemote:
			header = fmt.Sprintf("%-21s", header)
		case SortState:
			header = fmt.Sprintf("%-12s", header)
		}

		b.WriteString(header)
	}

	headerLine := TableHeaderStyle.Render(b.String())

	// Add separator line
	availableWidth := m.width
	if availableWidth == 0 {
		availableWidth = 80
	}
	separator := TableSeparatorStyle.Render(strings.Repeat("─", availableWidth))

	return headerLine + "\n" + separator
}
