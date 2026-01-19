package ui

import (
	"testing"

	"github.com/kostyay/netmon/internal/model"
)

// Test compare helpers

func TestCompareInt(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, -1},
		{2, 1, 1},
		{1, 1, 0},
		{0, 0, 0},
		{-1, 1, -1},
	}
	for _, tt := range tests {
		got := compareInt(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCompareInt32(t *testing.T) {
	tests := []struct {
		a, b int32
		want int
	}{
		{1, 2, -1},
		{2, 1, 1},
		{100, 100, 0},
	}
	for _, tt := range tests {
		got := compareInt32(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareInt32(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCompareUint64(t *testing.T) {
	tests := []struct {
		a, b uint64
		want int
	}{
		{1, 2, -1},
		{2, 1, 1},
		{0, 0, 0},
		{1000000, 1000000, 0},
	}
	for _, tt := range tests {
		got := compareUint64(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareUint64(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCompareString(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"a", "b", -1},
		{"b", "a", 1},
		{"abc", "abc", 0},
		{"", "", 0},
		{"", "a", -1},
	}
	for _, tt := range tests {
		got := compareString(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareString(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// Test sortProcessList

func TestSortProcessList_NilView(t *testing.T) {
	m := Model{stack: []ViewState{}} // empty stack = nil view
	apps := []model.Application{{Name: "B"}, {Name: "A"}}

	result := m.sortProcessList(apps)

	// Should return unchanged when view is nil
	if len(result) != 2 || result[0].Name != "B" {
		t.Error("sortProcessList with nil view should return unchanged slice")
	}
}

func TestSortProcessList_ByName_Ascending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortProcess,
			SortAscending: true,
		}},
	}
	apps := []model.Application{
		{Name: "Chrome"},
		{Name: "App"},
		{Name: "Zoom"},
	}

	result := m.sortProcessList(apps)

	if result[0].Name != "App" || result[1].Name != "Chrome" || result[2].Name != "Zoom" {
		t.Errorf("Expected [App, Chrome, Zoom], got [%s, %s, %s]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

func TestSortProcessList_ByName_Descending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortProcess,
			SortAscending: false,
		}},
	}
	apps := []model.Application{
		{Name: "Chrome"},
		{Name: "App"},
		{Name: "Zoom"},
	}

	result := m.sortProcessList(apps)

	if result[0].Name != "Zoom" || result[1].Name != "Chrome" || result[2].Name != "App" {
		t.Errorf("Expected [Zoom, Chrome, App], got [%s, %s, %s]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

func TestSortProcessList_ByPID_Ascending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortPID,
			SortAscending: true,
		}},
	}
	apps := []model.Application{
		{Name: "B", PIDs: []int32{200}},
		{Name: "A", PIDs: []int32{100}},
		{Name: "C", PIDs: []int32{300}},
	}

	result := m.sortProcessList(apps)

	if result[0].PIDs[0] != 100 || result[1].PIDs[0] != 200 || result[2].PIDs[0] != 300 {
		t.Errorf("Expected PIDs [100, 200, 300], got [%d, %d, %d]",
			result[0].PIDs[0], result[1].PIDs[0], result[2].PIDs[0])
	}
}

func TestSortProcessList_ByPID_Descending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortPID,
			SortAscending: false,
		}},
	}
	apps := []model.Application{
		{Name: "B", PIDs: []int32{200}},
		{Name: "A", PIDs: []int32{100}},
		{Name: "C", PIDs: []int32{300}},
	}

	result := m.sortProcessList(apps)

	if result[0].PIDs[0] != 300 || result[1].PIDs[0] != 200 || result[2].PIDs[0] != 100 {
		t.Errorf("Expected PIDs [300, 200, 100], got [%d, %d, %d]",
			result[0].PIDs[0], result[1].PIDs[0], result[2].PIDs[0])
	}
}

func TestSortProcessList_ByPID_EmptyPIDs(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortPID,
			SortAscending: true,
		}},
	}
	apps := []model.Application{
		{Name: "B", PIDs: []int32{200}},
		{Name: "A", PIDs: []int32{}}, // empty PIDs
		{Name: "C", PIDs: []int32{100}},
	}

	result := m.sortProcessList(apps)

	// Empty PIDs treated as 0
	if result[0].Name != "A" {
		t.Errorf("Expected app with empty PIDs first (PID=0), got %s", result[0].Name)
	}
}

