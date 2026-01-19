package ui

import (
	"github.com/kostyay/netmon/internal/model"
)

// findProcessIndex returns the index of a process by name in the filtered apps list.
// Returns -1 if not found.
func (m Model) findProcessIndex(name string) int {
	if name == "" {
		return -1
	}
	apps := m.filteredApps()
	apps = m.sortProcessList(apps)
	for i, app := range apps {
		if app.Name == name {
			return i
		}
	}
	return -1
}

// findConnectionIndex returns the index of a connection by key in the current connection list.
// Returns -1 if not found.
func (m Model) findConnectionIndex(key *model.ConnectionKey) int {
	if key == nil {
		return -1
	}

	view := m.CurrentView()
	if view == nil {
		return -1
	}

	switch view.Level {
	case LevelConnections:
		// Find the process and get its connections
		for i := range m.snapshot.Applications {
			if m.snapshot.Applications[i].Name == view.ProcessName {
				conns := m.filteredConnections(m.snapshot.Applications[i].Connections)
				conns = m.sortConnectionsForView(conns)
				for j, conn := range conns {
					if m.connectionMatchesKey(conn, key) {
						return j
					}
				}
				break
			}
		}
	case LevelAllConnections:
		conns := m.sortAllConnections(m.filteredAllConnections())
		for i, cwp := range conns {
			if cwp.ProcessName == key.ProcessName &&
				cwp.LocalAddr == key.LocalAddr &&
				cwp.RemoteAddr == key.RemoteAddr {
				return i
			}
		}
	}
	return -1
}

// connectionMatchesKey checks if a connection matches a ConnectionKey.
func (m Model) connectionMatchesKey(conn model.Connection, key *model.ConnectionKey) bool {
	if key == nil {
		return false
	}
	// For process match, we need to find the process name from the connection's PID
	processName := m.getProcessNameByPID(conn.PID)
	return processName == key.ProcessName &&
		conn.LocalAddr == key.LocalAddr &&
		conn.RemoteAddr == key.RemoteAddr
}

// getProcessNameByPID finds the process name for a given PID.
func (m Model) getProcessNameByPID(pid int32) string {
	if m.snapshot == nil {
		return ""
	}
	for _, app := range m.snapshot.Applications {
		for _, p := range app.PIDs {
			if p == pid {
				return app.Name
			}
		}
	}
	return ""
}

// resolveSelectionIndex resolves the current SelectedID to a cursor index.
// Falls back to Cursor if SelectedID is empty or not found.
func (m Model) resolveSelectionIndex() int {
	view := m.CurrentView()
	if view == nil {
		return 0
	}

	// Check if we have a valid SelectedID
	if view.SelectedID.ProcessName == "" {
		return view.Cursor
	}

	switch view.Level {
	case LevelProcessList:
		idx := m.findProcessIndex(view.SelectedID.ProcessName)
		if idx >= 0 {
			return idx
		}
	case LevelConnections, LevelAllConnections:
		idx := m.findConnectionIndex(view.SelectedID.ConnectionKey)
		if idx >= 0 {
			return idx
		}
	}

	return view.Cursor
}

// validateSelection clamps the cursor and updates SelectedID if the selected item is gone.
func (m *Model) validateSelection() {
	view := m.CurrentView()
	if view == nil || m.snapshot == nil {
		return
	}

	var itemCount int
	switch view.Level {
	case LevelProcessList:
		apps := m.filteredApps()
		itemCount = len(apps)
	case LevelConnections:
		for i := range m.snapshot.Applications {
			if m.snapshot.Applications[i].Name == view.ProcessName {
				conns := m.filteredConnections(m.snapshot.Applications[i].Connections)
				itemCount = len(conns)
				break
			}
		}
	case LevelAllConnections:
		itemCount = len(m.filteredAllConnections())
	}

	if itemCount == 0 {
		view.Cursor = 0
		view.SelectedID = model.SelectionID{}
		return
	}

	// For connection views, just clamp cursor - don't try ID resolution
	// (connections can have duplicate keys, so ID tracking doesn't work)
	if view.Level == LevelConnections || view.Level == LevelAllConnections {
		if view.Cursor >= itemCount {
			view.Cursor = itemCount - 1
		}
		if view.Cursor < 0 {
			view.Cursor = 0
		}
		return
	}

	// For process list, try to resolve SelectedID to follow item across re-sorts
	if view.SelectedID.ProcessName != "" {
		resolvedIdx := m.resolveSelectionIndex()
		if resolvedIdx >= 0 && resolvedIdx < itemCount {
			view.Cursor = resolvedIdx
			return
		}
	}

	// SelectedID not found or empty, clamp cursor
	if view.Cursor >= itemCount {
		view.Cursor = itemCount - 1
	}
	if view.Cursor < 0 {
		view.Cursor = 0
	}

	// Update SelectedID based on current cursor position
	m.updateSelectedIDFromCursor()
}

// updateSelectedIDFromCursor sets SelectedID based on the current cursor position.
func (m *Model) updateSelectedIDFromCursor() {
	view := m.CurrentView()
	if view == nil || m.snapshot == nil {
		return
	}

	switch view.Level {
	case LevelProcessList:
		apps := m.filteredApps()
		apps = m.sortProcessList(apps)
		if view.Cursor >= 0 && view.Cursor < len(apps) {
			view.SelectedID = model.SelectionIDFromProcess(apps[view.Cursor].Name)
		}
	case LevelConnections:
		for i := range m.snapshot.Applications {
			if m.snapshot.Applications[i].Name == view.ProcessName {
				conns := m.filteredConnections(m.snapshot.Applications[i].Connections)
				conns = m.sortConnectionsForView(conns)
				if view.Cursor >= 0 && view.Cursor < len(conns) {
					conn := conns[view.Cursor]
					processName := m.getProcessNameByPID(conn.PID)
					view.SelectedID = model.SelectionIDFromConnection(processName, conn.LocalAddr, conn.RemoteAddr)
				}
				break
			}
		}
	case LevelAllConnections:
		conns := m.sortAllConnections(m.filteredAllConnections())
		if view.Cursor >= 0 && view.Cursor < len(conns) {
			cwp := conns[view.Cursor]
			view.SelectedID = model.SelectionIDFromConnection(cwp.ProcessName, cwp.LocalAddr, cwp.RemoteAddr)
		}
	}
}
