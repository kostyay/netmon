package collector

import (
	"context"
	"testing"
	"time"
)

func TestCollectorInterface_New(t *testing.T) {
	c := New()

	if c == nil {
		t.Error("New() should return a non-nil Collector")
	}
}

func TestCollectorInterface_Collect(t *testing.T) {
	c := New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	snapshot, err := c.Collect(ctx)

	// Should not error in normal operation
	if err != nil {
		t.Logf("Collect() returned error (may be expected in some environments): %v", err)
	}

	// Snapshot should be returned (may be empty if no network connections)
	if snapshot == nil && err == nil {
		t.Error("Collect() should return non-nil snapshot when err is nil")
	}

	if snapshot != nil {
		// Timestamp should be set
		if snapshot.Timestamp.IsZero() {
			t.Error("Snapshot.Timestamp should not be zero")
		}

		// Applications should be a valid slice (may be empty)
		if snapshot.Applications == nil {
			t.Error("Snapshot.Applications should not be nil")
		}
	}
}

func TestCollectorInterface_ContextCancellation(t *testing.T) {
	c := New()

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	snapshot, err := c.Collect(ctx)

	// With a cancelled context, we expect either:
	// 1. An error indicating context was cancelled
	// 2. Or the function completed before checking context
	// Both are acceptable behaviors
	if err == nil && snapshot == nil {
		t.Error("Collect() should return snapshot or error, not both nil")
	}
}

func TestCollectorInterface_ReturnsSnapshot(t *testing.T) {
	c := New()
	ctx := context.Background()

	snapshot, err := c.Collect(ctx)

	if err != nil {
		t.Skipf("Skipping test due to collect error: %v", err)
	}

	// Verify snapshot structure
	if snapshot == nil {
		t.Fatal("Expected non-nil snapshot")
	}

	// Check that applications have valid structure
	for i, app := range snapshot.Applications {
		if app.Name == "" {
			t.Errorf("Application[%d].Name should not be empty", i)
		}
		if app.PIDs == nil {
			t.Errorf("Application[%d].PIDs should not be nil", i)
		}
		if app.Connections == nil {
			t.Errorf("Application[%d].Connections should not be nil", i)
		}

		// Check connections have valid structure
		for j, conn := range app.Connections {
			if conn.Protocol == "" {
				t.Errorf("Application[%d].Connections[%d].Protocol should not be empty", i, j)
			}
			if conn.LocalAddr == "" {
				t.Errorf("Application[%d].Connections[%d].LocalAddr should not be empty", i, j)
			}
		}
	}
}

func TestCollectorInterface_SortsApplications(t *testing.T) {
	c := New()
	ctx := context.Background()

	snapshot, err := c.Collect(ctx)
	if err != nil {
		t.Skipf("Skipping test due to collect error: %v", err)
	}

	if len(snapshot.Applications) < 2 {
		t.Skipf("Not enough applications to test sorting (%d)", len(snapshot.Applications))
	}

	// Verify applications are sorted by connection count (descending)
	for i := 0; i < len(snapshot.Applications)-1; i++ {
		if len(snapshot.Applications[i].Connections) < len(snapshot.Applications[i+1].Connections) {
			t.Errorf("Applications not sorted: app[%d] has %d connections < app[%d] has %d connections",
				i, len(snapshot.Applications[i].Connections),
				i+1, len(snapshot.Applications[i+1].Connections))
		}
	}
}

func TestCollectOnce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	snapshot, ioStats, err := CollectOnce(ctx)
	if err != nil {
		t.Fatalf("CollectOnce failed: %v", err)
	}

	if snapshot == nil {
		t.Fatal("Expected non-nil snapshot")
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("Snapshot timestamp should be set")
	}

	// ioStats may be nil or empty if nettop unavailable, but shouldn't error
	if ioStats == nil {
		t.Log("ioStats is nil (nettop may not be available)")
	}
}

func TestCollectOnce_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	snapshot, _, err := CollectOnce(ctx)

	// Either error or quick return is acceptable
	if err == nil && snapshot == nil {
		t.Error("CollectOnce should return snapshot or error")
	}
}
