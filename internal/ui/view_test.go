package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/kostyay/netmon/internal/model"
)

// initViewport initializes the viewport for testing (simulates WindowSizeMsg).
func initViewport(m *Model) {
	m.viewport = viewport.New(80, 20)
	m.ready = true
	m.width = 80
	m.height = 26
}

func TestFormatPIDs_Empty(t *testing.T) {
	got := formatPIDs([]int32{})
	if got != "" {
		t.Errorf("formatPIDs([]) = %q, want empty string", got)
	}
}

func TestFormatPIDs_SinglePID(t *testing.T) {
	got := formatPIDs([]int32{1234})
	want := "PID: 1234"
	if got != want {
		t.Errorf("formatPIDs([1234]) = %q, want %q", got, want)
	}
}

func TestFormatPIDs_TwoPIDs(t *testing.T) {
	got := formatPIDs([]int32{1234, 5678})
	want := "PIDs: 1234, 5678"
	if got != want {
		t.Errorf("formatPIDs([1234, 5678]) = %q, want %q", got, want)
	}
}

func TestFormatPIDs_ThreePIDs(t *testing.T) {
	got := formatPIDs([]int32{100, 200, 300})
	want := "PIDs: 100, 200, 300"
	if got != want {
		t.Errorf("formatPIDs([100, 200, 300]) = %q, want %q", got, want)
	}
}

func TestFormatPIDs_ManyPIDs(t *testing.T) {
	got := formatPIDs([]int32{100, 200, 300, 400, 500})
	want := "PIDs: 100, 200 +3 more"
	if got != want {
		t.Errorf("formatPIDs([100, 200, 300, 400, 500]) = %q, want %q", got, want)
	}
}

func TestFormatPIDs_FourPIDs(t *testing.T) {
	got := formatPIDs([]int32{1, 2, 3, 4})
	want := "PIDs: 1, 2 +2 more"
	if got != want {
		t.Errorf("formatPIDs([1, 2, 3, 4]) = %q, want %q", got, want)
	}
}

func TestTruncateAddr_Short(t *testing.T) {
	got := truncateAddr("127.0.0.1:80", 21)
	want := "127.0.0.1:80"
	if got != want {
		t.Errorf("truncateAddr() = %q, want %q", got, want)
	}
}

func TestTruncateAddr_ExactLength(t *testing.T) {
	// 21 characters exactly
	addr := "192.168.100.100:12345"
	got := truncateAddr(addr, 21)
	if got != addr {
		t.Errorf("truncateAddr() = %q, want %q", got, addr)
	}
}

func TestTruncateAddr_TooLong(t *testing.T) {
	// More than 21 characters
	addr := "255.255.255.255:123456789"
	got := truncateAddr(addr, 21)
	// Should be 18 chars + "..."
	if len(got) != 21 {
		t.Errorf("truncateAddr() len = %d, want 21", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Errorf("truncateAddr() should end with '...', got %q", got)
	}
}

func TestTruncateAddr_EmptyString(t *testing.T) {
	got := truncateAddr("", 21)
	if got != "" {
		t.Errorf("truncateAddr(\"\") = %q, want empty", got)
	}
}

func TestTruncateAddr_MaxLenOne(t *testing.T) {
	// Edge case with very small maxLen
	got := truncateAddr("abcd", 4)
	// maxLen 4, addr len 4, should return as is
	if got != "abcd" {
		t.Errorf("truncateAddr() = %q, want abcd", got)
	}
}

func TestTruncateAddr_MaxLenTruncates(t *testing.T) {
	got := truncateAddr("abcde", 4)
	// Should be 1 char + "..."
	want := "a..."
	if got != want {
		t.Errorf("truncateAddr() = %q, want %q", got, want)
	}
}

// View rendering tests

func TestView_Quitting(t *testing.T) {
	m := Model{
		quitting: true,
	}

	view := m.View()

	if view != "" {
		t.Errorf("View() when quitting should be empty, got %q", view)
	}
}

func TestView_NilSnapshot(t *testing.T) {
	m := Model{
		snapshot: nil,
		quitting: false,
	}
	initViewport(&m)

	view := m.View()

	if !strings.Contains(view, "Loading") {
		t.Error("View() with nil snapshot should show 'Loading'")
	}
}

func TestView_EmptyApplications(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{},
			Timestamp:    time.Now(),
		},
		quitting: false,
	}
	initViewport(&m)

	view := m.View()

	if !strings.Contains(view, "No network connections") {
		t.Error("View() with empty applications should show 'No network connections'")
	}
}

