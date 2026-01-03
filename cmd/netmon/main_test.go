package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/ui"
)

// TestNewModel_CanBeCreated verifies that the UI model can be created.
func TestNewModel_CanBeCreated(t *testing.T) {
	m := ui.NewModel()

	if m.View() == "" {
		t.Error("NewModel().View() should return non-empty string")
	}
}

// TestNewModel_ImplementsTeaModel verifies the model implements tea.Model.
func TestNewModel_ImplementsTeaModel(t *testing.T) {
	var _ tea.Model = ui.NewModel()
}

// TestNewModel_Init verifies initialization works.
func TestNewModel_Init(t *testing.T) {
	m := ui.NewModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

// TestProgramCreation verifies tea.Program can be created with our model.
func TestProgramCreation(t *testing.T) {
	m := ui.NewModel()

	p := tea.NewProgram(m)
	if p == nil {
		t.Error("tea.NewProgram should return non-nil program")
	}
}

// TestView_ShowsLoading verifies initial view shows loading state.
func TestView_ShowsLoading(t *testing.T) {
	m := ui.NewModel()
	view := m.View()

	// Initial view should show loading since no data has been fetched
	if view == "" {
		t.Error("View should return content")
	}

	// Check for expected UI elements (includes "Initializing" before viewport is ready)
	if !containsAny(view, []string{"Loading", "Initializing", "netmon", "Network"}) {
		t.Error("View should contain expected UI elements")
	}
}

// containsAny checks if s contains any of the substrings.
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