func TestSortProcessList_ByConns_Ascending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortConns,
			SortAscending: true,
		}},
	}
	apps := []model.Application{
		{Name: "B", Connections: make([]model.Connection, 5)},
		{Name: "A", Connections: make([]model.Connection, 1)},
		{Name: "C", Connections: make([]model.Connection, 10)},
	}

	result := m.sortProcessList(apps)

	if len(result[0].Connections) != 1 || len(result[1].Connections) != 5 || len(result[2].Connections) != 10 {
		t.Errorf("Expected conns [1, 5, 10], got [%d, %d, %d]",
			len(result[0].Connections), len(result[1].Connections), len(result[2].Connections))
	}
}

func TestSortProcessList_ByEstablished_Descending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortEstablished,
			SortAscending: false,
		}},
	}
	apps := []model.Application{
		{Name: "B", EstablishedCount: 5},
		{Name: "A", EstablishedCount: 10},
		{Name: "C", EstablishedCount: 1},
	}

	result := m.sortProcessList(apps)

	if result[0].EstablishedCount != 10 || result[1].EstablishedCount != 5 || result[2].EstablishedCount != 1 {
		t.Errorf("Expected established [10, 5, 1], got [%d, %d, %d]",
			result[0].EstablishedCount, result[1].EstablishedCount, result[2].EstablishedCount)
	}
}

func TestSortProcessList_ByListen_Ascending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortListen,
			SortAscending: true,
		}},
	}
	apps := []model.Application{
		{Name: "B", ListenCount: 3},
		{Name: "A", ListenCount: 1},
		{Name: "C", ListenCount: 5},
	}

	result := m.sortProcessList(apps)

	if result[0].ListenCount != 1 || result[1].ListenCount != 3 || result[2].ListenCount != 5 {
		t.Errorf("Expected listen [1, 3, 5], got [%d, %d, %d]",
			result[0].ListenCount, result[1].ListenCount, result[2].ListenCount)
	}
}

func TestSortProcessList_ByTX(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortTX,
			SortAscending: true,
		}},
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1000},
			200: {BytesSent: 500},
			300: {BytesSent: 2000},
		},
	}
	apps := []model.Application{
		{Name: "A", PIDs: []int32{100}},
		{Name: "B", PIDs: []int32{200}},
		{Name: "C", PIDs: []int32{300}},
	}

	result := m.sortProcessList(apps)

	if result[0].Name != "B" || result[1].Name != "A" || result[2].Name != "C" {
		t.Errorf("Expected [B(500), A(1000), C(2000)], got [%s, %s, %s]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

func TestSortProcessList_ByRX(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortRX,
			SortAscending: false,
		}},
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesRecv: 1000},
			200: {BytesRecv: 5000},
			300: {BytesRecv: 2000},
		},
	}
	apps := []model.Application{
		{Name: "A", PIDs: []int32{100}},
		{Name: "B", PIDs: []int32{200}},
		{Name: "C", PIDs: []int32{300}},
	}

	result := m.sortProcessList(apps)

	if result[0].Name != "B" || result[1].Name != "C" || result[2].Name != "A" {
		t.Errorf("Expected [B(5000), C(2000), A(1000)], got [%s, %s, %s]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

func TestSortProcessList_StableOrdering(t *testing.T) {
	// When primary keys are equal, should use name as secondary key
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortEstablished,
			SortAscending: true,
		}},
	}
	apps := []model.Application{
		{Name: "Zebra", EstablishedCount: 5},
		{Name: "Apple", EstablishedCount: 5},
		{Name: "Mango", EstablishedCount: 5},
	}

	result := m.sortProcessList(apps)

	// All have same EstablishedCount, should sort by name
	if result[0].Name != "Apple" || result[1].Name != "Mango" || result[2].Name != "Zebra" {
		t.Errorf("Stable ordering failed: expected [Apple, Mango, Zebra], got [%s, %s, %s]",
			result[0].Name, result[1].Name, result[2].Name)
	}
}

func TestSortProcessList_DefaultColumn(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortColumn(999), // invalid column
			SortAscending: true,
		}},
	}
	apps := []model.Application{
		{Name: "B"},
		{Name: "A"},
	}

	result := m.sortProcessList(apps)

	// Default should sort by name
	if result[0].Name != "A" || result[1].Name != "B" {
		t.Errorf("Default sort should use name, got [%s, %s]", result[0].Name, result[1].Name)
	}
}

// Test sortAllConnections

