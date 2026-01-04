package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/collector"
)

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.fetchData(),
		m.fetchNetIO(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate viewport height: total - header - footer - frame borders
		viewportHeight := msg.Height - headerHeight - footerHeight - frameHeight
		if viewportHeight < 1 {
			viewportHeight = 1
		}

		// Viewport width accounts for frame border and padding (2 border + 2 padding)
		viewportWidth := msg.Width - 4
		if viewportWidth < 1 {
			viewportWidth = 1
		}

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, viewportHeight)
			m.ready = true
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = viewportHeight
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			view := m.CurrentView()
			if view != nil && view.Cursor > 0 {
				view.Cursor--
			}
			return m, nil

		case "down", "j":
			view := m.CurrentView()
			if view == nil || m.snapshot == nil {
				return m, nil
			}
			maxCursor := m.maxCursorForLevel(view.Level)
			if view.Cursor < maxCursor-1 {
				view.Cursor++
			}
			return m, nil

		case "left", "h":
			view := m.CurrentView()
			if view == nil {
				return m, nil
			}
			// Move column selection left
			if view.SelectedColumn > 0 {
				view.SelectedColumn--
			}
			return m, nil

		case "right", "l":
			view := m.CurrentView()
			if view == nil {
				return m, nil
			}
			// Move column selection right
			maxCol := m.maxColumnForLevel(view.Level)
			if int(view.SelectedColumn) < maxCol-1 {
				view.SelectedColumn++
			}
			return m, nil

		case "enter", " ":
			view := m.CurrentView()
			if view == nil || m.snapshot == nil {
				return m, nil
			}
			if view.Level == LevelProcessList {
				// Push to connections view for selected process
				if view.Cursor < len(m.snapshot.Applications) {
					app := m.snapshot.Applications[view.Cursor]
					m.PushView(ViewState{
						Level:          LevelConnections,
						ProcessName:    app.Name,
						Cursor:         0,
						SortColumn:     SortLocal,
						SortAscending:  true,
						SelectedColumn: SortLocal,
					})
				}
			} else {
				// Sort by selected column
				if view.SortColumn == view.SelectedColumn {
					view.SortAscending = !view.SortAscending
				} else {
					view.SortColumn = view.SelectedColumn
					view.SortAscending = true
				}
			}
			return m, nil

		case "esc", "backspace":
			// Pop view (go back)
			m.PopView()
			return m, nil

		case "+", "=":
			// Decrease refresh interval (faster refresh)
			if m.refreshInterval > MinRefreshInterval {
				m.refreshInterval -= RefreshStep
			}
			return m, nil

		case "-", "_":
			// Increase refresh interval (slower refresh)
			if m.refreshInterval < MaxRefreshInterval {
				m.refreshInterval += RefreshStep
			}
			return m, nil

		case "v":
			// Toggle between grouped (process list) and ungrouped (all connections) view
			view := m.CurrentView()
			if view == nil {
				return m, nil
			}
			if view.Level == LevelAllConnections {
				// Toggle back to process list
				m.stack = []ViewState{{
					Level:          LevelProcessList,
					Cursor:         0,
					SortColumn:     SortProcess,
					SortAscending:  true,
					SelectedColumn: SortProcess,
				}}
			} else {
				// Toggle to all connections view
				m.stack = []ViewState{{
					Level:          LevelAllConnections,
					Cursor:         0,
					SortColumn:     SortProcess,
					SortAscending:  true,
					SelectedColumn: SortProcess,
				}}
			}
			return m, nil

		default:
			// Pass unhandled keys to viewport for page up/down, mouse scroll, etc.
			if m.ready {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}

	case TickMsg:
		// Schedule next tick and fetch new data
		return m, tea.Batch(
			m.tickCmd(),
			m.fetchData(),
			m.fetchNetIO(),
		)

	case DataMsg:
		if msg.Err != nil {
			// Store error for display in UI
			m.lastError = msg.Err
			m.lastErrorTime = time.Now()
			return m, nil
		}
		// Clear error on successful fetch
		m.lastError = nil
		m.snapshot = msg.Snapshot
		// Ensure cursor is valid for current view level
		view := m.CurrentView()
		if view != nil && m.snapshot != nil {
			maxCursor := m.maxCursorForLevel(view.Level)
			if maxCursor > 0 && view.Cursor >= maxCursor {
				view.Cursor = maxCursor - 1
			} else if maxCursor == 0 {
				view.Cursor = 0
			}
		}
		return m, nil

	case NetIOMsg:
		if msg.Err != nil {
			// Silently ignore network I/O errors - stats are optional
			return m, nil
		}
		// Update the netIOCache with new stats
		for pid, stats := range msg.Stats {
			m.netIOCache[pid] = stats
		}
		return m, nil
	}

	return m, nil
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		snapshot, err := m.collector.Collect(ctx)
		return DataMsg{Snapshot: snapshot, Err: err}
	}
}

func (m Model) fetchNetIO() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		netIOCollector := collector.NewNetIOCollector()
		stats, err := netIOCollector.Collect(ctx)
		return NetIOMsg{Stats: stats, Err: err}
	}
}

// maxCursorForLevel returns the maximum cursor position for the given view level.
func (m Model) maxCursorForLevel(level ViewLevel) int {
	if m.snapshot == nil {
		return 0
	}
	switch level {
	case LevelProcessList:
		return len(m.snapshot.Applications)
	case LevelConnections:
		view := m.CurrentView()
		if view == nil {
			return 0
		}
		for _, app := range m.snapshot.Applications {
			if app.Name == view.ProcessName {
				return len(app.Connections)
			}
		}
		return 0
	case LevelAllConnections:
		return m.snapshot.TotalConnections()
	default:
		return 0
	}
}

// maxColumnForLevel returns the number of columns for the given view level.
func (m Model) maxColumnForLevel(level ViewLevel) int {
	switch level {
	case LevelProcessList:
		return 6 // Process, Conns, ESTAB, LISTEN, TX, RX
	case LevelConnections:
		return 4 // Proto, Local, Remote, State (PID removed as redundant)
	case LevelAllConnections:
		return 5 // Process, Proto, Local, Remote, State
	default:
		return 1
	}
}

