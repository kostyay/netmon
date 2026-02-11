package ui

import (
	"context"

	"github.com/kostyay/netmon/internal/docker"
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

// mockNetIOCollector is a test double for collector.NetIOCollector.
type mockNetIOCollector struct {
	stats map[int32]*model.NetIOStats
	err   error
}

func (m *mockNetIOCollector) Collect(ctx context.Context) (map[int32]*model.NetIOStats, error) {
	return m.stats, m.err
}

// newMockNetIOCollector creates a mockNetIOCollector with the given stats.
func newMockNetIOCollector(stats map[int32]*model.NetIOStats) *mockNetIOCollector {
	return &mockNetIOCollector{stats: stats}
}

// mockDockerResolver is a test double for docker.Resolver.
type mockDockerResolver struct {
	containers map[int]*docker.ContainerPort
	err        error
}

func (m *mockDockerResolver) Resolve(ctx context.Context) (map[int]*docker.ContainerPort, error) {
	return m.containers, m.err
}

// newMockDockerResolver creates a mockDockerResolver with the given containers.
func newMockDockerResolver(containers map[int]*docker.ContainerPort) *mockDockerResolver {
	return &mockDockerResolver{containers: containers}
}