func TestSortAllConnections_NilView(t *testing.T) {
	m := Model{stack: []ViewState{}}
	conns := []connectionWithProcess{
		{ProcessName: "B"},
		{ProcessName: "A"},
	}

	result := m.sortAllConnections(conns)

	if result[0].ProcessName != "B" {
		t.Error("sortAllConnections with nil view should return unchanged")
	}
}

func TestSortAllConnections_ByPID_Ascending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortPID,
			SortAscending: true,
		}},
	}
	conns := []connectionWithProcess{
		{ProcessName: "A", Connection: model.Connection{PID: 200}},
		{ProcessName: "B", Connection: model.Connection{PID: 100}},
		{ProcessName: "C", Connection: model.Connection{PID: 300}},
	}

	result := m.sortAllConnections(conns)

	if result[0].PID != 100 || result[1].PID != 200 || result[2].PID != 300 {
		t.Errorf("Expected PIDs [100, 200, 300], got [%d, %d, %d]",
			result[0].PID, result[1].PID, result[2].PID)
	}
}

func TestSortAllConnections_ByProcess_Descending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortProcess,
			SortAscending: false,
		}},
	}
	conns := []connectionWithProcess{
		{ProcessName: "Chrome"},
		{ProcessName: "App"},
		{ProcessName: "Zoom"},
	}

	result := m.sortAllConnections(conns)

	if result[0].ProcessName != "Zoom" || result[1].ProcessName != "Chrome" || result[2].ProcessName != "App" {
		t.Errorf("Expected [Zoom, Chrome, App], got [%s, %s, %s]",
			result[0].ProcessName, result[1].ProcessName, result[2].ProcessName)
	}
}

func TestSortAllConnections_ByProtocol(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortProtocol,
			SortAscending: true,
		}},
	}
	conns := []connectionWithProcess{
		{ProcessName: "A", Connection: model.Connection{Protocol: model.ProtocolTCP}},
		{ProcessName: "B", Connection: model.Connection{Protocol: model.ProtocolUDP}},
		{ProcessName: "C", Connection: model.Connection{Protocol: model.ProtocolTCP}},
	}

	result := m.sortAllConnections(conns)

	// TCP < UDP alphabetically
	if result[0].Protocol != model.ProtocolTCP || result[2].Protocol != model.ProtocolUDP {
		t.Error("Protocol sort failed")
	}
}

func TestSortAllConnections_ByLocal(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortLocal,
			SortAscending: true,
		}},
	}
	conns := []connectionWithProcess{
		{ProcessName: "A", Connection: model.Connection{LocalAddr: "127.0.0.1:8080"}},
		{ProcessName: "B", Connection: model.Connection{LocalAddr: "127.0.0.1:443"}},
		{ProcessName: "C", Connection: model.Connection{LocalAddr: "127.0.0.1:9000"}},
	}

	result := m.sortAllConnections(conns)

	if result[0].LocalAddr != "127.0.0.1:443" {
		t.Errorf("Expected 443 first, got %s", result[0].LocalAddr)
	}
}

func TestSortAllConnections_ByRemote(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortRemote,
			SortAscending: true,
		}},
	}
	conns := []connectionWithProcess{
		{ProcessName: "A", Connection: model.Connection{RemoteAddr: "10.0.0.2:443"}},
		{ProcessName: "B", Connection: model.Connection{RemoteAddr: "10.0.0.1:443"}},
		{ProcessName: "C", Connection: model.Connection{RemoteAddr: "10.0.0.3:443"}},
	}

	result := m.sortAllConnections(conns)

	if result[0].RemoteAddr != "10.0.0.1:443" {
		t.Errorf("Expected 10.0.0.1 first, got %s", result[0].RemoteAddr)
	}
}

func TestSortAllConnections_ByState(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortState,
			SortAscending: true,
		}},
	}
	conns := []connectionWithProcess{
		{ProcessName: "A", Connection: model.Connection{State: model.StateTimeWait}},
		{ProcessName: "B", Connection: model.Connection{State: model.StateEstablished}},
		{ProcessName: "C", Connection: model.Connection{State: model.StateListen}},
	}

	result := m.sortAllConnections(conns)

	// Established < Listen < TimeWait alphabetically
	if result[0].State != model.StateEstablished {
		t.Errorf("Expected ESTABLISHED first, got %s", result[0].State)
	}
}

