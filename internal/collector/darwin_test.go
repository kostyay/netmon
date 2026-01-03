//go:build darwin

package collector

import (
	"testing"
)

func TestFormatAddr_WithIP(t *testing.T) {
	got := formatAddr("127.0.0.1", 8080)
	want := "127.0.0.1:8080"
	if got != want {
		t.Errorf("formatAddr() = %q, want %q", got, want)
	}
}

func TestFormatAddr_EmptyIP(t *testing.T) {
	got := formatAddr("", 8080)
	want := "*:8080"
	if got != want {
		t.Errorf("formatAddr() = %q, want %q", got, want)
	}
}

func TestFormatAddr_ZeroPort(t *testing.T) {
	got := formatAddr("192.168.1.1", 0)
	want := "192.168.1.1:0"
	if got != want {
		t.Errorf("formatAddr() = %q, want %q", got, want)
	}
}

func TestFormatAddr_IPv6(t *testing.T) {
	got := formatAddr("::1", 443)
	want := "::1:443"
	if got != want {
		t.Errorf("formatAddr() = %q, want %q", got, want)
	}
}

func TestContainsPID_Found(t *testing.T) {
	pids := []int32{100, 200, 300}
	if !containsPID(pids, 200) {
		t.Errorf("containsPID() = false, want true for existing PID")
	}
}

func TestContainsPID_NotFound(t *testing.T) {
	pids := []int32{100, 200, 300}
	if containsPID(pids, 400) {
		t.Errorf("containsPID() = true, want false for non-existing PID")
	}
}

func TestContainsPID_EmptySlice(t *testing.T) {
	pids := []int32{}
	if containsPID(pids, 100) {
		t.Errorf("containsPID() = true, want false for empty slice")
	}
}

func TestContainsPID_FirstElement(t *testing.T) {
	pids := []int32{100, 200, 300}
	if !containsPID(pids, 100) {
		t.Errorf("containsPID() = false, want true for first element")
	}
}

func TestContainsPID_LastElement(t *testing.T) {
	pids := []int32{100, 200, 300}
	if !containsPID(pids, 300) {
		t.Errorf("containsPID() = false, want true for last element")
	}
}

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

func TestNewPlatformCollector(t *testing.T) {
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

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Error("New() returned nil")
	}
}
