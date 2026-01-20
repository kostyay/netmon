package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kostyay/netmon/internal/model"
	"github.com/kostyay/netmon/internal/services"
)

// formatPIDList formats a slice of PIDs for display.
func formatPIDList(pids []int32) string {
	if len(pids) == 0 {
		return "-"
	}
	if len(pids) == 1 {
		return fmt.Sprintf("%d", pids[0])
	}
	if len(pids) <= 3 {
		strs := make([]string, len(pids))
		for i, p := range pids {
			strs[i] = fmt.Sprintf("%d", p)
		}
		return strings.Join(strs, ", ")
	}
	return fmt.Sprintf("%d, %d +%d more", pids[0], pids[1], len(pids)-2)
}

// truncateString truncates a string to maxLen with ellipsis if needed.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// truncateAddr is an alias for truncateString (kept for readability at call sites).
var truncateAddr = truncateString

// formatBytes formats bytes into human-readable units.
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// formatBytesOrDash formats bytes or returns '--' if nil.
func formatBytesOrDash(stats *model.NetIOStats, isSent bool) string {
	if stats == nil {
		return "--"
	}
	if isSent {
		return formatBytes(stats.BytesSent)
	}
	return formatBytes(stats.BytesRecv)
}

// formatAddr formats an address with optional DNS resolution and service name substitution.
// protocol should be "tcp" or "udp" for correct service name lookup.
// Pass nil for dnsCache to skip hostname resolution.
func formatAddr(addr string, protocol string, serviceNames bool, dnsCache ...map[string]string) string {
	if addr == "" || addr == "*" {
		return addr
	}

	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return addr
	}
	ip := addr[:idx]
	port := addr[idx+1:]

	// Replace port with service name if enabled
	if serviceNames {
		if portNum, err := strconv.Atoi(port); err == nil {
			if name := services.Lookup(portNum, strings.ToLower(protocol)); name != "" {
				port = name
			}
		}
	}

	// Check DNS cache if provided
	if len(dnsCache) > 0 && dnsCache[0] != nil {
		if hostname, ok := dnsCache[0][ip]; ok && hostname != "" {
			return hostname + ":" + port
		}
	}

	return ip + ":" + port
}

// formatRemoteAddr formats a remote address with DNS resolution and service names.
func formatRemoteAddr(addr string, protocol string, dnsCache map[string]string, serviceNames bool) string {
	return formatAddr(addr, protocol, serviceNames, dnsCache)
}
