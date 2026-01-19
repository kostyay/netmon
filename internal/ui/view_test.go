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

func initViewportWithStack(m *Model) {
	initViewport(m)
	if len(m.stack) == 0 {
		m.stack = []ViewState{{
			Level:          LevelProcessList,
			Cursor:         0,
			SortColumn:     SortProcess,
			SortAscending:  true,
			SelectedColumn: SortProcess,
		}}
	}
	m.updateViewportContent() // Pre-render content like Update() does
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
	initViewportWithStack(&m)

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
	initViewportWithStack(&m)

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
		quitting: false,
	}
	initViewportWithStack(&m)

	view := m.View()

	if !strings.Contains(view, "TestApp") {
		t.Error("View() should contain app name 'TestApp'")
	}
}

func TestView_ContainsHeader(t *testing.T) {
	m := Model{
		snapshot: nil,
		quitting: false,
	}
	initViewportWithStack(&m)

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
	initViewportWithStack(&m)

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
	initViewportWithStack(&m)

	view := m.View()

	if !strings.Contains(view, "Refresh:") {
		t.Error("View() should contain 'Refresh:' status")
	}
}

func TestRenderProcessList_Empty(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{},
		},
		stack: []ViewState{{Level: LevelProcessList, Cursor: 0}},
	}

	result := m.renderProcessList()

	// Should show empty message when no processes
	if !strings.Contains(result, "No processes found") {
		t.Errorf("renderProcessList() with empty apps should show empty message, got: %s", result)
	}
}

func TestRenderProcessList_SingleApp(t *testing.T) {
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
		stack: []ViewState{{Level: LevelProcessList, Cursor: 0}},
	}

	result := m.renderProcessList()

	if !strings.Contains(result, "App1") {
		t.Error("renderProcessList() should contain app name")
	}
}

func TestRenderConnectionsList(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished, PID: 100},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
			Cursor:      0,
			SortColumn:  SortLocal,
		}},
	}

	result := m.renderConnectionsList()

	if !strings.Contains(result, "TestApp") {
		t.Error("renderConnectionsList() should contain process name")
	}
	if !strings.Contains(result, "TCP") {
		t.Error("renderConnectionsList() should contain protocol")
	}
	if !strings.Contains(result, "127.0.0.1:80") {
		t.Error("renderConnectionsList() should contain local addr")
	}
}

func TestRenderConnectionsList_ProcessNotFound(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{Name: "OtherApp"},
			},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "MissingApp",
		}},
	}

	result := m.renderConnectionsList()

	if !strings.Contains(result, "Process not found") {
		t.Error("renderConnectionsList() should show 'Process not found' for missing process")
	}
}

func TestRenderConnectionsList_FilteredByLocalAddr(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished, PID: 100},
						{Protocol: model.ProtocolTCP, LocalAddr: "192.168.1.1:8080", RemoteAddr: "10.0.0.2:443", State: model.StateEstablished, PID: 100},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
			Cursor:      0,
			SortColumn:  SortLocal,
		}},
		activeFilter: "192.168",
	}

	result := m.renderConnectionsList()

	// Should contain the filtered connection
	if !strings.Contains(result, "192.168.1.1:8080") {
		t.Error("renderConnectionsList() should contain filtered local addr")
	}
	// Should NOT contain the excluded connection
	if strings.Contains(result, "127.0.0.1:80") {
		t.Error("renderConnectionsList() should NOT contain excluded local addr")
	}
	// Should show 1 connection count
	if !strings.Contains(result, "1 connections") {
		t.Error("renderConnectionsList() should show filtered count")
	}
}

func TestRenderConnectionsList_FilteredByRemoteAddr(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished, PID: 100},
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:81", RemoteAddr: "8.8.8.8:53", State: model.StateEstablished, PID: 100},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
			Cursor:      0,
			SortColumn:  SortLocal,
		}},
		activeFilter: "8.8.8.8",
	}

	result := m.renderConnectionsList()

	// Should contain the filtered connection
	if !strings.Contains(result, "8.8.8.8:53") {
		t.Error("renderConnectionsList() should contain filtered remote addr")
	}
	// Should NOT contain the excluded connection
	if strings.Contains(result, "10.0.0.1:443") {
		t.Error("renderConnectionsList() should NOT contain excluded remote addr")
	}
}

func TestRenderConnectionsList_FilterNoMatches(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "TestApp",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished, PID: 100},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
			Cursor:      0,
			SortColumn:  SortLocal,
		}},
		activeFilter: "nomatchxyz",
	}

	result := m.renderConnectionsList()

	if !strings.Contains(result, "No matches for 'nomatchxyz'") {
		t.Error("renderConnectionsList() should show 'No matches' when filter excludes all")
	}
}

func TestRenderBreadcrumbs_ProcessList(t *testing.T) {
	m := Model{
		refreshInterval: 2 * time.Second,
		stack:           []ViewState{{Level: LevelProcessList}},
	}

	result := m.renderBreadcrumbsText()

	if !strings.Contains(result, "Processes") {
		t.Error("Breadcrumbs should contain 'Processes'")
	}
	if !strings.Contains(result, "Refresh:") {
		t.Error("Breadcrumbs should contain refresh rate")
	}
}

