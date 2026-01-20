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
	headerHeight = 3 // double-line box header (top border + content + bottom border)
	footerHeight = 2 // crumbs + keybindings
	frameHeight  = 2 // top and bottom border
)

// renderHeader renders the industrial-style header with live indicator and stats.
func (m Model) renderHeader() string {
	borderStyle := BorderStyle()
	titleStyle := HeaderStyle()
	liveStyle := LiveIndicatorStyle()
	statsStyle := StatsStyle()
	warnStyle := WarnStyle()

	innerWidth := m.width - 2

	// Double-line box drawing
	topLeft := "╔"
	topRight := "╗"
	bottomLeft := "╚"
	bottomRight := "╝"
	horizontal := "═"
	vertical := "║"

	// Build top border with centered NETMON title
	title := " NETMON "
	titleLen := len(title)
	remainingWidth := innerWidth - titleLen
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	leftPad := remainingWidth / 2
	rightPad := remainingWidth - leftPad

	topBorder := borderStyle.Render(topLeft)
	topBorder += borderStyle.Render(strings.Repeat(horizontal, leftPad))
	topBorder += titleStyle.Render(title)
	topBorder += borderStyle.Render(strings.Repeat(horizontal, rightPad))
	topBorder += borderStyle.Render(topRight)

	// Build content line with live indicator and stats
	// Live indicator: ◉ (filled) or ○ (empty) based on animation frame
	liveIndicator := "◉"
	if m.animations && m.animationFrame == 1 {
		liveIndicator = "○"
	}
	liveText := liveStyle.Render(liveIndicator + " LIVE")

	// Connection count
	connCount := 0
	var totalTX, totalRX uint64
	if m.snapshot != nil {
		connCount = m.snapshot.TotalConnections()
		// Aggregate TX/RX from cache
		for _, stats := range m.netIOCache {
			totalTX += stats.BytesSent
			totalRX += stats.BytesRecv
		}
	}

	// Format stats
	statsText := statsStyle.Render(fmt.Sprintf("  %d connections", connCount))
	ioText := statsStyle.Render(fmt.Sprintf("   ▲ %s   ▼ %s", formatBytes(totalTX), formatBytes(totalRX)))
	refreshText := statsStyle.Render(fmt.Sprintf("   %.1fs", m.refreshInterval.Seconds()))

	// Error or update indicator
	rightContent := ""
	if m.lastError != nil {
		rightContent = warnStyle.Render(fmt.Sprintf("  ⚠ %s", truncateString(m.lastError.Error(), 30)))
	} else if m.updateAvailable != "" {
		rightContent = warnStyle.Render(fmt.Sprintf("  ▲ %s", m.updateAvailable))
	}

	content := liveText + statsText + ioText + refreshText + rightContent

	// Pad content to fill width
	contentWidth := lipgloss.Width(content)
	padding := innerWidth - contentWidth - 2 // -2 for vertical bars
	if padding < 0 {
		padding = 0
	}

	contentLine := borderStyle.Render(vertical)
	contentLine += " " + content + strings.Repeat(" ", padding) + " "
	contentLine += borderStyle.Render(vertical)

	// Build bottom border
	bottomBorder := borderStyle.Render(bottomLeft)
	bottomBorder += borderStyle.Render(strings.Repeat(horizontal, innerWidth))
	bottomBorder += borderStyle.Render(bottomRight)

	return topBorder + "\n" + contentLine + "\n" + bottomBorder
}

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

	// Render base content
	baseContent := m.renderBaseView()

	// Overlay modals if active
	if m.helpMode {
		return m.overlayModal(baseContent, m.renderHelpModalContent(), "Keyboard Shortcuts", 60)
	}
	if m.settingsMode {
		return m.overlayModal(baseContent, m.renderSettingsModalContent(), "Settings", 44)
	}
	if m.killMode && m.killTarget != nil {
		// Calculate modal width based on path length, max 70% of screen
		modalWidth := 50
		if m.killTarget.Exe != "" {
			// Path line: "  Path:    " (11 chars) + path + padding
			pathWidth := 11 + len(m.killTarget.Exe) + 4
			if pathWidth > modalWidth {
				modalWidth = pathWidth
			}
		}
		maxWidth := m.width * 70 / 100
		if modalWidth > maxWidth {
			modalWidth = maxWidth
		}
		return m.overlayDangerModal(baseContent, m.renderKillModalContent(), "Kill Process", modalWidth)
	}

	return baseContent
}

