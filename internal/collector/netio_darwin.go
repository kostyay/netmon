//go:build darwin

package collector

import (
	"bufio"
	"context"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kostyay/netmon/internal/model"
)

// NetIOCollector collects network I/O statistics using nettop.
type NetIOCollector struct{}

// NewNetIOCollector creates a new network I/O collector.
func NewNetIOCollector() *NetIOCollector {
	return &NetIOCollector{}
}

// Collect gathers network I/O stats for all processes using nettop.
// nettop uses the private NetworkStatistics.framework internally.
func (c *NetIOCollector) Collect(ctx context.Context) (map[int32]*model.NetIOStats, error) {
	// Run nettop with one sample (-l 1), extended output (-x), specific columns (-J)
	// Output format: "process.pid                    bytes_in     bytes_out"
	cmd := exec.CommandContext(ctx, "nettop", "-P", "-l", "1", "-x", "-J", "bytes_in,bytes_out")
	output, err := cmd.Output()
	if err != nil {
		// nettop might not be available or might require elevated privileges
		return make(map[int32]*model.NetIOStats), nil
	}

	return parseNettopOutput(string(output))
}

// pidRegex matches "process_name.PID" at the start of a line
var pidRegex = regexp.MustCompile(`\.(\d+)\s+`)

// parseNettopOutput parses nettop output.
// Format: "process_name.pid                    bytes_in     bytes_out"
func parseNettopOutput(output string) (map[int32]*model.NetIOStats, error) {
	stats := make(map[int32]*model.NetIOStats)
	scanner := bufio.NewScanner(strings.NewReader(output))
	now := time.Now()

	for scanner.Scan() {
		line := scanner.Text()

		// Skip header line
		if strings.HasPrefix(line, "bytes_in") || strings.HasPrefix(line, " ") {
			continue
		}

		// Extract PID from "process_name.PID" format
		matches := pidRegex.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}

		pid64, err := strconv.ParseInt(matches[1], 10, 32)
		if err != nil {
			continue
		}
		pid := int32(pid64)

		// Split remaining part to get bytes_in and bytes_out
		// The line format is: "name.pid     bytes_in     bytes_out"
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// Last two fields are bytes_in and bytes_out
		bytesIn, err1 := strconv.ParseUint(fields[len(fields)-2], 10, 64)
		bytesOut, err2 := strconv.ParseUint(fields[len(fields)-1], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}

		// Aggregate stats for same PID (process may have multiple entries)
		if existing, ok := stats[pid]; ok {
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

	return stats, nil
}
