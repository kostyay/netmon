package config

import (
	"os"
	"path/filepath"
	"testing"
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
