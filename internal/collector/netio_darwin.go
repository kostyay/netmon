//go:build darwin

package collector

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kostyay/netmon/internal/model"
)

// NetIOCollector collects network I/O statistics asynchronously.
type NetIOCollector struct{}

// NewNetIOCollector creates a new network I/O collector.
func NewNetIOCollector() *NetIOCollector {
	return &NetIOCollector{}
}

// Collect gathers network I/O stats for all processes.
// Uses nettop command on macOS which provides per-process network stats.
func (c *NetIOCollector) Collect(ctx context.Context) (map[int32]*model.NetIOStats, error) {
	// Run nettop with one sample (-l 1) and parseable output (-P)
	// nettop -P -l 1 gives us: time,interface,state,bytes_in,bytes_out,...
	cmd := exec.CommandContext(ctx, "nettop", "-P", "-l", "1", "-x", "-J", "bytes_in,bytes_out")
	output, err := cmd.Output()
	if err != nil {
		// nettop might not be available or might require elevated privileges
		// Return empty map on failure (graceful fallback)
		return make(map[int32]*model.NetIOStats), nil
	}

	return parseNettopOutput(string(output))
}

// parseNettopOutput parses the output of nettop command.
// Format: time,interface,state,bytes_in,bytes_out,process.pid
func parseNettopOutput(output string) (map[int32]*model.NetIOStats, error) {
	stats := make(map[int32]*model.NetIOStats)
	scanner := bufio.NewScanner(strings.NewReader(output))
	now := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		// Skip header line
		if strings.HasPrefix(line, "time") {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}

		// Parse the line - format varies, look for bytes_in, bytes_out, and PID
		// The format from -J is: process.pid,bytes_in,bytes_out
		var pid int32
		var bytesIn, bytesOut uint64

		// Try to parse each field
		for i, field := range fields {
			field = strings.TrimSpace(field)

			// Try to identify what this field is
			if strings.Contains(field, ".") {
				// Might be process.pid format
				parts := strings.Split(field, ".")
				if len(parts) >= 2 {
					if p, err := strconv.ParseInt(parts[len(parts)-1], 10, 32); err == nil {
						pid = int32(p)
					}
				}
			} else if i > 0 {
				// Numeric fields after the first are bytes_in and bytes_out
				if val, err := strconv.ParseUint(field, 10, 64); err == nil {
					if bytesIn == 0 {
						bytesIn = val
					} else if bytesOut == 0 {
						bytesOut = val
					}
				}
			}
		}

		if pid > 0 {
			existing, ok := stats[pid]
			if ok {
				existing.BytesRecv += bytesIn
				existing.BytesSent += bytesOut
			} else {
				stats[pid] = &model.NetIOStats{
					BytesRecv: bytesIn,
					BytesSent: bytesOut,
					UpdatedAt: now,
				}
			}
		}
	}

	return stats, nil
}
