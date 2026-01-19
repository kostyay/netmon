package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kostyay/netmon/internal/config"
	"github.com/kostyay/netmon/internal/model"
)

// Layout constants for fixed header/footer with scrollable content.
const (
	headerHeight = 1 // title row
	footerHeight = 2 // crumbs + keybindings
	frameHeight  = 2 // top and bottom border
)

// frozenHeaderHeight returns the number of lines for the frozen table header.
// This varies by view level.
func (m Model) frozenHeaderHeight() int {
	view := m.CurrentView()
	if view == nil {
		return 0
	}
	switch view.Level {
	case LevelConnections:
		// Process name (1) + [exe (1)] + stats line (1) + blank line (1) + table header (1)
		lines := 4
		if m.snapshot != nil {
			for _, app := range m.snapshot.Applications {
				if app.Name == view.ProcessName && app.Exe != "" {
					lines = 5
					break
				}
			}
		}
		return lines
	default:
		// ProcessList and AllConnections: just 1 table header line
		return 1
	}
}

// contentWidth returns the available width for table content.
// Accounts for frame border and padding.
func (m Model) contentWidth() int {
	// Frame has 2 chars border + 2 chars padding = 4 total
	return m.width - 4
}

// renderFrozenHeader returns the frozen header content for the current view.
// This is rendered outside the viewport so it stays visible when scrolling.
func (m Model) renderFrozenHeader() string {
	if m.snapshot == nil {
		return ""
	}

	view := m.CurrentView()
	if view == nil {
		return ""
	}

	var b strings.Builder

	switch view.Level {
	case LevelProcessList:
		columns := processListColumns()
		widths := calculateColumnWidths(columns, m.contentWidth())
		b.WriteString(m.renderProcessListHeader(widths))

	case LevelConnections:
		// Find the selected process
		var selectedApp *model.Application
		for i := range m.snapshot.Applications {
			if m.snapshot.Applications[i].Name == view.ProcessName {
				selectedApp = &m.snapshot.Applications[i]
				break
			}
		}
		if selectedApp == nil {
			return ""
		}

		// Process name (bold)
		b.WriteString(HeaderStyle().Render(selectedApp.Name))
		b.WriteString("\n")

		// Executable path (if available)
		if selectedApp.Exe != "" {
			b.WriteString(StatusStyle().Render(selectedApp.Exe))
			b.WriteString("\n")
		}

		// PIDs and TX/RX stats
		conns := m.filteredConnections(selectedApp.Connections)
		txStr, rxStr := m.getAggregatedNetIO(selectedApp.PIDs)
		statsLine := fmt.Sprintf("PIDs: %s  |  TX: %s  RX: %s  |  %d connections",
			formatPIDList(selectedApp.PIDs),
			txStr, rxStr,
			len(conns))
		b.WriteString(StatusStyle().Render(statsLine))
		b.WriteString("\n")

		// Table header
		columns := connectionsColumns()
		widths := calculateColumnWidths(columns, m.contentWidth())
		b.WriteString(m.renderConnectionsHeader(widths))

	case LevelAllConnections:
		columns := allConnectionsColumns()
		widths := calculateColumnWidths(columns, m.contentWidth())
		b.WriteString(m.renderAllConnectionsHeader(widths))
	}

	return b.String()
}

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

	// Help modal overlay
	if m.helpMode {
		return m.renderHelpModal()
	}

	// Settings modal overlay
	if m.settingsMode {
		return m.renderSettingsModal()
	}

	var b strings.Builder

	// === HEADER (fixed at top, full width) ===
	headerText := "netmon - Network Monitor"
	if m.lastError != nil {
		headerText = fmt.Sprintf("netmon - Network Monitor  │  Error: %s", m.lastError.Error())
	}
	header := HeaderStyle().Width(m.width).Render(headerText)
	b.WriteString(header)
	b.WriteString("\n")

	// === CONTENT (wrapped in frame with frozen header + scrollable viewport) ===
	// Calculate connection count for title
	connCount := 0
	if m.snapshot != nil {
		connCount = m.snapshot.TotalConnections()
	}
	frameTitle := fmt.Sprintf("connections: %d", connCount)

	// Render frame with frozen header outside viewport
	framedContent := m.renderFrameWithFrozenHeader(frameTitle)
	b.WriteString(framedContent)
	b.WriteString("\n")

	// === FOOTER (fixed at bottom, full width) ===
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderFrameWithFrozenHeader renders the frame with frozen header above scrollable viewport.
func (m Model) renderFrameWithFrozenHeader(title string) string {
	borderColor := lipgloss.Color(config.CurrentTheme.Styles.Table.HeaderFgColor)
	titleColor := lipgloss.Color(config.CurrentTheme.Styles.Header.TitleFg)

	// Border characters
	topLeft := "╭"
	topRight := "╮"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"
	vertical := "│"

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	// Inner width (content area without borders)
	innerWidth := m.width - 2

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

	// Build bottom border
	bottomBorder := borderStyle.Render(bottomLeft)
	bottomBorder += borderStyle.Render(strings.Repeat(horizontal, innerWidth))
	bottomBorder += borderStyle.Render(bottomRight)

	var result strings.Builder
	result.WriteString(topBorder)
	result.WriteString("\n")

	// Render frozen header lines (outside viewport, won't scroll)
	frozenHeader := m.renderFrozenHeader()
	if frozenHeader != "" {
		for _, line := range strings.Split(frozenHeader, "\n") {
			result.WriteString(borderStyle.Render(vertical))
			result.WriteString(" ") // left padding
			result.WriteString(padRight(line, innerWidth-2))
			result.WriteString(" ") // right padding
			result.WriteString(borderStyle.Render(vertical))
			result.WriteString("\n")
		}
	}

	// Render viewport content (scrollable data rows)
	viewportContent := m.viewport.View()
	for _, line := range strings.Split(viewportContent, "\n") {
		result.WriteString(borderStyle.Render(vertical))
		result.WriteString(" ") // left padding
		result.WriteString(padRight(line, innerWidth-2))
		result.WriteString(" ") // right padding
		result.WriteString(borderStyle.Render(vertical))
		result.WriteString("\n")
	}

	result.WriteString(bottomBorder)

	return result.String()
}

