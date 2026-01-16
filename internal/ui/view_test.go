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

	// Connections level should show Back and Sort
	if !strings.Contains(result, "Back") {
		t.Error("Connections keybindings should contain 'Back'")
	}
	if !strings.Contains(result, "Sort") {
		t.Error("Connections keybindings should contain 'Sort'")
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
	if !strings.Contains(result, "Navigate") {
		t.Error("Footer should contain 'Navigate' keybinding")
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
