package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/collector"
	"github.com/kostyay/netmon/internal/config"
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
	Level          ViewLevel         // Which view level (process list, connections)
	ProcessName    string            // Selected process name (empty at Level 0)
	Cursor         int               // Current cursor position in the list
	SelectedID     model.SelectionID // Stable selection identifier
	SortColumn     SortColumn        // Current sort column
	SortAscending  bool              // Sort direction
	SelectedColumn SortColumn        // Currently selected column for navigation
	SortMode       bool              // Whether sort mode is active
}

// Model is the Bubble Tea model for the network monitor.
type Model struct {
	// Data
	snapshot       *model.NetworkSnapshot
	prevSnapshot   *model.NetworkSnapshot // Previous snapshot for diff
	collector      collector.Collector
	netIOCollector collector.NetIOCollector
	netIOCache     map[int32]*model.NetIOStats // Network I/O stats keyed by PID

	// Change highlighting
	changes          map[ConnectionKey]Change // Recently changed connections
	highlightChanges bool                     // whether to show change highlights

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
	viewport        viewport.Model
	viewportContent string // Pre-rendered content for viewport (set in Update, used in View)
	ready           bool   // true after viewport initialized on first WindowSizeMsg

	// Search/filter state
	searchMode   bool   // true when search input is active
	searchQuery  string // current search text (live updates while typing)
	activeFilter string // confirmed filter (applied after Enter)
	cliFilter    string // CLI-provided filter (uses exact port matching)

	// Kill mode state
	killMode     bool            // true when kill confirmation dialog is active
	killTarget   *killTargetInfo // target process/connection to kill
	killResult   string          // result message from kill operation
	killResultAt time.Time       // when killResult was set (for auto-dismiss)

	// DNS resolution
	dnsCache   map[string]string // IP -> hostname cache
	dnsEnabled bool              // whether DNS resolution is enabled

	// Service names
	serviceNames bool // show service names instead of port numbers

	// Settings modal
	settingsMode   bool // true when settings modal is visible
	settingsCursor int  // which setting is selected (0-based)

	// Help modal
	helpMode bool // true when help modal is visible

	// PID targeting (from --pid flag)
	targetPID int32 // PID to drill into on first snapshot (0 = disabled)

	// Version string (set via WithVersion)
	version string

	// Update available (set via VersionCheckMsg)
	updateAvailable string // e.g., "v1.2.0" (empty if up-to-date)

	// Animation state
	animations     bool // whether animations are enabled
	animationFrame int  // current animation frame (for pulsing indicators)
}

// killTargetInfo holds info about the process to be killed.
type killTargetInfo struct {
	PID         int32   // primary PID (for single PID kills)
	PIDs        []int32 // all PIDs (for process-level kills)
	ProcessName string
	Exe         string // executable path
	Port        int    // optional, 0 if killing by PID only
	Signal      string // signal to send (default SIGTERM)
}

// NewModel creates a new Model with default settings.
func NewModel() Model {
	return Model{
		collector:        collector.New(),
		netIOCollector:   collector.NewNetIOCollector(),
		refreshInterval:  DefaultRefreshInterval,
		netIOCache:       make(map[int32]*model.NetIOStats),
		changes:          make(map[ConnectionKey]Change),
		highlightChanges: config.CurrentSettings.HighlightChanges,
		dnsCache:         make(map[string]string),
		dnsEnabled:       config.CurrentSettings.DNSEnabled,
		serviceNames:     config.CurrentSettings.ServiceNames,
		animations:       config.CurrentSettings.Animations,
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

// WithFilter returns a copy of the model with an initial filter applied.
// CLI filters use exact port matching (port 80 only matches port 80, not 8080).
func (m Model) WithFilter(filter string) Model {
	m.activeFilter = filter
	m.cliFilter = filter // CLI filters use exact port matching
	return m
}

// WithPID returns a copy of the model that will drill into the given PID on first snapshot.
func (m Model) WithPID(pid int32) Model {
	m.targetPID = pid
	return m
}

// WithVersion returns a copy of the model with version string set.
func (m Model) WithVersion(v string) Model {
	m.version = v
	return m
}

// useExactPortMatch returns true if the current filter should use exact port matching.
// This is true when the filter was set via CLI argument, not via interactive search.
func (m Model) useExactPortMatch() bool {
	// If there's a CLI filter and it matches the active filter, use exact matching
	// If user has cleared/changed the filter interactively, fall back to substring
	return m.cliFilter != "" && m.cliFilter == m.activeFilter
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

// filterFields holds all filterable fields for a connection or process.
type filterFields struct {
	ProcessName string
	PIDs        []int32
	LocalAddr   string
	RemoteAddr  string
	Protocol    string
	State       string
}

// matchesFilter checks if any field contains the search string (case-insensitive).
// When exactPortMatch is true (CLI filters), ONLY matches exact port numbers.
// When false (interactive search), matches process name, PID, addresses, protocol, state, and ports via substring.
func matchesFilter(filter string, fields filterFields, exactPortMatch bool) bool {
	if filter == "" {
		return true
	}

	// CLI exact port matching - only match port numbers
	if exactPortMatch {
		ports := extractPortsFromAddrs(fields.LocalAddr, fields.RemoteAddr)
		for _, port := range ports {
			if strconv.Itoa(port) == filter {
				return true
			}
		}
		return false
	}

	// Interactive search - substring match on all fields
	filterLower := strings.ToLower(filter)

	// Match process name
	if fields.ProcessName != "" && strings.Contains(strings.ToLower(fields.ProcessName), filterLower) {
		return true
	}

	// Match any PID
	for _, pid := range fields.PIDs {
		if strings.Contains(strconv.Itoa(int(pid)), filter) {
			return true
		}
	}

	// Match local address
	if fields.LocalAddr != "" && strings.Contains(strings.ToLower(fields.LocalAddr), filterLower) {
		return true
	}

	// Match remote address
	if fields.RemoteAddr != "" && strings.Contains(strings.ToLower(fields.RemoteAddr), filterLower) {
		return true
	}

	// Match protocol
	if fields.Protocol != "" && strings.Contains(strings.ToLower(fields.Protocol), filterLower) {
		return true
	}

	// Match state
	if fields.State != "" && strings.Contains(strings.ToLower(fields.State), filterLower) {
		return true
	}

	return false
}
