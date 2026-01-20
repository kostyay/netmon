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
		Name: "TestApp",
		PIDs: []int32{1234, 5678},
		Connections: []Connection{
			{Protocol: ProtocolTCP, LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: StateEstablished},
		},
	}

	if app.Name != "TestApp" {
		t.Errorf("Name = %s, want TestApp", app.Name)
	}
	if len(app.PIDs) != 2 {
		t.Errorf("PIDs length = %d, want 2", len(app.PIDs))
	}
}

// Tests for SelectionID helpers

func TestSelectionIDFromProcess(t *testing.T) {
	id := SelectionIDFromProcess("TestApp")

	if id.ProcessName != "TestApp" {
		t.Errorf("ProcessName = %q, want 'TestApp'", id.ProcessName)
	}
	if id.ConnectionKey != nil {
		t.Error("ConnectionKey should be nil for process selection")
	}
}

func TestSelectionIDFromProcess_EmptyName(t *testing.T) {
	id := SelectionIDFromProcess("")

	if id.ProcessName != "" {
		t.Errorf("ProcessName = %q, want ''", id.ProcessName)
	}
}

func TestSelectionIDFromConnection(t *testing.T) {
	id := SelectionIDFromConnection("TestApp", "127.0.0.1:80", "10.0.0.1:443")

	if id.ProcessName != "TestApp" {
		t.Errorf("ProcessName = %q, want 'TestApp'", id.ProcessName)
	}
	if id.ConnectionKey == nil {
		t.Fatal("ConnectionKey should not be nil")
	}
	if id.ConnectionKey.ProcessName != "TestApp" {
		t.Errorf("ConnectionKey.ProcessName = %q, want 'TestApp'", id.ConnectionKey.ProcessName)
	}
	if id.ConnectionKey.LocalAddr != "127.0.0.1:80" {
		t.Errorf("ConnectionKey.LocalAddr = %q, want '127.0.0.1:80'", id.ConnectionKey.LocalAddr)
	}
	if id.ConnectionKey.RemoteAddr != "10.0.0.1:443" {
		t.Errorf("ConnectionKey.RemoteAddr = %q, want '10.0.0.1:443'", id.ConnectionKey.RemoteAddr)
	}
}

func TestSelectionIDFromConnection_EmptyFields(t *testing.T) {
	id := SelectionIDFromConnection("", "", "")

	if id.ProcessName != "" {
		t.Errorf("ProcessName = %q, want ''", id.ProcessName)
	}
	if id.ConnectionKey == nil {
		t.Fatal("ConnectionKey should not be nil even with empty fields")
	}
}

// Tests for ExtractPort

func TestExtractPort_ValidIPv4(t *testing.T) {
	port := ExtractPort("127.0.0.1:8080")
	if port != 8080 {
		t.Errorf("ExtractPort('127.0.0.1:8080') = %d, want 8080", port)
	}
}

func TestExtractPort_ValidIPv6(t *testing.T) {
	port := ExtractPort("[::1]:443")
	if port != 443 {
		t.Errorf("ExtractPort('[::1]:443') = %d, want 443", port)
	}
}

func TestExtractPort_NoPort(t *testing.T) {
	port := ExtractPort("127.0.0.1")
	if port != 0 {
		t.Errorf("ExtractPort('127.0.0.1') = %d, want 0", port)
	}
}

func TestExtractPort_Asterisk(t *testing.T) {
	port := ExtractPort("*:80")
	if port != 80 {
		t.Errorf("ExtractPort('*:80') = %d, want 80", port)
	}
}

func TestExtractPort_EmptyString(t *testing.T) {
	port := ExtractPort("")
	if port != 0 {
		t.Errorf("ExtractPort('') = %d, want 0", port)
	}
}

func TestExtractPort_InvalidPort(t *testing.T) {
	port := ExtractPort("127.0.0.1:abc")
	if port != 0 {
		t.Errorf("ExtractPort('127.0.0.1:abc') = %d, want 0", port)
	}
}