func TestView_WithApplications(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{1234},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					},
				},
			},
			Timestamp: time.Now(),
		},
		cursor:       0,
		quitting:     false,
		expandedApps: make(map[string]bool),
	}
	initViewport(&m)

	view := m.View()

	if !strings.Contains(view, "TestApp") {
		t.Error("View() should contain app name 'TestApp'")
	}
	if !strings.Contains(view, "1 connections") {
		t.Error("View() should contain connection count")
	}
}

func TestView_ContainsHeader(t *testing.T) {
	m := Model{
		snapshot: nil,
		quitting: false,
	}
	initViewport(&m)

	view := m.View()

	if !strings.Contains(view, "netmon") {
		t.Error("View() should contain 'netmon' header")
	}
}

func TestView_ContainsFooter(t *testing.T) {
	m := Model{
		snapshot: nil,
		quitting: false,
	}
	initViewport(&m)

	view := m.View()

	if !strings.Contains(view, "Quit") {
		t.Error("View() should contain 'Quit' in footer")
	}
}

func TestView_ContainsRefreshRate(t *testing.T) {
	m := Model{
		snapshot:        nil,
		quitting:        false,
		refreshInterval: 2 * time.Second,
	}
	initViewport(&m)

	view := m.View()

	if !strings.Contains(view, "Refresh:") {
		t.Error("View() should contain 'Refresh:' status")
	}
}

func TestRenderApplications_Empty(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{},
		},
	}

	result := m.renderApplications()

	if result != "" {
		t.Errorf("renderApplications() with empty apps should be empty, got %q", result)
	}
}

func TestRenderApplications_SingleApp(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP},
					},
				},
			},
		},
		cursor:       0,
		expandedApps: make(map[string]bool),
	}

	result := m.renderApplications()

	if !strings.Contains(result, "App1") {
		t.Error("renderApplications() should contain app name")
	}
}

func TestRenderApplications_ExpandedApp(t *testing.T) {
	expandedApps := make(map[string]bool)
	expandedApps["ExpandedApp"] = true

	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "ExpandedApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					},
				},
			},
		},
		cursor:       0,
		expandedApps: expandedApps,
	}

	result := m.renderApplications()

	// Should show the expand icon ▼
	if !strings.Contains(result, "▼") {
		t.Error("Expanded app should show ▼ icon")
	}
	// Should show connection details
	if !strings.Contains(result, "TCP") {
		t.Error("Expanded app should show connection protocol")
	}
}

func TestRenderApplications_CollapsedApp(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "CollapsedApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
					},
				},
			},
		},
		cursor:       0,
		expandedApps: make(map[string]bool), // Not expanded
	}

	result := m.renderApplications()

	// Should show the collapsed icon ▶
	if !strings.Contains(result, "▶") {
		t.Error("Collapsed app should show ▶ icon")
	}
}

func TestRenderAppHeader_Collapsed(t *testing.T) {
	m := Model{cursor: 0, expandedApps: make(map[string]bool)}
	app := model.Application{
		Name:        "TestApp",
		PIDs:        []int32{1234},
		Connections: []model.Connection{{Protocol: model.ProtocolTCP}},
	}

	result := m.renderAppHeader(app, false)

	if !strings.Contains(result, "▶") {
		t.Error("Collapsed app header should contain ▶")
	}
	if !strings.Contains(result, "TestApp") {
		t.Error("App header should contain app name")
	}
}