func TestSortAllConnections_StableOrdering(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortPID,
			SortAscending: true,
		}},
	}
	// Same PID, different process names and addresses
	conns := []connectionWithProcess{
		{ProcessName: "Zebra", Connection: model.Connection{PID: 100, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.1:443"}},
		{ProcessName: "Apple", Connection: model.Connection{PID: 100, LocalAddr: "127.0.0.1:8000", RemoteAddr: "10.0.0.1:443"}},
		{ProcessName: "Apple", Connection: model.Connection{PID: 100, LocalAddr: "127.0.0.1:8000", RemoteAddr: "10.0.0.2:443"}},
	}

	result := m.sortAllConnections(conns)

	// Same PID → sort by ProcessName → same ProcessName → sort by LocalAddr → same LocalAddr → sort by RemoteAddr
	if result[0].ProcessName != "Apple" || result[0].RemoteAddr != "10.0.0.1:443" {
		t.Error("Stable ordering by secondary keys failed")
	}
	if result[1].ProcessName != "Apple" || result[1].RemoteAddr != "10.0.0.2:443" {
		t.Error("Stable ordering by tertiary key failed")
	}
	if result[2].ProcessName != "Zebra" {
		t.Error("Stable ordering failed for different process name")
	}
}

func TestSortAllConnections_DefaultColumn(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortColumn(999), // invalid
			SortAscending: true,
		}},
	}
	conns := []connectionWithProcess{
		{ProcessName: "A", Connection: model.Connection{PID: 200}},
		{ProcessName: "B", Connection: model.Connection{PID: 100}},
	}

	result := m.sortAllConnections(conns)

	// Default should sort by PID
	if result[0].PID != 100 {
		t.Error("Default sort should use PID")
	}
}

// Test sortConnectionsForView

func TestSortConnectionsForView_NilView(t *testing.T) {
	m := Model{stack: []ViewState{}}
	conns := []model.Connection{
		{LocalAddr: "127.0.0.1:9000"},
		{LocalAddr: "127.0.0.1:8000"},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].LocalAddr != "127.0.0.1:9000" {
		t.Error("sortConnectionsForView with nil view should return unchanged")
	}
}

func TestSortConnectionsForView_ByPID(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortPID,
			SortAscending: true,
		}},
	}
	conns := []model.Connection{
		{PID: 300},
		{PID: 100},
		{PID: 200},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].PID != 100 || result[1].PID != 200 || result[2].PID != 300 {
		t.Errorf("Expected PIDs [100, 200, 300], got [%d, %d, %d]",
			result[0].PID, result[1].PID, result[2].PID)
	}
}

func TestSortConnectionsForView_ByProtocol(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortProtocol,
			SortAscending: false,
		}},
	}
	conns := []model.Connection{
		{Protocol: model.ProtocolTCP},
		{Protocol: model.ProtocolUDP},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].Protocol != model.ProtocolUDP {
		t.Errorf("Expected UDP first (descending), got %s", result[0].Protocol)
	}
}

func TestSortConnectionsForView_ByLocal(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortLocal,
			SortAscending: true,
		}},
	}
	conns := []model.Connection{
		{LocalAddr: "127.0.0.1:9000"},
		{LocalAddr: "127.0.0.1:443"},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].LocalAddr != "127.0.0.1:443" {
		t.Errorf("Expected 443 first, got %s", result[0].LocalAddr)
	}
}

func TestSortConnectionsForView_ByRemote(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortRemote,
			SortAscending: true,
		}},
	}
	conns := []model.Connection{
		{RemoteAddr: "10.0.0.2:443"},
		{RemoteAddr: "10.0.0.1:443"},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].RemoteAddr != "10.0.0.1:443" {
		t.Errorf("Expected 10.0.0.1 first, got %s", result[0].RemoteAddr)
	}
}

func TestSortConnectionsForView_ByState(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortState,
			SortAscending: true,
		}},
	}
	conns := []model.Connection{
		{State: model.StateTimeWait},
		{State: model.StateEstablished},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].State != model.StateEstablished {
		t.Errorf("Expected ESTABLISHED first, got %s", result[0].State)
	}
}

func TestSortConnectionsForView_StableOrdering(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortState,
			SortAscending: true,
		}},
	}
	// Same state, different local/remote addresses
	conns := []model.Connection{
		{State: model.StateEstablished, LocalAddr: "127.0.0.1:9000", RemoteAddr: "10.0.0.1:443"},
		{State: model.StateEstablished, LocalAddr: "127.0.0.1:8000", RemoteAddr: "10.0.0.2:443"},
		{State: model.StateEstablished, LocalAddr: "127.0.0.1:8000", RemoteAddr: "10.0.0.1:443"},
	}

	result := m.sortConnectionsForView(conns)

	// Same State → sort by LocalAddr → same LocalAddr → sort by RemoteAddr
	if result[0].LocalAddr != "127.0.0.1:8000" || result[0].RemoteAddr != "10.0.0.1:443" {
		t.Error("Stable ordering failed")
	}
	if result[1].LocalAddr != "127.0.0.1:8000" || result[1].RemoteAddr != "10.0.0.2:443" {
		t.Error("Stable ordering by RemoteAddr failed")
	}
}