func TestRenderBreadcrumbs_ConnectionsLevel(t *testing.T) {
	m := Model{
		refreshInterval: 2 * time.Second,
		stack: []ViewState{
			{Level: LevelProcessList},
			{Level: LevelConnections, ProcessName: "Chrome"},
		},
	}

	result := m.renderBreadcrumbsText()

	if !strings.Contains(result, "Processes") {
		t.Error("Breadcrumbs should contain 'Processes'")
	}
	if !strings.Contains(result, "Chrome") {
		t.Error("Breadcrumbs should contain 'Chrome'")
	}
}

func TestRenderProcessListHeader(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:          LevelProcessList,
			SelectedColumn: SortPID,
		}},
	}

	columns := processListColumns()
	widths := calculateColumnWidths(columns, 100)
	result := m.renderProcessListHeader(widths)

	if !strings.Contains(result, "Process") {
		t.Error("Header should contain 'Process'")
	}
	if !strings.Contains(result, "Conns") {
		t.Error("Header should contain 'Conns'")
	}
	if !strings.Contains(result, "ESTAB") {
		t.Error("Header should contain 'ESTAB'")
	}
	if !strings.Contains(result, "LISTEN") {
		t.Error("Header should contain 'LISTEN'")
	}
}

func TestRenderConnectionsHeader(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:          LevelConnections,
			SortColumn:     SortLocal,
			SortAscending:  true,
			SelectedColumn: SortLocal,
		}},
	}

	columns := connectionsColumns()
	widths := calculateColumnWidths(columns, 100)
	result := m.renderConnectionsHeader(widths)

	if !strings.Contains(result, "Proto") {
		t.Error("Header should contain 'Proto'")
	}
	if !strings.Contains(result, "Local") {
		t.Error("Header should contain 'Local'")
	}
	if !strings.Contains(result, "Remote") {
		t.Error("Header should contain 'Remote'")
	}
	if !strings.Contains(result, "State") {
		t.Error("Header should contain 'State'")
	}
	// PID column removed - redundant in process detail view
	if !strings.Contains(result, "↑") {
		t.Error("Header should contain ascending sort indicator")
	}
}

func TestRenderConnectionsHeader_Descending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:          LevelConnections,
			SortColumn:     SortRemote,
			SortAscending:  false,
			SelectedColumn: SortRemote,
		}},
	}

	columns := connectionsColumns()
	widths := calculateColumnWidths(columns, 100)
	result := m.renderConnectionsHeader(widths)

	if !strings.Contains(result, "↓") {
		t.Error("Header should contain descending sort indicator")
	}
}

func TestSortConnectionsForView(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			SortColumn:    SortLocal,
			SortAscending: true,
		}},
	}

	conns := []model.Connection{
		{LocalAddr: "192.168.1.1:80", Protocol: "TCP"},
		{LocalAddr: "127.0.0.1:80", Protocol: "TCP"},
		{LocalAddr: "10.0.0.1:80", Protocol: "TCP"},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].LocalAddr != "10.0.0.1:80" {
		t.Errorf("First item should be 10.0.0.1:80, got %s", result[0].LocalAddr)
	}
	if result[1].LocalAddr != "127.0.0.1:80" {
		t.Errorf("Second item should be 127.0.0.1:80, got %s", result[1].LocalAddr)
	}
}

func TestSortConnectionsForView_Descending(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			SortColumn:    SortProtocol,
			SortAscending: false,
		}},
	}

	conns := []model.Connection{
		{Protocol: "TCP"},
		{Protocol: "UDP"},
	}

	result := m.sortConnectionsForView(conns)

	if result[0].Protocol != "UDP" {
		t.Errorf("First item should be UDP (descending), got %s", result[0].Protocol)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %s, want %s", tt.bytes, got, tt.want)
		}
	}
}

func TestFormatBytesOrDash_Nil(t *testing.T) {
	result := formatBytesOrDash(nil, true)
	if result != "--" {
		t.Errorf("formatBytesOrDash(nil) = %s, want --", result)
	}
}

func TestFormatBytesOrDash_Sent(t *testing.T) {
	stats := &model.NetIOStats{BytesSent: 1024, BytesRecv: 2048}
	result := formatBytesOrDash(stats, true)
	if result != "1.0 KB" {
		t.Errorf("formatBytesOrDash(stats, true) = %s, want 1.0 KB", result)
	}
}

func TestFormatBytesOrDash_Recv(t *testing.T) {
	stats := &model.NetIOStats{BytesSent: 1024, BytesRecv: 2048}
	result := formatBytesOrDash(stats, false)
	if result != "2.0 KB" {
		t.Errorf("formatBytesOrDash(stats, false) = %s, want 2.0 KB", result)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"toolongstring", 10, "toolong..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}

	for _, tt := range tests {
		got := truncateString(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}

// Tests for keybindings by level

func TestRenderKeybindings_ProcessList(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level: LevelProcessList,
		}},
	}

	result := m.renderKeybindingsText()

	// Process list should show Drill-in
	if !strings.Contains(result, "Drill-in") {
		t.Error("Process list keybindings should contain 'Drill-in'")
	}
	if strings.Contains(result, "Back") {
		t.Error("Process list keybindings should NOT contain 'Back'")
	}
}

