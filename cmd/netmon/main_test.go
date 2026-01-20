package main

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/model"
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

// Tests for filterSnapshotByPort

func TestFilterSnapshotByPort_LocalPortMatch(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					{PID: 100, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.2:443", State: model.StateEstablished},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "8080")

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	if len(result.Applications[0].Connections) != 1 {
		t.Errorf("expected 1 connection, got %d", len(result.Applications[0].Connections))
	}
	if result.Applications[0].Connections[0].LocalAddr != "127.0.0.1:8080" {
		t.Errorf("expected local port 8080, got %s", result.Applications[0].Connections[0].LocalAddr)
	}
}

func TestFilterSnapshotByPort_RemotePortMatch(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:55000", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					{PID: 100, LocalAddr: "127.0.0.1:55001", RemoteAddr: "10.0.0.2:80", State: model.StateEstablished},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "443")

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	if len(result.Applications[0].Connections) != 1 {
		t.Errorf("expected 1 connection, got %d", len(result.Applications[0].Connections))
	}
}

func TestFilterSnapshotByPort_NoMatch(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "9999")

	if len(result.Applications) != 0 {
		t.Errorf("expected 0 apps when no match, got %d", len(result.Applications))
	}
}

func TestFilterSnapshotByPort_MultiplePIDs(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100, 101, 102},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
					{PID: 101, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.2:443"}, // No match
					{PID: 102, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.3:443"},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "8080")

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	// Only PIDs 100 and 102 have matching connections
	if len(result.Applications[0].PIDs) != 2 {
		t.Errorf("expected 2 PIDs, got %d", len(result.Applications[0].PIDs))
	}
	if len(result.Applications[0].Connections) != 2 {
		t.Errorf("expected 2 connections, got %d", len(result.Applications[0].Connections))
	}
}

func TestFilterSnapshotByPort_StateCountsUpdated(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name:             "App1",
				PIDs:             []int32{100},
				EstablishedCount: 5, // Original count
				ListenCount:      3, // Original count
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "*", State: model.StateListen},
					{PID: 100, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.2:443", State: model.StateEstablished}, // No match
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "8080")

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	// State counts should be recalculated for filtered connections only
	if result.Applications[0].EstablishedCount != 1 {
		t.Errorf("expected EstablishedCount=1, got %d", result.Applications[0].EstablishedCount)
	}
	if result.Applications[0].ListenCount != 1 {
		t.Errorf("expected ListenCount=1, got %d", result.Applications[0].ListenCount)
	}
}

func TestFilterSnapshotByPort_PreservesMetadata(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		SkippedCount: 5,
		Applications: []model.Application{
			{
				Name: "App1",
				Exe:  "/usr/bin/app1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "8080")

	if result.SkippedCount != 5 {
		t.Errorf("SkippedCount should be preserved, got %d", result.SkippedCount)
	}
	if result.Applications[0].Exe != "/usr/bin/app1" {
		t.Errorf("Exe should be preserved, got %s", result.Applications[0].Exe)
	}
}

func TestFilterSnapshotByPort_MultipleApps(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
			{
				Name: "App2",
				PIDs: []int32{200},
				Connections: []model.Connection{
					{PID: 200, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.2:443"}, // No match
				},
			},
			{
				Name: "App3",
				PIDs: []int32{300},
				Connections: []model.Connection{
					{PID: 300, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.3:443"},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "8080")

	if len(result.Applications) != 2 {
		t.Errorf("expected 2 apps, got %d", len(result.Applications))
	}
	// Verify the right apps are included
	names := make(map[string]bool)
	for _, app := range result.Applications {
		names[app.Name] = true
	}
	if !names["App1"] || !names["App3"] {
		t.Errorf("expected App1 and App3, got %v", names)
	}
}

func TestFilterSnapshotByPort_PartialPortMatch(t *testing.T) {
	// Ensure "80" doesn't match "8080" (exact port suffix match only)
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "80")

	// :80 should not match :8080
	if len(result.Applications) != 0 {
		t.Errorf("expected 0 apps (partial port should not match), got %d", len(result.Applications))
	}
}

func TestFilterSnapshotByPort_ListenState(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "Server",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "*:80", RemoteAddr: "*", State: model.StateListen},
				},
			},
		},
	}

	result := filterSnapshotByPort(snapshot, "80")

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	if result.Applications[0].ListenCount != 1 {
		t.Errorf("expected ListenCount=1, got %d", result.Applications[0].ListenCount)
	}
}

