package services

import "testing"

func TestLookup(t *testing.T) {
	tests := []struct {
		port  int
		proto string
		want  string
	}{
		{22, "tcp", "ssh"},
		{80, "tcp", "http"},
		{443, "tcp", "https"},
		{53, "tcp", "dns"},
		{53, "udp", "dns"},
		{3306, "tcp", "mysql"},
		{5432, "tcp", "postgresql"},
		{6379, "tcp", "redis"},
		{12345, "tcp", ""}, // Unknown port
		{80, "udp", ""},    // HTTP is TCP only
	}

	for _, tt := range tests {
		got := Lookup(tt.port, tt.proto)
		if got != tt.want {
			t.Errorf("Lookup(%d, %q) = %q, want %q", tt.port, tt.proto, got, tt.want)
		}
	}
}

func TestLookupTCP(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{22, "ssh"},
		{80, "http"},
		{443, "https"},
		{8080, "http-alt"},
		{99999, ""}, // Unknown
	}

	for _, tt := range tests {
		got := LookupTCP(tt.port)
		if got != tt.want {
			t.Errorf("LookupTCP(%d) = %q, want %q", tt.port, got, tt.want)
		}
	}
}

func TestLookupUDP(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{53, "dns"},
		{123, "ntp"},
		{514, "syslog"},
		{67, "dhcp"},
		{80, ""}, // HTTP is TCP only
	}

	for _, tt := range tests {
		got := LookupUDP(tt.port)
		if got != tt.want {
			t.Errorf("LookupUDP(%d) = %q, want %q", tt.port, got, tt.want)
		}
	}
}
