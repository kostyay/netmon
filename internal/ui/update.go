package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/model"
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
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.snapshot != nil && m.cursor < len(m.snapshot.Applications)-1 {
				m.cursor++
			}
			return m, nil

		case "left", "h":
			// Collapse current app
			if m.snapshot != nil && m.cursor < len(m.snapshot.Applications) {
				m.snapshot.Applications[m.cursor].Expanded = false
			}
			return m, nil

		case "right", "l", "enter":
			// Expand current app
			if m.snapshot != nil && m.cursor < len(m.snapshot.Applications) {
				m.snapshot.Applications[m.cursor].Expanded = true
			}
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
		}

	case TickMsg:
		// Schedule next tick and fetch new data
		return m, tea.Batch(
			m.tickCmd(),
			m.fetchData(),
		)

	case DataMsg:
		if msg.Err != nil {
			// Handle error - for now just ignore and keep old data
			return m, nil
		}
		// Preserve expanded state from previous snapshot
		m.mergeExpandedState(msg.Snapshot)
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

func (m *Model) mergeExpandedState(newSnapshot *model.NetworkSnapshot) {
	if m.snapshot == nil || newSnapshot == nil {
		return
	}

	// Create map of old expanded states
	expandedMap := make(map[string]bool)
	for _, app := range m.snapshot.Applications {
		expandedMap[app.Name] = app.Expanded
	}

	// Apply to new snapshot
	for i := range newSnapshot.Applications {
		if expanded, ok := expandedMap[newSnapshot.Applications[i].Name]; ok {
			newSnapshot.Applications[i].Expanded = expanded
		}
	}
}
