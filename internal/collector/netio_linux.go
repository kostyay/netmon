//go:build linux

package collector

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kostyay/netmon/internal/model"
	"github.com/shirou/gopsutil/v3/process"
)

// netIOCollector collects network I/O statistics from /proc.
type netIOCollector struct{}

// NewNetIOCollector creates a new network I/O collector.
func NewNetIOCollector() NetIOCollector {
	return &netIOCollector{}
}

// Collect gathers network I/O stats for all processes using /proc/[pid]/net/dev.
func (c *netIOCollector) Collect(ctx context.Context) (map[int32]*model.NetIOStats, error) {
	stats := make(map[int32]*model.NetIOStats)
	now := time.Now()

	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return stats, nil
	}

	for _, p := range procs {
		if ctx.Err() != nil {
			break
		}

		bytesRecv, bytesSent := readProcNetDev(p.Pid)
		if bytesRecv > 0 || bytesSent > 0 {
			stats[p.Pid] = &model.NetIOStats{
				BytesRecv: bytesRecv,
				BytesSent: bytesSent,
				UpdatedAt: now,
			}
		}
	}

	return stats, nil
}

// readProcNetDev reads network stats from /proc/[pid]/net/dev.
func readProcNetDev(pid int32) (uint64, uint64) {
	path := "/proc/" + strconv.Itoa(int(pid)) + "/net/dev"
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	var totalRecv, totalSent uint64
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue // Skip header lines
		}

		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue // Skip loopback
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}

		recv, _ := strconv.ParseUint(fields[0], 10, 64)
		sent, _ := strconv.ParseUint(fields[8], 10, 64)
		totalRecv += recv
		totalSent += sent
	}

	return totalRecv, totalSent
}
