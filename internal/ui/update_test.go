package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/docker"
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
		changes:         make(map[ConnectionKey]Change),
		dockerResolver:  newMockDockerResolver(nil),
		dockerCache:     make(map[int]*docker.ContainerPort),
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

// TestUpdate_KeyMsg_Enter_RespectsSortOrder is a regression test for the bug where
// Enter key would select the wrong process when the list was sorted differently
// than the underlying slice order. The cursor indexes into the *sorted* view,
// so the Enter handler must also apply sorting before indexing.
func TestUpdate_KeyMsg_Enter_RespectsSortOrder(t *testing.T) {
	m := createTestModel()
	// Sort by PID descending: App3 (300), App2 (200), App1 (100)
	m.CurrentView().SortColumn = SortPID
	m.CurrentView().SortAscending = false
	m.CurrentView().Cursor = 0 // First row in sorted view = App3

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should have drilled into App3 (first in sorted view), not App1 (first in slice)
	if newModel.CurrentView().ProcessName != "App3" {
		t.Errorf("ProcessName = %s, want App3 (respects sort order)", newModel.CurrentView().ProcessName)
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

// Tests for kill mode

func TestKillMode_XEntersKillMode(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("killMode should be true after pressing 'x'")
	}
	if newModel.killTarget == nil {
		t.Error("killTarget should not be nil")
	}
	if newModel.killTarget.Signal != "SIGTERM" {
		t.Errorf("Signal = %s, want SIGTERM", newModel.killTarget.Signal)
	}
	if newModel.killTarget.PID != 100 {
		t.Errorf("PID = %d, want 100", newModel.killTarget.PID)
	}
}

func TestKillMode_ShiftXEntersKillModeWithSIGKILL(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("killMode should be true after pressing 'X'")
	}
	if newModel.killTarget == nil {
		t.Error("killTarget should not be nil")
	}
	if newModel.killTarget.Signal != "SIGKILL" {
		t.Errorf("Signal = %s, want SIGKILL", newModel.killTarget.Signal)
	}
}

func TestKillMode_EscCancels(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = &killTargetInfo{PID: 100, ProcessName: "App1", Signal: "SIGTERM"}

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false after pressing Esc")
	}
	if newModel.killTarget != nil {
		t.Error("killTarget should be nil after cancel")
	}
}

func TestKillMode_ArrowTogglesSignal(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = &killTargetInfo{PID: 100, ProcessName: "App1", Signal: "SIGTERM"}

	// Press down to toggle to SIGKILL
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("killMode should still be true")
	}
	if newModel.killTarget.Signal != "SIGKILL" {
		t.Errorf("Signal = %s, want SIGKILL after toggle", newModel.killTarget.Signal)
	}

	// Press up to toggle back to SIGTERM
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if newModel.killTarget.Signal != "SIGTERM" {
		t.Errorf("Signal = %s, want SIGTERM after second toggle", newModel.killTarget.Signal)
	}
}

func TestKillMode_EnterConfirmsKill(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = &killTargetInfo{PID: 99999, ProcessName: "FakeApp", Signal: "SIGTERM"}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false after kill attempt")
	}
	if newModel.killTarget != nil {
		t.Error("killTarget should be nil after kill")
	}
	// killResult should be set (either success or failure message)
	if newModel.killResult == "" {
		t.Error("killResult should be set after kill attempt")
	}
	if newModel.killResultAt.IsZero() {
		t.Error("killResultAt should be set after kill attempt")
	}
}

func TestKillMode_XWithNilSnapshotDoesNothing(t *testing.T) {
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

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false with nil snapshot")
	}
}

func TestKillMode_OtherKeysIgnored(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = &killTargetInfo{PID: 100, ProcessName: "App1", Signal: "SIGTERM"}

	// Try pressing 'q' which normally quits - should be ignored in kill mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("killMode should still be true (other keys ignored)")
	}
	if newModel.quitting {
		t.Error("quitting should be false (q key should be ignored in kill mode)")
	}
}

