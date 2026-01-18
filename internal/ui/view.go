package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/kostyay/netmon/internal/model"
)

// Layout constants for fixed header/footer with scrollable content.
const (
	headerHeight = 1 // title row
	footerHeight = 2 // crumbs + keybindings
	frameHeight  = 2 // top and bottom border
)

// contentWidth returns the available width for table content.
// Accounts for frame border and padding.
func (m Model) contentWidth() int {
	// Frame has 2 chars border + 2 chars padding = 4 total
	return m.width - 4
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

	// === CONTENT (scrollable via viewport, wrapped in frame) ===
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
		case LevelAllConnections:
			content = m.renderAllConnections()
		}
	}

	// Set content and render viewport
	// Note: scroll position is synced in Update() via syncViewportScroll()
	m.viewport.SetContent(content)

	// Calculate connection count for title
	connCount := 0
	if m.snapshot != nil {
		connCount = m.snapshot.TotalConnections()
	}
	frameTitle := fmt.Sprintf("connections: %d", connCount)

	// Wrap viewport in a frame with centered title
	framedContent := RenderFrameWithTitle(m.viewport.View(), frameTitle, m.width, m.viewport.Height+frameHeight)
	b.WriteString(framedContent)
	b.WriteString("\n")

	// === FOOTER (fixed at bottom, full width) ===
	b.WriteString(m.renderFooter())

	return b.String()
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
		prompt := fmt.Sprintf("Kill PID %d (%s) with %s? [y/n]",
			m.killTarget.PID, m.killTarget.ProcessName, m.killTarget.Signal)
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
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("Enter") + descStyle.Render(" Drill-in"),
			keyStyle.Render("/") + descStyle.Render(" Search"),
			keyStyle.Render("s") + descStyle.Render(" Sort"),
			keyStyle.Render("x/X") + descStyle.Render(" Kill"),
			keyStyle.Render("v") + descStyle.Render(" All"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	case LevelConnections:
		parts = []string{
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("/") + descStyle.Render(" Search"),
			keyStyle.Render("s") + descStyle.Render(" Sort"),
			keyStyle.Render("x/X") + descStyle.Render(" Kill"),
			keyStyle.Render("Esc") + descStyle.Render(" Back"),
			keyStyle.Render("v") + descStyle.Render(" All"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	case LevelAllConnections:
		parts = []string{
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("/") + descStyle.Render(" Search"),
			keyStyle.Render("s") + descStyle.Render(" Sort"),
			keyStyle.Render("x/X") + descStyle.Render(" Kill"),
			keyStyle.Render("v") + descStyle.Render(" Grouped"),
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
func (m Model) filteredApps() []model.Application {
	if m.snapshot == nil {
		return nil
	}
	filter := m.currentFilter()
	if filter == "" {
		return m.snapshot.Applications
	}

	var result []model.Application
	for _, app := range m.snapshot.Applications {
		ports := extractPorts(app.Connections)
		if matchesFilter(filter, app.Name, app.PIDs, ports) {
			result = append(result, app)
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
	for _, conn := range conns {
		if matchesConnection(filter, conn) {
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
	for _, app := range m.snapshot.Applications {
		for _, conn := range app.Connections {
			// No filter or matches filter - include connection
			if filter == "" || matchesFilter(filter, app.Name, app.PIDs, extractPortsFromAddrs(conn.LocalAddr, conn.RemoteAddr)) {
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

		remoteAddr := formatRemoteAddr(conn.RemoteAddr, m.dnsCache, m.serviceNames)
		localAddr := formatAddr(conn.LocalAddr, m.serviceNames)
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

		remoteAddr := formatRemoteAddr(conn.RemoteAddr, m.dnsCache, m.serviceNames)
		localAddr := formatAddr(conn.LocalAddr, m.serviceNames)
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
func (m Model) cursorLinePosition() int {
	view := m.CurrentView()
	if view == nil || m.snapshot == nil {
		return 0
	}

	// Different views have different header structures
	var headerLines int
	switch view.Level {
	case LevelConnections:
		// Process name (1) + [exe (1)] + stats line (1) + blank line (1) + table header (1)
		headerLines = 4
		// Add exe line if present
		for _, app := range m.snapshot.Applications {
			if app.Name == view.ProcessName && app.Exe != "" {
				headerLines = 5
				break
			}
		}
	default:
		// LevelProcessList and LevelAllConnections have just 1 table header line
		headerLines = 1
	}

	return view.Cursor + headerLines
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
