package ui

import (
	"testing"

	"github.com/kostyay/netmon/internal/model"
)

// Tests for findProcessIndex

func TestFindProcessIndex_Found(t *testing.T) {
	m := createTestModel()

	idx := m.findProcessIndex("App2")

	if idx != 1 {
		t.Errorf("findProcessIndex('App2') = %d, want 1", idx)
	}
}

func TestFindProcessIndex_NotFound(t *testing.T) {
	m := createTestModel()

	idx := m.findProcessIndex("NonExistent")

	if idx != -1 {
		t.Errorf("findProcessIndex('NonExistent') = %d, want -1", idx)
	}
}

func TestFindProcessIndex_EmptyName(t *testing.T) {
	m := createTestModel()

	idx := m.findProcessIndex("")

	if idx != -1 {
		t.Errorf("findProcessIndex('') = %d, want -1", idx)
	}
}

func TestFindProcessIndex_WithFilter(t *testing.T) {
	m := createTestModel()
	m.activeFilter = "App1" // Only App1 visible

	idx := m.findProcessIndex("App1")
	if idx != 0 {
		t.Errorf("findProcessIndex('App1') with filter = %d, want 0", idx)
	}

	idx = m.findProcessIndex("App2") // Filtered out
	if idx != -1 {
		t.Errorf("findProcessIndex('App2') with filter = %d, want -1 (filtered out)", idx)
	}
}

// Tests for findConnectionIndex

func TestFindConnectionIndex_NilKey(t *testing.T) {
	m := createTestModel()
	m.PushView(ViewState{Level: LevelConnections, ProcessName: "App1"})

	idx := m.findConnectionIndex(nil)

	if idx != -1 {
		t.Errorf("findConnectionIndex(nil) = %d, want -1", idx)
	}
}

func TestFindConnectionIndex_ConnectionsLevel(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443"},
						{PID: 100, LocalAddr: "127.0.0.1:81", RemoteAddr: "10.0.0.2:443"},
						{PID: 100, LocalAddr: "127.0.0.1:82", RemoteAddr: "10.0.0.3:443"},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:         LevelConnections,
			ProcessName:   "TestApp",
			SortColumn:    SortLocal,
			SortAscending: true,
		}},
		changes: make(map[ConnectionKey]Change),
	}

	key := &model.ConnectionKey{
		ProcessName: "TestApp",
		LocalAddr:   "127.0.0.1:81",
		RemoteAddr:  "10.0.0.2:443",
	}

	idx := m.findConnectionIndex(key)

	if idx != 1 {
		t.Errorf("findConnectionIndex() = %d, want 1", idx)
	}
}

func TestFindConnectionIndex_AllConnectionsLevel(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443"},
					},
				},
				{
					Name: "App2",
					PIDs: []int32{200},
					Connections: []model.Connection{
						{PID: 200, LocalAddr: "127.0.0.1:81", RemoteAddr: "10.0.0.2:443"},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortProcess,
			SortAscending: true,
		}},
		changes: make(map[ConnectionKey]Change),
	}

	key := &model.ConnectionKey{
		ProcessName: "App2",
		LocalAddr:   "127.0.0.1:81",
		RemoteAddr:  "10.0.0.2:443",
	}

	idx := m.findConnectionIndex(key)

	if idx != 1 {
		t.Errorf("findConnectionIndex() = %d, want 1", idx)
	}
}

func TestFindConnectionIndex_NotFound(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443"},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
		}},
		changes: make(map[ConnectionKey]Change),
	}

	key := &model.ConnectionKey{
		ProcessName: "TestApp",
		LocalAddr:   "192.168.1.1:99", // Non-existent
		RemoteAddr:  "10.0.0.9:443",
	}

	idx := m.findConnectionIndex(key)

	if idx != -1 {
		t.Errorf("findConnectionIndex() = %d, want -1 (not found)", idx)
	}
}

// Tests for getProcessNameByPID

func TestGetProcessNameByPID_Found(t *testing.T) {
	m := createTestModel()

	name := m.getProcessNameByPID(100)

	if name != "App1" {
		t.Errorf("getProcessNameByPID(100) = %q, want 'App1'", name)
	}
}

func TestGetProcessNameByPID_NotFound(t *testing.T) {
	m := createTestModel()

	name := m.getProcessNameByPID(999)

	if name != "" {
		t.Errorf("getProcessNameByPID(999) = %q, want '' (not found)", name)
	}
}

func TestGetProcessNameByPID_NilSnapshot(t *testing.T) {
	m := Model{snapshot: nil}

	name := m.getProcessNameByPID(100)

	if name != "" {
		t.Errorf("getProcessNameByPID with nil snapshot = %q, want ''", name)
	}
}