func TestRenderAppHeader_Expanded(t *testing.T) {
	expandedApps := make(map[string]bool)
	expandedApps["TestApp"] = true
	m := Model{cursor: 0, expandedApps: expandedApps}
	app := model.Application{
		Name:        "TestApp",
		PIDs:        []int32{1234},
		Connections: []model.Connection{{Protocol: model.ProtocolTCP}},
	}

	result := m.renderAppHeader(app, false)

	if !strings.Contains(result, "▼") {
		t.Error("Expanded app header should contain ▼")
	}
}

func TestRenderAppHeader_ConnectionCount(t *testing.T) {
	m := Model{cursor: 0}
	app := model.Application{
		Name: "TestApp",
		PIDs: []int32{1234},
		Connections: []model.Connection{
			{Protocol: "TCP"},
			{Protocol: "UDP"},
			{Protocol: "TCP"},
		},
	}

	result := m.renderAppHeader(app, false)

	if !strings.Contains(result, "3 connections") {
		t.Error("App header should show connection count")
	}
}

func TestRenderConnection_TCP(t *testing.T) {
	m := Model{}
	conn := model.Connection{
		Protocol:   "TCP",
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "10.0.0.1:443",
		State:      "ESTABLISHED",
	}

	result := m.renderConnection(conn, false)

	if !strings.Contains(result, "TCP") {
		t.Error("Connection should show protocol")
	}
	if !strings.Contains(result, "127.0.0.1:8080") {
		t.Error("Connection should show local addr")
	}
	if !strings.Contains(result, "10.0.0.1:443") {
		t.Error("Connection should show remote addr")
	}
	if !strings.Contains(result, "ESTABLISHED") {
		t.Error("Connection should show state")
	}
}

func TestRenderConnection_UDP(t *testing.T) {
	m := Model{}
	conn := model.Connection{
		Protocol:   "UDP",
		LocalAddr:  "0.0.0.0:53",
		RemoteAddr: "*",
		State:      "-",
	}

	result := m.renderConnection(conn, false)

	if !strings.Contains(result, "UDP") {
		t.Error("Connection should show UDP protocol")
	}
}

// Table view rendering tests

func TestRenderTableHeader(t *testing.T) {
	m := Model{
		sortColumn:     SortProcess,
		selectedColumn: SortPID,
		sortAscending:  true,
		width:          80,
	}

	result := m.renderTableHeader()

	// Should contain column headers (without [1], [2] prefixes)
	if !strings.Contains(result, "PID") {
		t.Error("Table header should contain PID")
	}
	if !strings.Contains(result, "Process") {
		t.Error("Table header should contain Process")
	}
	if !strings.Contains(result, "Proto") {
		t.Error("Table header should contain Proto")
	}
	if !strings.Contains(result, "Local") {
		t.Error("Table header should contain Local")
	}
	if !strings.Contains(result, "Remote") {
		t.Error("Table header should contain Remote")
	}
	if !strings.Contains(result, "State") {
		t.Error("Table header should contain State")
	}

	// Should contain ascending indicator (↑) for active sort column
	if !strings.Contains(result, "↑") {
		t.Error("Table header should contain ascending indicator ↑")
	}

	// Should NOT contain old-style prefixes
	if strings.Contains(result, "[1]") {
		t.Error("Table header should not contain [1] prefix")
	}
}

func TestRenderTableHeader_DescendingSort(t *testing.T) {
	m := Model{
		sortColumn:     SortProtocol,
		selectedColumn: SortProtocol,
		sortAscending:  false,
		width:          80,
	}

	result := m.renderTableHeader()

	// Should contain descending indicator (↓)
	if !strings.Contains(result, "↓") {
		t.Error("Table header should contain descending indicator ↓")
	}
}