func TestRenderKeybindings_Connections(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level: LevelConnections,
		}},
	}

	result := m.renderKeybindingsText()

	// Connections level should show Back and Help
	if !strings.Contains(result, "Back") {
		t.Error("Connections keybindings should contain 'Back'")
	}
	if !strings.Contains(result, "Help") {
		t.Error("Connections keybindings should contain 'Help'")
	}
	if strings.Contains(result, "Drill-in") {
		t.Error("Connections keybindings should NOT contain 'Drill-in'")
	}
}

// Tests for aggregated network I/O

func TestGetAggregatedNetIO_NoStats(t *testing.T) {
	m := Model{
		netIOCache: make(map[int32]*model.NetIOStats),
	}

	tx, rx := m.getAggregatedNetIO([]int32{1, 2, 3})

	if tx != "--" || rx != "--" {
		t.Errorf("No stats should return '--', got TX=%s, RX=%s", tx, rx)
	}
}

func TestGetAggregatedNetIO_SinglePID(t *testing.T) {
	m := Model{
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1024, BytesRecv: 2048},
		},
	}

	tx, rx := m.getAggregatedNetIO([]int32{100})

	if tx != "1.0 KB" {
		t.Errorf("TX should be 1.0 KB, got %s", tx)
	}
	if rx != "2.0 KB" {
		t.Errorf("RX should be 2.0 KB, got %s", rx)
	}
}

func TestGetAggregatedNetIO_MultiplePIDs(t *testing.T) {
	m := Model{
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1024, BytesRecv: 2048},
			200: {BytesSent: 1024, BytesRecv: 2048},
		},
	}

	tx, rx := m.getAggregatedNetIO([]int32{100, 200})

	if tx != "2.0 KB" {
		t.Errorf("TX should be 2.0 KB (aggregated), got %s", tx)
	}
	if rx != "4.0 KB" {
		t.Errorf("RX should be 4.0 KB (aggregated), got %s", rx)
	}
}

func TestGetAggregatedNetIO_PartialStats(t *testing.T) {
	m := Model{
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1024, BytesRecv: 2048},
			// 200 has no stats
		},
	}

	tx, rx := m.getAggregatedNetIO([]int32{100, 200})

	// Should still show stats for PIDs that have them
	if tx != "1.0 KB" {
		t.Errorf("TX should be 1.0 KB, got %s", tx)
	}
	if rx != "2.0 KB" {
		t.Errorf("RX should be 2.0 KB, got %s", rx)
	}
}

// Tests for PID list formatting

func TestFormatPIDList_Empty(t *testing.T) {
	result := formatPIDList([]int32{})
	if result != "-" {
		t.Errorf("Empty PIDs should return '-', got %s", result)
	}
}

func TestFormatPIDList_Single(t *testing.T) {
	result := formatPIDList([]int32{1234})
	if result != "1234" {
		t.Errorf("Single PID should return '1234', got %s", result)
	}
}

func TestFormatPIDList_Multiple(t *testing.T) {
	result := formatPIDList([]int32{1, 2, 3})
	if result != "1, 2, 3" {
		t.Errorf("Multiple PIDs should return '1, 2, 3', got %s", result)
	}
}

func TestFormatPIDList_ManyMore(t *testing.T) {
	result := formatPIDList([]int32{1, 2, 3, 4, 5})
	if result != "1, 2 +3 more" {
		t.Errorf("Many PIDs should return '1, 2 +3 more', got %s", result)
	}
}

// Tests for process detail view header

func TestRenderConnectionsList_HeaderContent(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "Chrome",
					PIDs: []int32{100, 101},
					Connections: []model.Connection{
						{Protocol: "TCP", LocalAddr: "127.0.0.1:80", State: "ESTABLISHED"},
					},
				},
			},
		},
		netIOCache: map[int32]*model.NetIOStats{
			100: {BytesSent: 1024, BytesRecv: 2048},
		},
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "Chrome",
		}},
	}

	result := m.renderConnectionsList()

	// Should contain process name
	if !strings.Contains(result, "Chrome") {
		t.Error("Detail view should contain process name 'Chrome'")
	}
	// Should contain PIDs
	if !strings.Contains(result, "100") {
		t.Error("Detail view should contain PID '100'")
	}
	// Should contain TX/RX
	if !strings.Contains(result, "TX:") {
		t.Error("Detail view should contain 'TX:'")
	}
	if !strings.Contains(result, "RX:") {
		t.Error("Detail view should contain 'RX:'")
	}
}

// Tests for two-row footer layout

func TestRenderFooter_ContainsBothRows(t *testing.T) {
	m := Model{
		refreshInterval: 2 * time.Second,
		stack: []ViewState{{
			Level: LevelProcessList,
		}},
	}

	result := m.renderFooter()

	// Should contain breadcrumbs (row 1)
	if !strings.Contains(result, "Processes") {
		t.Error("Footer should contain 'Processes' breadcrumb")
	}
	// Should contain keybindings (row 2)
	if !strings.Contains(result, "Help") {
		t.Error("Footer should contain 'Help' keybinding")
	}
	if !strings.Contains(result, "Quit") {
		t.Error("Footer should contain 'Quit' keybinding")
	}
}