// Tests for connectionMatchesKey

func TestConnectionMatchesKey_Match(t *testing.T) {
	m := createTestModel()

	conn := model.Connection{
		PID:        100,
		LocalAddr:  "127.0.0.1:80",
		RemoteAddr: "10.0.0.1:443",
	}
	key := &model.ConnectionKey{
		ProcessName: "App1",
		LocalAddr:   "127.0.0.1:80",
		RemoteAddr:  "10.0.0.1:443",
	}

	if !m.connectionMatchesKey(conn, key) {
		t.Error("connectionMatchesKey should return true for matching connection")
	}
}

func TestConnectionMatchesKey_NoMatch(t *testing.T) {
	m := createTestModel()

	conn := model.Connection{
		PID:        100,
		LocalAddr:  "127.0.0.1:80",
		RemoteAddr: "10.0.0.1:443",
	}
	key := &model.ConnectionKey{
		ProcessName: "App1",
		LocalAddr:   "127.0.0.1:99", // Different
		RemoteAddr:  "10.0.0.1:443",
	}

	if m.connectionMatchesKey(conn, key) {
		t.Error("connectionMatchesKey should return false for non-matching connection")
	}
}

func TestConnectionMatchesKey_NilKey(t *testing.T) {
	m := createTestModel()

	conn := model.Connection{PID: 100}

	if m.connectionMatchesKey(conn, nil) {
		t.Error("connectionMatchesKey should return false for nil key")
	}
}

// Tests for resolveSelectionIndex

func TestResolveSelectionIndex_NilView(t *testing.T) {
	m := Model{stack: nil}

	idx := m.resolveSelectionIndex()

	if idx != 0 {
		t.Errorf("resolveSelectionIndex() with nil stack = %d, want 0", idx)
	}
}

func TestResolveSelectionIndex_EmptySelectedID(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 2
	m.CurrentView().SelectedID = model.SelectionID{} // Empty

	idx := m.resolveSelectionIndex()

	if idx != 2 {
		t.Errorf("resolveSelectionIndex() with empty SelectedID = %d, want 2 (cursor fallback)", idx)
	}
}

func TestResolveSelectionIndex_ProcessListLevel(t *testing.T) {
	m := createTestModel()
	m.CurrentView().SelectedID = model.SelectionIDFromProcess("App2")

	idx := m.resolveSelectionIndex()

	if idx != 1 {
		t.Errorf("resolveSelectionIndex() = %d, want 1 (App2 position)", idx)
	}
}

func TestResolveSelectionIndex_ProcessNotFound(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 0
	m.CurrentView().SelectedID = model.SelectionIDFromProcess("NonExistent")

	idx := m.resolveSelectionIndex()

	if idx != 0 {
		t.Errorf("resolveSelectionIndex() = %d, want 0 (cursor fallback for not found)", idx)
	}
}

// Tests for validateSelection

func TestValidateSelection_NilView(t *testing.T) {
	m := Model{stack: nil, snapshot: createTestSnapshot()}

	// Should not panic
	m.validateSelection()
}

func TestValidateSelection_NilSnapshot(t *testing.T) {
	m := createTestModel()
	m.snapshot = nil

	// Should not panic
	m.validateSelection()
}

func TestValidateSelection_EmptyApps(t *testing.T) {
	m := createTestModel()
	m.snapshot = &model.NetworkSnapshot{Applications: []model.Application{}}
	m.CurrentView().Cursor = 5

	m.validateSelection()

	if m.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, want 0 (empty list)", m.CurrentView().Cursor)
	}
	if m.CurrentView().SelectedID.ProcessName != "" {
		t.Errorf("SelectedID should be empty for empty list")
	}
}

func TestValidateSelection_ClampsCursor(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 10 // Out of bounds (only 3 apps)

	m.validateSelection()

	if m.CurrentView().Cursor != 2 {
		t.Errorf("cursor = %d, want 2 (clamped to max index)", m.CurrentView().Cursor)
	}
}

func TestValidateSelection_FollowsSelectedID(t *testing.T) {
	m := createTestModel()
	// Sort by PID descending: App3, App2, App1
	m.CurrentView().SortColumn = SortPID
	m.CurrentView().SortAscending = false
	m.CurrentView().Cursor = 0
	m.CurrentView().SelectedID = model.SelectionIDFromProcess("App1")

	m.validateSelection()

	// App1 should be at index 2 when sorted by PID descending
	if m.CurrentView().Cursor != 2 {
		t.Errorf("cursor = %d, want 2 (followed SelectedID)", m.CurrentView().Cursor)
	}
}