// renderBreadcrumbsText returns the breadcrumbs text without styling.
func (m Model) renderBreadcrumbsText() string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}

	var parts []string
	switch view.Level {
	case LevelProcessList:
		parts = append(parts, "Processes")
	case LevelConnections:
		parts = append(parts, "Processes", view.ProcessName)
	case LevelAllConnections:
		parts = append(parts, "All Connections")
	}

	crumbs := strings.Join(parts, " > ")
	return fmt.Sprintf("📍 %s  |  Refresh: %.1fs", crumbs, m.refreshInterval.Seconds())
}

// renderFooter renders the two-row footer with crumbs and keybindings.
func (m Model) renderFooter() string {
	var b strings.Builder

	// Row 1: Kill mode, result message, search input, or breadcrumbs
	if m.killMode && m.killTarget != nil {
		// Show kill confirmation prompt
		var prompt string
		if len(m.killTarget.PIDs) > 1 {
			prompt = fmt.Sprintf("Kill %d PIDs of %s with %s? [y/n]",
				len(m.killTarget.PIDs), m.killTarget.ProcessName, m.killTarget.Signal)
		} else {
			prompt = fmt.Sprintf("Kill PID %d (%s) with %s? [y/n]",
				m.killTarget.PID, m.killTarget.ProcessName, m.killTarget.Signal)
		}
		b.WriteString(ErrorStyle().Width(m.width).Render(prompt))
	} else if m.killResult != "" && time.Since(m.killResultAt) < 2*time.Second {
		// Show kill result message (auto-dismiss after 2s)
		b.WriteString(StatusStyle().Width(m.width).Render(m.killResult))
	} else if m.searchMode {
		// Show search input with cursor
		b.WriteString(StatusStyle().Width(m.width).Render(fmt.Sprintf("/%s█", m.searchQuery)))
	} else if m.activeFilter != "" {
		// Show filter indicator + breadcrumbs
		filterText := fmt.Sprintf("[filter: %s]  %s", m.activeFilter, m.renderBreadcrumbsText())
		b.WriteString(StatusStyle().Width(m.width).Render(filterText))
	} else {
		b.WriteString(StatusStyle().Width(m.width).Render(m.renderBreadcrumbsText()))
	}
	b.WriteString("\n")

	// Row 2: Keybindings (full width)
	b.WriteString(FooterStyle().Width(m.width).Render(m.renderKeybindingsText()))

	return b.String()
}

