package ui

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/config"
	"github.com/kostyay/netmon/internal/dns"
	"github.com/kostyay/netmon/internal/model"
	"github.com/kostyay/netmon/internal/release"
)

// Animation tick interval (500ms for pulsing effect).
const animationInterval = 500 * time.Millisecond

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.tickCmd(),
		m.fetchData(),
		m.fetchNetIO(),
		m.checkVersion(),
	}
	if m.animations {
		cmds = append(cmds, m.animationTickCmd())
	}
	return tea.Batch(cmds...)
}

// Update handles messages and ensures viewport content/scroll is synced after any state change.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, cmd := m.update(msg)
	newModel := result.(Model)
	newModel.recalcViewportHeight() // Adjust for frozen header (varies by view level)
	newModel.updateViewportContent()
	newModel.syncViewportScroll()
	return newModel, cmd
}

// recalcViewportHeight recalculates viewport height based on current view's frozen header.
// Must be called after view switches since frozen header height varies by view level.
func (m *Model) recalcViewportHeight() {
	if !m.ready || m.height == 0 {
		return
	}
	frozenLines := m.frozenHeaderHeight()
	viewportHeight := m.height - headerHeight - footerHeight - frameHeight - frozenLines
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	m.viewport.Height = viewportHeight
}

// update handles messages (internal implementation).
func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate viewport height: total - header - footer - frame borders - frozen header
		// Frozen header varies by view level (1 for ProcessList/AllConns, 4-5 for Connections)
		frozenLines := m.frozenHeaderHeight()
		viewportHeight := msg.Height - headerHeight - footerHeight - frameHeight - frozenLines
		if viewportHeight < 1 {
			viewportHeight = 1
		}

		// Viewport width accounts for frame border and padding (2 border + 2 padding)
		viewportWidth := msg.Width - 4
		if viewportWidth < 1 {
			viewportWidth = 1
		}

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, viewportHeight)
			m.ready = true
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = viewportHeight
		}
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		// Kill mode intercepts all keys
		if m.killMode {
			if matchKey(key, KeyEnter) {
				return m.executeKill()
			}
			if matchKey(key, KeyEsc) {
				m.killMode = false
				m.killTarget = nil
				return m, nil
			}
			// Toggle signal with up/down/tab
			if matchKey(key, KeyUp, KeyUpAlt, KeyDown, KeyDownAlt) || key == "tab" {
				if m.killTarget != nil {
					if m.killTarget.Signal == "SIGTERM" {
						m.killTarget.Signal = "SIGKILL"
					} else {
						m.killTarget.Signal = "SIGTERM"
					}
				}
				return m, nil
			}
			return m, nil // Ignore other keys in kill mode
		}

		// Help mode intercepts all keys
		if m.helpMode {
			if matchKey(key, KeyEsc, KeyQuit, KeyHelp) {
				m.helpMode = false
				return m, nil
			}
			return m, nil // Ignore other keys in help mode
		}

		// Settings mode intercepts all keys
		if m.settingsMode {
			if matchKey(key, KeyEsc, KeySettings) {
				m.settingsMode = false
				return m, nil
			}
			if matchKey(key, KeyUp, KeyUpAlt) {
				if m.settingsCursor > 0 {
					m.settingsCursor--
				}
				return m, nil
			}
			if matchKey(key, KeyDown, KeyDownAlt) {
				maxCursor := 3 // Number of settings - 1
				if m.settingsCursor < maxCursor {
					m.settingsCursor++
				}
				return m, nil
			}
			if matchKey(key, KeyEnter, KeySpace) {
				// Toggle the selected setting
				switch m.settingsCursor {
				case 0: // DNS Resolution
					m.dnsEnabled = !m.dnsEnabled
					config.CurrentSettings.DNSEnabled = m.dnsEnabled
					_ = config.SaveSettings(config.CurrentSettings)
				case 1: // Service Names
					m.serviceNames = !m.serviceNames
					config.CurrentSettings.ServiceNames = m.serviceNames
					_ = config.SaveSettings(config.CurrentSettings)
				case 2: // Highlight Changes
					m.highlightChanges = !m.highlightChanges
					config.CurrentSettings.HighlightChanges = m.highlightChanges
					_ = config.SaveSettings(config.CurrentSettings)
				case 3: // Animations
					m.animations = !m.animations
					config.CurrentSettings.Animations = m.animations
					_ = config.SaveSettings(config.CurrentSettings)
					// Start animation tick if enabled
					if m.animations {
						return m, m.animationTickCmd()
					}
				}
				return m, nil
			}
			return m, nil // Ignore other keys in settings mode
		}

		// Search mode intercepts all keys
		if m.searchMode {
			if matchKey(key, KeyEnter) {
				m.activeFilter = m.searchQuery
				m.searchMode = false
				m.clampCursor()
				return m, nil
			}
			if matchKey(key, KeyEsc) {
				m.searchQuery = m.activeFilter // revert to confirmed
				m.searchMode = false
				return m, nil
			}
			if matchKey(key, KeyBack) {
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
				return m, nil
			}
			// Append printable characters
			r := msg.Runes
			if len(r) == 1 && r[0] >= 32 {
				m.searchQuery += string(r)
			}
			return m, nil
		}

		// Global keybindings
		if matchKey(key, KeyQuit, KeyQuitAlt) {
			m.quitting = true
			return m, tea.Quit
		}

		if matchKey(key, KeyUp, KeyUpAlt) {
			view := m.CurrentView()
			if view == nil {
				return m, nil
			}
			// Use cursor directly (not resolveSelectionIndex) to handle duplicate items
			if view.Cursor > 0 {
				view.Cursor--
				m.updateSelectedIDFromCursor()
			}
			return m, nil
		}

		if matchKey(key, KeyDown, KeyDownAlt) {
			view := m.CurrentView()
			if view == nil || m.snapshot == nil {
				return m, nil
			}
			maxCursor := m.filteredCount()
			// Use cursor directly (not resolveSelectionIndex) to handle duplicate items
			if maxCursor > 0 && view.Cursor < maxCursor-1 {
				view.Cursor++
				m.updateSelectedIDFromCursor()
			}
			return m, nil
		}

		if matchKey(key, KeyLeft, KeyLeftAlt) {
			view := m.CurrentView()
			if view == nil || !view.SortMode {
				return m, nil
			}
			// Move column selection left (only in sort mode)
			columns := m.columnsForLevel(view.Level)
			currentIdx := m.findColumnIndex(columns, view.SelectedColumn)
			if currentIdx > 0 {
				view.SelectedColumn = columns[currentIdx-1]
			}
			return m, nil
		}

		if matchKey(key, KeyRight, KeyRightAlt) {
			view := m.CurrentView()
			if view == nil || !view.SortMode {
				return m, nil
			}
			// Move column selection right (only in sort mode)
			columns := m.columnsForLevel(view.Level)
			currentIdx := m.findColumnIndex(columns, view.SelectedColumn)
			if currentIdx < len(columns)-1 {
				view.SelectedColumn = columns[currentIdx+1]
			}
			return m, nil
		}

		if matchKey(key, KeyEnter, KeySpace) {
			view := m.CurrentView()
			if view == nil || m.snapshot == nil {
				return m, nil
			}
			if view.SortMode {
				// Apply sort and exit sort mode
				if view.SortColumn == view.SelectedColumn {
					view.SortAscending = !view.SortAscending
				} else {
					view.SortColumn = view.SelectedColumn
					view.SortAscending = true
				}
				view.SortMode = false
				return m, nil
			}
			// Not in sort mode - drill down on process list
			if view.Level == LevelProcessList {
				apps := m.sortProcessList(m.filteredApps())
				// Use cursor directly for selection
				if view.Cursor >= 0 && view.Cursor < len(apps) {
					app := apps[view.Cursor]
					// Clear filter when drilling down (different search context)
					m.activeFilter = ""
					m.searchQuery = ""
					m.PushView(ViewState{
						Level:          LevelConnections,
						ProcessName:    app.Name,
						Cursor:         0,
						SelectedID:     model.SelectionID{}, // Start fresh in new view
						SortColumn:     SortLocal,
						SortAscending:  true,
						SelectedColumn: SortLocal,
					})
				}
			}
			return m, nil
		}

		if matchKey(key, KeyEsc, KeyBack) {
			view := m.CurrentView()
			if view != nil && view.SortMode {
				// Exit sort mode without changing sort
				view.SortMode = false
				return m, nil
			}
			// Pop view (go back)
			m.PopView()
			return m, nil
		}

		if matchKey(key, KeyRefreshUp) || key == "=" {
			// Decrease refresh interval (faster refresh)
			if m.refreshInterval > MinRefreshInterval {
				m.refreshInterval -= RefreshStep
			}
			return m, nil
		}

		if matchKey(key, KeyRefreshDown) || key == "_" {
			// Increase refresh interval (slower refresh)
			if m.refreshInterval < MaxRefreshInterval {
				m.refreshInterval += RefreshStep
			}
			return m, nil
		}

		if matchKey(key, KeySortMode) {
			// Enter sort mode
			view := m.CurrentView()
			if view == nil || view.SortMode {
				return m, nil
			}
			view.SortMode = true
			view.SelectedColumn = view.SortColumn // Start at current sort column
			return m, nil
		}

		if matchKey(key, KeyToggleView) {
			// Toggle between grouped (process list) and ungrouped (all connections) view
			view := m.CurrentView()
			if view == nil {
				return m, nil
			}
			if view.Level == LevelAllConnections {
				// Toggle back to process list
				m.stack = []ViewState{{
					Level:          LevelProcessList,
					Cursor:         0,
					SortColumn:     SortProcess,
					SortAscending:  true,
					SelectedColumn: SortProcess,
				}}
			} else {
				// Toggle to all connections view
				m.stack = []ViewState{{
					Level:          LevelAllConnections,
					Cursor:         0,
					SortColumn:     SortProcess,
					SortAscending:  true,
					SelectedColumn: SortProcess,
				}}
			}
			return m, nil
		}

		if matchKey(key, KeySearch) {
			// Enter search mode
			m.searchMode = true
			m.searchQuery = m.activeFilter // pre-fill with current filter
			return m, nil
		}

		if matchKey(key, KeyKillTerm) {
			return m.enterKillMode("SIGTERM")
		}

		if matchKey(key, KeyKillForce) {
			return m.enterKillMode("SIGKILL")
		}

		if matchKey(key, KeySettings) {
			// Open settings modal
			m.settingsMode = true
			m.settingsCursor = 0
			return m, nil
		}

		if matchKey(key, KeyHelp) {
			// Open help modal
			m.helpMode = true
			return m, nil
		}

		if matchKey(key, KeyPageUp) {
			view := m.CurrentView()
			if view == nil {
				return m, nil
			}
			pageSize := m.viewport.Height
			if pageSize < 1 {
				pageSize = 10
			}
			view.Cursor -= pageSize
			if view.Cursor < 0 {
				view.Cursor = 0
			}
			m.updateSelectedIDFromCursor()
			return m, nil
		}

		if matchKey(key, KeyPageDown) {
			view := m.CurrentView()
			if view == nil || m.snapshot == nil {
				return m, nil
			}
			pageSize := m.viewport.Height
			if pageSize < 1 {
				pageSize = 10
			}
			maxCursor := m.filteredCount()
			view.Cursor += pageSize
			if maxCursor > 0 && view.Cursor >= maxCursor {
				view.Cursor = maxCursor - 1
			}
			m.updateSelectedIDFromCursor()
			return m, nil
		}

		// Pass unhandled keys to viewport for page up/down, mouse scroll, etc.
		if m.ready {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case TickMsg:
		// Prune expired change highlights (older than 3s)
		m.pruneExpiredChanges(3 * time.Second)

		// Schedule next tick and fetch new data
		return m, tea.Batch(
			m.tickCmd(),
			m.fetchData(),
			m.fetchNetIO(),
		)

	case DataMsg:
		if msg.Err != nil {
			// Store error for display in UI
			m.lastError = msg.Err
			m.lastErrorTime = time.Now()
			return m, nil
		}
		// Clear error on successful fetch
		m.lastError = nil

		// Diff connections and merge new changes
		newChanges := diffConnections(m.snapshot, msg.Snapshot)
		for k, v := range newChanges {
			m.changes[k] = v
		}

		// Store current as previous for next diff
		m.prevSnapshot = m.snapshot
		m.snapshot = msg.Snapshot

		// Handle --pid: drill into target process on first snapshot
		if m.targetPID != 0 {
			drillIntoPID(&m, m.targetPID)
			m.targetPID = 0 // Clear so we don't re-drill on every update
		}

		// Validate selection using ID-based resolution (handles item reordering)
		m.validateSelection()

		// Queue DNS lookups for new IPs (if enabled)
		dnsCmd := m.queueDNSLookups(msg.Snapshot)
		return m, dnsCmd

	case NetIOMsg:
		if msg.Err != nil {
			// Silently ignore network I/O errors - stats are optional
			return m, nil
		}
		// Update the netIOCache with new stats
		for pid, stats := range msg.Stats {
			m.netIOCache[pid] = stats
		}
		return m, nil

	case DNSResolvedMsg:
		if msg.Err != nil {
			// Cache failed lookup to avoid repeated attempts
			m.dnsCache[msg.IP] = ""
			return m, nil
		}
		// Cache successful lookup
		m.dnsCache[msg.IP] = msg.Hostname
		return m, nil

	case VersionCheckMsg:
		if msg.Err == nil && msg.LatestVersion != "" {
			m.updateAvailable = msg.LatestVersion
		}
		return m, nil

	case AnimationTickMsg:
		if !m.animations {
			return m, nil
		}
		// Advance animation frame (cycles 0-1 for pulse effect)
		m.animationFrame = (m.animationFrame + 1) % 2
		return m, m.animationTickCmd()
	}

	return m, nil
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) animationTickCmd() tea.Cmd {
	return tea.Tick(animationInterval, func(t time.Time) tea.Msg {
		return AnimationTickMsg(t)
	})
}