// Tests for kill mode UI rendering

func TestRenderFooter_KillMode(t *testing.T) {
	m := Model{
		killMode: true,
		killTarget: &killTargetInfo{
			PID:         12345,
			ProcessName: "TestApp",
			Signal:      "SIGTERM",
		},
		stack: []ViewState{{
			Level: LevelProcessList,
		}},
	}

	result := m.renderFooter()

	// Should show kill confirmation prompt
	if !strings.Contains(result, "Kill PID 12345") {
		t.Error("Footer should contain kill PID")
	}
	if !strings.Contains(result, "TestApp") {
		t.Error("Footer should contain process name in kill prompt")
	}
	if !strings.Contains(result, "SIGTERM") {
		t.Error("Footer should contain signal in kill prompt")
	}
	if !strings.Contains(result, "[y/n]") {
		t.Error("Footer should contain [y/n] confirmation")
	}
}

func TestRenderFooter_KillResult(t *testing.T) {
	m := Model{
		killResult:   "Killed PID 12345 (TestApp)",
		killResultAt: time.Now(), // Recent
		stack: []ViewState{{
			Level: LevelProcessList,
		}},
	}

	result := m.renderFooter()

	// Should show kill result
	if !strings.Contains(result, "Killed PID 12345") {
		t.Error("Footer should contain kill result message")
	}
}

func TestRenderFooter_KillResultExpired(t *testing.T) {
	m := Model{
		killResult:      "Killed PID 12345 (TestApp)",
		killResultAt:    time.Now().Add(-3 * time.Second), // More than 2s ago
		refreshInterval: 2 * time.Second,
		stack: []ViewState{{
			Level: LevelProcessList,
		}},
	}

	result := m.renderFooter()

	// Should NOT show expired kill result (shows breadcrumbs instead)
	if strings.Contains(result, "Killed PID 12345") {
		t.Error("Footer should NOT contain expired kill result message")
	}
	if !strings.Contains(result, "Processes") {
		t.Error("Footer should show breadcrumbs when kill result expired")
	}
}

func TestRenderKeybindings_KillMode(t *testing.T) {
	m := Model{
		killMode: true,
		stack: []ViewState{{
			Level: LevelProcessList,
		}},
	}

	result := m.renderKeybindingsText()

	// Kill mode keybindings
	if !strings.Contains(result, "[KILL]") {
		t.Error("Keybindings should contain '[KILL]' label")
	}
	if !strings.Contains(result, "Confirm") {
		t.Error("Keybindings should contain 'Confirm'")
	}
	if !strings.Contains(result, "Cancel") {
		t.Error("Keybindings should contain 'Cancel'")
	}
}

func TestRenderKeybindings_ContainsKillKey(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level: LevelProcessList,
		}},
	}

	result := m.renderKeybindingsText()

	// Process list should show x/X for kill
	if !strings.Contains(result, "x/X") || !strings.Contains(result, "Kill") {
		t.Error("Keybindings should contain 'x/X Kill'")
	}
}

// Tests for viewport scrolling

func TestSyncViewportScroll_NotReady(t *testing.T) {
	m := &Model{
		ready: false,
		stack: []ViewState{{Level: LevelProcessList, Cursor: 5}},
	}

	// Should not panic when not ready
	m.syncViewportScroll()
}

func TestSyncViewportScroll_CursorAboveViewport(t *testing.T) {
	m := &Model{
		ready: true,
		stack: []ViewState{{Level: LevelProcessList, Cursor: 2}},
		snapshot: &model.NetworkSnapshot{
			Applications: make([]model.Application, 20),
		},
	}
	m.viewport = viewport.New(80, 10)
	m.viewport.SetYOffset(5) // Cursor at line 2+1=3 is above offset 5

	m.syncViewportScroll()

	// Viewport should scroll up to show cursor
	if m.viewport.YOffset > 3 {
		t.Errorf("YOffset = %d, should be <= 3 to show cursor at line 3", m.viewport.YOffset)
	}
}

func TestSyncViewportScroll_CursorBelowViewport(t *testing.T) {
	m := &Model{
		ready: true,
		stack: []ViewState{{Level: LevelProcessList, Cursor: 15}},
		snapshot: &model.NetworkSnapshot{
			Applications: make([]model.Application, 20),
		},
	}
	m.viewport = viewport.New(80, 5)
	// Set some content first so viewport can scroll
	lines := make([]string, 25)
	for i := range lines {
		lines[i] = "line"
	}
	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.SetYOffset(0) // Start at top

	m.syncViewportScroll()

	// Viewport should scroll down to show cursor
	// cursor line = 15 + 1 (header) = 16, viewport height = 5
	// visible end = YOffset + Height
	// Need: cursorLine < visibleEnd => YOffset + 5 > 16 => YOffset > 11
	if m.viewport.YOffset < 11 {
		t.Errorf("YOffset = %d, should be >= 11 to show cursor at line 16", m.viewport.YOffset)
	}
}