// renderKeybindingsText returns the keybindings text with inline styling.
func (m Model) renderKeybindingsText() string {
	keyStyle := FooterKeyStyle()
	descStyle := FooterDescStyle()

	view := m.CurrentView()
	var parts []string

	if view == nil {
		return ""
	}

	// Kill mode keybindings
	if m.killMode {
		parts = []string{
			descStyle.Render("[KILL]"),
			keyStyle.Render("y") + descStyle.Render(" Confirm"),
			keyStyle.Render("n/Esc") + descStyle.Render(" Cancel"),
		}
		return strings.Join(parts, "  ")
	}

	// Sort mode has its own keybindings
	if view.SortMode {
		parts = []string{
			descStyle.Render("[SORT MODE]"),
			keyStyle.Render("←→") + descStyle.Render(" Column"),
			keyStyle.Render("Enter") + descStyle.Render(" Sort"),
			keyStyle.Render("Esc") + descStyle.Render(" Cancel"),
		}
		return strings.Join(parts, "  ")
	}

	// Search mode keybindings
	if m.searchMode {
		parts = []string{
			descStyle.Render("[SEARCH]"),
			keyStyle.Render("Enter") + descStyle.Render(" Apply"),
			keyStyle.Render("Esc") + descStyle.Render(" Cancel"),
		}
		return strings.Join(parts, "  ")
	}

	switch view.Level {
	case LevelProcessList:
		parts = []string{
			keyStyle.Render("Enter") + descStyle.Render(" Drill-in"),
			keyStyle.Render("/") + descStyle.Render(" Search"),
			keyStyle.Render("x/X") + descStyle.Render(" Kill"),
			keyStyle.Render("v") + descStyle.Render(" All"),
			keyStyle.Render("?") + descStyle.Render(" Help"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	case LevelConnections:
		parts = []string{
			keyStyle.Render("/") + descStyle.Render(" Search"),
			keyStyle.Render("x/X") + descStyle.Render(" Kill"),
			keyStyle.Render("Esc") + descStyle.Render(" Back"),
			keyStyle.Render("v") + descStyle.Render(" All"),
			keyStyle.Render("?") + descStyle.Render(" Help"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	case LevelAllConnections:
		parts = []string{
			keyStyle.Render("/") + descStyle.Render(" Search"),
			keyStyle.Render("x/X") + descStyle.Render(" Kill"),
			keyStyle.Render("v") + descStyle.Render(" Grouped"),
			keyStyle.Render("?") + descStyle.Render(" Help"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	}

	return strings.Join(parts, "  ")
}

// currentFilter returns the active filter string, preferring searchQuery when in search mode.
func (m Model) currentFilter() string {
	if m.searchMode {
		return m.searchQuery
	}
	return m.activeFilter
}

// filteredApps returns applications matching the current filter.
// Matches if ANY connection in the app matches the filter.
func (m Model) filteredApps() []model.Application {
	if m.snapshot == nil {
		return nil
	}
	filter := m.currentFilter()
	if filter == "" {
		return m.snapshot.Applications
	}

	var result []model.Application
	exactMatch := m.useExactPortMatch()
	for _, app := range m.snapshot.Applications {
		// Check if process-level fields match
		if matchesFilter(filter, filterFields{ProcessName: app.Name, PIDs: app.PIDs}, exactMatch) {
			result = append(result, app)
			continue
		}
		// Check if any connection matches
		for _, conn := range app.Connections {
			if matchesFilter(filter, filterFields{
				LocalAddr:  conn.LocalAddr,
				RemoteAddr: conn.RemoteAddr,
				Protocol:   string(conn.Protocol),
				State:      string(conn.State),
			}, exactMatch) {
				result = append(result, app)
				break
			}
		}
	}
	return result
}

// filteredConnections returns connections matching the current filter for a specific process.
func (m Model) filteredConnections(conns []model.Connection) []model.Connection {
	filter := m.currentFilter()
	if filter == "" {
		return conns
	}

	var result []model.Connection
	exactMatch := m.useExactPortMatch()
	for _, conn := range conns {
		if matchesFilter(filter, filterFields{
			PIDs:       []int32{conn.PID},
			LocalAddr:  conn.LocalAddr,
			RemoteAddr: conn.RemoteAddr,
			Protocol:   string(conn.Protocol),
			State:      string(conn.State),
		}, exactMatch) {
			result = append(result, conn)
		}
	}
	return result
}

// filteredAllConnections returns connections matching the current filter.
func (m Model) filteredAllConnections() []connectionWithProcess {
	if m.snapshot == nil {
		return nil
	}
	filter := m.currentFilter()

	var result []connectionWithProcess
	exactMatch := m.useExactPortMatch()
	for _, app := range m.snapshot.Applications {
		for _, conn := range app.Connections {
			// No filter or matches filter - include connection
			if filter == "" || matchesFilter(filter, filterFields{
				ProcessName: app.Name,
				PIDs:        []int32{conn.PID},
				LocalAddr:   conn.LocalAddr,
				RemoteAddr:  conn.RemoteAddr,
				Protocol:    string(conn.Protocol),
				State:       string(conn.State),
			}, exactMatch) {
				result = append(result, connectionWithProcess{
					Connection:  conn,
					ProcessName: app.Name,
				})
			}
		}
	}
	return result
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

	// Get filtered applications
	apps := m.filteredApps()

	// Handle empty results
	if len(apps) == 0 {
		filter := m.currentFilter()
		if filter != "" {
			return EmptyStyle().Render(fmt.Sprintf("No matches for '%s'", filter))
		}
		return EmptyStyle().Render("No processes found")
	}

	var b strings.Builder

	// Calculate column widths
	columns := processListColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())

	// Render header
	b.WriteString(m.renderProcessListHeader(widths))
	b.WriteString("\n")

	// Sort applications
	apps = m.sortProcessList(apps)

	// Use view.Cursor directly for selection (view already defined above)
	cursorIdx := view.Cursor

	// Render each process row
	for i, app := range apps {
		isSelected := i == cursorIdx

		// Aggregate TX/RX stats for all PIDs of this app
		txStr, rxStr := m.getAggregatedNetIO(app.PIDs)

		// Get primary PID (first in list)
		var primaryPID int32
		if len(app.PIDs) > 0 {
			primaryPID = app.PIDs[0]
		}

		// Build row content with dynamic widths
		row := fmt.Sprintf("%*d %-*s %*d %*d %*d %*s %*s",
			widths[0], primaryPID,
			widths[1], truncateString(app.Name, widths[1]),
			widths[2], len(app.Connections),
			widths[3], app.EstablishedCount,
			widths[4], app.ListenCount,
			widths[5], txStr,
			widths[6], rxStr,
		)

		b.WriteString(renderRow(row, isSelected))
	}

	return b.String()
}

// renderProcessListHeader renders the header for process list table.
func (m Model) renderProcessListHeader(widths []int) string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}
	columns := processListColumns()
	return renderTableHeader(columns, widths, view.SelectedColumn, view.SortColumn, view.SortAscending, true)
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

	// Get filtered connections
	conns := m.filteredConnections(selectedApp.Connections)

	// Handle empty results
	if len(conns) == 0 {
		filter := m.currentFilter()
		if filter != "" {
			return EmptyStyle().Render(fmt.Sprintf("No matches for '%s'", filter))
		}
		return EmptyStyle().Render("No connections found")
	}

	var b strings.Builder

	// === HEADER SECTION ===
	// Process name (bold)
	b.WriteString(HeaderStyle().Render(selectedApp.Name))
	b.WriteString("\n")

	// Executable path (if available)
	if selectedApp.Exe != "" {
		b.WriteString(StatusStyle().Render(selectedApp.Exe))
		b.WriteString("\n")
	}

	// PIDs and TX/RX stats
	txStr, rxStr := m.getAggregatedNetIO(selectedApp.PIDs)
	statsLine := fmt.Sprintf("PIDs: %s  |  TX: %s  RX: %s  |  %d connections",
		formatPIDList(selectedApp.PIDs),
		txStr, rxStr,
		len(conns))
	b.WriteString(StatusStyle().Render(statsLine))
	b.WriteString("\n\n")

	// === CONNECTIONS TABLE ===
	// Calculate column widths
	columns := connectionsColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())

	// Header
	b.WriteString(m.renderConnectionsHeader(widths))
	b.WriteString("\n")

	// Sort connections
	conns = m.sortConnectionsForView(conns)

	// Use view.Cursor directly for selection (view already defined above)
	cursorIdx := view.Cursor

	// Render each connection (no PID column - redundant at this level)
	for i, conn := range conns {
		isSelected := i == cursorIdx

		proto := string(conn.Protocol)
		remoteAddr := formatRemoteAddr(conn.RemoteAddr, proto, m.dnsCache, m.serviceNames)
		localAddr := formatAddr(conn.LocalAddr, proto, m.serviceNames)
		row := fmt.Sprintf("%-*s %-*s %-*s %-*s",
			widths[0], conn.Protocol,
			widths[1], truncateAddr(localAddr, widths[1]),
			widths[2], truncateAddr(remoteAddr, widths[2]),
			widths[3], conn.State,
		)

		change := m.GetChange(conn)
		b.WriteString(renderRowWithHighlight(row, isSelected, change))
	}

	return b.String()
}

// renderConnectionsHeader renders the header for connections table.
func (m Model) renderConnectionsHeader(widths []int) string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}
	columns := connectionsColumns()
	return renderTableHeader(columns, widths, view.SelectedColumn, view.SortColumn, view.SortAscending, true)
}