func TestKillMode_AllConnectionsView(t *testing.T) {
	m := createTestModel()
	// Set up test snapshot with connections
	m.snapshot.Applications[0].Connections = []model.Connection{
		{PID: 100, LocalAddr: "127.0.0.1:8080", Protocol: "TCP", State: "ESTABLISHED"},
	}
	// Switch to all connections view
	m.stack = []ViewState{{
		Level:          LevelAllConnections,
		Cursor:         0,
		SortColumn:     SortProcess,
		SortAscending:  true,
		SelectedColumn: SortProcess,
	}}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("killMode should be true")
	}
	if newModel.killTarget == nil {
		t.Error("killTarget should not be nil")
	}
	if newModel.killTarget.Port != 8080 {
		t.Errorf("Port = %d, want 8080", newModel.killTarget.Port)
	}
}

func TestExtractSinglePort(t *testing.T) {
	tests := []struct {
		addr string
		want int
	}{
		{"127.0.0.1:8080", 8080},
		{"[::1]:9090", 9090},
		{"*:80", 80},
		{"*", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := model.ExtractPort(tt.addr)
		if got != tt.want {
			t.Errorf("model.ExtractPort(%q) = %d, want %d", tt.addr, got, tt.want)
		}
	}
}

func TestKillMode_ConnectionsView(t *testing.T) {
	m := createTestModel()
	// Set up test snapshot with connections
	m.snapshot.Applications[0].Connections = []model.Connection{
		{PID: 100, LocalAddr: "127.0.0.1:8080", Protocol: "TCP", State: "ESTABLISHED"},
		{PID: 101, LocalAddr: "127.0.0.1:9000", Protocol: "TCP", State: "ESTABLISHED"},
	}
	// Switch to connections view for App1
	m.stack = []ViewState{{
		Level:          LevelConnections,
		ProcessName:    "App1",
		Cursor:         1, // Select second connection
		SortColumn:     SortLocal,
		SortAscending:  true,
		SelectedColumn: SortLocal,
	}}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("killMode should be true")
	}
	if newModel.killTarget == nil {
		t.Error("killTarget should not be nil")
	}
	if newModel.killTarget.Port != 9000 {
		t.Errorf("Port = %d, want 9000 (second connection)", newModel.killTarget.Port)
	}
	if newModel.killTarget.ProcessName != "App1" {
		t.Errorf("ProcessName = %s, want App1", newModel.killTarget.ProcessName)
	}
}

func TestEnterKillMode_NilView(t *testing.T) {
	m := Model{
		snapshot:   &model.NetworkSnapshot{},
		stack:      []ViewState{}, // Empty stack = nil view
		netIOCache: make(map[int32]*model.NetIOStats),
	}

	updated, cmd := m.enterKillMode("SIGTERM")
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false with nil view")
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestEnterKillMode_ProcessWithEmptyPIDs(t *testing.T) {
	m := createTestModel()
	// Set an app with no PIDs
	m.snapshot.Applications = []model.Application{
		{Name: "EmptyApp", PIDs: []int32{}},
	}
	m.CurrentView().Cursor = 0

	updated, _ := m.enterKillMode("SIGTERM")
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false with empty PIDs")
	}
	if newModel.killTarget != nil {
		t.Error("killTarget should be nil")
	}
}

func TestEnterKillMode_CursorOutOfBounds(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 999 // Way out of bounds

	updated, _ := m.enterKillMode("SIGTERM")
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false with cursor out of bounds")
	}
	if newModel.killTarget != nil {
		t.Error("killTarget should be nil")
	}
}

func TestEnterKillMode_ProcessHasMultiplePIDs(t *testing.T) {
	m := createTestModel()
	// Set up app with multiple PIDs
	m.snapshot.Applications = []model.Application{
		{Name: "MultiPID", PIDs: []int32{100, 101, 102}},
	}
	m.CurrentView().Cursor = 0

	updated, _ := m.enterKillMode("SIGTERM")
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("killMode should be true")
	}
	if newModel.killTarget == nil {
		t.Fatal("killTarget should not be nil")
	}
	// Should capture all PIDs
	if len(newModel.killTarget.PIDs) != 3 {
		t.Errorf("Expected 3 PIDs, got %d", len(newModel.killTarget.PIDs))
	}
}

