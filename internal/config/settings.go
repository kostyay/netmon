package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Settings holds user-configurable options.
type Settings struct {
	DNSEnabled       bool `yaml:"dnsEnabled"`
	ServiceNames     bool `yaml:"serviceNames"`
	HighlightChanges bool `yaml:"highlightChanges"`
	Animations       bool `yaml:"animations"`        // Enable UI animations (live pulse, spinners)
	DockerContainers bool `yaml:"dockerContainers"`   // Show Docker containers as virtual rows
}

// DefaultSettings returns the default settings.
func DefaultSettings() *Settings {
	return &Settings{
		DNSEnabled:       true, // On by default
		ServiceNames:     true, // On by default (no overhead)
		HighlightChanges: true, // On by default
		Animations:       true, // On by default
		DockerContainers: true, // On by default
	}
}

// settingsPath returns the path to the settings file.
func settingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "netmon", "settings.yaml"), nil
}

// LoadSettings loads settings from disk, returning defaults if not found.
func LoadSettings() (*Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings(), nil
	}

	// #nosec G304 - path is constructed from trusted sources
	data, err := os.ReadFile(path)
	if err != nil {
		// Return defaults if file doesn't exist
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return DefaultSettings(), err
	}

	var settings Settings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return DefaultSettings(), err
	}

	return &settings, nil
}

// SaveSettings writes settings to disk.
func SaveSettings(s *Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// CurrentSettings holds the loaded settings (singleton).
var CurrentSettings *Settings

// InitSettings initializes the global settings.
func InitSettings() error {
	settings, err := LoadSettings()
	if err != nil {
		return err
	}
	CurrentSettings = settings
	return nil
}

func init() {
	// Initialize with default settings on package load
	CurrentSettings = DefaultSettings()
}
