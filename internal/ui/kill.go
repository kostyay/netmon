package ui

import (
	"context"
	"fmt"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kostyay/netmon/internal/docker"
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
		if idx < len(apps) {
			app := apps[idx]
			if len(app.PIDs) == 0 {
				return m, nil
			}
			target = &killTargetInfo{
				PID:         app.PIDs[0],
				PIDs:        app.PIDs,
				ProcessName: app.Name,
				Exe:         app.Exe,
				Signal:      signal,
			}
		} else {
			// Virtual container row
			vcs := m.filteredVirtualContainers()
			vcIdx := idx - len(apps)
			if vcIdx < 0 || vcIdx >= len(vcs) {
				return m, nil
			}
			vc := vcs[vcIdx]
			target = &killTargetInfo{
				ProcessName: containerDisplayName(vc),
				Exe:         vc.Info.Image,
				Signal:      signal,
				ContainerID: vc.Info.ID,
			}
		}

	case LevelConnections:
		selectedApp := m.findSelectedApp(view.ProcessName)
		if selectedApp == nil {
			return m, nil
		}
		conns := m.sortConnectionsForView(m.filteredConnections(selectedApp.Connections))
		if idx >= len(conns) {
			return m, nil
		}
		conn := conns[idx]
		target = &killTargetInfo{
			PID:         conn.PID,
			ProcessName: selectedApp.Name,
			Exe:         selectedApp.Exe,
			Port:        model.ExtractPort(conn.LocalAddr),
			Signal:      signal,
		}
		// If viewing a virtual container, set ContainerID for docker stop
		if vc := m.findVirtualContainer(view.ProcessName); vc != nil {
			target.ContainerID = vc.Info.ID
		}

	case LevelAllConnections:
		conns := m.sortAllConnections(m.filteredAllConnections())
		if idx >= len(conns) {
			return m, nil
		}
		conn := conns[idx]
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

// finishKill sets common fields after kill execution.
func (m *Model) finishKill() {
	m.killMode = false
	m.killResultAt = time.Now()
	m.killTarget = nil
}

// executeKill sends the signal to the target process(es) or stops a Docker container.
func (m Model) executeKill() (tea.Model, tea.Cmd) {
	if m.killTarget == nil {
		m.killMode = false
		return m, nil
	}

	// Docker container stop/kill
	if m.killTarget.ContainerID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		if m.killTarget.Signal == "SIGKILL" {
			err = docker.KillContainer(ctx, m.killTarget.ContainerID)
		} else {
			err = docker.StopContainer(ctx, m.killTarget.ContainerID, 10)
		}
		if err != nil {
			m.killResult = fmt.Sprintf("Failed to stop container %s: %v", m.killTarget.ContainerID, err)
		} else {
			m.killResult = fmt.Sprintf("Stopped container %s", m.killTarget.ContainerID)
		}
		m.finishKill()
		return m, nil
	}

	// Process kill via syscall
	sig, ok := process.SignalMap[m.killTarget.Signal]
	if !ok {
		sig = syscall.SIGTERM
	}

	pidsToKill := m.killTarget.PIDs
	if len(pidsToKill) == 0 {
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

	m.finishKill()
	return m, nil
}