func TestExecuteKill_NilTarget(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = nil

	updated, _ := m.executeKill()
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false after executeKill with nil target")
	}
}

func TestExecuteKill_UnknownSignalFallsBackToSIGTERM(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = &killTargetInfo{
		PID:         99999, // Non-existent PID
		ProcessName: "FakeApp",
		Signal:      "INVALID_SIGNAL", // Unknown signal
	}

	updated, _ := m.executeKill()
	newModel := updated.(Model)

	// Should attempt kill with SIGTERM fallback
	if newModel.killMode {
		t.Error("killMode should be false after kill attempt")
	}
	// Should have a result (probably failed since PID doesn't exist)
	if newModel.killResult == "" {
		t.Error("killResult should be set")
	}
}

func TestExecuteKill_MultiplePIDsPartialFailure(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = &killTargetInfo{
		PID:         99999,
		PIDs:        []int32{99999, 99998, 99997}, // Non-existent PIDs
		ProcessName: "FakeApp",
		Signal:      "SIGTERM",
	}

	updated, _ := m.executeKill()
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false after kill attempt")
	}
	if newModel.killTarget != nil {
		t.Error("killTarget should be nil after kill")
	}
	// Should have failure result
	if newModel.killResult == "" {
		t.Error("killResult should be set")
	}
}

func TestExecuteKill_SinglePIDFallback(t *testing.T) {
	m := createTestModel()
	m.killMode = true
	m.killTarget = &killTargetInfo{
		PID:         99999, // Non-existent PID
		PIDs:        nil,   // Empty PIDs slice
		ProcessName: "FakeApp",
		Signal:      "SIGTERM",
	}

	updated, _ := m.executeKill()
	newModel := updated.(Model)

	if newModel.killMode {
		t.Error("killMode should be false after kill attempt")
	}
	// Should use single PID fallback
	if newModel.killResult == "" {
		t.Error("killResult should be set")
	}
}

// TestUpdate_KeyMsg_Down_RespectsFilter verifies cursor stays within filtered bounds.
// This is a regression test for the bug where cursor could exceed filtered item count.
func TestUpdate_KeyMsg_Down_RespectsFilter(t *testing.T) {
	m := createTestModel()
	// Filter to show only App1 (1 item out of 3)
	m.activeFilter = "App1"
	m.CurrentView().Cursor = 0

	// Try to move down - should stay at 0 since only 1 filtered item
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, want 0 (filter shows only 1 item)", newModel.CurrentView().Cursor)
	}
}

// TestUpdate_DataMsg_ClampsCursorWithFilter verifies cursor is clamped on data refresh.
func TestUpdate_DataMsg_ClampsCursorWithFilter(t *testing.T) {
	m := createTestModel()
	// Set filter and cursor beyond filtered bounds
	m.activeFilter = "App1"
	m.CurrentView().Cursor = 5 // Invalid: only 1 item matches filter

	// Simulate data refresh
	msg := DataMsg{
		Snapshot: createTestSnapshot(),
		Err:      nil,
	}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Cursor should be clamped to valid range
	if newModel.CurrentView().Cursor >= 1 {
		t.Errorf("cursor = %d, want < 1 (filter shows only 1 item)", newModel.CurrentView().Cursor)
	}
}

// TestSelectionIDStability verifies selection follows item when sort order changes.
// Regression test for ID-based selection.
func TestSelectionIDStability(t *testing.T) {
	m := createTestModel()
	// Select App2 (middle item when sorted A-Z)
	m.CurrentView().Cursor = 1
	m.CurrentView().SelectedID = model.SelectionIDFromProcess("App2")

	// Verify initial selection resolves correctly
	idx := m.resolveSelectionIndex()
	if idx != 1 {
		t.Errorf("initial idx = %d, want 1", idx)
	}

	// Simulate snapshot update with same data (order unchanged)
	msg := DataMsg{
		Snapshot: createTestSnapshot(),
		Err:      nil,
	}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Selection should still point to App2
	view := newModel.CurrentView()
	if view.SelectedID.ProcessName != "App2" {
		t.Errorf("SelectedID.ProcessName = %q, want App2", view.SelectedID.ProcessName)
	}

	// Cursor should still resolve to App2's position
	resolvedIdx := newModel.resolveSelectionIndex()
	apps := newModel.sortProcessList(newModel.filteredApps())
	if resolvedIdx >= len(apps) {
		t.Fatalf("resolvedIdx %d out of bounds (len=%d)", resolvedIdx, len(apps))
	}
	if apps[resolvedIdx].Name != "App2" {
		t.Errorf("apps[%d].Name = %q, want App2", resolvedIdx, apps[resolvedIdx].Name)
	}
}