func TestRenderTableHeader_SelectedColumn(t *testing.T) {
	m := Model{
		sortColumn:     SortProcess,
		selectedColumn: SortRemote, // Different column selected
		sortAscending:  true,
		width:          80,
	}

	result := m.renderTableHeader()

	// Result should contain Remote column (selected) styled differently
	// Both columns should be present
	if !strings.Contains(result, "Process") {
		t.Error("Table header should contain Process")
	}
	if !strings.Contains(result, "Remote") {
		t.Error("Table header should contain Remote")
	}
}

func TestRenderTable(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "Chrome",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: "ESTABLISHED"},
					},
				},
				{
					Name: "Firefox",
					PIDs: []int32{200},
					Connections: []model.Connection{
						{Protocol: "UDP", LocalAddr: "0.0.0.0:53", RemoteAddr: "*", State: "-"},
					},
				},
			},
		},
		sortColumn:    SortProcess,
		sortAscending: true,
		tableCursor:   0,
		width:         80,
	}

	result := m.renderTable()

	// Should contain process names
	if !strings.Contains(result, "Chrome") {
		t.Error("Table should contain 'Chrome'")
	}
	if !strings.Contains(result, "Firefox") {
		t.Error("Table should contain 'Firefox'")
	}

	// Should contain connection details
	if !strings.Contains(result, "TCP") {
		t.Error("Table should contain protocol")
	}

	// Should contain selection marker on first row
	if !strings.Contains(result, "▶") {
		t.Error("Table should contain selection marker")
	}

	// Should contain summary
	if !strings.Contains(result, "Showing 2 connections") {
		t.Error("Table should contain connection count summary")
	}
}

func TestFlattenConnections(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: "TCP"},
						{Protocol: "UDP"},
					},
				},
				{
					Name: "App2",
					PIDs: []int32{200},
					Connections: []model.Connection{
						{Protocol: "TCP"},
					},
				},
			},
		},
	}

	result := m.flattenConnections()

	if len(result) != 3 {
		t.Errorf("flattenConnections() returned %d items, want 3", len(result))
	}

	// Verify process names are attached
	if result[0].ProcessName != "App1" || result[1].ProcessName != "App1" {
		t.Error("First two connections should have ProcessName 'App1'")
	}
	if result[2].ProcessName != "App2" {
		t.Error("Third connection should have ProcessName 'App2'")
	}
}

func TestFlattenConnections_NilSnapshot(t *testing.T) {
	m := Model{snapshot: nil}

	result := m.flattenConnections()

	if result != nil {
		t.Error("flattenConnections() with nil snapshot should return nil")
	}
}

func TestSortConnections_ByProcess(t *testing.T) {
	m := Model{
		sortColumn:    SortProcess,
		sortAscending: true,
	}

	conns := []FlatConnection{
		{ProcessName: "Zebra", Connection: model.Connection{Protocol: "TCP"}},
		{ProcessName: "Alpha", Connection: model.Connection{Protocol: "UDP"}},
		{ProcessName: "Beta", Connection: model.Connection{Protocol: "TCP"}},
	}

	result := m.sortConnections(conns)

	if result[0].ProcessName != "Alpha" {
		t.Errorf("First item should be Alpha, got %s", result[0].ProcessName)
	}
	if result[1].ProcessName != "Beta" {
		t.Errorf("Second item should be Beta, got %s", result[1].ProcessName)
	}
	if result[2].ProcessName != "Zebra" {
		t.Errorf("Third item should be Zebra, got %s", result[2].ProcessName)
	}
}

func TestSortConnections_ByProcessDescending(t *testing.T) {
	m := Model{
		sortColumn:    SortProcess,
		sortAscending: false,
	}

	conns := []FlatConnection{
		{ProcessName: "Alpha", Connection: model.Connection{}},
		{ProcessName: "Zebra", Connection: model.Connection{}},
		{ProcessName: "Beta", Connection: model.Connection{}},
	}

	result := m.sortConnections(conns)

	if result[0].ProcessName != "Zebra" {
		t.Errorf("First item should be Zebra (descending), got %s", result[0].ProcessName)
	}
}

