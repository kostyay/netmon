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

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Error("New() returned nil")
	}
}