// connectionWithProcess holds a connection along with its process name for the all-connections view.
type connectionWithProcess struct {
	model.Connection
	ProcessName string
}

// renderAllConnections renders a flat list of all connections from all processes.
func (m Model) renderAllConnections() string {
	if m.snapshot == nil {
		return LoadingStyle().Render("Loading...")
	}

	view := m.CurrentView()
	if view == nil {
		return ""
	}

	// Get filtered connections
	allConns := m.filteredAllConnections()

	// Handle empty results
	if len(allConns) == 0 {
		filter := m.currentFilter()
		if filter != "" {
			return EmptyStyle().Render(fmt.Sprintf("No matches for '%s'", filter))
		}
		return EmptyStyle().Render("No connections found")
	}

	var b strings.Builder

	// === CONNECTIONS TABLE ===
	// Calculate column widths
	columns := allConnectionsColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())

	// Header
	b.WriteString(m.renderAllConnectionsHeader(widths))
	b.WriteString("\n")

	// Sort connections
	allConns = m.sortAllConnections(allConns)

	// Use view.Cursor directly for selection (view already defined above)
	cursorIdx := view.Cursor

	// Render each connection
	for i, conn := range allConns {
		isSelected := i == cursorIdx

		proto := string(conn.Protocol)
		remoteAddr := formatRemoteAddr(conn.RemoteAddr, proto, m.dnsCache, m.serviceNames)
		localAddr := formatAddr(conn.LocalAddr, proto, m.serviceNames)
		row := fmt.Sprintf("%*d %-*s %-*s %-*s %-*s %-*s",
			widths[0], conn.PID,
			widths[1], truncateString(conn.ProcessName, widths[1]),
			widths[2], conn.Protocol,
			widths[3], truncateAddr(localAddr, widths[3]),
			widths[4], truncateAddr(remoteAddr, widths[4]),
			widths[5], conn.State,
		)

		change := m.GetChange(conn.Connection)
		b.WriteString(renderRowWithHighlight(row, isSelected, change))
	}

	return b.String()
}

