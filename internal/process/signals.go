package process

import "syscall"

// SignalMap maps signal names to syscall.Signal values.
// Supports both full names (SIGTERM) and short names (TERM).
var SignalMap = map[string]syscall.Signal{
	"SIGTERM": syscall.SIGTERM,
	"SIGKILL": syscall.SIGKILL,
	"SIGHUP":  syscall.SIGHUP,
	"SIGINT":  syscall.SIGINT,
	"SIGQUIT": syscall.SIGQUIT,
	"TERM":    syscall.SIGTERM,
	"KILL":    syscall.SIGKILL,
	"HUP":     syscall.SIGHUP,
	"INT":     syscall.SIGINT,
	"QUIT":    syscall.SIGQUIT,
	"9":       syscall.SIGKILL,
	"15":      syscall.SIGTERM,
}
