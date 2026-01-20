package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/kostyay/netmon/internal/model"
)

func TestRenderJSON(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Applications: []model.Application{
			{
				Name:             "testapp",
				PIDs:             []int32{123, 456},
				EstablishedCount: 2,
				ListenCount:      1,
				Connections: []model.Connection{
					{
						PID:        123,
						Protocol:   model.ProtocolTCP,
						LocalAddr:  "127.0.0.1:8080",
						RemoteAddr: "10.0.0.1:443",
						State:      model.StateEstablished,
					},
					{
						PID:        123,
						Protocol:   model.ProtocolTCP,
						LocalAddr:  "0.0.0.0:8080",
						RemoteAddr: "*",
						State:      model.StateListen,
					},
					{
						PID:        456,
						Protocol:   model.ProtocolUDP,
						LocalAddr:  "127.0.0.1:5353",
						RemoteAddr: "*",
						State:      model.StateNone,
					},
				},
			},
		},
		SkippedCount: 5,
	}

	ioStats := map[int32]*model.NetIOStats{
		123: {BytesSent: 1000, BytesRecv: 2000},
		456: {BytesSent: 500, BytesRecv: 750},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, snapshot, ioStats)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	// Verify structure
	if len(output.Applications) != 1 {
		t.Errorf("Expected 1 application, got %d", len(output.Applications))
	}

	app := output.Applications[0]
	if app.Name != "testapp" {
		t.Errorf("Expected name 'testapp', got '%s'", app.Name)
	}
	if app.ConnectionCount != 3 {
		t.Errorf("Expected 3 connections, got %d", app.ConnectionCount)
	}
	if app.EstablishedCount != 2 {
		t.Errorf("Expected 2 established, got %d", app.EstablishedCount)
	}
	if app.ListenCount != 1 {
		t.Errorf("Expected 1 listen, got %d", app.ListenCount)
	}

	// Verify I/O stats aggregation (123 + 456)
	if app.BytesSent != 1500 {
		t.Errorf("Expected BytesSent 1500, got %d", app.BytesSent)
	}
	if app.BytesRecv != 2750 {
		t.Errorf("Expected BytesRecv 2750, got %d", app.BytesRecv)
	}

	if output.SkippedCount != 5 {
		t.Errorf("Expected SkippedCount 5, got %d", output.SkippedCount)
	}
}

func TestRenderJSON_EmptySnapshot(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Timestamp:    time.Now(),
		Applications: []model.Application{},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, snapshot, nil)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if len(output.Applications) != 0 {
		t.Errorf("Expected 0 applications, got %d", len(output.Applications))
	}
}

func TestRenderJSON_NilIOStats(t *testing.T) {
	snapshot := &model.NetworkSnapshot{
		Timestamp: time.Now(),
		Applications: []model.Application{
			{
				Name: "app",
				PIDs: []int32{100},
			},
		},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, snapshot, nil)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	// Should have zero bytes when no IO stats
	if output.Applications[0].BytesSent != 0 {
		t.Errorf("Expected 0 BytesSent, got %d", output.Applications[0].BytesSent)
	}
}