func TestSortConnectionsForView_DefaultColumn(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortColumn(999), // invalid
			SortAscending: true,
		}},
	}
	conns := []model.Connection{
		{LocalAddr: "127.0.0.1:9000"},
		{LocalAddr: "127.0.0.1:8000"},
	}

	result := m.sortConnectionsForView(conns)

	// Default should sort by LocalAddr
	if result[0].LocalAddr != "127.0.0.1:8000" {
		t.Errorf("Default sort should use LocalAddr, got %s first", result[0].LocalAddr)
	}
}

// Test getAggregatedBytes

func TestGetAggregatedBytes_TX(t *testing.T) {
	m := Model{
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1000, BytesRecv: 500},
			200: {BytesSent: 2000, BytesRecv: 1000},
		},
	}

	total := m.getAggregatedBytes([]int32{100, 200}, true)

	if total != 3000 {
		t.Errorf("Expected TX total 3000, got %d", total)
	}
}

func TestGetAggregatedBytes_RX(t *testing.T) {
	m := Model{
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1000, BytesRecv: 500},
			200: {BytesSent: 2000, BytesRecv: 1500},
		},
	}

	total := m.getAggregatedBytes([]int32{100, 200}, false)

	if total != 2000 {
		t.Errorf("Expected RX total 2000, got %d", total)
	}
}

func TestGetAggregatedBytes_MissingPID(t *testing.T) {
	m := Model{
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1000},
		},
	}

	// PID 200 not in cache
	total := m.getAggregatedBytes([]int32{100, 200}, true)

	if total != 1000 {
		t.Errorf("Expected 1000 (ignoring missing PID), got %d", total)
	}
}

func TestGetAggregatedBytes_EmptyPIDs(t *testing.T) {
	m := Model{
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1000},
		},
	}

	total := m.getAggregatedBytes([]int32{}, true)

	if total != 0 {
		t.Errorf("Expected 0 for empty PIDs, got %d", total)
	}
}

func TestGetAggregatedBytes_NilCache(t *testing.T) {
	m := Model{
		netIOCache: nil,
	}

	// Should not panic
	total := m.getAggregatedBytes([]int32{100}, true)

	if total != 0 {
		t.Errorf("Expected 0 for nil cache, got %d", total)
	}
}

// Test empty slice handling

func TestSortProcessList_EmptySlice(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortProcess,
			SortAscending: true,
		}},
	}

	result := m.sortProcessList([]model.Application{})

	if len(result) != 0 {
		t.Error("Empty slice should return empty result")
	}
}

func TestSortAllConnections_EmptySlice(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortPID,
			SortAscending: true,
		}},
	}

	result := m.sortAllConnections([]connectionWithProcess{})

	if len(result) != 0 {
		t.Error("Empty slice should return empty result")
	}
}

func TestSortConnectionsForView_EmptySlice(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelConnections,
			SortColumn:    SortLocal,
			SortAscending: true,
		}},
	}

	result := m.sortConnectionsForView([]model.Connection{})

	if len(result) != 0 {
		t.Error("Empty slice should return empty result")
	}
}

// Test does not modify original slice

func TestSortProcessList_DoesNotModifyOriginal(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelProcessList,
			SortColumn:    SortProcess,
			SortAscending: true,
		}},
	}
	original := []model.Application{
		{Name: "C"},
		{Name: "A"},
		{Name: "B"},
	}
	originalFirstName := original[0].Name

	_ = m.sortProcessList(original)

	if original[0].Name != originalFirstName {
		t.Error("sortProcessList should not modify original slice")
	}
}

func TestSortAllConnections_DoesNotModifyOriginal(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:         LevelAllConnections,
			SortColumn:    SortProcess,
			SortAscending: true,
		}},
	}
	original := []connectionWithProcess{
		{ProcessName: "C"},
		{ProcessName: "A"},
	}
	originalFirstName := original[0].ProcessName

	_ = m.sortAllConnections(original)

	if original[0].ProcessName != originalFirstName {
		t.Error("sortAllConnections should not modify original slice")
	}
}
