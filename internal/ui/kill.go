package ui

import (
	"fmt"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kostyay/netmon/internal/model"
	"github.com/kostyay/netmon/internal/process"
)

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
	idx := m.resolveSelectionIndex()

	switch view.Level {
	case LevelProcessList:
		apps := m.sortProcessList(m.filteredApps())
		if idx >= len(apps) {
			return m, nil
		}
		app := apps[idx]
		if len(app.PIDs) == 0 {
			return m, nil
		}
		target = &killTargetInfo{
			PID:         app.PIDs[0],
			PIDs:        app.PIDs, // Store all PIDs for multi-process apps
			ProcessName: app.Name,
			Exe:         app.Exe,
			Signal:      signal,
		}

	case LevelConnections:
		for _, app := range m.snapshot.Applications {
			if app.Name != view.ProcessName {
				continue
			}
			conns := m.sortConnectionsForView(m.filteredConnections(app.Connections))
			if idx >= len(conns) {
				return m, nil
			}
			conn := conns[idx]
			target = &killTargetInfo{
				PID:         conn.PID,
				ProcessName: app.Name,
				Exe:         app.Exe,
				Port:        model.ExtractPort(conn.LocalAddr),
				Signal:      signal,
			}
			break
		}

	case LevelAllConnections:
		conns := m.sortAllConnections(m.filteredAllConnections())
		if idx >= len(conns) {
			return m, nil
		}
		conn := conns[idx]
		// Look up the Exe from the application
		var exe string
		for _, app := range m.snapshot.Applications {
			if app.Name == conn.ProcessName {
				exe = app.Exe
				break
			}
		}
		target = &killTargetInfo{
			PID:         conn.PID,
			ProcessName: conn.ProcessName,
			Exe:         exe,
			Port:        model.ExtractPort(conn.LocalAddr),
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

// executeKill sends the signal to the target process(es).
func (m Model) executeKill() (tea.Model, tea.Cmd) {
	if m.killTarget == nil {
		m.killMode = false
		return m, nil
	}

	sig, ok := process.SignalMap[m.killTarget.Signal]
	if !ok {
		sig = syscall.SIGTERM
	}

	// If we have multiple PIDs (from process list), kill all of them
	pidsToKill := m.killTarget.PIDs
	if len(pidsToKill) == 0 {
		// Fall back to single PID for connection-level kills
		pidsToKill = []int32{m.killTarget.PID}
	}

	var killed, failed int
	var lastErr error
	for _, pid := range pidsToKill {
		if err := syscall.Kill(int(pid), sig); err != nil {
			failed++
			lastErr = err
		} else {
			killed++
		}
	}

	m.killMode = false

	// Format result message
	if failed == 0 {
		if len(pidsToKill) == 1 {
			m.killResult = fmt.Sprintf("Killed PID %d (%s)", pidsToKill[0], m.killTarget.ProcessName)
		} else {
			m.killResult = fmt.Sprintf("Killed %d PIDs (%s)", killed, m.killTarget.ProcessName)
		}
	} else if killed == 0 {
		m.killResult = fmt.Sprintf("Failed to kill %s: %v", m.killTarget.ProcessName, lastErr)
	} else {
		m.killResult = fmt.Sprintf("Killed %d PIDs, %d failed (%s)", killed, failed, m.killTarget.ProcessName)
	}

	m.killResultAt = time.Now()
	m.killTarget = nil

	return m, nil
}
