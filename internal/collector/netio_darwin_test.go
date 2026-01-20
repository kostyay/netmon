//go:build darwin

package collector

import (
	"context"
	"testing"
)

func TestParseNettopOutput_Basic(t *testing.T) {
	output := `bytes_in	bytes_out
process1.123	1000	2000
process2.456	3000	4000
`
	stats, err := parseNettopOutput(output)
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}

	if stats[123] == nil {
		t.Fatal("expected stats for PID 123")
	}
	if stats[123].BytesRecv != 1000 {
		t.Errorf("PID 123 BytesRecv = %d, want 1000", stats[123].BytesRecv)
	}
	if stats[123].BytesSent != 2000 {
		t.Errorf("PID 123 BytesSent = %d, want 2000", stats[123].BytesSent)
	}

	if stats[456] == nil {
		t.Fatal("expected stats for PID 456")
	}
	if stats[456].BytesRecv != 3000 {
		t.Errorf("PID 456 BytesRecv = %d, want 3000", stats[456].BytesRecv)
	}
}

func TestParseNettopOutput_LeadingSpaces(t *testing.T) {
	// nettop commonly outputs lines with leading whitespace
	output := `bytes_in	bytes_out
 process1.123	1000	2000
  process2.456	3000	4000
   process3.789	5000	6000
`
	stats, err := parseNettopOutput(output)
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}

	// Fixed: lines with leading spaces should be parsed correctly
	if len(stats) != 3 {
		t.Errorf("expected 3 stats after trimming leading spaces, got %d", len(stats))
	}
	if stats[123] == nil {
		t.Error("expected stats for PID 123")
	}
	if stats[456] == nil {
		t.Error("expected stats for PID 456")
	}
	if stats[789] == nil {
		t.Error("expected stats for PID 789")
	}
}

func TestParseNettopOutput_MixedLeadingSpaces(t *testing.T) {
	// Mix of lines with and without leading spaces
	output := `bytes_in	bytes_out
process1.123	1000	2000
 process2.456	3000	4000
process3.789	5000	6000
`
	stats, err := parseNettopOutput(output)
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}

	// All processes should be parsed correctly now
	if len(stats) != 3 {
		t.Errorf("expected 3 stats, got %d", len(stats))
	}
	if _, ok := stats[123]; !ok {
		t.Error("expected stats for PID 123")
	}
	if _, ok := stats[456]; !ok {
		t.Error("expected stats for PID 456 (leading space trimmed)")
	}
	if _, ok := stats[789]; !ok {
		t.Error("expected stats for PID 789")
	}
}

func TestParseNettopOutput_EmptyOutput(t *testing.T) {
	stats, err := parseNettopOutput("")
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats for empty output, got %d", len(stats))
	}
}

func TestParseNettopOutput_HeaderOnly(t *testing.T) {
	output := "bytes_in	bytes_out\n"
	stats, err := parseNettopOutput(output)
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats for header-only output, got %d", len(stats))
	}
}

func TestParseNettopOutput_AggregatesSamePID(t *testing.T) {
	// Same PID with multiple entries (different interfaces/connections)
	output := `bytes_in	bytes_out
process1.123	1000	2000
process1.123	500	300
`
	stats, err := parseNettopOutput(output)
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("expected 1 stat (aggregated), got %d", len(stats))
	}

	if stats[123].BytesRecv != 1500 {
		t.Errorf("aggregated BytesRecv = %d, want 1500", stats[123].BytesRecv)
	}
	if stats[123].BytesSent != 2300 {
		t.Errorf("aggregated BytesSent = %d, want 2300", stats[123].BytesSent)
	}
}

func TestParseNettopOutput_MalformedLines(t *testing.T) {
	output := `bytes_in	bytes_out
process1.123	1000	2000
malformed_no_pid	1000	2000
process2	not_a_number	2000
process3.456	1000
`
	stats, err := parseNettopOutput(output)
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}

	// Only valid line (process1.123) should be parsed
	if len(stats) != 1 {
		t.Errorf("expected 1 valid stat, got %d", len(stats))
	}
	if stats[123] == nil {
		t.Error("expected stats for PID 123")
	}
}

func TestParseNettopOutput_LargeNumbers(t *testing.T) {
	output := `bytes_in	bytes_out
process1.123	18446744073709551615	18446744073709551615
`
	stats, err := parseNettopOutput(output)
	if err != nil {
		t.Fatalf("parseNettopOutput failed: %v", err)
	}

	if stats[123] == nil {
		t.Fatal("expected stats for PID 123")
	}
	// Max uint64
	expected := uint64(18446744073709551615)
	if stats[123].BytesRecv != expected {
		t.Errorf("BytesRecv = %d, want %d", stats[123].BytesRecv, expected)
	}
}

func TestNetIOCollector_Collect(t *testing.T) {
	c := NewNetIOCollector()
	ctx := context.Background()

	stats, err := c.Collect(ctx)
	if err != nil {
		t.Logf("Collect returned error (may be expected): %v", err)
	}

	// Stats may be empty if nettop is unavailable or no network activity
	if stats == nil {
		t.Error("Collect should return non-nil map")
	}
}

func TestNetIOCollector_Interface(t *testing.T) {
	var _ NetIOCollector = &netIOCollector{}
}
