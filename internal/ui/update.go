package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/model"
)

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.fetchData(),
		m.fetchNetIO(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate viewport height: total - header - footer - frame borders
		viewportHeight := msg.Height - headerHeight - footerHeight - frameHeight
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
		// Kill mode intercepts all keys
		if m.killMode {
			switch msg.String() {
			case "y", "Y":
				return m.executeKill()
			case "n", "N", "esc":
				m.killMode = false
				m.killTarget = nil
				return m, nil
			}
			return m, nil // Ignore other keys in kill mode
		}

		// Search mode intercepts all keys
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.activeFilter = m.searchQuery
				m.searchMode = false
				m.clampCursor()
				return m, nil
			case "esc":
				m.searchQuery = m.activeFilter // revert to confirmed
				m.searchMode = false
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
				return m, nil
			default:
				// Append printable characters
				r := msg.Runes
				if len(r) == 1 && r[0] >= 32 {
					m.searchQuery += string(r)
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			view := m.CurrentView()
			if view != nil && view.Cursor > 0 {
				view.Cursor--
			}
			return m, nil

		case "down", "j":
			view := m.CurrentView()
			if view == nil || m.snapshot == nil {
				return m, nil
			}
			maxCursor := m.maxCursorForLevel(view.Level)
			if view.Cursor < maxCursor-1 {
				view.Cursor++
			}
			return m, nil

		case "left", "h":
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

		case "right", "l":
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

		case "enter", " ":
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
				if view.Cursor < len(m.snapshot.Applications) {
					app := m.snapshot.Applications[view.Cursor]
					m.PushView(ViewState{
						Level:          LevelConnections,
						ProcessName:    app.Name,
						Cursor:         0,
						SortColumn:     SortLocal,
						SortAscending:  true,
						SelectedColumn: SortLocal,
					})
				}
			}
			return m, nil

		case "esc", "backspace":
			view := m.CurrentView()
			if view != nil && view.SortMode {
				// Exit sort mode without changing sort
				view.SortMode = false
				return m, nil
			}
			// Pop view (go back)
			m.PopView()
			return m, nil

		case "+", "=":
			// Decrease refresh interval (faster refresh)
			if m.refreshInterval > MinRefreshInterval {
				m.refreshInterval -= RefreshStep
			}
			return m, nil

		case "-", "_":
			// Increase refresh interval (slower refresh)
			if m.refreshInterval < MaxRefreshInterval {
				m.refreshInterval += RefreshStep
			}
			return m, nil

		case "s":
			// Enter sort mode
			view := m.CurrentView()
			if view == nil || view.SortMode {
				return m, nil
			}
			view.SortMode = true
			view.SelectedColumn = view.SortColumn // Start at current sort column
			return m, nil

		case "v":
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

		case "/":
			// Enter search mode
			m.searchMode = true
			m.searchQuery = m.activeFilter // pre-fill with current filter
			return m, nil

		case "x":
			// Kill with SIGTERM
			return m.enterKillMode("SIGTERM")

		case "X":
			// Kill with SIGKILL
			return m.enterKillMode("SIGKILL")

		default:
			// Pass unhandled keys to viewport for page up/down, mouse scroll, etc.
			if m.ready {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}

	case TickMsg:
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
		m.snapshot = msg.Snapshot
		// Ensure cursor is valid for current view level
		view := m.CurrentView()
		if view != nil && m.snapshot != nil {
			maxCursor := m.maxCursorForLevel(view.Level)
			if maxCursor > 0 && view.Cursor >= maxCursor {
				view.Cursor = maxCursor - 1
			} else if maxCursor == 0 {
				view.Cursor = 0
			}
		}
		return m, nil

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
	}

	return m, nil
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
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

