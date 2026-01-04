package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kostyay/netmon/internal/model"
)

// Layout constants for fixed header/footer with scrollable content.
const (
	headerHeight = 1 // title row
	footerHeight = 2 // crumbs + keybindings
	frameHeight  = 2 // top and bottom border
)

// columnDef defines a table column with sizing properties.
type columnDef struct {
	label      string
	id         SortColumn // column identifier for selection/sorting
	minWidth   int        // minimum width
	flex       int        // flex weight for extra space distribution (0 = fixed)
	rightAlign bool       // true for right-aligned columns (numbers)
}

// renderRow renders a table row with selection styling.
func renderRow(content string, isSelected bool) string {
	row := "  " + content
	if isSelected {
		return SelectedConnStyle().Render(row) + "\n"
	}
	return ConnStyle().Render(row) + "\n"
}

// renderTableHeader renders a table header with optional sort indicators.
// If showSort is false, sort indicators are not displayed (for process list).
func renderTableHeader(columns []columnDef, widths []int, selectedCol, sortCol SortColumn, sortAsc, showSort bool) string {
	var b strings.Builder

	// Add 2-space prefix to align with data rows (which have "  " prefix from renderRow)
	b.WriteString("  ")

	headerStyle := TableHeaderStyle()
	selectedStyle := TableHeaderSelectedStyle()
	sortStyle := SortIndicatorStyle()

	for i, col := range columns {
		if i > 0 {
			b.WriteString(" ")
		}

		isSelected := selectedCol == col.id
		isSorted := showSort && sortCol == col.id

		header := col.label

		var sortIndicator string
		if isSorted {
			if sortAsc {
				sortIndicator = "↑"
			} else {
				sortIndicator = "↓"
			}
		}

		padWidth := widths[i] - len(header)
		if isSorted {
			padWidth -= 1
		}
		if padWidth < 0 {
			padWidth = 0
		}
		var paddedHeader string
		if col.rightAlign {
			paddedHeader = strings.Repeat(" ", padWidth) + header
		} else {
			paddedHeader = header + strings.Repeat(" ", padWidth)
		}

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

// calculateColumnWidths distributes available width among columns.
// Fixed columns (flex=0) get their minWidth, remaining space goes to flex columns.
func calculateColumnWidths(columns []columnDef, availableWidth int) []int {
	widths := make([]int, len(columns))

	// Account for spaces between columns and selection marker
	separators := len(columns) - 1
	selectionMarker := 2 // "  " prefix for all rows
	availableWidth -= separators + selectionMarker

	// First pass: assign minimum widths and calculate total flex
	totalMinWidth := 0
	totalFlex := 0
	for i, col := range columns {
		widths[i] = col.minWidth
		totalMinWidth += col.minWidth
		totalFlex += col.flex
	}

	// Distribute remaining space to flex columns
	extraSpace := availableWidth - totalMinWidth
	if extraSpace > 0 && totalFlex > 0 {
		for i, col := range columns {
			if col.flex > 0 {
				extra := (extraSpace * col.flex) / totalFlex
				widths[i] += extra
			}
		}
	}

	return widths
}

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

	// Set content, sync scroll position, and render viewport
	m.viewport.SetContent(content)
	m.ensureCursorVisible()

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

	// Row 1: Breadcrumbs (full width)
	b.WriteString(StatusStyle().Width(m.width).Render(m.renderBreadcrumbsText()))
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

	switch view.Level {
	case LevelProcessList:
		parts = []string{
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("Enter") + descStyle.Render(" Drill-in"),
			keyStyle.Render("s") + descStyle.Render(" Sort"),
			keyStyle.Render("v") + descStyle.Render(" All"),
			keyStyle.Render("+/-") + descStyle.Render(" Refresh"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	case LevelConnections:
		parts = []string{
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("s") + descStyle.Render(" Sort"),
			keyStyle.Render("Esc") + descStyle.Render(" Back"),
			keyStyle.Render("v") + descStyle.Render(" All"),
			keyStyle.Render("+/-") + descStyle.Render(" Refresh"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	case LevelAllConnections:
		parts = []string{
			keyStyle.Render("↑↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("s") + descStyle.Render(" Sort"),
			keyStyle.Render("v") + descStyle.Render(" Grouped"),
			keyStyle.Render("+/-") + descStyle.Render(" Refresh"),
			keyStyle.Render("q") + descStyle.Render(" Quit"),
		}
	}

	return strings.Join(parts, "  ")
}

// processListColumns returns the column definitions for the process list.
func processListColumns() []columnDef {
	return []columnDef{
		{label: "Process", id: SortProcess, minWidth: 15, flex: 3, rightAlign: false},
		{label: "Conns", id: SortConns, minWidth: 6, flex: 0, rightAlign: true},
		{label: "ESTAB", id: SortEstablished, minWidth: 6, flex: 0, rightAlign: true},
		{label: "LISTEN", id: SortListen, minWidth: 7, flex: 0, rightAlign: true},
		{label: "TX", id: SortTX, minWidth: 8, flex: 1, rightAlign: true},
		{label: "RX", id: SortRX, minWidth: 8, flex: 1, rightAlign: true},
	}
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

	// Calculate column widths
	columns := processListColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())

	// Render header
	b.WriteString(m.renderProcessListHeader(widths))
	b.WriteString("\n")

	// Sort applications
	apps := m.sortProcessList(m.snapshot.Applications)

	// Render each process row
	for i, app := range apps {
		isSelected := i == view.Cursor

		// Aggregate TX/RX stats for all PIDs of this app
		txStr, rxStr := m.getAggregatedNetIO(app.PIDs)

		// Build row content with dynamic widths
		row := fmt.Sprintf("%-*s %*d %*d %*d %*s %*s",
			widths[0], truncateString(app.Name, widths[0]),
			widths[1], len(app.Connections),
			widths[2], app.EstablishedCount,
			widths[3], app.ListenCount,
			widths[4], txStr,
			widths[5], rxStr,
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

// connectionsColumns returns the column definitions for the connections list.
func connectionsColumns() []columnDef {
	return []columnDef{
		{label: "Proto", id: SortProtocol, minWidth: 6, flex: 0},
		{label: "Local", id: SortLocal, minWidth: 20, flex: 2},
		{label: "Remote", id: SortRemote, minWidth: 20, flex: 2},
		{label: "State", id: SortState, minWidth: 11, flex: 1},
	}
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
	// Calculate column widths
	columns := connectionsColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())

	// Header
	b.WriteString(m.renderConnectionsHeader(widths))
	b.WriteString("\n")

	// Sort connections
	conns := m.sortConnectionsForView(selectedApp.Connections)

	// Render each connection (no PID column - redundant at this level)
	for i, conn := range conns {
		isSelected := i == view.Cursor

		row := fmt.Sprintf("%-*s %-*s %-*s %-*s",
			widths[0], conn.Protocol,
			widths[1], truncateAddr(conn.LocalAddr, widths[1]),
			widths[2], truncateAddr(conn.RemoteAddr, widths[2]),
			widths[3], conn.State,
		)

		b.WriteString(renderRow(row, isSelected))
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

// allConnectionsColumns returns the column definitions for the all-connections list.
func allConnectionsColumns() []columnDef {
	return []columnDef{
		{label: "PID", id: SortPID, minWidth: 6, flex: 0, rightAlign: true},
		{label: "Process", id: SortProcess, minWidth: 12, flex: 2},
		{label: "Proto", id: SortProtocol, minWidth: 6, flex: 0},
		{label: "Local", id: SortLocal, minWidth: 18, flex: 2},
		{label: "Remote", id: SortRemote, minWidth: 18, flex: 2},
		{label: "State", id: SortState, minWidth: 11, flex: 1},
	}
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

	var b strings.Builder

	// === CONNECTIONS TABLE ===
	// Calculate column widths
	columns := allConnectionsColumns()
	widths := calculateColumnWidths(columns, m.contentWidth())

	// Header
	b.WriteString(m.renderAllConnectionsHeader(widths))
	b.WriteString("\n")

	// Collect all connections with process names
	var allConns []connectionWithProcess
	for _, app := range m.snapshot.Applications {
		for _, conn := range app.Connections {
			allConns = append(allConns, connectionWithProcess{
				Connection:  conn,
				ProcessName: app.Name,
			})
		}
	}

	// Sort connections
	allConns = m.sortAllConnections(allConns)

	// Render each connection
	for i, conn := range allConns {
		isSelected := i == view.Cursor

		row := fmt.Sprintf("%*d %-*s %-*s %-*s %-*s %-*s",
			widths[0], conn.PID,
			widths[1], truncateString(conn.ProcessName, widths[1]),
			widths[2], conn.Protocol,
			widths[3], truncateAddr(conn.LocalAddr, widths[3]),
			widths[4], truncateAddr(conn.RemoteAddr, widths[4]),
			widths[5], conn.State,
		)

		b.WriteString(renderRow(row, isSelected))
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

// sortAllConnections sorts connections with process names based on current view state.
func (m Model) sortAllConnections(conns []connectionWithProcess) []connectionWithProcess {
	view := m.CurrentView()
	if view == nil {
		return conns
	}

	sorted := make([]connectionWithProcess, len(conns))
	copy(sorted, conns)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool
		switch view.SortColumn {
		case SortPID:
			less = sorted[i].PID < sorted[j].PID
		case SortProcess:
			less = sorted[i].ProcessName < sorted[j].ProcessName
		case SortProtocol:
			less = sorted[i].Protocol < sorted[j].Protocol
		case SortLocal:
			less = sorted[i].LocalAddr < sorted[j].LocalAddr
		case SortRemote:
			less = sorted[i].RemoteAddr < sorted[j].RemoteAddr
		case SortState:
			less = sorted[i].State < sorted[j].State
		default:
			less = sorted[i].PID < sorted[j].PID
		}

		if view.SortAscending {
			return less
		}
		return !less
	})

	return sorted
}

// sortProcessList sorts applications based on current view state.
func (m Model) sortProcessList(apps []model.Application) []model.Application {
	view := m.CurrentView()
	if view == nil {
		return apps
	}

	sorted := make([]model.Application, len(apps))
	copy(sorted, apps)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool
		switch view.SortColumn {
		case SortProcess:
			less = sorted[i].Name < sorted[j].Name
		case SortConns:
			less = len(sorted[i].Connections) < len(sorted[j].Connections)
		case SortEstablished:
			less = sorted[i].EstablishedCount < sorted[j].EstablishedCount
		case SortListen:
			less = sorted[i].ListenCount < sorted[j].ListenCount
		case SortTX:
			// Sort by aggregated TX bytes
			txI := m.getAggregatedTXBytes(sorted[i].PIDs)
			txJ := m.getAggregatedTXBytes(sorted[j].PIDs)
			less = txI < txJ
		case SortRX:
			// Sort by aggregated RX bytes
			rxI := m.getAggregatedRXBytes(sorted[i].PIDs)
			rxJ := m.getAggregatedRXBytes(sorted[j].PIDs)
			less = rxI < rxJ
		default:
			less = sorted[i].Name < sorted[j].Name
		}

		if view.SortAscending {
			return less
		}
		return !less
	})

	return sorted
}

// getAggregatedTXBytes returns the total TX bytes for all PIDs.
func (m Model) getAggregatedTXBytes(pids []int32) uint64 {
	var total uint64
	for _, pid := range pids {
		if stats, ok := m.netIOCache[pid]; ok {
			total += stats.BytesSent
		}
	}
	return total
}

// getAggregatedRXBytes returns the total RX bytes for all PIDs.
func (m Model) getAggregatedRXBytes(pids []int32) uint64 {
	var total uint64
	for _, pid := range pids {
		if stats, ok := m.netIOCache[pid]; ok {
			total += stats.BytesRecv
		}
	}
	return total
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

	// Account for header row
	const tableHeaderLines = 1
	return view.Cursor + tableHeaderLines
}

