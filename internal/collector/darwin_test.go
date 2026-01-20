//go:build darwin

package collector

import (
	"testing"
)

func TestDarwinCollector_GetProtocol_TCP(t *testing.T) {
	c := &darwinCollector{}
	got := c.getProtocol(1)
	if got != "TCP" {
		t.Errorf("getProtocol(1) = %q, want TCP", got)
	}
}

func TestDarwinCollector_GetProtocol_UDP(t *testing.T) {
	c := &darwinCollector{}
	got := c.getProtocol(2)
	if got != "UDP" {
		t.Errorf("getProtocol(2) = %q, want UDP", got)
	}
}

func TestDarwinCollector_GetProtocol_Unknown(t *testing.T) {
	c := &darwinCollector{}
	got := c.getProtocol(99)
	if got != "UNK" {
		t.Errorf("getProtocol(99) = %q, want UNK", got)
	}
}

func TestDarwinCollector_GetProtocol_Zero(t *testing.T) {
	c := &darwinCollector{}
	got := c.getProtocol(0)
	if got != "UNK" {
		t.Errorf("getProtocol(0) = %q, want UNK", got)
	}
}

func TestNewPlatformCollector_Darwin(t *testing.T) {
	c := newPlatformCollector()
	if c == nil {
		t.Error("newPlatformCollector() returned nil")
	}

	dc, ok := c.(*darwinCollector)
	if !ok {
		t.Error("newPlatformCollector() did not return *darwinCollector")
	}

	if dc.processCache == nil {
		t.Error("darwinCollector.processCache is nil")
	}
}
