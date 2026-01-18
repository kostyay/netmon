package ui

import (
	"time"

	"github.com/kostyay/netmon/internal/model"
)

// ChangeType indicates whether a connection was added or removed.
type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeRemoved
)

// ConnectionKey uniquely identifies a connection for diffing.
type ConnectionKey struct {
	PID        int32
	Protocol   model.Protocol
	LocalAddr  string
	RemoteAddr string
}

// Change represents a detected connection change.
type Change struct {
	Type      ChangeType
	Timestamp time.Time
}

// KeyFromConnection creates a ConnectionKey from a Connection.
func KeyFromConnection(c model.Connection) ConnectionKey {
	return ConnectionKey{
		PID:        c.PID,
		Protocol:   c.Protocol,
		LocalAddr:  c.LocalAddr,
		RemoteAddr: c.RemoteAddr,
	}
}

// GetChange returns a pointer to the Change for a connection, or nil if no change.
func (m Model) GetChange(c model.Connection) *Change {
	key := KeyFromConnection(c)
	if change, ok := m.changes[key]; ok {
		return &change
	}
	return nil
}

// pruneExpiredChanges removes changes older than maxAge.
func (m *Model) pruneExpiredChanges(maxAge time.Duration) {
	if m.changes == nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for key, change := range m.changes {
		if change.Timestamp.Before(cutoff) {
			delete(m.changes, key)
		}
	}
}

// diffConnections compares previous and current snapshots, returning changes map.
// The returned map contains new changes; caller should merge with existing changes.
func diffConnections(prev, curr *model.NetworkSnapshot) map[ConnectionKey]Change {
	if prev == nil || curr == nil {
		return nil
	}

	now := time.Now()
	changes := make(map[ConnectionKey]Change)

	// Build sets of connections
	prevSet := make(map[ConnectionKey]struct{})
	currSet := make(map[ConnectionKey]struct{})

	for _, app := range prev.Applications {
		for _, conn := range app.Connections {
			prevSet[KeyFromConnection(conn)] = struct{}{}
		}
	}

	for _, app := range curr.Applications {
		for _, conn := range app.Connections {
			currSet[KeyFromConnection(conn)] = struct{}{}
		}
	}

	// Find added connections (in curr but not in prev)
	for key := range currSet {
		if _, found := prevSet[key]; !found {
			changes[key] = Change{Type: ChangeAdded, Timestamp: now}
		}
	}

	// Find removed connections (in prev but not in curr)
	for key := range prevSet {
		if _, found := currSet[key]; !found {
			changes[key] = Change{Type: ChangeRemoved, Timestamp: now}
		}
	}

	return changes
}
