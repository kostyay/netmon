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
