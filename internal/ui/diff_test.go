package ui

import (
	"testing"
	"time"

	"github.com/kostyay/netmon/internal/model"
)

func TestDiffConnections_NilInputs(t *testing.T) {
	// Nil prev
	changes := diffConnections(nil, &model.NetworkSnapshot{})
	if changes != nil {
		t.Error("expected nil changes for nil prev")
	}

	// Nil curr
	changes = diffConnections(&model.NetworkSnapshot{}, nil)
	if changes != nil {
		t.Error("expected nil changes for nil curr")
	}

	// Both nil
	changes = diffConnections(nil, nil)
	if changes != nil {
		t.Error("expected nil changes for both nil")
	}
}

func TestDiffConnections_EmptySnapshots(t *testing.T) {
	prev := &model.NetworkSnapshot{
		Applications: []model.Application{},
		Timestamp:    time.Now(),
	}
	curr := &model.NetworkSnapshot{
		Applications: []model.Application{},
		Timestamp:    time.Now(),
	}

	changes := diffConnections(prev, curr)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestDiffConnections_NewConnections(t *testing.T) {
	prev := &model.NetworkSnapshot{
		Applications: []model.Application{},
		Timestamp:    time.Now(),
	}
	curr := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
				},
			},
		},
		Timestamp: time.Now(),
	}

	changes := diffConnections(prev, curr)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	key := ConnectionKey{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"}
	change, ok := changes[key]
	if !ok {
		t.Fatal("expected change for new connection")
	}
	if change.Type != ChangeAdded {
		t.Errorf("expected ChangeAdded, got %v", change.Type)
	}
}

func TestDiffConnections_RemovedConnections(t *testing.T) {
	prev := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished},
				},
			},
		},
		Timestamp: time.Now(),
	}
	curr := &model.NetworkSnapshot{
		Applications: []model.Application{},
		Timestamp:    time.Now(),
	}

	changes := diffConnections(prev, curr)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	key := ConnectionKey{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"}
	change, ok := changes[key]
	if !ok {
		t.Fatal("expected change for removed connection")
	}
	if change.Type != ChangeRemoved {
		t.Errorf("expected ChangeRemoved, got %v", change.Type)
	}
}

func TestDiffConnections_NoChanges(t *testing.T) {
	conn := model.Connection{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443", State: model.StateEstablished}
	prev := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", PIDs: []int32{100}, Connections: []model.Connection{conn}},
		},
		Timestamp: time.Now(),
	}
	curr := &model.NetworkSnapshot{
		Applications: []model.Application{
			{Name: "App1", PIDs: []int32{100}, Connections: []model.Connection{conn}},
		},
		Timestamp: time.Now(),
	}

	changes := diffConnections(prev, curr)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical snapshots, got %d", len(changes))
	}
}

func TestDiffConnections_MixedChanges(t *testing.T) {
	prev := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"},
					{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:9090", RemoteAddr: "10.0.0.2:80"},
				},
			},
		},
		Timestamp: time.Now(),
	}
	curr := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "App1",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:8080", RemoteAddr: "10.0.0.1:443"}, // unchanged
					{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:7070", RemoteAddr: "10.0.0.3:22"},  // new
				},
			},
		},
		Timestamp: time.Now(),
	}

	changes := diffConnections(prev, curr)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes (1 added, 1 removed), got %d", len(changes))
	}

	// Check added
	addedKey := ConnectionKey{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:7070", RemoteAddr: "10.0.0.3:22"}
	if change, ok := changes[addedKey]; !ok || change.Type != ChangeAdded {
		t.Error("expected ChangeAdded for new connection")
	}

	// Check removed
	removedKey := ConnectionKey{PID: 100, Protocol: "TCP", LocalAddr: "127.0.0.1:9090", RemoteAddr: "10.0.0.2:80"}
	if change, ok := changes[removedKey]; !ok || change.Type != ChangeRemoved {
		t.Error("expected ChangeRemoved for removed connection")
	}
}

func TestKeyFromConnection(t *testing.T) {
	conn := model.Connection{
		PID:        123,
		Protocol:   model.ProtocolTCP,
		LocalAddr:  "192.168.1.1:8080",
		RemoteAddr: "10.0.0.1:443",
		State:      model.StateEstablished,
	}

	key := KeyFromConnection(conn)

	if key.PID != 123 {
		t.Errorf("PID = %d, want 123", key.PID)
	}
	if key.Protocol != model.ProtocolTCP {
		t.Errorf("Protocol = %v, want TCP", key.Protocol)
	}
	if key.LocalAddr != "192.168.1.1:8080" {
		t.Errorf("LocalAddr = %s, want 192.168.1.1:8080", key.LocalAddr)
	}
	if key.RemoteAddr != "10.0.0.1:443" {
		t.Errorf("RemoteAddr = %s, want 10.0.0.1:443", key.RemoteAddr)
	}
}

func TestPruneExpiredChanges(t *testing.T) {
	m := Model{
		changes: make(map[ConnectionKey]Change),
	}

	now := time.Now()
	oldKey := ConnectionKey{PID: 1, Protocol: "TCP", LocalAddr: "a", RemoteAddr: "b"}
	newKey := ConnectionKey{PID: 2, Protocol: "TCP", LocalAddr: "c", RemoteAddr: "d"}

	m.changes[oldKey] = Change{Type: ChangeAdded, Timestamp: now.Add(-5 * time.Second)}
	m.changes[newKey] = Change{Type: ChangeAdded, Timestamp: now.Add(-1 * time.Second)}

	m.pruneExpiredChanges(3 * time.Second)

	if _, ok := m.changes[oldKey]; ok {
		t.Error("expected old change to be pruned")
	}
	if _, ok := m.changes[newKey]; !ok {
		t.Error("expected new change to remain")
	}
}

func TestPruneExpiredChanges_NilMap(t *testing.T) {
	m := Model{changes: nil}
	// Should not panic
	m.pruneExpiredChanges(3 * time.Second)
}