// renderAllConnectionsHeader renders the header for the all-connections table.
func (m Model) renderAllConnectionsHeader(widths []int) string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}
	columns := allConnectionsColumns()
	return renderTableHeader(columns, widths, view.SelectedColumn, view.SortColumn, view.SortAscending, true)
}

// renderProcessListData renders only the data rows for process list (no header).
func (m Model) renderProcessListData() string {
	if m.snapshot == nil {
		return LoadingStyle().Render("Loading...")
	}

	view := m.CurrentView()
	if view == nil {
		return ""
	}

	apps := m.filteredApps()
	if len(apps) == 0 {
		filter := m.currentFilter()
		if filter != "" {
			return EmptyStyle().Render(fmt.Sprintf("No matches for '%s'", filter))
		}
		return EmptyStyle().Render("No processes found")
	}

	var b strings.Builder
	columns := processListColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())
	apps = m.sortProcessList(apps)
	cursorIdx := view.Cursor

	for i, app := range apps {
		isSelected := i == cursorIdx
		txStr, rxStr := m.getAggregatedNetIO(app.PIDs)
		var primaryPID int32
		if len(app.PIDs) > 0 {
			primaryPID = app.PIDs[0]
		}

		row := fmt.Sprintf("%*d %-*s %*d %*d %*d %*s %*s",
			widths[0], primaryPID,
			widths[1], truncateString(app.Name, widths[1]),
			widths[2], len(app.Connections),
			widths[3], app.EstablishedCount,
			widths[4], app.ListenCount,
			widths[5], txStr,
			widths[6], rxStr,
		)
		b.WriteString(renderRow(row, isSelected))
	}

	return b.String()
}