// Tests for filterSnapshotByPID

func TestFilterSnapshotByPID_BasicMatch(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100, 101},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					{PID: 101, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.2:443", State: model.StateEstablished},
				},
			},
		},
	}

	result := filterSnapshotByPID(snapshot, 100)

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	if len(result.Applications[0].Connections) != 1 {
		t.Errorf("expected 1 connection, got %d", len(result.Applications[0].Connections))
	}
	if result.Applications[0].Connections[0].PID != 100 {
		t.Errorf("expected PID 100, got %d", result.Applications[0].Connections[0].PID)
	}
	// Only the matched PID should be in PIDs list
	if len(result.Applications[0].PIDs) != 1 || result.Applications[0].PIDs[0] != 100 {
		t.Errorf("expected PIDs=[100], got %v", result.Applications[0].PIDs)
	}
}

func TestFilterSnapshotByPID_NoMatch(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
		},
	}

	result := filterSnapshotByPID(snapshot, 999)

	if len(result.Applications) != 0 {
		t.Errorf("expected 0 apps when no match, got %d", len(result.Applications))
	}
}

func TestFilterSnapshotByPID_StateCountsUpdated(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name:             "App1",
				PIDs:             []int32{100, 101},
				EstablishedCount: 5,
				ListenCount:      3,
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					{PID: 100, LocalAddr: "127.0.0.1:8081", RemoteAddr: "*", State: model.StateListen},
					{PID: 101, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.2:443", State: model.StateEstablished},
				},
			},
		},
	}

	result := filterSnapshotByPID(snapshot, 100)

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	if result.Applications[0].EstablishedCount != 1 {
		t.Errorf("expected EstablishedCount=1, got %d", result.Applications[0].EstablishedCount)
	}
	if result.Applications[0].ListenCount != 1 {
		t.Errorf("expected ListenCount=1, got %d", result.Applications[0].ListenCount)
	}
}

func TestFilterSnapshotByPID_PreservesMetadata(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		SkippedCount: 7,
		Applications: []model.Application{
			{
				Name: "App1",
				Exe:  "/usr/bin/app1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
		},
	}

	result := filterSnapshotByPID(snapshot, 100)

	if result.SkippedCount != 7 {
		t.Errorf("SkippedCount should be preserved, got %d", result.SkippedCount)
	}
	if result.Applications[0].Exe != "/usr/bin/app1" {
		t.Errorf("Exe should be preserved, got %s", result.Applications[0].Exe)
	}
	if result.Applications[0].Name != "App1" {
		t.Errorf("Name should be preserved, got %s", result.Applications[0].Name)
	}
}

func TestFilterSnapshotByPID_MultipleApps(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
			{
				Name: "App2",
				PIDs: []int32{200},
				Connections: []model.Connection{
					{PID: 200, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.2:443"},
				},
			},
		},
	}

	result := filterSnapshotByPID(snapshot, 200)

	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	if result.Applications[0].Name != "App2" {
		t.Errorf("expected App2, got %s", result.Applications[0].Name)
	}
}

func TestFilterSnapshotByPID_PIDWithNoConnections(t *testing.T) {
	// PID exists in app but has no connections (edge case)
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100, 101},
				Connections: []model.Connection{
					{PID: 101, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
				},
			},
		},
	}

	result := filterSnapshotByPID(snapshot, 100)

	// Should still return the app but with no connections
	if len(result.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result.Applications))
	}
	if len(result.Applications[0].Connections) != 0 {
		t.Errorf("expected 0 connections, got %d", len(result.Applications[0].Connections))
	}
}

// Tests for pidExists

func TestPidExists_CurrentProcess(t *testing.T) {
	// Current process should always exist
	if !pidExists(int32(os.Getpid())) {
		t.Error("pidExists should return true for current process")
	}
}

func TestPidExists_InvalidPID(t *testing.T) {
	// Very large PID is unlikely to exist
	if pidExists(999999999) {
		t.Error("pidExists should return false for non-existent PID")
	}
}

// Tests for WithPID

func TestWithPID_SetsTargetPID(t *testing.T) {
	m := ui.NewModel()
	m = m.WithPID(12345)

	// We can't directly check targetPID since it's unexported,
	// but we can verify the model was returned properly
	if m.View() == "" {
		t.Error("WithPID should return a valid model")
	}
}