func TestSyncViewportScroll_CursorWithinViewport(t *testing.T) {
	m := &Model{
		ready: true,
		stack: []ViewState{{Level: LevelProcessList, Cursor: 3}},
		snapshot: &model.NetworkSnapshot{
			Applications: make([]model.Application, 20),
		},
	}
	m.viewport = viewport.New(80, 10)
	m.viewport.SetYOffset(2) // Cursor at line 4 is within range 2-11

	originalOffset := m.viewport.YOffset
	m.syncViewportScroll()

	// Viewport should not change
	if m.viewport.YOffset != originalOffset {
		t.Errorf("YOffset changed from %d to %d, should stay same", originalOffset, m.viewport.YOffset)
	}
}

func TestCursorLinePosition_ProcessList(t *testing.T) {
	m := Model{
		stack: []ViewState{{Level: LevelProcessList, Cursor: 5}},
		snapshot: &model.NetworkSnapshot{
			Applications: make([]model.Application, 10),
		},
	}

	// Headers are now frozen outside viewport, so cursor position = cursor index
	pos := m.cursorLinePosition()
	expected := 5
	if pos != expected {
		t.Errorf("cursorLinePosition = %d, want %d", pos, expected)
	}
}

func TestCursorLinePosition_ConnectionsNoExe(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
			Cursor:      3,
		}},
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{Name: "TestApp", Exe: ""}, // No exe path
			},
		},
	}

	// Headers are now frozen outside viewport, so cursor position = cursor index
	pos := m.cursorLinePosition()
	expected := 3
	if pos != expected {
		t.Errorf("cursorLinePosition = %d, want %d", pos, expected)
	}
}

func TestCursorLinePosition_ConnectionsWithExe(t *testing.T) {
	m := Model{
		stack: []ViewState{{
			Level:       LevelConnections,
			ProcessName: "TestApp",
			Cursor:      3,
		}},
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{Name: "TestApp", Exe: "/usr/bin/testapp"}, // Has exe path
			},
		},
	}

	// Headers are now frozen outside viewport, so cursor position = cursor index
	pos := m.cursorLinePosition()
	expected := 3
	if pos != expected {
		t.Errorf("cursorLinePosition = %d, want %d", pos, expected)
	}
}

func TestCursorLinePosition_AllConnections(t *testing.T) {
	m := Model{
		stack: []ViewState{{Level: LevelAllConnections, Cursor: 7}},
		snapshot: &model.NetworkSnapshot{
			Applications: make([]model.Application, 10),
		},
	}

	// Headers are now frozen outside viewport, so cursor position = cursor index
	pos := m.cursorLinePosition()
	expected := 7
	if pos != expected {
		t.Errorf("cursorLinePosition = %d, want %d", pos, expected)
	}
}

// Tests for all-connections view filtering

func TestRenderAllConnections_Basic(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443", State: "ESTABLISHED"},
					},
				},
			},
		},
		stack: []ViewState{{
			Level:          LevelAllConnections,
			Cursor:         0,
			SortColumn:     SortProcess,
			SortAscending:  true,
			SelectedColumn: SortProcess,
		}},
		dnsCache: make(map[string]string),
		changes:  make(map[ConnectionKey]Change),
	}

	result := m.renderAllConnections()

	if !strings.Contains(result, "App1") {
		t.Error("All connections should contain process name 'App1'")
	}
	if !strings.Contains(result, "TCP") {
		t.Error("All connections should contain protocol 'TCP'")
	}
	if !strings.Contains(result, "127.0.0.1:80") {
		t.Error("All connections should contain local addr")
	}
}

func TestFilteredAllConnections_ByProcessName(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "Chrome",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:80"},
					},
				},
				{
					Name: "Firefox",
					PIDs: []int32{200},
					Connections: []model.Connection{
						{PID: 200, Protocol: "TCP", LocalAddr: "127.0.0.1:81"},
					},
				},
			},
		},
		activeFilter: "Chrome",
		stack:        []ViewState{{Level: LevelAllConnections}},
	}

	conns := m.filteredAllConnections()

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if conns[0].ProcessName != "Chrome" {
		t.Errorf("expected Chrome, got %s", conns[0].ProcessName)
	}
}

func TestFilteredAllConnections_ByPID(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{12345},
					Connections: []model.Connection{
						{PID: 12345, Protocol: "TCP", LocalAddr: "127.0.0.1:80"},
					},
				},
				{
					Name: "App2",
					PIDs: []int32{67890},
					Connections: []model.Connection{
						{PID: 67890, Protocol: "TCP", LocalAddr: "127.0.0.1:81"},
					},
				},
			},
		},
		activeFilter: "12345",
		stack:        []ViewState{{Level: LevelAllConnections}},
	}

	conns := m.filteredAllConnections()

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if conns[0].PID != 12345 {
		t.Errorf("expected PID 12345, got %d", conns[0].PID)
	}
}

