package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/collector"
	"github.com/kostyay/netmon/internal/model"
)

// Refresh interval bounds.
const (
	MinRefreshInterval     = 500 * time.Millisecond
	MaxRefreshInterval     = 10 * time.Second
	DefaultRefreshInterval = 2 * time.Second
	RefreshStep            = 500 * time.Millisecond
)

// Model is the Bubble Tea model for the network monitor.
type Model struct {
	// Data
	snapshot  *model.NetworkSnapshot
	collector collector.Collector

	// UI State
	cursor   int  // Current selected app index
	quitting bool

	// Configuration
	refreshInterval time.Duration

	// Dimensions
	width  int
	height int
}

// NewModel creates a new Model with default settings.
func NewModel() Model {
	return Model{
		collector:       collector.New(),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        nil,
	}
}

var _ tea.Model = Model{}