// renderConnectionsListData renders only the data rows for connections list (no header).
func (m Model) renderConnectionsListData() string {
	if m.snapshot == nil {
		return LoadingStyle().Render("Loading...")
	}

	view := m.CurrentView()
	if view == nil {
		return ""
	}

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

	conns := m.filteredConnections(selectedApp.Connections)
	if len(conns) == 0 {
		filter := m.currentFilter()
		if filter != "" {
			return EmptyStyle().Render(fmt.Sprintf("No matches for '%s'", filter))
		}
		return EmptyStyle().Render("No connections found")
	}

	var b strings.Builder
	columns := connectionsColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())
	conns = m.sortConnectionsForView(conns)
	cursorIdx := view.Cursor

	for i, conn := range conns {
		isSelected := i == cursorIdx
		proto := string(conn.Protocol)
		remoteAddr := formatRemoteAddr(conn.RemoteAddr, proto, m.dnsCache, m.serviceNames)
		localAddr := formatAddr(conn.LocalAddr, proto, m.serviceNames)
		row := fmt.Sprintf("%-*s %-*s %-*s %-*s",
			widths[0], conn.Protocol,
			widths[1], truncateAddr(localAddr, widths[1]),
			widths[2], truncateAddr(remoteAddr, widths[2]),
			widths[3], conn.State,
		)
		change := m.GetChange(conn)
		b.WriteString(renderRowWithHighlight(row, isSelected, change))
	}

	return b.String()
}

