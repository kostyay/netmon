package ui

import (
	"fmt"
	"strconv"
	"strings"
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

// ViewLevel represents which level of navigation the user is at.
type ViewLevel int

const (
	LevelProcessList    ViewLevel = iota // Level 0: list of processes
	LevelConnections                     // Level 1: connections for a specific process
	LevelAllConnections                  // Level 2: flat view of all connections
)

// String returns a human-readable name for the ViewLevel.
func (v ViewLevel) String() string {
	switch v {
	case LevelProcessList:
		return "Processes"
	case LevelConnections:
		return "Connections"
	case LevelAllConnections:
		return "All Connections"
	default:
		return fmt.Sprintf("ViewLevel(%d)", v)
	}
}

// SortColumn represents the column to sort by in table view.
type SortColumn int

const (
	SortPID SortColumn = iota
	SortProcess
	SortProtocol
	SortLocal
	SortRemote
	SortState
	// Process list specific columns
	SortConns
	SortEstablished
	SortListen
	SortTX
	SortRX
)

// String returns a human-readable name for the SortColumn.
func (s SortColumn) String() string {
	switch s {
	case SortPID:
		return "PID"
	case SortProcess:
		return "Process"
	case SortProtocol:
		return "Protocol"
	case SortLocal:
		return "Local"
	case SortRemote:
		return "Remote"
	case SortState:
		return "State"
	case SortConns:
		return "Conns"
	case SortEstablished:
		return "Established"
	case SortListen:
		return "Listen"
	case SortTX:
		return "TX"
	case SortRX:
		return "RX"
	default:
		return fmt.Sprintf("SortColumn(%d)", s)
	}
}

// ViewState captures the navigation state at a given level.
type ViewState struct {
	Level          ViewLevel  // Which view level (process list, connections)
	ProcessName    string     // Selected process name (empty at Level 0)
	Cursor         int        // Current cursor position in the list
	SortColumn     SortColumn // Current sort column
	SortAscending  bool       // Sort direction
	SelectedColumn SortColumn // Currently selected column for navigation
	SortMode       bool       // Whether sort mode is active
}

// Model is the Bubble Tea model for the network monitor.
type Model struct {
	// Data
	snapshot       *model.NetworkSnapshot
	collector      collector.Collector
	netIOCollector collector.NetIOCollector
	netIOCache     map[int32]*model.NetIOStats // Network I/O stats keyed by PID

	// Navigation stack (replaces viewMode, expandedApps, cursor, tableCursor)
	stack []ViewState

	// UI State
	quitting bool

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

	// Search/filter state
	searchMode   bool   // true when search input is active
	searchQuery  string // current search text (live updates while typing)
	activeFilter string // confirmed filter (applied after Enter)

	// Kill mode state
	killMode     bool            // true when kill confirmation dialog is active
	killTarget   *killTargetInfo // target process/connection to kill
	killResult   string          // result message from kill operation
	killResultAt time.Time       // when killResult was set (for auto-dismiss)
}

// killTargetInfo holds info about the process to be killed.
type killTargetInfo struct {
	PID         int32
	ProcessName string
	Port        int    // optional, 0 if killing by PID only
	Signal      string // signal to send (default SIGTERM)
}

// NewModel creates a new Model with default settings.
func NewModel() Model {
	return Model{
		collector:       collector.New(),
		netIOCollector:  collector.NewNetIOCollector(),
		refreshInterval: DefaultRefreshInterval,
		netIOCache:      make(map[int32]*model.NetIOStats),
		stack: []ViewState{{
			Level:          LevelProcessList,
			ProcessName:    "",
			Cursor:         0,
			SortColumn:     SortProcess,
			SortAscending:  true,
			SelectedColumn: SortProcess,
		}},
	}
}

// CurrentView returns the current view state (top of stack).
func (m *Model) CurrentView() *ViewState {
	if len(m.stack) == 0 {
		return nil
	}
	return &m.stack[len(m.stack)-1]
}

// PushView pushes a new view state onto the stack.
func (m *Model) PushView(state ViewState) {
	m.stack = append(m.stack, state)
}

// PopView pops the current view state from the stack.
// Returns false if already at the root level.
func (m *Model) PopView() bool {
	if len(m.stack) <= 1 {
		return false
	}
	m.stack = m.stack[:len(m.stack)-1]
	return true
}

// AtRootLevel returns true if at the root navigation level.
func (m *Model) AtRootLevel() bool {
	return len(m.stack) <= 1
}

var _ tea.Model = Model{}

// matchesFilter checks if any field contains the search string (case-insensitive).
func matchesFilter(filter, processName string, pids []int32, ports []int) bool {
	if filter == "" {
		return true
	}
	filter = strings.ToLower(filter)

	// Match process name
	if strings.Contains(strings.ToLower(processName), filter) {
		return true
	}

	// Match any PID as string
	for _, pid := range pids {
		if strings.Contains(strconv.Itoa(int(pid)), filter) {
			return true
		}
	}

	// Match any port
	for _, port := range ports {
		if strings.Contains(strconv.Itoa(port), filter) {
			return true
		}
	}

	return false
}
