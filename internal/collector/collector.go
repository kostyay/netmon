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

// New returns the appropriate Collector for the current platform.
func New() Collector {
	return newPlatformCollector()
}