// renderAllConnectionsData renders only the data rows for all connections (no header).
func (m Model) renderAllConnectionsData() string {
	if m.snapshot == nil {
		return LoadingStyle().Render("Loading...")
	}

	view := m.CurrentView()
	if view == nil {
		return ""
	}

	allConns := m.filteredAllConnections()
	if len(allConns) == 0 {
		filter := m.currentFilter()
		if filter != "" {
			return EmptyStyle().Render(fmt.Sprintf("No matches for '%s'", filter))
		}
		return EmptyStyle().Render("No connections found")
	}

	var b strings.Builder
	columns := allConnectionsColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())
	allConns = m.sortAllConnections(allConns)
	cursorIdx := view.Cursor

	for i, conn := range allConns {
		isSelected := i == cursorIdx
		proto := string(conn.Protocol)
		remoteAddr := formatRemoteAddr(conn.RemoteAddr, proto, m.dnsCache, m.serviceNames)
		localAddr := formatAddr(conn.LocalAddr, proto, m.serviceNames)
		row := fmt.Sprintf("%*d %-*s %-*s %-*s %-*s %-*s",
			widths[0], conn.PID,
			widths[1], truncateString(conn.ProcessName, widths[1]),
			widths[2], conn.Protocol,
			widths[3], truncateAddr(localAddr, widths[3]),
			widths[4], truncateAddr(remoteAddr, widths[4]),
			widths[5], conn.State,
		)
		change := m.GetChange(conn.Connection)
		b.WriteString(renderRowWithHighlight(row, isSelected, change))
	}

	return b.String()
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

// updateViewportContent renders the current view and sets viewport content.
// MUST be called from Update() (not View()) so viewport knows content height for scrolling.
// Uses data-only rendering since headers are rendered outside viewport (frozen).
func (m *Model) updateViewportContent() {
	if !m.ready {
		return
	}

	view := m.CurrentView()
	if view == nil {
		m.viewportContent = ""
		return
	}

	var content string
	if m.snapshot == nil {
		content = LoadingStyle().Render("Loading...")
	} else if len(m.snapshot.Applications) == 0 {
		content = EmptyStyle().Render("No network connections found")
	} else {
		// Use data-only methods (headers are rendered separately outside viewport)
		switch view.Level {
		case LevelProcessList:
			content = m.renderProcessListData()
		case LevelConnections:
			content = m.renderConnectionsListData()
		case LevelAllConnections:
			content = m.renderAllConnectionsData()
		}
	}

	m.viewportContent = content
	m.viewport.SetContent(content)
}

