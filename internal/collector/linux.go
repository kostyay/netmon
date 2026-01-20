//go:build linux

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

type linuxCollector struct {
	processCache map[int32]processInfo
	cacheMu      sync.RWMutex
}

func newPlatformCollector() Collector {
	return &linuxCollector{
		processCache: make(map[int32]processInfo),
	}
}

func (c *linuxCollector) Collect(ctx context.Context) (*model.NetworkSnapshot, error) {
	c.cacheMu.Lock()
	c.processCache = make(map[int32]processInfo)
	c.cacheMu.Unlock()

	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	appMap := make(map[string]*model.Application)
	skippedCount := 0

	for _, conn := range connections {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if conn.Pid == 0 {
			continue
		}

		info := c.getProcessInfo(ctx, conn.Pid)
		if info.name == "" {
			skippedCount++
			continue
		}

		app, exists := appMap[info.name]
		if !exists {
			app = &model.Application{
				Name: info.name,
				Exe:  info.exe,
			}
			appMap[info.name] = app
		}

		if !containsPID(app.PIDs, conn.Pid) {
			app.PIDs = append(app.PIDs, conn.Pid)
		}

		mc := model.Connection{
			PID:        conn.Pid,
			Protocol:   c.getProtocol(conn.Type),
			LocalAddr:  formatAddr(conn.Laddr.IP, conn.Laddr.Port),
			RemoteAddr: c.formatRemoteAddr(conn),
			State:      c.getState(conn),
		}
		app.Connections = append(app.Connections, mc)
	}

	apps := make([]model.Application, 0, len(appMap))
	for _, app := range appMap {
		sort.Slice(app.PIDs, func(i, j int) bool {
			return app.PIDs[i] < app.PIDs[j]
		})
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

func (c *linuxCollector) getProcessInfo(ctx context.Context, pid int32) processInfo {
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

	exe, _ := proc.ExeWithContext(ctx)

	info := processInfo{name: name, exe: exe}

	c.cacheMu.Lock()
	c.processCache[pid] = info
	c.cacheMu.Unlock()

	return info
}

func (c *linuxCollector) getProtocol(connType uint32) model.Protocol {
	switch connType {
	case 1:
		return model.ProtocolTCP
	case 2:
		return model.ProtocolUDP
	default:
		return model.ProtocolUnknown
	}
}

func (c *linuxCollector) formatRemoteAddr(conn net.ConnectionStat) string {
	if conn.Raddr.IP == "" || conn.Raddr.Port == 0 {
		return "*"
	}
	return formatAddr(conn.Raddr.IP, conn.Raddr.Port)
}

func (c *linuxCollector) getState(conn net.ConnectionStat) model.ConnectionState {
	if conn.Status == "" {
		return model.StateNone
	}
	return model.ConnectionState(conn.Status)
}

