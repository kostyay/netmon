package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
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

// ViewMode represents the display mode for network connections.
type ViewMode int

const (
	ViewGrouped ViewMode = iota
	ViewTable
)

// SortColumn represents the column to sort by in table view.
type SortColumn int

const (
	SortProcess SortColumn = iota
	SortProtocol
	SortLocal
	SortRemote
	SortState
)

// Model is the Bubble Tea model for the network monitor.
type Model struct {
	// Data
	snapshot  *model.NetworkSnapshot
	collector collector.Collector

	// UI State
	cursor       int             // Current selected app index
	quitting     bool
	expandedApps map[string]bool // Track expanded state by app name

	// Error tracking
	lastError     error
	lastErrorTime time.Time

	// Configuration
	refreshInterval time.Duration

	// Dimensions
	width  int
	height int

	// Viewport for scrollable content
	viewport viewport.Model
	ready    bool // true after viewport initialized on first WindowSizeMsg

	// Table view state
	viewMode      ViewMode
	sortColumn    SortColumn
	sortAscending bool
	tableCursor   int
}

// NewModel creates a new Model with default settings.
func NewModel() Model {
	return Model{
		collector:       collector.New(),
		refreshInterval: DefaultRefreshInterval,
		expandedApps:    make(map[string]bool),
		viewMode:        ViewGrouped,
		sortColumn:      SortProcess,
		sortAscending:   true,
		tableCursor:     0,
	}
}

var _ tea.Model = Model{}