// syncViewportScroll adjusts the viewport scroll position to keep the cursor visible.
// MUST be called from Update() (not View()) to persist the scroll position.
func (m *Model) syncViewportScroll() {
	if !m.ready {
		return
	}

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
// Since headers are now frozen outside the viewport, cursor position is just the index.
func (m Model) cursorLinePosition() int {
	view := m.CurrentView()
	if view == nil || m.snapshot == nil {
		return 0
	}
	return view.Cursor
}

// centerModal renders a framed modal centered on screen.
func (m Model) centerModal(content, title string, modalWidth int) string {
	if m.width < modalWidth+4 {
		modalWidth = m.width - 4
	}

	lines := strings.Split(content, "\n")
	modalHeight := len(lines) + 4 // content + border + padding

	framedModal := RenderFrameWithTitle(content, title, modalWidth, modalHeight)

	leftPad := max((m.width-modalWidth)/2, 0)
	topPad := max((m.height-modalHeight)/2, 0)

	var b strings.Builder
	for i := 0; i < topPad; i++ {
		b.WriteString(DimmedStyle().Render(strings.Repeat(" ", m.width)))
		b.WriteString("\n")
	}
	for _, line := range strings.Split(framedModal, "\n") {
		b.WriteString(strings.Repeat(" ", leftPad))
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

// renderHelpModal renders the help modal overlay with keyboard shortcuts.
func (m Model) renderHelpModal() string {
	keyStyle := FooterKeyStyle()
	descStyle := FooterDescStyle()

	formatKey := func(k Keybinding) string {
		return keyStyle.Render(k.Key) + descStyle.Render(" "+k.Desc)
	}

	lines := []string{
		// Navigation
		HeaderStyle().Render("Navigation"),
		formatKey(KeyUp) + ", " + formatKey(KeyUpAlt),
		formatKey(KeyDown) + ", " + formatKey(KeyDownAlt),
		formatKey(KeyPageUp) + ", " + formatKey(KeyPageDown),
		formatKey(KeyEnter) + descStyle.Render(" Select/drill-down"),
		formatKey(KeyEsc) + ", " + formatKey(KeyBack) + descStyle.Render(" Back/cancel"),
		"",
		// Views
		HeaderStyle().Render("Views"),
		formatKey(KeyToggleView),
		formatKey(KeySortMode),
		keyStyle.Render("←→") + descStyle.Render(" Select column (sort mode)"),
		"",
		// Search
		HeaderStyle().Render("Search"),
		formatKey(KeySearch),
		"",
		// Actions
		HeaderStyle().Render("Actions"),
		formatKey(KeyKillTerm),
		formatKey(KeyKillForce),
		formatKey(KeyRefreshUp) + ", " + keyStyle.Render("=") + descStyle.Render(" Faster refresh"),
		formatKey(KeyRefreshDown) + ", " + keyStyle.Render("_") + descStyle.Render(" Slower refresh"),
		"",
		// Other
		HeaderStyle().Render("Other"),
		formatKey(KeySettings),
		formatKey(KeyHelp),
		formatKey(KeyQuit) + ", " + keyStyle.Render("ctrl+c") + descStyle.Render(" Quit"),
	}

	return m.centerModal(strings.Join(lines, "\n"), "Keyboard Shortcuts", 60)
}

// renderSettingsModal renders the settings modal overlay.
func (m Model) renderSettingsModal() string {
	var b strings.Builder

	// Header
	header := HeaderStyle().Width(m.width).Render("netmon - Settings")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Settings options
	settings := []struct {
		name    string
		enabled bool
		desc    string
	}{
		{"DNS Resolution", m.dnsEnabled, "Resolve IP addresses to hostnames"},
		{"Service Names", m.serviceNames, "Show service names instead of port numbers"},
		{"Highlight Changes", m.highlightChanges, "Highlight added/removed connections"},
	}

	for i, s := range settings {
		// Selection indicator
		cursor := "  "
		if i == m.settingsCursor {
			cursor = "> "
		}

		// Toggle state
		toggle := "[ ]"
		if s.enabled {
			toggle = "[x]"
		}

		// Row
		row := fmt.Sprintf("%s%s %s - %s", cursor, toggle, s.name, s.desc)
		if i == m.settingsCursor {
			b.WriteString(SelectedConnStyle().Render(row))
		} else {
			b.WriteString(ConnStyle().Render(row))
		}
		b.WriteString("\n")
	}

	// Footer with keybindings
	b.WriteString("\n")
	keyStyle := FooterKeyStyle()
	descStyle := FooterDescStyle()
	footer := keyStyle.Render("↑↓") + descStyle.Render(" Navigate  ") +
		keyStyle.Render("Space") + descStyle.Render(" Toggle  ") +
		keyStyle.Render("Esc") + descStyle.Render(" Close")
	b.WriteString(FooterStyle().Width(m.width).Render(footer))

	return b.String()
}