func TestSortConnections_ByProtocol(t *testing.T) {
	m := Model{
		sortColumn:    SortProtocol,
		sortAscending: true,
	}

	conns := []FlatConnection{
		{ProcessName: "App", Connection: model.Connection{Protocol: "UDP"}},
		{ProcessName: "App", Connection: model.Connection{Protocol: "TCP"}},
	}

	result := m.sortConnections(conns)

	if result[0].Connection.Protocol != "TCP" {
		t.Errorf("First item should be TCP, got %s", result[0].Connection.Protocol)
	}
}

func TestSortConnections_ByState(t *testing.T) {
	m := Model{
		sortColumn:    SortState,
		sortAscending: true,
	}

	conns := []FlatConnection{
		{ProcessName: "App", Connection: model.Connection{State: "TIME_WAIT"}},
		{ProcessName: "App", Connection: model.Connection{State: "ESTABLISHED"}},
		{ProcessName: "App", Connection: model.Connection{State: "LISTEN"}},
	}

	result := m.sortConnections(conns)

	if result[0].Connection.State != "ESTABLISHED" {
		t.Errorf("First item should be ESTABLISHED, got %s", result[0].Connection.State)
	}
}

func TestSortConnections_DoesNotMutateInput(t *testing.T) {
	m := Model{
		sortColumn:    SortProcess,
		sortAscending: true,
	}

	conns := []FlatConnection{
		{ProcessName: "Zebra", Connection: model.Connection{}},
		{ProcessName: "Alpha", Connection: model.Connection{}},
	}

	m.sortConnections(conns)

	// Original slice should be unchanged
	if conns[0].ProcessName != "Zebra" {
		t.Error("sortConnections should not mutate input slice")
	}
}

func TestTableViewSelection(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: "TCP", State: "ESTABLISHED"},
						{Protocol: "UDP", State: "-"},
					},
				},
			},
		},
		sortColumn:    SortProcess,
		sortAscending: true,
		tableCursor:   1, // Second row selected
		width:         80,
	}

	result := m.renderTable()

	// Count occurrences of ▶
	count := strings.Count(result, "▶")
	if count != 1 {
		t.Errorf("Expected exactly 1 selection marker, got %d", count)
	}
}

func TestSortConnections_ByLocalAddress(t *testing.T) {
	m := Model{
		sortColumn:    SortLocal,
		sortAscending: true,
	}

	conns := []FlatConnection{
		{ProcessName: "App", Connection: model.Connection{LocalAddr: "192.168.1.100:8080"}},
		{ProcessName: "App", Connection: model.Connection{LocalAddr: "127.0.0.1:80"}},
		{ProcessName: "App", Connection: model.Connection{LocalAddr: "10.0.0.1:443"}},
	}

	result := m.sortConnections(conns)

	if result[0].Connection.LocalAddr != "10.0.0.1:443" {
		t.Errorf("First item should be 10.0.0.1:443, got %s", result[0].Connection.LocalAddr)
	}
}

func TestSortConnections_ByRemoteAddress(t *testing.T) {
	m := Model{
		sortColumn:    SortRemote,
		sortAscending: true,
	}

	conns := []FlatConnection{
		{ProcessName: "App", Connection: model.Connection{RemoteAddr: "google.com:443"}},
		{ProcessName: "App", Connection: model.Connection{RemoteAddr: "api.github.com:443"}},
	}

	result := m.sortConnections(conns)

	if result[0].Connection.RemoteAddr != "api.github.com:443" {
		t.Errorf("First item should be api.github.com:443, got %s", result[0].Connection.RemoteAddr)
	}
}