func (m Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		snapshot, err := m.collector.Collect(ctx)
		return DataMsg{Snapshot: snapshot, Err: err}
	}
}

func (m Model) fetchNetIO() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stats, err := m.netIOCollector.Collect(ctx)
		return NetIOMsg{Stats: stats, Err: err}
	}
}

func (m Model) checkVersion() tea.Cmd {
	return func() tea.Msg {
		latest, err := release.CheckLatest("kostyay", "netmon", m.version)
		return VersionCheckMsg{LatestVersion: latest, Err: err}
	}
}

// resolveDNS returns a command to resolve an IP address.
func (m Model) resolveDNS(ip string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		result := <-dns.ResolveAsync(ctx, ip)
		return DNSResolvedMsg{IP: result.IP, Hostname: result.Hostname, Err: result.Err}
	}
}

// queueDNSLookups returns commands for IPs that need resolution.
func (m Model) queueDNSLookups(snapshot *model.NetworkSnapshot) tea.Cmd {
	if !m.dnsEnabled || snapshot == nil {
		return nil
	}

	var cmds []tea.Cmd
	seen := make(map[string]bool)

	for _, app := range snapshot.Applications {
		for _, conn := range app.Connections {
			// Extract IP from remote address (skip port)
			ip := extractIP(conn.RemoteAddr)
			if ip == "" || ip == "*" || seen[ip] {
				continue
			}
			seen[ip] = true

			// Skip if already cached
			if _, ok := m.dnsCache[ip]; ok {
				continue
			}

			// Queue resolution (limit to avoid flooding)
			if len(cmds) < 10 {
				cmds = append(cmds, m.resolveDNS(ip))
			}
		}
	}

	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// extractIP extracts the IP portion from an address like "192.168.1.1:8080".
func extractIP(addr string) string {
	if addr == "" || addr == "*" {
		return ""
	}
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return addr
	}
	return addr[:idx]
}

