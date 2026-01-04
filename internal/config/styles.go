package config

import (
	"embed"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed skins/dracula.yaml
var defaultSkin embed.FS

// Color represents a hex color string.
type Color string

// TableStyle defines colors for the table view.
type TableStyle struct {
	FgColor         Color `yaml:"fgColor"`
	BgColor         Color `yaml:"bgColor"`
	CursorFgColor   Color `yaml:"cursorFgColor"`
	CursorBgColor   Color `yaml:"cursorBgColor"`
	HeaderFgColor   Color `yaml:"headerFgColor"`
	HeaderBgColor   Color `yaml:"headerBgColor"`
	SortIndicator   Color `yaml:"sortIndicator"`
	SelectedColumn  Color `yaml:"selectedColumn"`
}

// HeaderStyle defines colors for the header section.
type HeaderStyle struct {
	FgColor   Color `yaml:"fgColor"`
	BgColor   Color `yaml:"bgColor"`
	TitleFg   Color `yaml:"titleFg"`
}

// FooterStyle defines colors for the footer section.
type FooterStyle struct {
	FgColor     Color `yaml:"fgColor"`
	BgColor     Color `yaml:"bgColor"`
	KeyFgColor  Color `yaml:"keyFgColor"`
	DescFgColor Color `yaml:"descFgColor"`
}

// StatusStyle defines colors for status lines.
type StatusStyle struct {
	FgColor Color `yaml:"fgColor"`
	BgColor Color `yaml:"bgColor"`
}

// Styles holds all the theme colors.
type Styles struct {
	Table  TableStyle  `yaml:"table"`
	Header HeaderStyle `yaml:"header"`
	Footer FooterStyle `yaml:"footer"`
	Status StatusStyle `yaml:"status"`
}

// Theme is the top-level theme configuration.
type Theme struct {
	Name   string `yaml:"name"`
	Styles Styles `yaml:"styles"`
}

// DefaultTheme returns the built-in Dracula theme.
func DefaultTheme() *Theme {
	return &Theme{
		Name: "dracula",
		Styles: Styles{
			Table: TableStyle{
				FgColor:         "#f8f8f2",
				BgColor:         "#282a36",
				CursorFgColor:   "#282a36",
				CursorBgColor:   "#bd93f9",
				HeaderFgColor:   "#bd93f9",
				HeaderBgColor:   "#282a36",
				SortIndicator:   "#8be9fd",
				SelectedColumn:  "#f8f8f2",
			},
			Header: HeaderStyle{
				FgColor: "#f8f8f2",
				BgColor: "#282a36",
				TitleFg: "#bd93f9",
			},
			Footer: FooterStyle{
				FgColor:     "#f8f8f2",
				BgColor:     "#282a36",
				KeyFgColor:  "#ff79c6",
				DescFgColor: "#f8f8f2",
			},
			Status: StatusStyle{
				FgColor: "#6272a4",
				BgColor: "#282a36",
			},
		},
	}
}

// LoadTheme loads a theme from the user's config directory or returns the default.
func LoadTheme() (*Theme, error) {
	// Try user config first
	configDir, err := os.UserConfigDir()
	if err == nil {
		userSkinPath := filepath.Join(configDir, "netmon", "skin.yaml")
		// #nosec G304 - userSkinPath is constructed from trusted sources (UserConfigDir + hardcoded path)
		if data, err := os.ReadFile(userSkinPath); err == nil {
			var theme Theme
			if err := yaml.Unmarshal(data, &theme); err == nil {
				return &theme, nil
			}
		}
	}

	// Fall back to embedded default
	data, err := defaultSkin.ReadFile("skins/dracula.yaml")
	if err != nil {
		// If embedded file not found, return hardcoded default
		return DefaultTheme(), nil
	}

	var theme Theme
	if err := yaml.Unmarshal(data, &theme); err != nil {
		return DefaultTheme(), nil
	}

	return &theme, nil
}

// CurrentTheme holds the loaded theme (singleton).
var CurrentTheme *Theme

// InitTheme initializes the global theme.
func InitTheme() error {
	theme, err := LoadTheme()
	if err != nil {
		return err
	}
	CurrentTheme = theme
	return nil
}

func init() {
	// Initialize with default theme on package load
	CurrentTheme = DefaultTheme()
}
