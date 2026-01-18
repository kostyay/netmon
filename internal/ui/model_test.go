package ui

import (
	"context"
	"strings"
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

func TestNewModel_DefaultNetIOCollector(t *testing.T) {
	m := NewModel()

	if m.netIOCollector == nil {
		t.Error("NewModel() should set a netIOCollector")
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

func TestNewModel_StackInitialized(t *testing.T) {
	m := NewModel()

	if len(m.stack) != 1 {
		t.Errorf("stack length = %d, want 1", len(m.stack))
	}
	if m.CurrentView().Level != LevelProcessList {
		t.Errorf("initial level = %v, want LevelProcessList", m.CurrentView().Level)
	}
	if m.CurrentView().Cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.CurrentView().Cursor)
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

func TestPushView(t *testing.T) {
	m := NewModel()

	m.PushView(ViewState{
		Level:       LevelConnections,
		ProcessName: "TestApp",
		Cursor:      5,
	})

	if len(m.stack) != 2 {
		t.Errorf("stack length = %d, want 2", len(m.stack))
	}
	if m.CurrentView().Level != LevelConnections {
		t.Errorf("current level = %v, want LevelConnections", m.CurrentView().Level)
	}
	if m.CurrentView().ProcessName != "TestApp" {
		t.Errorf("ProcessName = %s, want TestApp", m.CurrentView().ProcessName)
	}
}

func TestPopView(t *testing.T) {
	m := NewModel()
	m.PushView(ViewState{
		Level:       LevelConnections,
		ProcessName: "TestApp",
	})

	popped := m.PopView()

	if !popped {
		t.Error("PopView should return true when stack has more than 1 item")
	}
	if len(m.stack) != 1 {
		t.Errorf("stack length = %d, want 1", len(m.stack))
	}
	if m.CurrentView().Level != LevelProcessList {
		t.Errorf("current level = %v, want LevelProcessList", m.CurrentView().Level)
	}
}

func TestPopView_AtRoot(t *testing.T) {
	m := NewModel()

	popped := m.PopView()

	if popped {
		t.Error("PopView should return false when at root level")
	}
	if len(m.stack) != 1 {
		t.Errorf("stack length = %d, want 1 (unchanged)", len(m.stack))
	}
}

func TestAtRootLevel(t *testing.T) {
	m := NewModel()

	if !m.AtRootLevel() {
		t.Error("AtRootLevel should return true for new model")
	}

	m.PushView(ViewState{Level: LevelConnections})

	if m.AtRootLevel() {
		t.Error("AtRootLevel should return false after pushing")
	}
}

func TestViewLevelString(t *testing.T) {
	if LevelProcessList.String() != "Processes" {
		t.Errorf("LevelProcessList.String() = %s, want Processes", LevelProcessList.String())
	}
	if LevelConnections.String() != "Connections" {
		t.Errorf("LevelConnections.String() = %s, want Connections", LevelConnections.String())
	}
}

func TestSortColumnString(t *testing.T) {
	tests := []struct {
		col  SortColumn
		want string
	}{
		{SortPID, "PID"},
		{SortProcess, "Process"},
		{SortProtocol, "Protocol"},
		{SortLocal, "Local"},
		{SortRemote, "Remote"},
		{SortState, "State"},
	}

	for _, tt := range tests {
		if got := tt.col.String(); got != tt.want {
			t.Errorf("%v.String() = %s, want %s", tt.col, got, tt.want)
		}
	}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }

// Additional navigation stack tests

func TestNavigationStack_CursorPreservation(t *testing.T) {
	m := NewModel()

	// Set cursor in root view
	m.CurrentView().Cursor = 5
	m.CurrentView().SortColumn = SortProcess
	m.CurrentView().SortAscending = false

	// Push new view
	m.PushView(ViewState{
		Level:       LevelConnections,
		ProcessName: "Chrome",
		Cursor:      3,
	})

	// Pop back
	m.PopView()

	// Verify original cursor and sort settings preserved
	if m.CurrentView().Cursor != 5 {
		t.Errorf("cursor = %d, want 5 (preserved)", m.CurrentView().Cursor)
	}
	if m.CurrentView().SortColumn != SortProcess {
		t.Errorf("sortColumn = %v, want SortProcess", m.CurrentView().SortColumn)
	}
	if m.CurrentView().SortAscending != false {
		t.Error("sortAscending should be false (preserved)")
	}
}

func TestNavigationStack_MultiLevelNavigation(t *testing.T) {
	m := NewModel()

	// Start at root
	if len(m.stack) != 1 {
		t.Fatalf("initial stack = %d, want 1", len(m.stack))
	}

	// Push to connections view
	m.PushView(ViewState{
		Level:       LevelConnections,
		ProcessName: "Chrome",
		Cursor:      2,
	})

	if len(m.stack) != 2 {
		t.Fatalf("stack after push = %d, want 2", len(m.stack))
	}

	// Verify current view
	cv := m.CurrentView()
	if cv.Level != LevelConnections {
		t.Errorf("level = %v, want LevelConnections", cv.Level)
	}
	if cv.ProcessName != "Chrome" {
		t.Errorf("processName = %s, want Chrome", cv.ProcessName)
	}

	// Pop back
	m.PopView()

	if len(m.stack) != 1 {
		t.Fatalf("stack after pop = %d, want 1", len(m.stack))
	}
	if m.CurrentView().Level != LevelProcessList {
		t.Errorf("after pop level = %v, want LevelProcessList", m.CurrentView().Level)
	}
}

func TestNavigationStack_ColumnSelection(t *testing.T) {
	m := NewModel()

	// Initial column selection
	if m.CurrentView().SelectedColumn != SortProcess {
		t.Errorf("initial column = %v, want SortProcess", m.CurrentView().SelectedColumn)
	}

	// Move column selection
	m.CurrentView().SelectedColumn = SortPID

	// Push view and verify it has its own column selection
	m.PushView(ViewState{
		Level:          LevelConnections,
		ProcessName:    "Firefox",
		SelectedColumn: SortLocal,
	})

	if m.CurrentView().SelectedColumn != SortLocal {
		t.Errorf("pushed view column = %v, want SortLocal", m.CurrentView().SelectedColumn)
	}

	// Pop and verify original column selection preserved
	m.PopView()

	if m.CurrentView().SelectedColumn != SortPID {
		t.Errorf("after pop column = %v, want SortPID", m.CurrentView().SelectedColumn)
	}
}

func TestCurrentView_EmptyStack(t *testing.T) {
	m := Model{
		stack: []ViewState{}, // Empty stack
	}

	cv := m.CurrentView()
	if cv != nil {
		t.Error("CurrentView should return nil for empty stack")
	}
}

func TestBreadcrumbs_AtRoot(t *testing.T) {
	m := createTestModel()
	m.ready = true

	crumbs := m.renderBreadcrumbsText()

	if !strings.Contains(crumbs, "Processes") {
		t.Errorf("breadcrumbs at root should contain 'Processes', got: %s", crumbs)
	}
}

func TestBreadcrumbs_AtConnections(t *testing.T) {
	m := createTestModel()
	m.ready = true

	m.PushView(ViewState{
		Level:       LevelConnections,
		ProcessName: "Chrome",
	})

	crumbs := m.renderBreadcrumbsText()

	if !strings.Contains(crumbs, "Processes") {
		t.Errorf("breadcrumbs should contain 'Processes', got: %s", crumbs)
	}
	if !strings.Contains(crumbs, "Chrome") {
		t.Errorf("breadcrumbs should contain 'Chrome', got: %s", crumbs)
	}
}

func TestViewState_Defaults(t *testing.T) {
	m := NewModel()
	view := m.CurrentView()

	if view.Level != LevelProcessList {
		t.Errorf("default level = %v, want LevelProcessList", view.Level)
	}
	if view.ProcessName != "" {
		t.Errorf("default processName = %s, want empty", view.ProcessName)
	}
	if view.Cursor != 0 {
		t.Errorf("default cursor = %d, want 0", view.Cursor)
	}
	if view.SortColumn != SortProcess {
		t.Errorf("default sortColumn = %v, want SortProcess", view.SortColumn)
	}
	if view.SortAscending != true {
		t.Error("default sortAscending should be true")
	}
	if view.SelectedColumn != SortProcess {
		t.Errorf("default selectedColumn = %v, want SortProcess", view.SelectedColumn)
	}
}

// Tests for matchesFilter function

func TestMatchesFilter_EmptyFilter(t *testing.T) {
	if !matchesFilter("", "Chrome", []int32{1234}, []int{8080}) {
		t.Error("Empty filter should match everything")
	}
}

func TestMatchesFilter_ProcessName(t *testing.T) {
	tests := []struct {
		filter      string
		processName string
		want        bool
	}{
		{"chrome", "Chrome", true},
		{"CHROME", "Chrome", true},
		{"Chr", "Chrome", true},
		{"firefox", "Chrome", false},
		{"ch", "Chrome", true},
	}

	for _, tt := range tests {
		got := matchesFilter(tt.filter, tt.processName, nil, nil)
		if got != tt.want {
			t.Errorf("matchesFilter(%q, %q, nil, nil) = %v, want %v",
				tt.filter, tt.processName, got, tt.want)
		}
	}
}

func TestMatchesFilter_PID(t *testing.T) {
	tests := []struct {
		filter string
		pids   []int32
		want   bool
	}{
		{"1234", []int32{1234}, true},
		{"123", []int32{1234}, true},      // partial match
		{"12", []int32{1234, 5678}, true}, // matches first
		{"56", []int32{1234, 5678}, true}, // matches second
		{"9999", []int32{1234}, false},
	}

	for _, tt := range tests {
		got := matchesFilter(tt.filter, "NoMatch", tt.pids, nil)
		if got != tt.want {
			t.Errorf("matchesFilter(%q, 'NoMatch', %v, nil) = %v, want %v",
				tt.filter, tt.pids, got, tt.want)
		}
	}
}

func TestMatchesFilter_Port(t *testing.T) {
	tests := []struct {
		filter string
		ports  []int
		want   bool
	}{
		{"8080", []int{8080}, true},
		{"80", []int{8080}, true},     // partial match
		{"443", []int{80, 443}, true}, // matches second
		{"9999", []int{8080}, false},
	}

	for _, tt := range tests {
		got := matchesFilter(tt.filter, "NoMatch", nil, tt.ports)
		if got != tt.want {
			t.Errorf("matchesFilter(%q, 'NoMatch', nil, %v) = %v, want %v",
				tt.filter, tt.ports, got, tt.want)
		}
	}
}

func TestMatchesFilter_Combined(t *testing.T) {
	// Should match if ANY of the criteria match
	if !matchesFilter("chrome", "Chrome", []int32{1234}, []int{8080}) {
		t.Error("Should match on process name")
	}
	if !matchesFilter("1234", "Firefox", []int32{1234}, []int{8080}) {
		t.Error("Should match on PID")
	}
	if !matchesFilter("8080", "Firefox", []int32{5678}, []int{8080}) {
		t.Error("Should match on port")
	}
	if matchesFilter("nomatch", "Firefox", []int32{5678}, []int{8080}) {
		t.Error("Should not match when nothing matches")
	}
}

func TestMatchesConnection_EmptyFilter(t *testing.T) {
	conn := model.Connection{
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "10.0.0.1:443",
		Protocol:   "tcp",
		State:      "ESTABLISHED",
	}
	if !matchesConnection("", conn) {
		t.Error("Empty filter should match all connections")
	}
}

func TestMatchesConnection_LocalAddr(t *testing.T) {
	conn := model.Connection{
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "10.0.0.1:443",
		Protocol:   "tcp",
		State:      "ESTABLISHED",
	}
	if !matchesConnection("127.0.0.1", conn) {
		t.Error("Should match local address")
	}
	if !matchesConnection("8080", conn) {
		t.Error("Should match local port")
	}
}

func TestMatchesConnection_RemoteAddr(t *testing.T) {
	conn := model.Connection{
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "10.0.0.1:443",
		Protocol:   "tcp",
		State:      "ESTABLISHED",
	}
	if !matchesConnection("10.0.0.1", conn) {
		t.Error("Should match remote address")
	}
	if !matchesConnection("443", conn) {
		t.Error("Should match remote port")
	}
}

func TestMatchesConnection_Protocol(t *testing.T) {
	conn := model.Connection{
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "10.0.0.1:443",
		Protocol:   "tcp",
		State:      "ESTABLISHED",
	}
	if !matchesConnection("tcp", conn) {
		t.Error("Should match protocol")
	}
	if !matchesConnection("TCP", conn) {
		t.Error("Should match protocol case-insensitively")
	}
}

func TestMatchesConnection_State(t *testing.T) {
	conn := model.Connection{
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "10.0.0.1:443",
		Protocol:   "tcp",
		State:      "ESTABLISHED",
	}
	if !matchesConnection("established", conn) {
		t.Error("Should match state case-insensitively")
	}
	if !matchesConnection("ESTAB", conn) {
		t.Error("Should match partial state")
	}
}

func TestMatchesConnection_NoMatch(t *testing.T) {
	conn := model.Connection{
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "10.0.0.1:443",
		Protocol:   "tcp",
		State:      "ESTABLISHED",
	}
	if matchesConnection("nomatch", conn) {
		t.Error("Should not match when nothing matches")
	}
}
