package ui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/model"
)

func createTestSnapshot() *model.NetworkSnapshot {
	return &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", PIDs: []int32{100}, Connections: []model.Connection{{Protocol: "TCP"}}},
			{Name: "App2", PIDs: []int32{200}, Connections: []model.Connection{{Protocol: "UDP"}}},
			{Name: "App3", PIDs: []int32{300}, Connections: []model.Connection{{Protocol: "TCP"}}},
		},
		Timestamp: time.Now(),
	}
}

func createTestModel() Model {
	snapshot := createTestSnapshot()
	m := Model{
		collector:       newMockCollector(snapshot),
		netIOCollector:  newMockNetIOCollector(nil),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        snapshot,
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
	return m
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := createTestModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.width != 100 {
		t.Errorf("width = %d, want 100", newModel.width)
	}
	if newModel.height != 50 {
		t.Errorf("height = %d, want 50", newModel.height)
	}
	if cmd != nil {
		t.Errorf("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Quit_Q(t *testing.T) {
	m := createTestModel()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.quitting {
		t.Error("quitting should be true after 'q'")
	}
	if cmd == nil {
		t.Error("cmd should not be nil (should be tea.Quit)")
	}
}

func TestUpdate_KeyMsg_Quit_CtrlC(t *testing.T) {
	m := createTestModel()
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.quitting {
		t.Error("quitting should be true after ctrl+c")
	}
	if cmd == nil {
		t.Error("cmd should not be nil")
	}
}

func TestUpdate_KeyMsg_Up(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 2

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 1 {
		t.Errorf("cursor = %d, want 1", newModel.CurrentView().Cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Up_K(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, want 0", newModel.CurrentView().Cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Up_AtTop(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should not go negative)", newModel.CurrentView().Cursor)
	}
}

func TestUpdate_KeyMsg_Down(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 1 {
		t.Errorf("cursor = %d, want 1", newModel.CurrentView().Cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Down_J(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 1 {
		t.Errorf("cursor = %d, want 1", newModel.CurrentView().Cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Down_AtBottom(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 2 // Last item (3 items total)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 2 {
		t.Errorf("cursor = %d, want 2 (should not exceed bounds)", newModel.CurrentView().Cursor)
	}
}

func TestUpdate_KeyMsg_Left_DoesNothingOutsideSortMode(t *testing.T) {
	m := createTestModel()
	m.CurrentView().SelectedColumn = SortProcess // Column index 1

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	// Left should do nothing outside sort mode
	if newModel.CurrentView().SelectedColumn != SortProcess {
		t.Errorf("SelectedColumn = %v, want SortProcess (unchanged) outside sort mode", newModel.CurrentView().SelectedColumn)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Right_DoesNothingOutsideSortMode(t *testing.T) {
	m := createTestModel()
	m.CurrentView().SelectedColumn = SortPID // Column index 0

	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	// Right should do nothing outside sort mode
	if newModel.CurrentView().SelectedColumn != SortPID {
		t.Errorf("SelectedColumn = %v, want SortPID (unchanged) outside sort mode", newModel.CurrentView().SelectedColumn)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Left_MovesColumnInSortMode(t *testing.T) {
	m := createTestModel()
	// Process list columns: [SortProcess, SortConns, SortEstablished, SortListen, SortTX, SortRX]
	m.CurrentView().SelectedColumn = SortConns // Second column
	m.CurrentView().SortMode = true

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().SelectedColumn != SortProcess {
		t.Errorf("SelectedColumn = %v, want SortProcess after left in sort mode", newModel.CurrentView().SelectedColumn)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Right_MovesColumnInSortMode(t *testing.T) {
	m := createTestModel()
	// Process list columns: [SortProcess, SortConns, SortEstablished, SortListen, SortTX, SortRX]
	m.CurrentView().SelectedColumn = SortProcess // First column
	m.CurrentView().SortMode = true

	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().SelectedColumn != SortConns {
		t.Errorf("SelectedColumn = %v, want SortConns after right in sort mode", newModel.CurrentView().SelectedColumn)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Enter_DrillsIn(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should have pushed a new view onto the stack
	if len(newModel.stack) != 2 {
		t.Errorf("stack length = %d, want 2 after drilling in", len(newModel.stack))
	}
	if newModel.CurrentView().Level != LevelConnections {
		t.Errorf("level = %v, want LevelConnections", newModel.CurrentView().Level)
	}
	if newModel.CurrentView().ProcessName != "App1" {
		t.Errorf("ProcessName = %s, want App1", newModel.CurrentView().ProcessName)
	}
}

func TestUpdate_KeyMsg_Esc_GoesBack(t *testing.T) {
	m := createTestModel()
	// Drill into App1
	m.PushView(ViewState{
		Level:       LevelConnections,
		ProcessName: "App1",
		Cursor:      0,
	})

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if len(newModel.stack) != 1 {
		t.Errorf("stack length = %d, want 1 after going back", len(newModel.stack))
	}
	if newModel.CurrentView().Level != LevelProcessList {
		t.Errorf("level = %v, want LevelProcessList", newModel.CurrentView().Level)
	}
}

func TestUpdate_KeyMsg_Plus_IncreaseRefresh(t *testing.T) {
	m := createTestModel()
	m.refreshInterval = 2 * time.Second

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	expected := 2*time.Second - RefreshStep
	if newModel.refreshInterval != expected {
		t.Errorf("refreshInterval = %v, want %v", newModel.refreshInterval, expected)
	}
}

func TestUpdate_KeyMsg_Minus_DecreaseRefresh(t *testing.T) {
	m := createTestModel()
	m.refreshInterval = 2 * time.Second

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	expected := 2*time.Second + RefreshStep
	if newModel.refreshInterval != expected {
		t.Errorf("refreshInterval = %v, want %v", newModel.refreshInterval, expected)
	}
}

func TestUpdate_KeyMsg_Plus_AtMinInterval(t *testing.T) {
	m := createTestModel()
	m.refreshInterval = MinRefreshInterval

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.refreshInterval != MinRefreshInterval {
		t.Errorf("refreshInterval = %v, should stay at min %v", newModel.refreshInterval, MinRefreshInterval)
	}
}

func TestUpdate_KeyMsg_Minus_AtMaxInterval(t *testing.T) {
	m := createTestModel()
	m.refreshInterval = MaxRefreshInterval

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.refreshInterval != MaxRefreshInterval {
		t.Errorf("refreshInterval = %v, should stay at max %v", newModel.refreshInterval, MaxRefreshInterval)
	}
}

func TestUpdate_TickMsg_SchedulesNextTick(t *testing.T) {
	m := createTestModel()
	msg := TickMsg(time.Now())

	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("TickMsg should return a command (batch of tick + fetch)")
	}
}

func TestUpdate_DataMsg_Success(t *testing.T) {
	m := createTestModel()
	m.snapshot = nil

	newSnapshot := createTestSnapshot()
	msg := DataMsg{Snapshot: newSnapshot, Err: nil}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.snapshot == nil {
		t.Error("snapshot should be set after DataMsg")
	}
	if newModel.snapshot.Applications[0].Name != "App1" {
		t.Error("snapshot should contain correct data")
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_DataMsg_Error(t *testing.T) {
	m := createTestModel()
	originalSnapshot := m.snapshot

	msg := DataMsg{Snapshot: nil, Err: errors.New("test error")}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.snapshot != originalSnapshot {
		t.Error("snapshot should remain unchanged on error")
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_DataMsg_CursorBounds(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 5 // Out of bounds

	smallSnapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", PIDs: []int32{100}, Connections: []model.Connection{{Protocol: "TCP"}}},
		},
		Timestamp: time.Now(),
	}
	msg := DataMsg{Snapshot: smallSnapshot, Err: nil}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, should be adjusted to 0 (max index)", newModel.CurrentView().Cursor)
	}
}

func TestUpdate_NilSnapshot_Down(t *testing.T) {
	m := Model{
		collector:       newMockCollector(nil),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        nil,
		netIOCache:      make(map[int32]*model.NetIOStats),
		stack: []ViewState{{
			Level:  LevelProcessList,
			Cursor: 0,
		}},
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, should stay at 0 with nil snapshot", newModel.CurrentView().Cursor)
	}
}

func TestConnectionsLevel_SortOnEnterInSortMode(t *testing.T) {
	m := createTestModel()
	// Push to connections level in sort mode
	m.PushView(ViewState{
		Level:          LevelConnections,
		ProcessName:    "App1",
		Cursor:         0,
		SortColumn:     SortLocal,
		SortAscending:  true,
		SelectedColumn: SortRemote,
		SortMode:       true,
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should have changed sort column
	if newModel.CurrentView().SortColumn != SortRemote {
		t.Errorf("sortColumn = %v, want SortRemote after Enter in sort mode", newModel.CurrentView().SortColumn)
	}
	// Should have exited sort mode
	if newModel.CurrentView().SortMode {
		t.Error("SortMode should be false after applying sort")
	}
}

func TestConnectionsLevel_SortToggleDirectionInSortMode(t *testing.T) {
	m := createTestModel()
	// Push to connections level with sort column same as selected, in sort mode
	m.PushView(ViewState{
		Level:          LevelConnections,
		ProcessName:    "App1",
		Cursor:         0,
		SortColumn:     SortLocal,
		SortAscending:  true,
		SelectedColumn: SortLocal, // Same as sort column
		SortMode:       true,
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().SortAscending {
		t.Error("sortAscending should be false after toggling")
	}
	// Should have exited sort mode
	if newModel.CurrentView().SortMode {
		t.Error("SortMode should be false after applying sort")
	}
}

func TestConnectionsLevel_EnterDoesNothingOutsideSortMode(t *testing.T) {
	m := createTestModel()
	// Push to connections level without sort mode
	m.PushView(ViewState{
		Level:          LevelConnections,
		ProcessName:    "App1",
		Cursor:         0,
		SortColumn:     SortLocal,
		SortAscending:  true,
		SelectedColumn: SortRemote,
		SortMode:       false,
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should NOT have changed sort column (Enter does nothing on connections without sort mode)
	if newModel.CurrentView().SortColumn != SortLocal {
		t.Errorf("sortColumn = %v, want SortLocal (unchanged) outside sort mode", newModel.CurrentView().SortColumn)
	}
}

func TestSortMode_SKeyEntersSortMode(t *testing.T) {
	m := createTestModel()
	m.CurrentView().SortColumn = SortProcess

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.CurrentView().SortMode {
		t.Error("SortMode should be true after pressing 's'")
	}
	// SelectedColumn should be set to current SortColumn
	if newModel.CurrentView().SelectedColumn != SortProcess {
		t.Errorf("SelectedColumn = %v, want SortProcess (same as SortColumn)", newModel.CurrentView().SelectedColumn)
	}
}

func TestSortMode_EscCancelsSortMode(t *testing.T) {
	m := createTestModel()
	m.CurrentView().SortMode = true
	m.CurrentView().SortColumn = SortProcess
	m.CurrentView().SelectedColumn = SortConns // Different from SortColumn

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().SortMode {
		t.Error("SortMode should be false after pressing Esc")
	}
	// SortColumn should be unchanged (cancel didn't apply the new column)
	if newModel.CurrentView().SortColumn != SortProcess {
		t.Errorf("SortColumn = %v, want SortProcess (unchanged after cancel)", newModel.CurrentView().SortColumn)
	}
	// Should NOT pop view - still at root level
	if len(newModel.stack) != 1 {
		t.Errorf("stack length = %d, want 1 (Esc in sort mode shouldn't pop)", len(newModel.stack))
	}
}

func TestSortMode_SKeyIsNoOpWhenAlreadyInSortMode(t *testing.T) {
	m := createTestModel()
	m.CurrentView().SortMode = true
	m.CurrentView().SelectedColumn = SortConns

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should still be in sort mode
	if !newModel.CurrentView().SortMode {
		t.Error("SortMode should still be true")
	}
	// SelectedColumn should be unchanged
	if newModel.CurrentView().SelectedColumn != SortConns {
		t.Errorf("SelectedColumn = %v, want SortConns (unchanged)", newModel.CurrentView().SelectedColumn)
	}
}

func TestNetIOMsg_Success(t *testing.T) {
	m := createTestModel()
	stats := map[int32]*model.NetIOStats{
		100: {BytesSent: 1000, BytesRecv: 2000},
		200: {BytesSent: 3000, BytesRecv: 4000},
	}
	msg := NetIOMsg{Stats: stats, Err: nil}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if len(newModel.netIOCache) != 2 {
		t.Errorf("netIOCache length = %d, want 2", len(newModel.netIOCache))
	}
	if newModel.netIOCache[100].BytesSent != 1000 {
		t.Errorf("BytesSent = %d, want 1000", newModel.netIOCache[100].BytesSent)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestNetIOMsg_Error(t *testing.T) {
	m := createTestModel()
	m.netIOCache[100] = &model.NetIOStats{BytesSent: 500}

	msg := NetIOMsg{Stats: nil, Err: errors.New("test error")}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	// Should still have existing cache
	if newModel.netIOCache[100].BytesSent != 500 {
		t.Error("netIOCache should be unchanged on error")
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

// Tests for extractPorts functions

func TestExtractPortsFromAddrs_SingleAddr(t *testing.T) {
	ports := extractPortsFromAddrs("127.0.0.1:8080")
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
	if ports[0] != 8080 {
		t.Errorf("expected port 8080, got %d", ports[0])
	}
}

func TestExtractPortsFromAddrs_MultipleAddrs(t *testing.T) {
	ports := extractPortsFromAddrs("127.0.0.1:8080", "10.0.0.1:443")
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(ports))
	}
	if ports[0] != 8080 || ports[1] != 443 {
		t.Errorf("expected ports [8080, 443], got %v", ports)
	}
}

func TestExtractPortsFromAddrs_IPv6(t *testing.T) {
	ports := extractPortsFromAddrs("[::1]:9090")
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
	if ports[0] != 9090 {
		t.Errorf("expected port 9090, got %d", ports[0])
	}
}

func TestExtractPortsFromAddrs_Wildcard(t *testing.T) {
	ports := extractPortsFromAddrs("*:80")
	if len(ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(ports))
	}
	if ports[0] != 80 {
		t.Errorf("expected port 80, got %d", ports[0])
	}
}

func TestExtractPortsFromAddrs_NoPort(t *testing.T) {
	ports := extractPortsFromAddrs("*")
	if len(ports) != 0 {
		t.Errorf("expected 0 ports for '*', got %d", len(ports))
	}
}

func TestExtractPortsFromAddrs_Empty(t *testing.T) {
	ports := extractPortsFromAddrs()
	if len(ports) != 0 {
		t.Errorf("expected 0 ports for empty input, got %d", len(ports))
	}
}

func TestExtractPorts_FromConnections(t *testing.T) {
	conns := []model.Connection{
		{LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
		{LocalAddr: "0.0.0.0:22", RemoteAddr: "*"},
	}

	ports := extractPorts(conns)
	if len(ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(ports))
	}

	// Should contain 8080, 443, 22 (wildcard * has no port)
	expected := map[int]bool{8080: true, 443: true, 22: true}
	for _, p := range ports {
		if !expected[p] {
			t.Errorf("unexpected port %d", p)
		}
	}
}

func TestExtractPorts_EmptyConnections(t *testing.T) {
	ports := extractPorts(nil)
	if len(ports) != 0 {
		t.Errorf("expected 0 ports for nil connections, got %d", len(ports))
	}

	ports = extractPorts([]model.Connection{})
	if len(ports) != 0 {
		t.Errorf("expected 0 ports for empty connections, got %d", len(ports))
	}
}

// NetIOCollector tests

func TestUpdate_NetIOMsg_Success(t *testing.T) {
	m := createTestModel()
	m.netIOCache = make(map[int32]*model.NetIOStats)

	stats := map[int32]*model.NetIOStats{
		100: {BytesSent: 1000, BytesRecv: 2000},
		200: {BytesSent: 3000, BytesRecv: 4000},
	}
	msg := NetIOMsg{Stats: stats, Err: nil}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if len(newModel.netIOCache) != 2 {
		t.Errorf("netIOCache length = %d, want 2", len(newModel.netIOCache))
	}
	if newModel.netIOCache[100].BytesSent != 1000 {
		t.Errorf("BytesSent for PID 100 = %d, want 1000", newModel.netIOCache[100].BytesSent)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_NetIOMsg_Error(t *testing.T) {
	m := createTestModel()
	originalCache := map[int32]*model.NetIOStats{
		100: {BytesSent: 1000, BytesRecv: 2000},
	}
	m.netIOCache = originalCache

	msg := NetIOMsg{Stats: nil, Err: errors.New("nettop failed")}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	// Cache should remain unchanged on error
	if len(newModel.netIOCache) != 1 {
		t.Errorf("netIOCache should remain unchanged on error, got length %d", len(newModel.netIOCache))
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestModel_WithMockNetIOCollector(t *testing.T) {
	stats := map[int32]*model.NetIOStats{
		1234: {BytesSent: 5000, BytesRecv: 10000},
	}
	mock := newMockNetIOCollector(stats)

	m := Model{
		collector:       newMockCollector(nil),
		netIOCollector:  mock,
		refreshInterval: DefaultRefreshInterval,
		netIOCache:      make(map[int32]*model.NetIOStats),
		stack:           []ViewState{{Level: LevelProcessList}},
	}

	// Simulate what fetchNetIO() does internally
	result, err := m.netIOCollector.Collect(context.Background())

	if err != nil {
		t.Errorf("Collect() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Collect() returned nil stats")
	}
	if result[1234].BytesSent != 5000 {
		t.Errorf("BytesSent = %d, want 5000", result[1234].BytesSent)
	}
}
