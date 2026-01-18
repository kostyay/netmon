//go:build darwin

package collector

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/kostyay/netmon/internal/model"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// processInfo holds cached process name and executable path.
type processInfo struct {
	name string
	exe  string
}

type darwinCollector struct {
	processCache map[int32]processInfo
	cacheMu      sync.RWMutex
}

func newPlatformCollector() Collector {
	return &darwinCollector{
		processCache: make(map[int32]processInfo),
	}
}

func (c *darwinCollector) Collect(ctx context.Context) (*model.NetworkSnapshot, error) {
	// Clear process cache at the start of each cycle to prevent stale entries
	// when PIDs are reused by different processes
	c.cacheMu.Lock()
	c.processCache = make(map[int32]processInfo)
	c.cacheMu.Unlock()

	// Get all network connections (TCP and UDP)
	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	// Group connections by process name
	appMap := make(map[string]*model.Application)
	skippedCount := 0

	for _, conn := range connections {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Skip connections without PID (kernel/system)
		if conn.Pid == 0 {
			continue
		}

		// Get process info (with caching)
		info := c.getProcessInfo(ctx, conn.Pid)
		if info.name == "" {
			skippedCount++
			continue // Skip if we can't get process name
		}

		// Create or get application entry
		app, exists := appMap[info.name]
		if !exists {
			app = &model.Application{
				Name: info.name,
				Exe:  info.exe,
			}
			appMap[info.name] = app
		}

		// Add PID if not already present
		if !containsPID(app.PIDs, conn.Pid) {
			app.PIDs = append(app.PIDs, conn.Pid)
		}

		// Convert connection
		mc := model.Connection{
			PID:        conn.Pid,
			Protocol:   c.getProtocol(conn.Type),
			LocalAddr:  formatAddr(conn.Laddr.IP, conn.Laddr.Port),
			RemoteAddr: c.formatRemoteAddr(conn),
			State:      c.getState(conn),
		}
		app.Connections = append(app.Connections, mc)
	}

	// Convert map to slice, calculate state counts, and sort PIDs
	apps := make([]model.Application, 0, len(appMap))
	for _, app := range appMap {
		sort.Slice(app.PIDs, func(i, j int) bool {
			return app.PIDs[i] < app.PIDs[j]
		})
		// Calculate connection state counts
		for _, conn := range app.Connections {
			switch conn.State {
			case model.StateEstablished:
				app.EstablishedCount++
			case model.StateListen:
				app.ListenCount++
			}
		}
		apps = append(apps, *app)
	}

	snapshot := &model.NetworkSnapshot{
		Applications: apps,
		Timestamp:    time.Now(),
		SkippedCount: skippedCount,
	}
	snapshot.SortByConnectionCount()

	return snapshot, nil
}

func (c *darwinCollector) getProcessInfo(ctx context.Context, pid int32) processInfo {
	c.cacheMu.RLock()
	if info, ok := c.processCache[pid]; ok {
		c.cacheMu.RUnlock()
		return info
	}
	c.cacheMu.RUnlock()

	proc, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return processInfo{}
	}

	name, err := proc.NameWithContext(ctx)
	if err != nil {
		return processInfo{}
	}

	// Get executable path (may fail for some processes)
	exe, _ := proc.ExeWithContext(ctx)

	info := processInfo{name: name, exe: exe}

	c.cacheMu.Lock()
	c.processCache[pid] = info
	c.cacheMu.Unlock()

	return info
}

func (c *darwinCollector) getProtocol(connType uint32) model.Protocol {
	// net.SOCK_STREAM = 1 (TCP), net.SOCK_DGRAM = 2 (UDP)
	switch connType {
	case 1:
		return model.ProtocolTCP
	case 2:
		return model.ProtocolUDP
	default:
		return model.ProtocolUnknown
	}
}

func (c *darwinCollector) formatRemoteAddr(conn net.ConnectionStat) string {
	if conn.Raddr.IP == "" || conn.Raddr.Port == 0 {
		return "*"
	}
	return formatAddr(conn.Raddr.IP, conn.Raddr.Port)
}

func (c *darwinCollector) getState(conn net.ConnectionStat) model.ConnectionState {
	if conn.Status == "" {
		return model.StateNone // UDP has no state
	}
	return model.ConnectionState(conn.Status)
}

func formatAddr(ip string, port uint32) string {
	if ip == "" {
		ip = "*"
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

func containsPID(pids []int32, pid int32) bool {
	for _, p := range pids {
		if p == pid {
			return true
		}
	}
	return false
}