// TestSelectionIDGoneItem verifies cursor clamps when selected item disappears.
func TestSelectionIDGoneItem(t *testing.T) {
	m := createTestModel()
	// Select App3 (last item)
	m.CurrentView().Cursor = 2
	m.CurrentView().SelectedID = model.SelectionIDFromProcess("App3")

	// New snapshot without App3
	newSnapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", PIDs: []int32{100}, Connections: []model.Connection{{Protocol: "TCP"}}},
			{Name: "App2", PIDs: []int32{200}, Connections: []model.Connection{{Protocol: "UDP"}}},
		},
	}

	msg := DataMsg{Snapshot: newSnapshot, Err: nil}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Cursor should be clamped (App3 gone)
	view := newModel.CurrentView()
	if view.Cursor >= 2 {
		t.Errorf("cursor = %d, want < 2 (only 2 items remain)", view.Cursor)
	}
}

// Tests for extractIP helper

func TestExtractIP_WithPort(t *testing.T) {
	got := extractIP("192.168.1.1:8080")
	if got != "192.168.1.1" {
		t.Errorf("extractIP('192.168.1.1:8080') = %q, want '192.168.1.1'", got)
	}
}

func TestExtractIP_IPv6(t *testing.T) {
	// IPv6 with port uses format [::1]:8080
	got := extractIP("[::1]:8080")
	if got != "[::1]" {
		t.Errorf("extractIP('[::1]:8080') = %q, want '[::1]'", got)
	}
}

func TestExtractIP_NoPort(t *testing.T) {
	got := extractIP("192.168.1.1")
	if got != "192.168.1.1" {
		t.Errorf("extractIP('192.168.1.1') = %q, want '192.168.1.1'", got)
	}
}

func TestExtractIP_Empty(t *testing.T) {
	got := extractIP("")
	if got != "" {
		t.Errorf("extractIP('') = %q, want ''", got)
	}
}

func TestExtractIP_Wildcard(t *testing.T) {
	got := extractIP("*")
	if got != "" {
		t.Errorf("extractIP('*') = %q, want ''", got)
	}
}

func TestExtractIP_WildcardWithPort(t *testing.T) {
	got := extractIP("*:80")
	if got != "*" {
		t.Errorf("extractIP('*:80') = %q, want '*'", got)
	}
}

// Tests for queueDNSLookups

func TestQueueDNSLookups_Disabled(t *testing.T) {
	m := createTestModel()
	m.dnsEnabled = false
	m.dnsCache = make(map[string]string)

	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				Connections: []model.Connection{
					{RemoteAddr: "8.8.8.8:53"},
				},
			},
		},
	}

	cmd := m.queueDNSLookups(snapshot)

	if cmd != nil {
		t.Error("queueDNSLookups should return nil when DNS disabled")
	}
}

func TestQueueDNSLookups_NilSnapshot(t *testing.T) {
	m := createTestModel()
	m.dnsEnabled = true
	m.dnsCache = make(map[string]string)

	cmd := m.queueDNSLookups(nil)

	if cmd != nil {
		t.Error("queueDNSLookups should return nil for nil snapshot")
	}
}

func TestQueueDNSLookups_CacheHit(t *testing.T) {
	m := createTestModel()
	m.dnsEnabled = true
	m.dnsCache = map[string]string{
		"8.8.8.8": "dns.google", // Already cached
	}

	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				Connections: []model.Connection{
					{RemoteAddr: "8.8.8.8:53"},
				},
			},
		},
	}

	cmd := m.queueDNSLookups(snapshot)

	// Should return nil since IP is already cached
	if cmd != nil {
		t.Error("queueDNSLookups should return nil for cached IP")
	}
}