// maxCursorForLevel returns the maximum cursor position for the given view level.
func (m Model) maxCursorForLevel(level ViewLevel) int {
	if m.snapshot == nil {
		return 0
	}
	switch level {
	case LevelProcessList:
		return len(m.snapshot.Applications)
	case LevelConnections:
		view := m.CurrentView()
		if view == nil {
			return 0
		}
		for _, app := range m.snapshot.Applications {
			if app.Name == view.ProcessName {
				return len(app.Connections)
			}
		}
		return 0
	case LevelAllConnections:
		return m.snapshot.TotalConnections()
	default:
		return 0
	}
}

// columnsForLevel returns the SortColumn IDs for the given view level.
func (m Model) columnsForLevel(level ViewLevel) []SortColumn {
	var cols []columnDef
	switch level {
	case LevelProcessList:
		cols = processListColumns()
	case LevelConnections:
		cols = connectionsColumns()
	case LevelAllConnections:
		cols = allConnectionsColumns()
	default:
		return nil
	}
	result := make([]SortColumn, len(cols))
	for i, col := range cols {
		result[i] = col.id
	}
	return result
}

// findColumnIndex returns the index of the given SortColumn in the columns slice.
// Returns 0 if not found.
func (m Model) findColumnIndex(columns []SortColumn, col SortColumn) int {
	for i, c := range columns {
		if c == col {
			return i
		}
	}
	return 0
}

