package ui

import (
	"context"

	"github.com/kostyay/netmon/internal/model"
)

// mockCollector is a test double for collector.Collector.
type mockCollector struct {
	snapshot *model.NetworkSnapshot
	err      error
}

func (m *mockCollector) Collect(ctx context.Context) (*model.NetworkSnapshot, error) {
	return m.snapshot, m.err
}

// newMockCollector creates a mockCollector with the given snapshot.
func newMockCollector(snapshot *model.NetworkSnapshot) *mockCollector {
	return &mockCollector{snapshot: snapshot}
}

// newMockCollectorWithError creates a mockCollector that returns an error.
func newMockCollectorWithError(err error) *mockCollector {
	return &mockCollector{err: err}
}