// renderBaseView renders the main UI without modals.
func (m Model) renderBaseView() string {
	var b strings.Builder

	// === HEADER (industrial double-line box with live stats) ===
	b.WriteString(m.renderHeader())
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

	// Helper to render a content line with borders
	renderLine := func(line string) {
		result.WriteString(borderStyle.Render(vertical))
		result.WriteString(" ")
		result.WriteString(padRight(line, innerWidth-2))
		result.WriteString(" ")
		result.WriteString(borderStyle.Render(vertical))
		result.WriteString("\n")
	}

	// Render frozen header lines (outside viewport, won't scroll)
	if frozenHeader := m.renderFrozenHeader(); frozenHeader != "" {
		for _, line := range strings.Split(frozenHeader, "\n") {
			renderLine(line)
		}
	}

	// Render viewport content (scrollable data rows)
	for _, line := range strings.Split(m.viewport.View(), "\n") {
		renderLine(line)
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
	switch view.Level {
	case LevelProcessList:
		return "PROCESSES"
	case LevelConnections:
		return "PROCESSES > " + view.ProcessName
	case LevelAllConnections:
		return "ALL CONNECTIONS"
	default:
		return ""
	}
}

// renderFooter renders the two-row footer with crumbs and keybindings.
func (m Model) renderFooter() string {
	var b strings.Builder
	statusStyle := StatusStyle()

	// Row 1: Status line (result, search, or breadcrumbs)
	if m.killResult != "" && time.Since(m.killResultAt) < 2*time.Second {
		b.WriteString(statusStyle.Width(m.width).Render(m.killResult))
	} else if m.searchMode {
		b.WriteString(statusStyle.Width(m.width).Render(fmt.Sprintf("/%s█", m.searchQuery)))
	} else {
		// Breadcrumbs + filter indicator
		statusLine := m.renderBreadcrumbsText()
		if m.activeFilter != "" {
			statusLine = fmt.Sprintf("[%s] %s", m.activeFilter, statusLine)
		}
		b.WriteString(statusStyle.Width(m.width).Render(statusLine))
	}
	b.WriteString("\n")

	// Row 2: Keybindings
	b.WriteString(FooterStyle().Width(m.width).Render(m.renderKeybindingsText()))

	return b.String()
}

// renderKeybindingsText returns keybindings in modern minimal style.
// Clean keys with soft separators, no backgrounds, natural flow.
func (m Model) renderKeybindingsText() string {
	view := m.CurrentView()
	if view == nil {
		return ""
	}

	keyStyle := FooterKeyStyle()
	descStyle := FooterDescStyle()
	sepStyle := FooterDescStyle()

	// Helper to format a key-label pair
	btn := func(key, label string) string {
		return keyStyle.Render(key) + " " + descStyle.Render(label)
	}

	sep := sepStyle.Render("  ·  ")

	var parts []string

	if m.killMode {
		parts = []string{
			btn("↵", "confirm"),
			btn("↑↓", "signal"),
			btn("esc", "cancel"),
		}
	} else if view.SortMode {
		parts = []string{
			btn("←→", "column"),
			btn("↵", "apply"),
			btn("esc", "cancel"),
		}
	} else if m.searchMode {
		parts = []string{
			btn("↵", "apply"),
			btn("esc", "cancel"),
		}
	} else {
		// Normal mode - contextual keys
		switch view.Level {
		case LevelProcessList:
			parts = []string{
				btn("↵", "drill"),
				btn("/", "search"),
				btn("s", "sort"),
				btn("v", "flat"),
				btn("x", "kill"),
				btn("S", "settings"),
				btn("?", "help"),
				btn("q", "quit"),
			}
		case LevelConnections:
			parts = []string{
				btn("esc", "back"),
				btn("/", "search"),
				btn("s", "sort"),
				btn("v", "flat"),
				btn("x", "kill"),
				btn("S", "settings"),
				btn("?", "help"),
				btn("q", "quit"),
			}
		case LevelAllConnections:
			parts = []string{
				btn("/", "search"),
				btn("s", "sort"),
				btn("v", "grouped"),
				btn("x", "kill"),
				btn("S", "settings"),
				btn("?", "help"),
				btn("q", "quit"),
			}
		}
	}

	return strings.Join(parts, sep)
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

// overlayModal renders a modal on top of background content with dimmed backdrop.
func (m Model) overlayModal(background, content, title string, modalWidth int) string {
	return m.overlayModalWithRenderer(background, content, title, modalWidth, RenderFrameWithTitle)
}

// overlayDangerModal renders a danger-styled modal (red borders) on top of background content.
func (m Model) overlayDangerModal(background, content, title string, modalWidth int) string {
	return m.overlayModalWithRenderer(background, content, title, modalWidth, RenderDangerFrameWithTitle)
}

// overlayModalWithRenderer renders a modal using the provided frame renderer.
func (m Model) overlayModalWithRenderer(background, content, title string, modalWidth int, frameRenderer func(string, string, int, int) string) string {
	if m.width < modalWidth+4 {
		modalWidth = m.width - 4
	}

	contentLines := strings.Split(content, "\n")
	modalHeight := len(contentLines) + 4

	framedModal := frameRenderer(content, title, modalWidth, modalHeight)
	modalLines := strings.Split(framedModal, "\n")

	leftPad := max((m.width-modalWidth-4)/2, 0)
	topPad := max((m.height-modalHeight)/2, 0)

	bgLines := strings.Split(background, "\n")
	for len(bgLines) < m.height {
		bgLines = append(bgLines, "")
	}

	dimStyle := DimmedStyle()
	for i := range bgLines {
		bgLines[i] = dimStyle.Render(stripAnsi(bgLines[i]))
	}

	for i, modalLine := range modalLines {
		bgIdx := topPad + i
		if bgIdx >= 0 && bgIdx < len(bgLines) {
			leftBg := ""
			if leftPad > 0 {
				leftBg = dimStyle.Render(strings.Repeat(" ", leftPad))
			}
			bgLines[bgIdx] = leftBg + modalLine
		}
	}

	return strings.Join(bgLines[:m.height], "\n")
}

// renderKillModalContent returns the kill confirmation modal content.
func (m Model) renderKillModalContent() string {
	if m.killTarget == nil {
		return ""
	}

	dangerStyle := ErrorStyle()
	descStyle := FooterDescStyle()

	var lines []string

	// Warning icon and message
	lines = append(lines, "")

	if len(m.killTarget.PIDs) > 1 {
		lines = append(lines, dangerStyle.Render(fmt.Sprintf("  Kill %d processes?", len(m.killTarget.PIDs))))
		lines = append(lines, "")
		lines = append(lines, descStyle.Render(fmt.Sprintf("  Process: %s", m.killTarget.ProcessName)))
		if m.killTarget.Exe != "" {
			lines = append(lines, DimmedStyle().Render(fmt.Sprintf("  Path:    %s", m.killTarget.Exe)))
		}
		lines = append(lines, descStyle.Render(fmt.Sprintf("  PIDs:    %s", formatPIDList(m.killTarget.PIDs))))
	} else {
		lines = append(lines, dangerStyle.Render("  Kill this process?"))
		lines = append(lines, "")
		lines = append(lines, descStyle.Render(fmt.Sprintf("  Process: %s", m.killTarget.ProcessName)))
		if m.killTarget.Exe != "" {
			lines = append(lines, DimmedStyle().Render(fmt.Sprintf("  Path:    %s", m.killTarget.Exe)))
		}
		lines = append(lines, descStyle.Render(fmt.Sprintf("  PID:     %d", m.killTarget.PID)))
	}

	// Signal radio options
	lines = append(lines, "")
	termSelected := m.killTarget.Signal == "SIGTERM"
	if termSelected {
		lines = append(lines, "  "+dangerStyle.Render("(●)")+" "+descStyle.Render("SIGTERM")+"  "+descStyle.Render("graceful"))
		lines = append(lines, "  "+descStyle.Render("( ) SIGKILL")+"  "+DimmedStyle().Render("force"))
	} else {
		lines = append(lines, "  "+descStyle.Render("( ) SIGTERM")+"  "+DimmedStyle().Render("graceful"))
		lines = append(lines, "  "+dangerStyle.Render("(●)")+" "+descStyle.Render("SIGKILL")+"  "+descStyle.Render("force"))
	}

	lines = append(lines, "")

	// Footer keybindings - all red
	footer := dangerStyle.Render("↵") + descStyle.Render(" Confirm  ") +
		dangerStyle.Render("Esc") + descStyle.Render(" Cancel  ") +
		dangerStyle.Render("↑↓") + descStyle.Render(" Signal")
	lines = append(lines, "  "+footer)

	return strings.Join(lines, "\n")
}

// stripAnsi removes ANSI escape codes from a string.
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// renderHelpModalContent returns the help modal content.
func (m Model) renderHelpModalContent() string {
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

	return strings.Join(lines, "\n")
}

// renderSettingsModalContent returns the settings modal content.
func (m Model) renderSettingsModalContent() string {
	var lines []string

	settings := []struct {
		name    string
		enabled bool
		desc    string
	}{
		{"DNS Resolution", m.dnsEnabled, "Reverse lookup IPs to hostnames"},
		{"Service Names", m.serviceNames, "Show http/https instead of 80/443"},
		{"Highlight Changes", m.highlightChanges, "Flash new/removed connections"},
		{"Animations", m.animations, "Enable UI animations (pulse, spinners)"},
	}

	for i, s := range settings {
		cursor := "  "
		if i == m.settingsCursor {
			cursor = "▸ "
		}
		toggle := "[ ]"
		if s.enabled {
			toggle = "[■]"
		}
		row := fmt.Sprintf("%s%s %s", cursor, toggle, s.name)
		if i == m.settingsCursor {
			row = SelectedConnStyle().Render(row)
		}
		lines = append(lines, row)
		// Description line (dimmed, indented)
		lines = append(lines, DimmedStyle().Render("      "+s.desc))
	}

	// Footer keybindings
	lines = append(lines, "")
	keyStyle := FooterKeyStyle()
	descStyle := FooterDescStyle()
	footer := keyStyle.Render("↑↓") + descStyle.Render(" Navigate  ") +
		keyStyle.Render("Space") + descStyle.Render(" Toggle  ") +
		keyStyle.Render("Esc") + descStyle.Render(" Close")
	lines = append(lines, footer)

	return strings.Join(lines, "\n")
}
