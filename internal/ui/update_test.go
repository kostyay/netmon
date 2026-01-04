package ui

import (
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
	return Model{
		collector:       newMockCollector(snapshot),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        snapshot,
		cursor:          0,
		expandedApps:    make(map[string]bool),
	}
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
	m.cursor = 2

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 1 {
		t.Errorf("cursor = %d, want 1", newModel.cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Up_K(t *testing.T) {
	m := createTestModel()
	m.cursor = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 0 {
		t.Errorf("cursor = %d, want 0", newModel.cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Up_AtTop(t *testing.T) {
	m := createTestModel()
	m.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should not go negative)", newModel.cursor)
	}
}

func TestUpdate_KeyMsg_Down(t *testing.T) {
	m := createTestModel()
	m.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 1 {
		t.Errorf("cursor = %d, want 1", newModel.cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Down_J(t *testing.T) {
	m := createTestModel()
	m.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 1 {
		t.Errorf("cursor = %d, want 1", newModel.cursor)
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Down_AtBottom(t *testing.T) {
	m := createTestModel()
	m.cursor = 2 // Last item (3 items total)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (should not exceed bounds)", newModel.cursor)
	}
}

func TestUpdate_KeyMsg_Left_Collapse(t *testing.T) {
	m := createTestModel()
	m.expandedApps["App1"] = true

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if newModel.expandedApps["App1"] {
		t.Error("Application should be collapsed after left key")
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Left_H(t *testing.T) {
	m := createTestModel()
	m.expandedApps["App1"] = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.expandedApps["App1"] {
		t.Error("Application should be collapsed after 'h' key")
	}
}

func TestUpdate_KeyMsg_Right_Expand(t *testing.T) {
	m := createTestModel()
	m.expandedApps["App1"] = false

	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, cmd := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.expandedApps["App1"] {
		t.Error("Application should be expanded after right key")
	}
	if cmd != nil {
		t.Error("cmd should be nil")
	}
}

func TestUpdate_KeyMsg_Right_L(t *testing.T) {
	m := createTestModel()
	m.expandedApps["App1"] = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.expandedApps["App1"] {
		t.Error("Application should be expanded after 'l' key")
	}
}

func TestUpdate_KeyMsg_Enter_Expand(t *testing.T) {
	m := createTestModel()
	m.expandedApps["App1"] = false

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.expandedApps["App1"] {
		t.Error("Application should be expanded after Enter key")
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
	m.cursor = 5 // Out of bounds

	smallSnapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", PIDs: []int32{100}, Connections: []model.Connection{{Protocol: "TCP"}}},
		},
		Timestamp: time.Now(),
	}
	msg := DataMsg{Snapshot: smallSnapshot, Err: nil}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 0 {
		t.Errorf("cursor = %d, should be adjusted to 0 (max index)", newModel.cursor)
	}
}

func TestExpandedState_PreservedAcrossDataUpdates(t *testing.T) {
	m := createTestModel()
	m.expandedApps["App1"] = true
	m.expandedApps["App2"] = false
	m.expandedApps["App3"] = true

	// Simulate new data coming in
	newSnapshot := createTestSnapshot()
	msg := DataMsg{Snapshot: newSnapshot, Err: nil}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Expanded state should be preserved in the map
	if !newModel.expandedApps["App1"] {
		t.Error("App1 should preserve expanded state")
	}
	if newModel.expandedApps["App2"] {
		t.Error("App2 should preserve non-expanded state")
	}
	if !newModel.expandedApps["App3"] {
		t.Error("App3 should preserve expanded state")
	}
}

func TestUpdate_NilSnapshot_Down(t *testing.T) {
	m := Model{
		collector:       newMockCollector(nil),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        nil,
		cursor:          0,
		expandedApps:    make(map[string]bool),
	}

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 0 {
		t.Errorf("cursor = %d, should stay at 0 with nil snapshot", newModel.cursor)
	}
}

func TestUpdate_NilSnapshot_Collapse(t *testing.T) {
	m := Model{
		collector:       newMockCollector(nil),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        nil,
		cursor:          0,
		expandedApps:    make(map[string]bool),
	}

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should not panic
	if newModel.snapshot != nil {
		t.Error("snapshot should remain nil")
	}
}

func TestUpdate_NilSnapshot_Expand(t *testing.T) {
	m := Model{
		collector:       newMockCollector(nil),
		refreshInterval: DefaultRefreshInterval,
		snapshot:        nil,
		cursor:          0,
		expandedApps:    make(map[string]bool),
	}

	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should not panic
	if newModel.snapshot != nil {
		t.Error("snapshot should remain nil")
	}
}

func TestViewToggle(t *testing.T) {
	m := createTestModel()

	// Initially in grouped view
	if m.viewMode != ViewGrouped {
		t.Errorf("initial viewMode = %v, want ViewGrouped", m.viewMode)
	}

	// Press 'v' to switch to table view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.viewMode != ViewTable {
		t.Errorf("viewMode = %v, want ViewTable after 'v'", newModel.viewMode)
	}

	// Press 'v' again to switch back
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if newModel.viewMode != ViewGrouped {
		t.Errorf("viewMode = %v, want ViewGrouped after second 'v'", newModel.viewMode)
	}
}

func TestColumnNavigation(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.selectedColumn = SortPID // Start at first column

	// Move right
	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.selectedColumn != SortProcess {
		t.Errorf("selectedColumn = %v, want SortProcess after right", newModel.selectedColumn)
	}

	// Move right again
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if newModel.selectedColumn != SortProtocol {
		t.Errorf("selectedColumn = %v, want SortProtocol after second right", newModel.selectedColumn)
	}

	// Move left
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if newModel.selectedColumn != SortProcess {
		t.Errorf("selectedColumn = %v, want SortProcess after left", newModel.selectedColumn)
	}
}

func TestColumnNavigation_Bounds(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable

	// At first column, left should not go negative
	m.selectedColumn = SortPID
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.selectedColumn != SortPID {
		t.Errorf("selectedColumn = %v, should stay at SortPID at left bound", newModel.selectedColumn)
	}

	// At last column, right should not exceed
	m.selectedColumn = SortState
	msg = tea.KeyMsg{Type: tea.KeyRight}
	updated, _ = m.Update(msg)
	newModel = updated.(Model)

	if newModel.selectedColumn != SortState {
		t.Errorf("selectedColumn = %v, should stay at SortState at right bound", newModel.selectedColumn)
	}
}

func TestColumnNavigation_WithHLKeys(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.selectedColumn = SortProcess

	// 'l' moves right
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.selectedColumn != SortProtocol {
		t.Errorf("selectedColumn = %v, want SortProtocol after 'l'", newModel.selectedColumn)
	}

	// 'h' moves left
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if newModel.selectedColumn != SortProcess {
		t.Errorf("selectedColumn = %v, want SortProcess after 'h'", newModel.selectedColumn)
	}
}

func TestSortOnEnter(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.selectedColumn = SortProtocol
	m.sortColumn = SortProcess
	m.sortAscending = true

	// Press Enter to sort by selected column
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.sortColumn != SortProtocol {
		t.Errorf("sortColumn = %v, want SortProtocol after Enter", newModel.sortColumn)
	}
	if !newModel.sortAscending {
		t.Error("sortAscending should be true for new column")
	}
}

func TestSortOnSpace(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.selectedColumn = SortLocal
	m.sortColumn = SortProcess
	m.sortAscending = true

	// Press Space to sort by selected column
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.sortColumn != SortLocal {
		t.Errorf("sortColumn = %v, want SortLocal after Space", newModel.sortColumn)
	}
}

func TestSortToggleDirection(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.selectedColumn = SortProcess
	m.sortColumn = SortProcess // Same column
	m.sortAscending = true

	// Press Enter to toggle direction
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.sortAscending {
		t.Error("sortAscending should be false after toggling")
	}

	// Toggle again
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if !newModel.sortAscending {
		t.Error("sortAscending should be true after toggling back")
	}
}

func TestTableNavigation(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.tableCursor = 0

	// Down in table view
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.tableCursor != 1 {
		t.Errorf("tableCursor = %d, want 1 after down", newModel.tableCursor)
	}

	// Up in table view
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if newModel.tableCursor != 0 {
		t.Errorf("tableCursor = %d, want 0 after up", newModel.tableCursor)
	}

	// Verify grouped cursor unchanged
	if newModel.cursor != 0 {
		t.Errorf("grouped cursor = %d, should be unchanged", newModel.cursor)
	}
}

func TestExpandCollapseIgnoredInTableView(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.expandedApps["App1"] = true
	m.selectedColumn = SortProcess

	// Left in table view should move column selection, not collapse
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Expand state should be unchanged
	if !newModel.expandedApps["App1"] {
		t.Error("left key should not affect expand state in table view")
	}
	// Column should have moved
	if newModel.selectedColumn != SortPID {
		t.Errorf("selectedColumn = %v, want SortPID after left in table view", newModel.selectedColumn)
	}

	// Right in table view should move column selection, not expand
	m.expandedApps["App2"] = false
	m.selectedColumn = SortProcess
	msg = tea.KeyMsg{Type: tea.KeyRight}
	updated, _ = m.Update(msg)
	newModel = updated.(Model)

	// Expand state should be unchanged
	if newModel.expandedApps["App2"] {
		t.Error("right key should not affect expand state in table view")
	}
	// Column should have moved
	if newModel.selectedColumn != SortProtocol {
		t.Errorf("selectedColumn = %v, want SortProtocol after right in table view", newModel.selectedColumn)
	}
}

func TestCursorPersistence(t *testing.T) {
	m := createTestModel()
	m.cursor = 2
	m.tableCursor = 1

	// Switch to table view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.cursor != 2 {
		t.Errorf("grouped cursor = %d, should be preserved", newModel.cursor)
	}
	if newModel.tableCursor != 1 {
		t.Errorf("table cursor = %d, should be preserved", newModel.tableCursor)
	}

	// Switch back to grouped view
	updated, _ = newModel.Update(msg)
	newModel = updated.(Model)

	if newModel.cursor != 2 {
		t.Errorf("grouped cursor = %d, should be preserved after switch back", newModel.cursor)
	}
}

func TestTableNavigation_DownAtBottom(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable

	flatConns := m.flattenConnections()
	m.tableCursor = len(flatConns) - 1 // Last connection

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if newModel.tableCursor != len(flatConns)-1 {
		t.Errorf("tableCursor = %d, should stay at %d (last item)",
			newModel.tableCursor, len(flatConns)-1)
	}
}

func TestUpdate_DataMsg_TableCursorBounds(t *testing.T) {
	m := createTestModel()
	m.viewMode = ViewTable
	m.tableCursor = 10 // Beyond what smaller snapshot would have

	// Smaller snapshot with only 1 connection
	smallSnapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", PIDs: []int32{100}, Connections: []model.Connection{{Protocol: "TCP"}}},
		},
	}
	msg := DataMsg{Snapshot: smallSnapshot, Err: nil}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	flatConns := newModel.flattenConnections()
	if newModel.tableCursor >= len(flatConns) {
		t.Errorf("tableCursor = %d, should be < %d", newModel.tableCursor, len(flatConns))
	}
}
