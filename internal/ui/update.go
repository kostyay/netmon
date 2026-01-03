package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.fetchData(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		viewportHeight := msg.Height - headerHeight - footerHeight
		if viewportHeight < 1 {
			viewportHeight = 1
		}

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.viewMode == ViewGrouped {
				if m.cursor > 0 {
					m.cursor--
					m.syncViewportToCursor()
				}
			} else {
				// Table view: use tableCursor
				if m.tableCursor > 0 {
					m.tableCursor--
					m.syncViewportToCursor()
				}
			}
			return m, nil

		case "down", "j":
			if m.viewMode == ViewGrouped {
				if m.snapshot != nil && m.cursor < len(m.snapshot.Applications)-1 {
					m.cursor++
					m.syncViewportToCursor()
				}
			} else {
				// Table view: use tableCursor
				flatConns := m.flattenConnections()
				if m.tableCursor < len(flatConns)-1 {
					m.tableCursor++
					m.syncViewportToCursor()
				}
			}
			return m, nil

		case "left", "h":
			// Collapse current app (grouped view only)
			if m.viewMode == ViewGrouped {
				if m.snapshot != nil && m.cursor < len(m.snapshot.Applications) {
					m.expandedApps[m.snapshot.Applications[m.cursor].Name] = false
					m.syncViewportToCursor()
				}
			}
			// No-op in table view
			return m, nil

		case "right", "l", "enter":
			// Expand current app (grouped view only)
			if m.viewMode == ViewGrouped {
				if m.snapshot != nil && m.cursor < len(m.snapshot.Applications) {
					m.expandedApps[m.snapshot.Applications[m.cursor].Name] = true
					m.syncViewportToCursor()
				}
			}
			// No-op in table view
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
			// Toggle between grouped and table view
			if m.viewMode == ViewGrouped {
				m.viewMode = ViewTable
			} else {
				m.viewMode = ViewGrouped
			}
			return m, nil

		case "1", "2", "3", "4", "5":
			// Column sorting (only in table view)
			if m.viewMode == ViewTable {
				var newColumn SortColumn
				switch msg.String() {
				case "1":
					newColumn = SortProcess
				case "2":
					newColumn = SortProtocol
				case "3":
					newColumn = SortLocal
				case "4":
					newColumn = SortRemote
				case "5":
					newColumn = SortState
				}

				if m.sortColumn == newColumn {
					// Toggle sort direction
					m.sortAscending = !m.sortAscending
				} else {
					// New column, default to ascending
					m.sortColumn = newColumn
					m.sortAscending = true
				}
				return m, nil
			}

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
		// Ensure cursor is valid
		if m.snapshot != nil && m.cursor >= len(m.snapshot.Applications) {
			m.cursor = max(0, len(m.snapshot.Applications)-1)
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

// syncViewportToCursor ensures the cursor line is visible in the viewport.
func (m *Model) syncViewportToCursor() {
	if !m.ready {
		return
	}

	lineNumber := m.cursorLinePosition()

	// If cursor is above visible area, scroll up
	if lineNumber < m.viewport.YOffset {
		m.viewport.SetYOffset(lineNumber)
	}

	// If cursor is below visible area, scroll down
	if lineNumber >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(lineNumber - m.viewport.Height + 1)
	}
}