func TestFilteredAllConnections_ByPort(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "*"},
						{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:9090", RemoteAddr: "*"},
					},
				},
			},
		},
		activeFilter: "8080",
		stack:        []ViewState{{Level: LevelAllConnections}},
	}

	conns := m.filteredAllConnections()

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if !strings.Contains(conns[0].LocalAddr, "8080") {
		t.Errorf("expected port 8080, got %s", conns[0].LocalAddr)
	}
}

func TestFilteredAllConnections_NoFilter(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, Protocol: "TCP"},
						{PID: 100, Protocol: "UDP"},
					},
				},
				{
					Name: "App2",
					PIDs: []int32{200},
					Connections: []model.Connection{
						{PID: 200, Protocol: "TCP"},
					},
				},
			},
		},
		activeFilter: "",
		stack:        []ViewState{{Level: LevelAllConnections}},
	}

	conns := m.filteredAllConnections()

	if len(conns) != 3 {
		t.Errorf("expected 3 connections (no filter), got %d", len(conns))
	}
}

func TestFilteredAllConnections_NilSnapshot(t *testing.T) {
	m := Model{
		snapshot:     nil,
		activeFilter: "test",
		stack:        []ViewState{{Level: LevelAllConnections}},
	}

	conns := m.filteredAllConnections()

	if conns != nil {
		t.Errorf("expected nil for nil snapshot, got %v", conns)
	}
}

func TestFilteredAllConnections_ByState(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, Protocol: "TCP", State: "ESTABLISHED"},
						{PID: 100, Protocol: "TCP", State: "LISTEN"},
					},
				},
			},
		},
		activeFilter: "LISTEN",
		stack:        []ViewState{{Level: LevelAllConnections}},
	}

	conns := m.filteredAllConnections()

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if string(conns[0].State) != "LISTEN" {
		t.Errorf("expected LISTEN state, got %s", conns[0].State)
	}
}

func TestFilteredAllConnections_ByRemoteAddr(t *testing.T) {
	m := Model{
		snapshot: &model.NetworkSnapshot{
			Applications: []model.Application{
				{
					Name: "App1",
					PIDs: []int32{100},
					Connections: []model.Connection{
						{PID: 100, Protocol: "TCP", RemoteAddr: "8.8.8.8:443"},
						{PID: 100, Protocol: "TCP", RemoteAddr: "1.1.1.1:53"},
					},
				},
			},
		},
		activeFilter: "8.8.8",
		stack:        []ViewState{{Level: LevelAllConnections}},
	}

	conns := m.filteredAllConnections()

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	if !strings.Contains(conns[0].RemoteAddr, "8.8.8.8") {
		t.Errorf("expected remote 8.8.8.8, got %s", conns[0].RemoteAddr)
	}
}

// Tests for padRight helper

func TestPadRight_Nopad(t *testing.T) {
	got := padRight("hello", 5)
	if got != "hello" {
		t.Errorf("padRight('hello', 5) = %q, want %q", got, "hello")
	}
}

func TestPadRight_AddsPadding(t *testing.T) {
	got := padRight("hi", 5)
	if got != "hi   " {
		t.Errorf("padRight('hi', 5) = %q, want %q", got, "hi   ")
	}
}

func TestPadRight_LongerThanWidth(t *testing.T) {
	got := padRight("hello world", 5)
	if got != "hello world" {
		t.Errorf("padRight longer than width should not truncate, got %q", got)
	}
}

func TestPadRight_EmptyString(t *testing.T) {
	got := padRight("", 3)
	if got != "   " {
		t.Errorf("padRight('', 3) = %q, want %q", got, "   ")
	}
}

// Tests for RenderFrameWithTitle

func TestRenderFrameWithTitle_ContainsTitle(t *testing.T) {
	result := RenderFrameWithTitle("content", "Test Title", 40, 5)

	if !strings.Contains(result, "Test Title") {
		t.Error("RenderFrameWithTitle should contain the title")
	}
}

func TestRenderFrameWithTitle_ContainsContent(t *testing.T) {
	result := RenderFrameWithTitle("my content here", "Title", 40, 5)

	if !strings.Contains(result, "my content here") {
		t.Error("RenderFrameWithTitle should contain the content")
	}
}

func TestRenderFrameWithTitle_HasBorders(t *testing.T) {
	result := RenderFrameWithTitle("content", "Title", 40, 5)

	// Check for rounded border characters
	if !strings.Contains(result, "╭") {
		t.Error("Should contain top-left border")
	}
	if !strings.Contains(result, "╮") {
		t.Error("Should contain top-right border")
	}
	if !strings.Contains(result, "╰") {
		t.Error("Should contain bottom-left border")
	}
	if !strings.Contains(result, "╯") {
		t.Error("Should contain bottom-right border")
	}
	if !strings.Contains(result, "│") {
		t.Error("Should contain vertical border")
	}
}

func TestRenderFrameWithTitle_NarrowWidth(t *testing.T) {
	// Title longer than available space
	result := RenderFrameWithTitle("x", "Very Long Title Here", 10, 3)

	// Should not panic, title gets truncated
	if result == "" {
		t.Error("Should render something even with narrow width")
	}
}

// Tests for splitLines helper

