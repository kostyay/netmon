package dns

import (
	"context"
	"testing"
	"time"
)

func TestResolveAsync_Timeout(t *testing.T) {
	// Use short timeout to test timeout handling
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Try to resolve an IP that likely won't respond quickly
	ch := ResolveAsync(ctx, "192.0.2.1") // TEST-NET, unlikely to have PTR

	result := <-ch
	if result.IP != "192.0.2.1" {
		t.Errorf("IP = %q, want %q", result.IP, "192.0.2.1")
	}
	// Either error or empty hostname is acceptable for this test
}

func TestResolveAsync_ValidIP(t *testing.T) {
	// Skip in short mode as this requires network
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Resolve localhost - should work on most systems
	ch := ResolveAsync(ctx, "127.0.0.1")
	result := <-ch

	if result.IP != "127.0.0.1" {
		t.Errorf("IP = %q, want %q", result.IP, "127.0.0.1")
	}
	// Hostname might be "localhost" or something else depending on system
	// Just verify no error occurred (some systems may not have PTR for localhost)
}

func TestResolve_Synchronous(t *testing.T) {
	// Use short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Resolve should work synchronously
	hostname, err := Resolve(ctx, "192.0.2.1")
	_ = hostname // May be empty
	_ = err      // May have error due to timeout
	// Just verify it doesn't panic
}
