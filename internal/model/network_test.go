package model

import (
	"testing"
	"time"
)

func TestApplicationConnectionCount_Empty(t *testing.T) {
	app := Application{
		Name:        "TestApp",
		PIDs:        []int32{1234},
		Connections: []Connection{},
	}

	if got := app.ConnectionCount(); got != 0 {
		t.Errorf("ConnectionCount() = %d, want 0", got)
	}
}

func TestApplicationConnectionCount_SingleConnection(t *testing.T) {
	app := Application{
		Name: "TestApp",
		PIDs: []int32{1234},
		Connections: []Connection{
			{Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: "ESTABLISHED"},
		},
	}

	if got := app.ConnectionCount(); got != 1 {
		t.Errorf("ConnectionCount() = %d, want 1", got)
	}
}

func TestApplicationConnectionCount_MultipleConnections(t *testing.T) {
	app := Application{
		Name: "TestApp",
		PIDs: []int32{1234, 5678},
		Connections: []Connection{
			{Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: "ESTABLISHED"},
			{Protocol: "TCP", LocalAddr: "127.0.0.1:8081", RemoteAddr: "10.0.0.2:443", State: "ESTABLISHED"},
			{Protocol: "UDP", LocalAddr: "127.0.0.1:53", RemoteAddr: "*", State: "-"},
		},
	}

	if got := app.ConnectionCount(); got != 3 {
		t.Errorf("ConnectionCount() = %d, want 3", got)
	}
}

func TestNetworkSnapshotSortByConnectionCount_Empty(t *testing.T) {
	snapshot := NetworkSnapshot{
		Applications: []Application{},
		Timestamp:    time.Now(),
	}

	snapshot.SortByConnectionCount()

	if len(snapshot.Applications) != 0 {
		t.Errorf("SortByConnectionCount() modified empty slice, got len %d", len(snapshot.Applications))
	}
}

func TestNetworkSnapshotSortByConnectionCount_SingleApp(t *testing.T) {
	snapshot := NetworkSnapshot{
		Applications: []Application{
			{Name: "App1", Connections: []Connection{{Protocol: "TCP"}}},
		},
		Timestamp: time.Now(),
	}

	snapshot.SortByConnectionCount()

	if len(snapshot.Applications) != 1 || snapshot.Applications[0].Name != "App1" {
		t.Errorf("SortByConnectionCount() modified single app slice unexpectedly")
	}
}

func TestNetworkSnapshotSortByConnectionCount_MultipleApps(t *testing.T) {
	snapshot := NetworkSnapshot{
		Applications: []Application{
			{Name: "SmallApp", Connections: []Connection{{Protocol: "TCP"}}},
			{Name: "BigApp", Connections: []Connection{{Protocol: "TCP"}, {Protocol: "TCP"}, {Protocol: "UDP"}}},
			{Name: "MediumApp", Connections: []Connection{{Protocol: "TCP"}, {Protocol: "UDP"}}},
		},
		Timestamp: time.Now(),
	}

	snapshot.SortByConnectionCount()

	expected := []string{"BigApp", "MediumApp", "SmallApp"}
	for i, app := range snapshot.Applications {
		if app.Name != expected[i] {
			t.Errorf("SortByConnectionCount() at index %d: got %s, want %s", i, app.Name, expected[i])
		}
	}
}

func TestNetworkSnapshotSortByConnectionCount_EqualCounts(t *testing.T) {
	snapshot := NetworkSnapshot{
		Applications: []Application{
			{Name: "App1", Connections: []Connection{{Protocol: "TCP"}}},
			{Name: "App2", Connections: []Connection{{Protocol: "UDP"}}},
		},
		Timestamp: time.Now(),
	}

	snapshot.SortByConnectionCount()

	// With equal counts, order is stable (implementation detail, but we just check counts are equal)
	if len(snapshot.Applications) != 2 {
		t.Errorf("SortByConnectionCount() changed number of apps, got %d, want 2", len(snapshot.Applications))
	}
	if snapshot.Applications[0].ConnectionCount() != snapshot.Applications[1].ConnectionCount() {
		t.Errorf("Apps should have equal connection counts")
	}
}

func TestNetworkSnapshotTotalConnections_Empty(t *testing.T) {
	snapshot := NetworkSnapshot{
		Applications: []Application{},
		Timestamp:    time.Now(),
	}

	if got := snapshot.TotalConnections(); got != 0 {
		t.Errorf("TotalConnections() = %d, want 0", got)
	}
}

func TestNetworkSnapshotTotalConnections_SingleApp(t *testing.T) {
	snapshot := NetworkSnapshot{
		Applications: []Application{
			{Name: "App1", Connections: []Connection{{Protocol: "TCP"}, {Protocol: "UDP"}}},
		},
		Timestamp: time.Now(),
	}

	if got := snapshot.TotalConnections(); got != 2 {
		t.Errorf("TotalConnections() = %d, want 2", got)
	}
}

func TestNetworkSnapshotTotalConnections_MultipleApps(t *testing.T) {
	snapshot := NetworkSnapshot{
		Applications: []Application{
			{Name: "App1", Connections: []Connection{{Protocol: "TCP"}, {Protocol: "UDP"}}},
			{Name: "App2", Connections: []Connection{{Protocol: "TCP"}}},
			{Name: "App3", Connections: []Connection{{Protocol: "TCP"}, {Protocol: "TCP"}, {Protocol: "TCP"}}},
		},
		Timestamp: time.Now(),
	}

	if got := snapshot.TotalConnections(); got != 6 {
		t.Errorf("TotalConnections() = %d, want 6", got)
	}
}

func TestConnectionStruct(t *testing.T) {
	conn := Connection{
		Protocol:   "TCP",
		LocalAddr:  "127.0.0.1:8080",
		RemoteAddr: "192.168.1.1:443",
		State:      "ESTABLISHED",
	}

	if conn.Protocol != "TCP" {
		t.Errorf("Protocol = %s, want TCP", conn.Protocol)
	}
	if conn.LocalAddr != "127.0.0.1:8080" {
		t.Errorf("LocalAddr = %s, want 127.0.0.1:8080", conn.LocalAddr)
	}
	if conn.RemoteAddr != "192.168.1.1:443" {
		t.Errorf("RemoteAddr = %s, want 192.168.1.1:443", conn.RemoteAddr)
	}
	if conn.State != "ESTABLISHED" {
		t.Errorf("State = %s, want ESTABLISHED", conn.State)
	}
}

func TestApplicationStruct(t *testing.T) {
	app := Application{
		Name:     "TestApp",
		PIDs:     []int32{1234, 5678},
		Expanded: true,
		Connections: []Connection{
			{Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: "ESTABLISHED"},
		},
	}

	if app.Name != "TestApp" {
		t.Errorf("Name = %s, want TestApp", app.Name)
	}
	if len(app.PIDs) != 2 {
		t.Errorf("PIDs length = %d, want 2", len(app.PIDs))
	}
	if !app.Expanded {
		t.Errorf("Expanded = false, want true")
	}
}