func TestValidateSelection_ConnectionsLevel(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, LocalAddr: "127.0.0.1:80"},
						{PID: 100, LocalAddr: "127.0.0.1:81"},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
			Cursor:      5, // Out of bounds
		}},
		changes: make(map[ConnectionKey]Change),
	}

	m.validateSelection()

	if m.CurrentView().Cursor != 1 {
		t.Errorf("cursor = %d, want 1 (clamped for connections)", m.CurrentView().Cursor)
	}
}

func TestValidateSelection_AllConnectionsLevel(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:  LevelAllConnections,
			Cursor: 5, // Out of bounds
		}},
		changes: make(map[ConnectionKey]Change),
	}

	m.validateSelection()

	if m.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, want 0 (only 1 connection)", m.CurrentView().Cursor)
	}
}

// Tests for updateSelectedIDFromCursor

func TestUpdateSelectedIDFromCursor_ProcessList(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 1 // App2

	m.updateSelectedIDFromCursor()

	if m.CurrentView().SelectedID.ProcessName != "App2" {
		t.Errorf("SelectedID.ProcessName = %q, want 'App2'", m.CurrentView().SelectedID.ProcessName)
	}
}

func TestUpdateSelectedIDFromCursor_ConnectionsLevel(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443"},
						{PID: 100, LocalAddr: "127.0.0.1:81", RemoteAddr: "10.0.0.2:443"},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:         LevelConnections,
			ProcessName:   "TestApp",
			Cursor:        1,
			SortColumn:    SortLocal,
			SortAscending: true,
		}},
		changes: make(map[ConnectionKey]Change),
	}

	m.updateSelectedIDFromCursor()

	key := m.CurrentView().SelectedID.ConnectionKey
	if key == nil {
		t.Fatal("ConnectionKey should not be nil")
	}
	if key.LocalAddr != "127.0.0.1:81" {
		t.Errorf("ConnectionKey.LocalAddr = %q, want '127.0.0.1:81'", key.LocalAddr)
	}
}

func TestUpdateSelectedIDFromCursor_AllConnectionsLevel(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443"},
					},
				},
				{
					Name: "App2",
					PIDs: []int32{200},
					Connections: []model.Connection{
						{PID: 200, LocalAddr: "127.0.0.1:81", RemoteAddr: "10.0.0.2:443"},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:         LevelAllConnections,
			Cursor:        1,
			SortColumn:    SortProcess,
			SortAscending: true,
		}},
		changes: make(map[ConnectionKey]Change),
	}

	m.updateSelectedIDFromCursor()

	key := m.CurrentView().SelectedID.ConnectionKey
	if key == nil {
		t.Fatal("ConnectionKey should not be nil")
	}
	if key.ProcessName != "App2" {
		t.Errorf("ConnectionKey.ProcessName = %q, want 'App2'", key.ProcessName)
	}
}

func TestUpdateSelectedIDFromCursor_NilView(t *testing.T) {
	m := Model{stack: nil, snapshot: createTestSnapshot()}

	// Should not panic
	m.updateSelectedIDFromCursor()
}

func TestUpdateSelectedIDFromCursor_NilSnapshot(t *testing.T) {
	m := createTestModel()
	m.snapshot = nil

	// Should not panic
	m.updateSelectedIDFromCursor()
}

func TestUpdateSelectedIDFromCursor_CursorOutOfBounds(t *testing.T) {
	m := createTestModel()
	m.CurrentView().Cursor = 100 // Out of bounds
	originalID := m.CurrentView().SelectedID

	m.updateSelectedIDFromCursor()

	// Should not update when cursor is out of bounds
	if m.CurrentView().SelectedID != originalID {
		t.Error("SelectedID should not change when cursor is out of bounds")
	}
}

// Tests for validateSelection with filters

func TestValidateSelection_WithFilter(t *testing.T) {
	m := createTestModel()
	m.activeFilter = "App1"    // Only App1 visible
	m.CurrentView().Cursor = 5 // Out of bounds for filtered list

	m.validateSelection()

	if m.CurrentView().Cursor != 0 {
		t.Errorf("cursor = %d, want 0 (only 1 item in filtered list)", m.CurrentView().Cursor)
	}
}

func TestFindProcessIndex_SortedByPIDDescending(t *testing.T) {
	m := createTestModel()
	m.CurrentView().SortColumn = SortPID
	m.CurrentView().SortAscending = false

	// When sorted by PID descending: App3 (300), App2 (200), App1 (100)
	idx := m.findProcessIndex("App1")

	if idx != 2 {
		t.Errorf("findProcessIndex('App1') with PID desc sort = %d, want 2", idx)
	}
}