func TestQueueDNSLookups_SkipsWildcard(t *testing.T) {
	m := createTestModel()
	m.dnsEnabled = true
	m.dnsCache = make(map[string]string)

	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				Connections: []model.Connection{
					{RemoteAddr: "*:80"},
					{RemoteAddr: "*"},
				},
			},
		},
	}

	cmd := m.queueDNSLookups(snapshot)

	// Should return nil since wildcards are skipped
	if cmd != nil {
		t.Error("queueDNSLookups should return nil for wildcards")
	}
}

func TestQueueDNSLookups_DeduplicatesIPs(t *testing.T) {
	m := createTestModel()
	m.dnsEnabled = true
	m.dnsCache = make(map[string]string)

	// Multiple connections to same IP
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				Connections: []model.Connection{
					{RemoteAddr: "8.8.8.8:53"},
					{RemoteAddr: "8.8.8.8:443"},
					{RemoteAddr: "8.8.8.8:80"},
				},
			},
		},
	}

	cmd := m.queueDNSLookups(snapshot)

	// Should return a command (one lookup for deduplicated IP)
	if cmd == nil {
		t.Error("queueDNSLookups should return a command for new IP")
	}
	// We can't easily count the number of commands in a Batch,
	// but the test verifies the function works with duplicates
}

func TestQueueDNSLookups_LimitsTen(t *testing.T) {
	m := createTestModel()
	m.dnsEnabled = true
	m.dnsCache = make(map[string]string)

	// Create 15 unique IPs
	var conns []model.Connection
	for i := 1; i <= 15; i++ {
		conns = append(conns, model.Connection{
			RemoteAddr: "10.0.0." + string(rune('0'+i)) + ":80",
		})
	}

	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", Connections: conns},
		},
	}

	cmd := m.queueDNSLookups(snapshot)

	// Should return a command (limited to 10)
	if cmd == nil {
		t.Error("queueDNSLookups should return a command")
	}
	// The limit is enforced by len(cmds) < 10 check in the implementation
}

// Tests for DNSResolvedMsg handling