// maxColumnForLevel returns the number of columns for the given view level.
func (m Model) maxColumnForLevel(level ViewLevel) int {
	switch level {
	case LevelProcessList:
		return len(processListColumns())
	case LevelConnections:
		return len(connectionsColumns())
	case LevelAllConnections:
		return len(allConnectionsColumns())
	default:
		return 1
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
	filter := m.activeFilter
	if m.searchMode {
		filter = m.searchQuery
	}
	if filter == "" {
		return m.maxCursorForLevel(m.CurrentView().Level)
	}

	view := m.CurrentView()
	switch view.Level {
	case LevelProcessList:
		count := 0
		for _, app := range m.snapshot.Applications {
			ports := extractPorts(app.Connections)
			if matchesFilter(filter, app.Name, app.PIDs, ports) {
				count++
			}
		}
		return count
	case LevelAllConnections:
		count := 0
		for _, app := range m.snapshot.Applications {
			for _, conn := range app.Connections {
				ports := extractPortsFromAddrs(conn.LocalAddr, conn.RemoteAddr)
				if matchesFilter(filter, app.Name, app.PIDs, ports) {
					count++
				}
			}
		}
		return count
	default:
		return m.maxCursorForLevel(view.Level)
	}
}

// extractPorts gets all ports from connections for filter matching.
func extractPorts(conns []model.Connection) []int {
	var ports []int
	for _, c := range conns {
		ports = append(ports, extractPortsFromAddrs(c.LocalAddr, c.RemoteAddr)...)
	}
	return ports
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

// enterKillMode sets up kill mode with the currently selected target.
func (m Model) enterKillMode(signal string) (tea.Model, tea.Cmd) {
	if m.snapshot == nil {
		return m, nil
	}
	view := m.CurrentView()
	if view == nil {
		return m, nil
	}

	var target *killTargetInfo

	switch view.Level {
	case LevelProcessList:
		// Get PID from selected process
		apps := m.filteredApps()
		if view.Cursor >= len(apps) {
			return m, nil
		}
		app := apps[view.Cursor]
		if len(app.PIDs) == 0 {
			return m, nil
		}
		target = &killTargetInfo{
			PID:         app.PIDs[0], // Use first PID
			ProcessName: app.Name,
			Signal:      signal,
		}

	case LevelConnections:
		// Get PID from selected connection within the process
		for _, app := range m.snapshot.Applications {
			if app.Name != view.ProcessName {
				continue
			}
			if view.Cursor >= len(app.Connections) {
				return m, nil
			}
			conn := app.Connections[view.Cursor]
			port := extractSinglePort(conn.LocalAddr)
			target = &killTargetInfo{
				PID:         conn.PID,
				ProcessName: app.Name,
				Port:        port,
				Signal:      signal,
			}
			break
		}

	case LevelAllConnections:
		// Get PID from selected connection in flat view
		conns := m.filteredAllConnections()
		if view.Cursor >= len(conns) {
			return m, nil
		}
		conn := conns[view.Cursor]
		port := extractSinglePort(conn.LocalAddr)
		target = &killTargetInfo{
			PID:         conn.PID,
			ProcessName: conn.ProcessName,
			Port:        port,
			Signal:      signal,
		}
	}

	if target == nil {
		return m, nil
	}

	m.killMode = true
	m.killTarget = target
	return m, nil
}

// extractSinglePort extracts port from an address like "127.0.0.1:8080".
func extractSinglePort(addr string) int {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		if port, err := strconv.Atoi(addr[idx+1:]); err == nil {
			return port
		}
	}
	return 0
}

// signalMap maps signal names to syscall.Signal values.
var signalMap = map[string]syscall.Signal{
	"SIGTERM": syscall.SIGTERM,
	"SIGKILL": syscall.SIGKILL,
}

// executeKill sends the signal to the target process.
func (m Model) executeKill() (tea.Model, tea.Cmd) {
	if m.killTarget == nil {
		m.killMode = false
		return m, nil
	}

	sig, ok := signalMap[m.killTarget.Signal]
	if !ok {
		sig = syscall.SIGTERM
	}

	err := syscall.Kill(int(m.killTarget.PID), sig)
	m.killMode = false

	if err != nil {
		m.killResult = fmt.Sprintf("Failed to kill PID %d: %v", m.killTarget.PID, err)
	} else {
		m.killResult = fmt.Sprintf("Killed PID %d (%s)", m.killTarget.PID, m.killTarget.ProcessName)
	}
	m.killResultAt = time.Now()
	m.killTarget = nil

	return m, nil
}
