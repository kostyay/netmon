package ui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/model"
)

func TestNewModel_DefaultCollector(t *testing.T) {
	m := NewModel()

	if m.collector == nil {
		t.Error("NewModel() should set a collector")
	}
}

func TestNewModel_DefaultRefreshInterval(t *testing.T) {
	m := NewModel()

	if m.refreshInterval != DefaultRefreshInterval {
		t.Errorf("refreshInterval = %v, want %v", m.refreshInterval, DefaultRefreshInterval)
	}
}

func TestNewModel_NilSnapshot(t *testing.T) {
	m := NewModel()

	if m.snapshot != nil {
		t.Error("NewModel() should have nil snapshot initially")
	}
}

func TestNewModel_ZeroCursor(t *testing.T) {
	m := NewModel()

	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestNewModel_NotQuitting(t *testing.T) {
	m := NewModel()

	if m.quitting {
		t.Error("quitting should be false initially")
	}
}

func TestModelImplementsTeaModel(t *testing.T) {
	var _ tea.Model = Model{}
}

func TestInit_ReturnsBatchCommand(t *testing.T) {
	m := createTestModel()

	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestConstants(t *testing.T) {
	if MinRefreshInterval > DefaultRefreshInterval {
		t.Errorf("MinRefreshInterval (%v) > DefaultRefreshInterval (%v)", MinRefreshInterval, DefaultRefreshInterval)
	}
	if DefaultRefreshInterval > MaxRefreshInterval {
		t.Errorf("DefaultRefreshInterval (%v) > MaxRefreshInterval (%v)", DefaultRefreshInterval, MaxRefreshInterval)
	}
	if RefreshStep <= 0 {
		t.Errorf("RefreshStep (%v) should be positive", RefreshStep)
	}
}

func TestMinRefreshInterval(t *testing.T) {
	expected := 500 * time.Millisecond
	if MinRefreshInterval != expected {
		t.Errorf("MinRefreshInterval = %v, want %v", MinRefreshInterval, expected)
	}
}

func TestMaxRefreshInterval(t *testing.T) {
	expected := 10 * time.Second
	if MaxRefreshInterval != expected {
		t.Errorf("MaxRefreshInterval = %v, want %v", MaxRefreshInterval, expected)
	}
}

func TestDefaultRefreshInterval(t *testing.T) {
	expected := 2 * time.Second
	if DefaultRefreshInterval != expected {
		t.Errorf("DefaultRefreshInterval = %v, want %v", DefaultRefreshInterval, expected)
	}
}

func TestRefreshStep(t *testing.T) {
	expected := 500 * time.Millisecond
	if RefreshStep != expected {
		t.Errorf("RefreshStep = %v, want %v", RefreshStep, expected)
	}
}

func TestModel_WithMockCollector(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "TestApp", PIDs: []int32{1234}},
		},
		Timestamp: time.Now(),
	}
	mock := newMockCollector(snapshot)

	m := Model{
		collector:       mock,
		refreshInterval: DefaultRefreshInterval,
	}

	// Simulate what fetchData() does
	result, err := m.collector.Collect(context.Background())

	if err != nil {
		t.Errorf("Collect() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Collect() returned nil snapshot")
	}
	if len(result.Applications) != 1 {
		t.Errorf("Expected 1 application, got %d", len(result.Applications))
	}
}

func TestTickMsg(t *testing.T) {
	now := time.Now()
	msg := TickMsg(now)

	// Verify it's the same time
	if time.Time(msg) != now {
		t.Error("TickMsg should preserve time value")
	}
}

func TestDataMsg(t *testing.T) {
	snapshot := createTestSnapshot()
	msg := DataMsg{
		Snapshot: snapshot,
		Err:      nil,
	}

	if msg.Snapshot != snapshot {
		t.Error("DataMsg.Snapshot should be set")
	}
	if msg.Err != nil {
		t.Error("DataMsg.Err should be nil")
	}
}

func TestDataMsg_WithError(t *testing.T) {
	err := &testError{}
	msg := DataMsg{
		Snapshot: nil,
		Err:      err,
	}

	if msg.Snapshot != nil {
		t.Error("DataMsg.Snapshot should be nil")
	}
	if msg.Err != err {
		t.Error("DataMsg.Err should be set")
	}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }
