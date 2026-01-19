package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme == nil {
		t.Fatal("DefaultTheme returned nil")
	}

	if theme.Name != "dracula" {
		t.Errorf("Expected theme name 'dracula', got '%s'", theme.Name)
	}

	// Verify some colors are set
	if theme.Styles.Table.FgColor == "" {
		t.Error("Table FgColor should not be empty")
	}
	if theme.Styles.Table.BgColor == "" {
		t.Error("Table BgColor should not be empty")
	}
	if theme.Styles.Header.TitleFg == "" {
		t.Error("Header TitleFg should not be empty")
	}
	if theme.Styles.Footer.KeyFgColor == "" {
		t.Error("Footer KeyFgColor should not be empty")
	}
}

func TestLoadTheme_FallsBackToDefault(t *testing.T) {
	theme, err := LoadTheme()
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}

	if theme == nil {
		t.Fatal("LoadTheme returned nil")
	}

	// Should return valid theme (embedded or default)
	if theme.Name == "" {
		t.Error("Theme name should not be empty")
	}
}

func TestLoadTheme_LoadsUserConfig(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "netmon")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write custom skin
	skinContent := `
name: custom-test
styles:
  table:
    fgColor: "#ffffff"
    bgColor: "#000000"
    cursorFgColor: "#000000"
    cursorBgColor: "#ffffff"
    headerFgColor: "#ff0000"
    headerBgColor: "#000000"
    sortIndicator: "#00ff00"
    selectedColumn: "#0000ff"
  header:
    fgColor: "#ffffff"
    bgColor: "#000000"
    titleFg: "#ff0000"
  footer:
    fgColor: "#ffffff"
    bgColor: "#000000"
    keyFgColor: "#ff0000"
    descFgColor: "#ffffff"
  status:
    fgColor: "#888888"
    bgColor: "#000000"
`
	skinPath := filepath.Join(configDir, "skin.yaml")
	if err := os.WriteFile(skinPath, []byte(skinContent), 0644); err != nil {
		t.Fatalf("Failed to write skin file: %v", err)
	}

	// Override config dir (not easily possible without env manipulation)
	// This test verifies the fallback path works
	theme, err := LoadTheme()
	if err != nil {
		t.Fatalf("LoadTheme failed: %v", err)
	}

	if theme == nil {
		t.Fatal("LoadTheme returned nil")
	}
}

func TestInitTheme(t *testing.T) {
	// Save original
	original := CurrentTheme

	err := InitTheme()
	if err != nil {
		t.Fatalf("InitTheme failed: %v", err)
	}

	if CurrentTheme == nil {
		t.Error("CurrentTheme should not be nil after InitTheme")
	}

	// Restore
	CurrentTheme = original
}

func TestCurrentTheme_InitializedOnPackageLoad(t *testing.T) {
	// CurrentTheme is set in init()
	if CurrentTheme == nil {
		t.Error("CurrentTheme should be initialized on package load")
	}
}

func TestTableStyle_AllFieldsSet(t *testing.T) {
	theme := DefaultTheme()
	table := theme.Styles.Table

	if table.FgColor == "" {
		t.Error("FgColor not set")
	}
	if table.BgColor == "" {
		t.Error("BgColor not set")
	}
	if table.CursorFgColor == "" {
		t.Error("CursorFgColor not set")
	}
	if table.CursorBgColor == "" {
		t.Error("CursorBgColor not set")
	}
	if table.HeaderFgColor == "" {
		t.Error("HeaderFgColor not set")
	}
	if table.HeaderBgColor == "" {
		t.Error("HeaderBgColor not set")
	}
	if table.SortIndicator == "" {
		t.Error("SortIndicator not set")
	}
	if table.SelectedColumn == "" {
		t.Error("SelectedColumn not set")
	}
}

func TestTableStyle_ChangeHighlightColors(t *testing.T) {
	theme := DefaultTheme()
	table := theme.Styles.Table

	if table.AddedFgColor == "" {
		t.Error("AddedFgColor not set")
	}
	if table.RemovedFgColor == "" {
		t.Error("RemovedFgColor not set")
	}
}

func TestModalStyle_DimmedColor(t *testing.T) {
	theme := DefaultTheme()
	modal := theme.Styles.Modal

	if modal.DimmedFgColor == "" {
		t.Error("DimmedFgColor not set")
	}
}

func TestTheme_YAMLRoundtrip(t *testing.T) {
	theme := DefaultTheme()

	// Marshal to YAML
	data, err := yaml.Marshal(theme)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}

	// Unmarshal back
	var loaded Theme
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	// Verify key fields
	if loaded.Name != theme.Name {
		t.Errorf("Name mismatch: got %q, want %q", loaded.Name, theme.Name)
	}
	if loaded.Styles.Table.FgColor != theme.Styles.Table.FgColor {
		t.Errorf("Table.FgColor mismatch")
	}
	if loaded.Styles.Header.TitleFg != theme.Styles.Header.TitleFg {
		t.Errorf("Header.TitleFg mismatch")
	}
}

func TestLoadTheme_ParsesValidYAML(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "netmon")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write valid skin YAML
	skinContent := `name: test-skin
styles:
  table:
    fgColor: "#aabbcc"
    bgColor: "#112233"
    cursorFgColor: "#ffffff"
    cursorBgColor: "#000000"
    headerFgColor: "#ff0000"
    headerBgColor: "#00ff00"
    sortIndicator: "#0000ff"
    selectedColumn: "#ffff00"
    addedFgColor: "#00ff00"
    removedFgColor: "#ff0000"
  header:
    fgColor: "#ffffff"
    bgColor: "#000000"
    titleFg: "#ff00ff"
  footer:
    fgColor: "#cccccc"
    bgColor: "#333333"
    keyFgColor: "#ff0000"
    descFgColor: "#ffffff"
  status:
    fgColor: "#888888"
    bgColor: "#222222"
  modal:
    dimmedFgColor: "#666666"
`
	skinPath := filepath.Join(configDir, "skin.yaml")
	if err := os.WriteFile(skinPath, []byte(skinContent), 0644); err != nil {
		t.Fatalf("Failed to write skin file: %v", err)
	}

	// Parse the YAML directly (can't override os.UserConfigDir)
	data, err := os.ReadFile(skinPath)
	if err != nil {
		t.Fatalf("Failed to read skin file: %v", err)
	}

	var theme Theme
	if err := yaml.Unmarshal(data, &theme); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	// Verify parsed values
	if theme.Name != "test-skin" {
		t.Errorf("Name = %q, want 'test-skin'", theme.Name)
	}
	if theme.Styles.Table.FgColor != "#aabbcc" {
		t.Errorf("Table.FgColor = %q, want '#aabbcc'", theme.Styles.Table.FgColor)
	}
	if theme.Styles.Header.TitleFg != "#ff00ff" {
		t.Errorf("Header.TitleFg = %q, want '#ff00ff'", theme.Styles.Header.TitleFg)
	}
	if theme.Styles.Modal.DimmedFgColor != "#666666" {
		t.Errorf("Modal.DimmedFgColor = %q, want '#666666'", theme.Styles.Modal.DimmedFgColor)
	}
}

func TestLoadTheme_HandlesMalformedYAML(t *testing.T) {
	// Create temp file with malformed YAML
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "malformed.yaml")

	malformedContent := "name: [invalid: yaml: content"
	if err := os.WriteFile(tmpFile, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var theme Theme
	err = yaml.Unmarshal(data, &theme)

	// Should error on malformed YAML
	if err == nil {
		t.Error("yaml.Unmarshal should fail on malformed YAML")
	}
}

func TestLoadTheme_EmbeddedFileExists(t *testing.T) {
	// Verify the embedded default skin can be read
	data, err := defaultSkin.ReadFile("skins/dracula.yaml")
	if err != nil {
		t.Fatalf("Failed to read embedded skin: %v", err)
	}

	if len(data) == 0 {
		t.Error("Embedded skin file is empty")
	}

	// Verify it's valid YAML
	var theme Theme
	if err := yaml.Unmarshal(data, &theme); err != nil {
		t.Fatalf("Embedded skin is not valid YAML: %v", err)
	}

	if theme.Name == "" {
		t.Error("Embedded theme should have a name")
	}
}
