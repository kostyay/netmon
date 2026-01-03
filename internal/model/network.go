package model

import (
	"sort"
	"time"
)

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
	Protocol   Protocol        // TCP or UDP
	LocalAddr  string          // e.g., 127.0.0.1:52341
	RemoteAddr string          // e.g., 142.250.80.46:443 or * for listening
	State      ConnectionState // e.g., ESTABLISHED, LISTEN, - for UDP
}

// Application represents a grouped set of connections by app name.
type Application struct {
	Name        string       // Process name (e.g., Chrome)
	PIDs        []int32      // All PIDs running this app
	Connections []Connection // All connections across all PIDs
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
