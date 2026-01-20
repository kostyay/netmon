package collector

import "fmt"

// formatAddr formats an IP address and port as "ip:port".
func formatAddr(ip string, port uint32) string {
	if ip == "" {
		ip = "*"
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

// containsPID checks if a PID is in the slice.
func containsPID(pids []int32, pid int32) bool {
	for _, p := range pids {
		if p == pid {
			return true
		}
	}
	return false
}