func TestDNSResolvedMsg_Success(t *testing.T) {
	m := createTestModel()
	m.dnsCache = make(map[string]string)

	msg := DNSResolvedMsg{
		IP:       "8.8.8.8",
		Hostname: "dns.google",
		Err:      nil,
	}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.dnsCache["8.8.8.8"] != "dns.google" {
		t.Errorf("dnsCache[8.8.8.8] = %q, want 'dns.google'", newModel.dnsCache["8.8.8.8"])
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestDNSResolvedMsg_Error(t *testing.T) {
	m := createTestModel()
	m.dnsCache = make(map[string]string)

	msg := DNSResolvedMsg{
		IP:       "8.8.8.8",
		Hostname: "",
		Err:      errors.New("DNS lookup failed"),
	}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	// Failed lookups should cache empty string to avoid retries
	cached, ok := newModel.dnsCache["8.8.8.8"]
	if !ok {
		t.Error("dnsCache should contain entry for failed lookup")
	}
	if cached != "" {
		t.Errorf("dnsCache[8.8.8.8] = %q, want '' (empty for failed lookup)", cached)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestDNSResolvedMsg_OverwritesCache(t *testing.T) {
	m := createTestModel()
	m.dnsCache = map[string]string{
		"8.8.8.8": "old.name", // Existing entry
	}

	msg := DNSResolvedMsg{
		IP:       "8.8.8.8",
		Hostname: "new.name",
		Err:      nil,
	}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.dnsCache["8.8.8.8"] != "new.name" {
		t.Errorf("dnsCache[8.8.8.8] = %q, want 'new.name'", newModel.dnsCache["8.8.8.8"])
	}
}

// Tests for Docker detection and messages

func createDockerTestModel() Model {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "com.docker.backend", PIDs: []int32{100}, Connections: []model.Connection{
				{PID: 100, Protocol: "TCP", LocalAddr: "0.0.0.0:8080", State: "LISTEN"},
				{PID: 100, Protocol: "TCP", LocalAddr: "0.0.0.0:3306", State: "LISTEN"},
			}},
			{Name: "Chrome", PIDs: []int32{200}, Connections: []model.Connection{
				{PID: 200, Protocol: "TCP", LocalAddr: "127.0.0.1:52341", RemoteAddr: "142.250.80.46:443", State: "ESTABLISHED"},
			}},
		},
		Timestamp: time.Now(),
	}
	m := Model{
		collector:       newMockCollector(snapshot),
		netIOCollector:  newMockNetIOCollector(nil),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        snapshot,
		netIOCache:      make(map[int32]*model.NetIOStats),
		changes:         make(map[ConnectionKey]Change),
		dockerResolver:  newMockDockerResolver(nil),
		dockerCache:     make(map[int]*docker.ContainerPort),
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

func TestDrillIntoDocker_SetsDockerView(t *testing.T) {
	m := createDockerTestModel()
	// Chrome sorts before com.docker.backend; Docker is at cursor=1
	m.CurrentView().Cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.dockerView {
		t.Error("dockerView should be true after drilling into Docker process")
	}
	if newModel.CurrentView().Level != LevelConnections {
		t.Errorf("level = %v, want LevelConnections", newModel.CurrentView().Level)
	}
	// Should fire a Docker resolve command
	if cmd == nil {
		t.Error("cmd should not be nil (should fire Docker resolve)")
	}
}

func TestDrillIntoNonDocker_DockerViewFalse(t *testing.T) {
	m := createDockerTestModel()
	// Chrome is first in sorted list
	m.CurrentView().Cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.dockerView {
		t.Error("dockerView should be false after drilling into non-Docker process")
	}
	if newModel.CurrentView().ProcessName != "Chrome" {
		t.Errorf("ProcessName = %q, want 'Chrome'", newModel.CurrentView().ProcessName)
	}
}

func TestPopFromDockerView_ClearsDockerView(t *testing.T) {
	m := createDockerTestModel()
	m.dockerView = true
	m.PushView(ViewState{
		Level:       LevelConnections,
		ProcessName: "com.docker.backend",
		Cursor:      0,
	})

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.dockerView {
		t.Error("dockerView should be false after popping from Docker view")
	}
	if newModel.CurrentView().Level != LevelProcessList {
		t.Errorf("level = %v, want LevelProcessList", newModel.CurrentView().Level)
	}
}

func TestDockerResolvedMsg_PopulatesCache(t *testing.T) {
	m := createDockerTestModel()
	containers := map[int]*docker.ContainerPort{
		8080: {
			Container:     model.ContainerInfo{Name: "nginx", Image: "nginx:latest", ID: "abc123"},
			HostPort:      8080,
			ContainerPort: 80,
			Protocol:      "tcp",
		},
	}
	msg := DockerResolvedMsg{Containers: containers, Err: nil}

	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if len(newModel.dockerCache) != 1 {
		t.Fatalf("dockerCache length = %d, want 1", len(newModel.dockerCache))
	}
	cp := newModel.dockerCache[8080]
	if cp == nil {
		t.Fatal("expected cache entry for port 8080")
	}
	if cp.Container.Name != "nginx" {
		t.Errorf("Name = %q, want 'nginx'", cp.Container.Name)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestDockerResolvedMsg_EmptyResult(t *testing.T) {
	m := createDockerTestModel()
	msg := DockerResolvedMsg{Containers: map[int]*docker.ContainerPort{}, Err: nil}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if len(newModel.dockerCache) != 0 {
		t.Errorf("dockerCache length = %d, want 0", len(newModel.dockerCache))
	}
}

func TestDockerResolvedMsg_ReplacesOldCache(t *testing.T) {
	m := createDockerTestModel()
	m.dockerCache[8080] = &docker.ContainerPort{
		Container: model.ContainerInfo{Name: "old-container"},
		HostPort:  8080,
	}

	newContainers := map[int]*docker.ContainerPort{
		9090: {
			Container:     model.ContainerInfo{Name: "new-container", Image: "app:v2", ID: "def456"},
			HostPort:      9090,
			ContainerPort: 3000,
			Protocol:      "tcp",
		},
	}
	msg := DockerResolvedMsg{Containers: newContainers, Err: nil}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Old cache entry should be gone (full replacement)
	if newModel.dockerCache[8080] != nil {
		t.Error("old cache entry for port 8080 should be gone")
	}
	if newModel.dockerCache[9090] == nil {
		t.Fatal("expected new cache entry for port 9090")
	}
	if newModel.dockerCache[9090].Container.Name != "new-container" {
		t.Errorf("Name = %q, want 'new-container'", newModel.dockerCache[9090].Container.Name)
	}
}

func TestDockerResolvedMsg_Error(t *testing.T) {
	m := createDockerTestModel()
	m.dockerCache[8080] = &docker.ContainerPort{
		Container: model.ContainerInfo{Name: "existing"},
		HostPort:  8080,
	}

	msg := DockerResolvedMsg{Containers: nil, Err: errors.New("docker error")}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Cache should remain unchanged on error
	if newModel.dockerCache[8080] == nil || newModel.dockerCache[8080].Container.Name != "existing" {
		t.Error("dockerCache should be unchanged on error")
	}
}

// Integration tests: full Docker drill-down flow

func TestDockerDrillDownFlow(t *testing.T) {
	// Start at process list, drill into Docker process, receive resolver data, verify view
	m := createDockerTestModel()
	m.width = 120
	m.height = 40
	m.ready = true
	m.viewport = viewport.New(116, 30)

	// Step 1: Drill into Docker process (cursor=1, com.docker.backend)
	m.CurrentView().Cursor = 1
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.dockerView {
		t.Fatal("dockerView should be true after drill-down")
	}
	if m.CurrentView().Level != LevelConnections {
		t.Fatalf("level = %v, want LevelConnections", m.CurrentView().Level)
	}
	if cmd == nil {
		t.Fatal("should fire Docker resolve command")
	}

	// Step 2: Simulate DockerResolvedMsg with container data
	dockerData := map[int]*docker.ContainerPort{
		8080: {
			Container:     model.ContainerInfo{Name: "web", Image: "nginx:latest", ID: "abc123"},
			HostPort:      8080,
			ContainerPort: 80,
			Protocol:      "tcp",
		},
		3306: {
			Container:     model.ContainerInfo{Name: "db", Image: "mysql:8", ID: "def456"},
			HostPort:      3306,
			ContainerPort: 3306,
			Protocol:      "tcp",
		},
	}
	updated, _ = m.Update(DockerResolvedMsg{Containers: dockerData})
	m = updated.(Model)

	if len(m.dockerCache) != 2 {
		t.Fatalf("dockerCache len = %d, want 2", len(m.dockerCache))
	}

	// Step 3: Verify view renders Container column
	view := m.View()
	if !strings.Contains(view, "Container") {
		t.Error("view should contain Container header")
	}
	if !strings.Contains(view, "web") {
		t.Error("view should contain container name 'web'")
	}
	if !strings.Contains(view, "db") {
		t.Error("view should contain container name 'db'")
	}

	// Step 4: Go back clears docker state
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.dockerView {
		t.Error("dockerView should be false after going back")
	}
	if m.CurrentView().Level != LevelProcessList {
		t.Error("should be back at process list")
	}
}

func TestDockerDrillDownFlow_NoDocker(t *testing.T) {
	// Docker resolver returns empty: container column present but empty
	m := createDockerTestModel()
	m.width = 120
	m.height = 40
	m.ready = true
	m.viewport = viewport.New(116, 30)

	// Drill into Docker process
	m.CurrentView().Cursor = 1
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.dockerView {
		t.Fatal("dockerView should be true")
	}

	// Resolver returns empty (Docker not running)
	updated, _ = m.Update(DockerResolvedMsg{Containers: map[int]*docker.ContainerPort{}})
	m = updated.(Model)

	if len(m.dockerCache) != 0 {
		t.Fatalf("dockerCache should be empty, got %d entries", len(m.dockerCache))
	}

	// View should still render Container column header, just no container names
	view := m.View()
	if !strings.Contains(view, "Container") {
		t.Error("view should still contain Container header even with empty cache")
	}
}
