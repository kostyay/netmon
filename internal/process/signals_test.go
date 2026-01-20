package process

import (
	"syscall"
	"testing"
)

func TestSignalMap_ContainsCommonSignals(t *testing.T) {
	tests := []struct {
		name string
		want syscall.Signal
	}{
		{"SIGTERM", syscall.SIGTERM},
		{"SIGKILL", syscall.SIGKILL},
		{"SIGHUP", syscall.SIGHUP},
		{"SIGINT", syscall.SIGINT},
		{"SIGQUIT", syscall.SIGQUIT},
	}

	for _, tt := range tests {
		got, ok := SignalMap[tt.name]
		if !ok {
			t.Errorf("SignalMap missing %s", tt.name)
			continue
		}
		if got != tt.want {
			t.Errorf("SignalMap[%s] = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSignalMap_ShortNames(t *testing.T) {
	tests := []struct {
		name string
		want syscall.Signal
	}{
		{"TERM", syscall.SIGTERM},
		{"KILL", syscall.SIGKILL},
		{"HUP", syscall.SIGHUP},
		{"INT", syscall.SIGINT},
		{"QUIT", syscall.SIGQUIT},
	}

	for _, tt := range tests {
		got, ok := SignalMap[tt.name]
		if !ok {
			t.Errorf("SignalMap missing short name %s", tt.name)
			continue
		}
		if got != tt.want {
			t.Errorf("SignalMap[%s] = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSignalMap_NumericSignals(t *testing.T) {
	tests := []struct {
		name string
		want syscall.Signal
	}{
		{"9", syscall.SIGKILL},
		{"15", syscall.SIGTERM},
	}

	for _, tt := range tests {
		got, ok := SignalMap[tt.name]
		if !ok {
			t.Errorf("SignalMap missing numeric signal %s", tt.name)
			continue
		}
		if got != tt.want {
			t.Errorf("SignalMap[%s] = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSignalMap_InvalidSignal(t *testing.T) {
	invalidSignals := []string{
		"INVALID",
		"sigterm", // lowercase not supported
		"SIG",
		"99",
		"",
	}

	for _, name := range invalidSignals {
		_, ok := SignalMap[name]
		if ok {
			t.Errorf("SignalMap should not contain %q", name)
		}
	}
}

func TestSignalMap_ConsistentMapping(t *testing.T) {
	// Verify that long names and short names map to the same signal
	pairs := [][2]string{
		{"SIGTERM", "TERM"},
		{"SIGKILL", "KILL"},
		{"SIGHUP", "HUP"},
		{"SIGINT", "INT"},
		{"SIGQUIT", "QUIT"},
	}

	for _, pair := range pairs {
		longSig := SignalMap[pair[0]]
		shortSig := SignalMap[pair[1]]
		if longSig != shortSig {
			t.Errorf("%s != %s (%v != %v)", pair[0], pair[1], longSig, shortSig)
		}
	}
}
