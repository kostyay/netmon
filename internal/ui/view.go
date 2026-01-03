package ui

import (
	"fmt"
	"strings"

	"github.com/kostyay/netmon/internal/model"
)

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header
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
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Network data
	if m.snapshot == nil {
		b.WriteString(LoadingStyle.Render("Loading..."))
	} else if len(m.snapshot.Applications) == 0 {
		b.WriteString(EmptyStyle.Render("No network connections found"))
	} else {
		b.WriteString(m.renderApplications())
	}

	b.WriteString("\n")

	// Footer with controls
	footer := FooterStyle.Render(
		"↑↓/jk Navigate  ←→/hl Collapse/Expand  +/- Refresh rate  q Quit",
	)
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