// clampCursor ensures cursor is within bounds after filter changes.
func (m *Model) clampCursor() {
	view := m.CurrentView()
	if view == nil {
		return
	}
	max := m.filteredCount()
	if max == 0 {
		view.Cursor = 0
	} else if view.Cursor >= max {
		view.Cursor = max - 1
	}
}

// filteredCount returns the number of items after filtering for current view.
func (m *Model) filteredCount() int {
	if m.snapshot == nil {
		return 0
	}
	filter := m.currentFilter()
	if filter == "" {
		return m.maxCursorForLevel(m.CurrentView().Level)
	}

	view := m.CurrentView()
	switch view.Level {
	case LevelProcessList:
		return len(m.filteredApps())
	case LevelConnections:
		for _, app := range m.snapshot.Applications {
			if app.Name == view.ProcessName {
				return len(m.filteredConnections(app.Connections))
			}
		}
		return 0
	case LevelAllConnections:
		return len(m.filteredAllConnections())
	default:
		return m.maxCursorForLevel(view.Level)
	}
}

// drillIntoPID finds the application containing the given PID and pushes to its connections view.
func drillIntoPID(m *Model, pid int32) {
	if m.snapshot == nil {
		return
	}
	for _, app := range m.snapshot.Applications {
		for _, p := range app.PIDs {
			if p == pid {
				m.PushView(ViewState{
					Level:          LevelConnections,
					ProcessName:    app.Name,
					Cursor:         0,
					SortColumn:     SortLocal,
					SortAscending:  true,
					SelectedColumn: SortLocal,
				})
				return
			}
		}
	}
}

// extractPortsFromAddrs parses port numbers from address strings like "127.0.0.1:8080".
func extractPortsFromAddrs(addrs ...string) []int {
	var ports []int
	for _, addr := range addrs {
		if idx := strings.LastIndex(addr, ":"); idx != -1 {
			if port, err := strconv.Atoi(addr[idx+1:]); err == nil {
				ports = append(ports, port)
			}
		}
	}
	return ports
}