func TestExtractPort_OnlyColon(t *testing.T) {
	port := ExtractPort(":")
	if port != 0 {
		t.Errorf("ExtractPort(':') = %d, want 0", port)
	}
}

func TestExtractPort_TrailingColon(t *testing.T) {
	port := ExtractPort("127.0.0.1:")
	if port != 0 {
		t.Errorf("ExtractPort('127.0.0.1:') = %d, want 0", port)
	}
}

// Tests for ConnectionKey

func TestConnectionKey_Struct(t *testing.T) {
	key := ConnectionKey{
		ProcessName: "App",
		LocalAddr:   "127.0.0.1:80",
		RemoteAddr:  "10.0.0.1:443",
	}

	if key.ProcessName != "App" {
		t.Errorf("ProcessName = %q, want 'App'", key.ProcessName)
	}
	if key.LocalAddr != "127.0.0.1:80" {
		t.Errorf("LocalAddr = %q, want '127.0.0.1:80'", key.LocalAddr)
	}
	if key.RemoteAddr != "10.0.0.1:443" {
		t.Errorf("RemoteAddr = %q, want '10.0.0.1:443'", key.RemoteAddr)
	}
}

func TestConnectionKey_Equality(t *testing.T) {
	key1 := ConnectionKey{ProcessName: "App", LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443"}
	key2 := ConnectionKey{ProcessName: "App", LocalAddr: "127.0.0.1:80", RemoteAddr: "10.0.0.1:443"}
	key3 := ConnectionKey{ProcessName: "App", LocalAddr: "127.0.0.1:81", RemoteAddr: "10.0.0.1:443"}

	if key1 != key2 {
		t.Error("Identical ConnectionKeys should be equal")
	}
	if key1 == key3 {
		t.Error("Different ConnectionKeys should not be equal")
	}
}

// Tests for NetIOStats

func TestNetIOStats_Struct(t *testing.T) {
	now := time.Now()
	stats := NetIOStats{
		BytesSent: 1024,
		BytesRecv: 2048,
		UpdatedAt: now,
	}

	if stats.BytesSent != 1024 {
		t.Errorf("BytesSent = %d, want 1024", stats.BytesSent)
	}
	if stats.BytesRecv != 2048 {
		t.Errorf("BytesRecv = %d, want 2048", stats.BytesRecv)
	}
	if stats.UpdatedAt != now {
		t.Error("UpdatedAt should match")
	}
}

// Tests for Protocol constants

func TestProtocolConstants(t *testing.T) {
	if ProtocolTCP != "TCP" {
		t.Errorf("ProtocolTCP = %q, want 'TCP'", ProtocolTCP)
	}
	if ProtocolUDP != "UDP" {
		t.Errorf("ProtocolUDP = %q, want 'UDP'", ProtocolUDP)
	}
	if ProtocolUnknown != "UNK" {
		t.Errorf("ProtocolUnknown = %q, want 'UNK'", ProtocolUnknown)
	}
}

// Tests for ConnectionState constants

func TestConnectionStateConstants(t *testing.T) {
	if StateEstablished != "ESTABLISHED" {
		t.Errorf("StateEstablished = %q, want 'ESTABLISHED'", StateEstablished)
	}
	if StateListen != "LISTEN" {
		t.Errorf("StateListen = %q, want 'LISTEN'", StateListen)
	}
	if StateTimeWait != "TIME_WAIT" {
		t.Errorf("StateTimeWait = %q, want 'TIME_WAIT'", StateTimeWait)
	}
	if StateCloseWait != "CLOSE_WAIT" {
		t.Errorf("StateCloseWait = %q, want 'CLOSE_WAIT'", StateCloseWait)
	}
	if StateNone != "-" {
		t.Errorf("StateNone = %q, want '-'", StateNone)
	}
}
