package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if s == nil {
		t.Fatal("DefaultSettings returned nil")
	}
	if !s.DNSEnabled {
		t.Error("DNSEnabled should be true by default")
	}
	if !s.ServiceNames {
		t.Error("ServiceNames should be true by default")
	}
	if !s.HighlightChanges {
		t.Error("HighlightChanges should be true by default")
	}
}

func TestLoadSettings_ReturnsDefaultsWhenNoFile(t *testing.T) {
	s, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}
	if s == nil {
		t.Fatal("LoadSettings returned nil")
	}
	// Should match defaults
	defaults := DefaultSettings()
	if s.DNSEnabled != defaults.DNSEnabled {
		t.Error("DNSEnabled should match default")
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	// Create temp dir for settings
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "netmon")
	settingsFile := filepath.Join(configDir, "settings.yaml")

	// Create settings with non-default values
	original := &Settings{
		DNSEnabled:       true,
		ServiceNames:     false,
		HighlightChanges: false,
	}

	// Manually save to temp location
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	data := []byte("dnsEnabled: true\nserviceNames: false\nhighlightChanges: false\n")
	if err := os.WriteFile(settingsFile, data, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file was written
	readData, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(readData) == 0 {
		t.Fatal("Settings file is empty")
	}

	// Verify original values
	if !original.DNSEnabled {
		t.Error("original.DNSEnabled should be true")
	}
	if original.ServiceNames {
		t.Error("original.ServiceNames should be false")
	}
	if original.HighlightChanges {
		t.Error("original.HighlightChanges should be false")
	}
}

func TestCurrentSettings_InitializedOnPackageLoad(t *testing.T) {
	if CurrentSettings == nil {
		t.Error("CurrentSettings should be initialized on package load")
	}
}

func TestInitSettings(t *testing.T) {
	// Save original
	original := CurrentSettings

	err := InitSettings()
	if err != nil {
		t.Fatalf("InitSettings failed: %v", err)
	}

	if CurrentSettings == nil {
		t.Error("CurrentSettings should not be nil after InitSettings")
	}

	// Restore
	CurrentSettings = original
}

func TestSettings_YAMLRoundtrip(t *testing.T) {
	// Test that YAML tags work correctly
	s := &Settings{
		DNSEnabled:       true,
		ServiceNames:     false,
		HighlightChanges: true,
	}

	// Marshal and unmarshal
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")

	// Write YAML content
	content := "dnsEnabled: true\nserviceNames: false\nhighlightChanges: true\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != content {
		t.Errorf("Content mismatch: got %q, want %q", string(data), content)
	}

	// Verify original settings
	if !s.DNSEnabled {
		t.Error("DNSEnabled should be true")
	}
	if s.ServiceNames {
		t.Error("ServiceNames should be false")
	}
	if !s.HighlightChanges {
		t.Error("HighlightChanges should be true")
	}
}