func TestSplitLines_Empty(t *testing.T) {
	got := splitLines("")
	if len(got) != 1 || got[0] != "" {
		t.Errorf("splitLines('') = %v, want ['']", got)
	}
}

func TestSplitLines_SingleLine(t *testing.T) {
	got := splitLines("hello")
	if len(got) != 1 || got[0] != "hello" {
		t.Errorf("splitLines('hello') = %v, want ['hello']", got)
	}
}

func TestSplitLines_MultipleLines(t *testing.T) {
	got := splitLines("a\nb\nc")
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("splitLines('a\\nb\\nc') = %v", got)
	}
}

// Tests for footer display priority

func TestRenderFooter_PriorityKillOverResult(t *testing.T) {
	m := Model{
		killMode: true,
		killTarget: &killTargetInfo{
			PID:         999,
			ProcessName: "KillMe",
			Signal:      "SIGTERM",
		},
		killResult:   "Some old result",
		killResultAt: time.Now(),
		stack:        []ViewState{{Level: LevelProcessList}},
	}

	result := m.renderFooter()

	// Kill mode should take precedence over result
	if !strings.Contains(result, "Kill PID 999") {
		t.Error("Kill mode should show kill prompt, not result")
	}
	if strings.Contains(result, "Some old result") {
		t.Error("Kill result should not appear when kill mode is active")
	}
}

func TestRenderFooter_PriorityResultOverSearch(t *testing.T) {
	m := Model{
		killMode:        false,
		killResult:      "Killed successfully",
		killResultAt:    time.Now(),
		searchMode:      true,
		searchQuery:     "chrome",
		refreshInterval: 2 * time.Second,
		stack:           []ViewState{{Level: LevelProcessList}},
	}

	result := m.renderFooter()

	// Result should take precedence over search
	if !strings.Contains(result, "Killed successfully") {
		t.Error("Kill result should show when recent")
	}
	if strings.Contains(result, "/chrome") {
		t.Error("Search input should not appear when result is shown")
	}
}

func TestRenderFooter_PrioritySearchOverFilter(t *testing.T) {
	m := Model{
		killMode:        false,
		killResult:      "",
		searchMode:      true,
		searchQuery:     "firefox",
		activeFilter:    "chrome",
		refreshInterval: 2 * time.Second,
		stack:           []ViewState{{Level: LevelProcessList}},
	}

	result := m.renderFooter()

	// Search mode should take precedence over filter
	if !strings.Contains(result, "/firefox") {
		t.Error("Search input should show in search mode")
	}
	if strings.Contains(result, "[filter: chrome]") {
		t.Error("Active filter should not appear when in search mode")
	}
}

func TestRenderFooter_PriorityFilterOverBreadcrumbs(t *testing.T) {
	m := Model{
		killMode:        false,
		killResult:      "",
		searchMode:      false,
		activeFilter:    "ssh",
		refreshInterval: 2 * time.Second,
		stack:           []ViewState{{Level: LevelProcessList}},
	}

	result := m.renderFooter()

	// Filter indicator should appear with breadcrumbs
	if !strings.Contains(result, "[filter: ssh]") {
		t.Error("Filter indicator should show")
	}
	if !strings.Contains(result, "Processes") {
		t.Error("Breadcrumbs should still appear with filter")
	}
}

func TestRenderFooter_BreadcrumbsOnly(t *testing.T) {
	m := Model{
		killMode:        false,
		killResult:      "",
		searchMode:      false,
		activeFilter:    "",
		refreshInterval: 2 * time.Second,
		stack:           []ViewState{{Level: LevelProcessList}},
	}

	result := m.renderFooter()

	// Only breadcrumbs should appear
	if !strings.Contains(result, "Processes") {
		t.Error("Breadcrumbs should show")
	}
	if strings.Contains(result, "[filter:") {
		t.Error("No filter indicator when filter is empty")
	}
}

// Tests for formatAddr

func TestFormatAddr_Empty(t *testing.T) {
	result := formatAddr("", "tcp", false)
	if result != "" {
		t.Errorf("formatAddr('') = %q, want ''", result)
	}
}

func TestFormatAddr_Asterisk(t *testing.T) {
	result := formatAddr("*", "tcp", false)
	if result != "*" {
		t.Errorf("formatAddr('*') = %q, want '*'", result)
	}
}

func TestFormatAddr_NoColon(t *testing.T) {
	result := formatAddr("127.0.0.1", "tcp", false)
	if result != "127.0.0.1" {
		t.Errorf("formatAddr('127.0.0.1') = %q, want '127.0.0.1'", result)
	}
}

func TestFormatAddr_BasicIPv4(t *testing.T) {
	result := formatAddr("127.0.0.1:8080", "tcp", false)
	if result != "127.0.0.1:8080" {
		t.Errorf("formatAddr('127.0.0.1:8080') = %q, want '127.0.0.1:8080'", result)
	}
}

func TestFormatAddr_WithServiceNames(t *testing.T) {
	result := formatAddr("127.0.0.1:80", "tcp", true)
	if result != "127.0.0.1:http" {
		t.Errorf("formatAddr with service names = %q, want '127.0.0.1:http'", result)
	}
}

