//go:build linux

package collector

import (
	"testing"
)

func TestLinuxCollector_GetProtocol_TCP(t *testing.T) {
	c := &linuxCollector{}
	got := c.getProtocol(1)
	if got != "TCP" {
		t.Errorf("getProtocol(1) = %q, want TCP", got)
	}
}

func TestLinuxCollector_GetProtocol_UDP(t *testing.T) {
	c := &linuxCollector{}
	got := c.getProtocol(2)
	if got != "UDP" {
		t.Errorf("getProtocol(2) = %q, want UDP", got)
	}
}

func TestLinuxCollector_GetProtocol_Unknown(t *testing.T) {
	c := &linuxCollector{}
	got := c.getProtocol(99)
	if got != "UNK" {
		t.Errorf("getProtocol(99) = %q, want UNK", got)
	}
}

func TestLinuxCollector_GetProtocol_Zero(t *testing.T) {
	c := &linuxCollector{}
	got := c.getProtocol(0)
	if got != "UNK" {
		t.Errorf("getProtocol(0) = %q, want UNK", got)
	}
}

func TestNewPlatformCollector_Linux(t *testing.T) {
	c := newPlatformCollector()
	if c == nil {
		t.Error("newPlatformCollector() returned nil")
	}

	lc, ok := c.(*linuxCollector)
	if !ok {
		t.Error("newPlatformCollector() did not return *linuxCollector")
	}

	if lc.processCache == nil {
		t.Error("linuxCollector.processCache is nil")
	}
}
