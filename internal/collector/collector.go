package collector

import (
	"context"

	"github.com/kostyay/netmon/internal/model"
)

// Collector is the interface for collecting network data.
// Implementations are platform-specific.
type Collector interface {
	// Collect gathers all network connections and groups them by application.
	Collect(ctx context.Context) (*model.NetworkSnapshot, error)
}

// NetIOCollector is the interface for collecting network I/O statistics.
// Implementations are platform-specific.
type NetIOCollector interface {
	// Collect gathers network I/O stats (bytes sent/received) per process.
	Collect(ctx context.Context) (map[int32]*model.NetIOStats, error)
}

// New returns the appropriate Collector for the current platform.
func New() Collector {
	return newPlatformCollector()
}

// CollectOnce performs a single snapshot collection including NetIO stats.
// This is a convenience function for one-shot data collection without goroutines.
func CollectOnce(ctx context.Context) (*model.NetworkSnapshot, map[int32]*model.NetIOStats, error) {
	c := New()
	snapshot, err := c.Collect(ctx)
	if err != nil {
		return nil, nil, err
	}

	netio := NewNetIOCollector()
	ioStats, err := netio.Collect(ctx)
	if err != nil {
		return snapshot, nil, err
	}

	return snapshot, ioStats, nil
}