func TestFormatAddr_WithServiceNames_HTTPS(t *testing.T) {
	result := formatAddr("10.0.0.1:443", "tcp", true)
	if result != "10.0.0.1:https" {
		t.Errorf("formatAddr with service names = %q, want '10.0.0.1:https'", result)
	}
}

func TestFormatAddr_WithServiceNames_DNS(t *testing.T) {
	result := formatAddr("8.8.8.8:53", "udp", true)
	if result != "8.8.8.8:dns" {
		t.Errorf("formatAddr with service names = %q, want '8.8.8.8:dns'", result)
	}
}

func TestFormatAddr_WithServiceNames_UnknownPort(t *testing.T) {
	result := formatAddr("127.0.0.1:12345", "tcp", true)
	// Unknown port should stay as number
	if result != "127.0.0.1:12345" {
		t.Errorf("formatAddr with unknown port = %q, want '127.0.0.1:12345'", result)
	}
}

func TestFormatAddr_WithDNSCache(t *testing.T) {
	cache := map[string]string{
		"8.8.8.8": "dns.google",
	}
	result := formatAddr("8.8.8.8:443", "tcp", false, cache)
	if result != "dns.google:443" {
		t.Errorf("formatAddr with DNS cache = %q, want 'dns.google:443'", result)
	}
}

func TestFormatAddr_WithDNSCacheAndServiceNames(t *testing.T) {
	cache := map[string]string{
		"1.1.1.1": "one.cloudflare",
	}
	result := formatAddr("1.1.1.1:443", "tcp", true, cache)
	if result != "one.cloudflare:https" {
		t.Errorf("formatAddr with DNS + service names = %q, want 'one.cloudflare:https'", result)
	}
}

func TestFormatAddr_DNSCacheMiss(t *testing.T) {
	cache := map[string]string{
		"8.8.8.8": "dns.google",
	}
	result := formatAddr("1.1.1.1:443", "tcp", false, cache)
	// IP not in cache, should return IP
	if result != "1.1.1.1:443" {
		t.Errorf("formatAddr DNS cache miss = %q, want '1.1.1.1:443'", result)
	}
}

func TestFormatAddr_EmptyDNSCache(t *testing.T) {
	cache := map[string]string{}
	result := formatAddr("8.8.8.8:443", "tcp", false, cache)
	if result != "8.8.8.8:443" {
		t.Errorf("formatAddr empty cache = %q, want '8.8.8.8:443'", result)
	}
}

func TestFormatAddr_NilDNSCache(t *testing.T) {
	result := formatAddr("8.8.8.8:443", "tcp", false, nil)
	if result != "8.8.8.8:443" {
		t.Errorf("formatAddr nil cache = %q, want '8.8.8.8:443'", result)
	}
}

// Tests for formatRemoteAddr (wrapper)

func TestFormatRemoteAddr_Basic(t *testing.T) {
	result := formatRemoteAddr("10.0.0.1:443", "tcp", nil, false)
	if result != "10.0.0.1:443" {
		t.Errorf("formatRemoteAddr = %q, want '10.0.0.1:443'", result)
	}
}

func TestFormatRemoteAddr_WithServiceNames(t *testing.T) {
	result := formatRemoteAddr("10.0.0.1:22", "tcp", nil, true)
	if result != "10.0.0.1:ssh" {
		t.Errorf("formatRemoteAddr with service names = %q, want '10.0.0.1:ssh'", result)
	}
}

func TestFormatRemoteAddr_WithDNSCache(t *testing.T) {
	cache := map[string]string{
		"10.0.0.1": "server.local",
	}
	result := formatRemoteAddr("10.0.0.1:22", "tcp", cache, true)
	if result != "server.local:ssh" {
		t.Errorf("formatRemoteAddr = %q, want 'server.local:ssh'", result)
	}
}

// Tests for truncateString

func TestTruncateString_Short(t *testing.T) {
	result := truncateString("hello", 10)
	if result != "hello" {
		t.Errorf("truncateString short = %q, want 'hello'", result)
	}
}

func TestTruncateString_ExactLength(t *testing.T) {
	result := truncateString("hello", 5)
	if result != "hello" {
		t.Errorf("truncateString exact = %q, want 'hello'", result)
	}
}

func TestTruncateString_NeedsTruncation(t *testing.T) {
	result := truncateString("hello world", 8)
	if result != "hello..." {
		t.Errorf("truncateString truncate = %q, want 'hello...'", result)
	}
}

func TestTruncateString_VeryShortMaxLen(t *testing.T) {
	result := truncateString("hello", 3)
	if result != "hel" {
		t.Errorf("truncateString very short = %q, want 'hel'", result)
	}
}

func TestTruncateString_MaxLen4(t *testing.T) {
	result := truncateString("hello", 4)
	if result != "h..." {
		t.Errorf("truncateString maxLen 4 = %q, want 'h...'", result)
	}
}

func TestTruncateString_EmptyString(t *testing.T) {
	result := truncateString("", 10)
	if result != "" {
		t.Errorf("truncateString empty = %q, want ''", result)
	}
}
