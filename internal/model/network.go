package model

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// NetIOStats represents network I/O statistics for a process.
type NetIOStats struct {
	BytesSent uint64    // Total bytes sent
	BytesRecv uint64    // Total bytes received
	UpdatedAt time.Time // When these stats were last updated
}

// Protocol represents a network protocol.
type Protocol string

const (
	ProtocolTCP     Protocol = "TCP"
	ProtocolUDP     Protocol = "UDP"
	ProtocolUnknown Protocol = "UNK"
)

// ConnectionState represents a TCP connection state.
type ConnectionState string

const (
	StateEstablished ConnectionState = "ESTABLISHED"
	StateListen      ConnectionState = "LISTEN"
	StateTimeWait    ConnectionState = "TIME_WAIT"
	StateCloseWait   ConnectionState = "CLOSE_WAIT"
	StateNone        ConnectionState = "-"
)

// Connection represents a single network connection.
type Connection struct {
	PID        int32           // Process ID owning this connection
	Protocol   Protocol        // TCP or UDP
	LocalAddr  string          // e.g., 127.0.0.1:52341
	RemoteAddr string          // e.g., 142.250.80.46:443 or * for listening
	State      ConnectionState // e.g., ESTABLISHED, LISTEN, - for UDP
}

// Application represents a grouped set of connections by app name.
type Application struct {
	Name             string       // Process name (e.g., Chrome)
	Exe              string       // Full path to executable (e.g., /usr/bin/chrome)
	PIDs             []int32      // All PIDs running this app
	Connections      []Connection // All connections across all PIDs
	EstablishedCount int          // Number of ESTABLISHED connections
	ListenCount      int          // Number of LISTEN connections
}

// ConnectionCount returns the number of connections for this application.
func (a *Application) ConnectionCount() int {
	return len(a.Connections)
}

// NetworkSnapshot represents all network data at a point in time.
type NetworkSnapshot struct {
	Applications []Application
	Timestamp    time.Time
	SkippedCount int // Number of connections skipped due to unknown process
}

// SortByConnectionCount sorts applications by number of connections (descending).
func (s *NetworkSnapshot) SortByConnectionCount() {
	sort.Slice(s.Applications, func(i, j int) bool {
		return len(s.Applications[i].Connections) > len(s.Applications[j].Connections)
	})
}

// TotalConnections returns the total number of connections across all apps.
func (s *NetworkSnapshot) TotalConnections() int {
	total := 0
	for _, app := range s.Applications {
		total += len(app.Connections)
	}
	return total
}

// ConnectionKey uniquely identifies a connection.
type ConnectionKey struct {
	ProcessName string
	LocalAddr   string
	RemoteAddr  string
}

// SelectionID identifies a selected item (process or connection).
type SelectionID struct {
	ProcessName   string
	ConnectionKey *ConnectionKey
}

// SelectionIDFromProcess creates a SelectionID for a process.
func SelectionIDFromProcess(name string) SelectionID {
	return SelectionID{ProcessName: name}
}

// SelectionIDFromConnection creates a SelectionID for a connection.
func SelectionIDFromConnection(processName, localAddr, remoteAddr string) SelectionID {
	return SelectionID{
		ProcessName: processName,
		ConnectionKey: &ConnectionKey{
			ProcessName: processName,
			LocalAddr:   localAddr,
			RemoteAddr:  remoteAddr,
		},
	}
}

// ExtractPort extracts port number from an address string like "127.0.0.1:8080".
// Returns 0 if the address doesn't contain a valid port.
func ExtractPort(addr string) int {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0
	}
	port, err := strconv.Atoi(addr[idx+1:])
	if err != nil {
		return 0
	}
	return port
}
