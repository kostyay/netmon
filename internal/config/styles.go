package config

import (
	"embed"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed skins/dracula.yaml skins/industrial.yaml
var defaultSkin embed.FS

// Color represents a hex color string.
type Color string

// TableStyle defines colors for the table view.
type TableStyle struct {
	FgColor        Color `yaml:"fgColor"`
	BgColor        Color `yaml:"bgColor"`
	CursorFgColor  Color `yaml:"cursorFgColor"`
	CursorBgColor  Color `yaml:"cursorBgColor"`
	HeaderFgColor  Color `yaml:"headerFgColor"`
	HeaderBgColor  Color `yaml:"headerBgColor"`
	SortIndicator  Color `yaml:"sortIndicator"`
	SelectedColumn Color `yaml:"selectedColumn"`
	AddedFgColor   Color `yaml:"addedFgColor"`   // New connections
	RemovedFgColor Color `yaml:"removedFgColor"` // Removed connections
}

// HeaderStyle defines colors for the header section.
type HeaderStyle struct {
	FgColor Color `yaml:"fgColor"`
	BgColor Color `yaml:"bgColor"`
	TitleFg Color `yaml:"titleFg"`
	LiveFg  Color `yaml:"liveFg"`  // Live indicator (green pulse)
	WarnFg  Color `yaml:"warnFg"`  // Warnings/attention (amber)
	StatsFg Color `yaml:"statsFg"` // Stats text (muted)
}

// FooterStyle defines colors for the footer section.
type FooterStyle struct {
	FgColor      Color `yaml:"fgColor"`
	BgColor      Color `yaml:"bgColor"`
	KeyFgColor   Color `yaml:"keyFgColor"`
	DescFgColor  Color `yaml:"descFgColor"`
	GroupFgColor Color `yaml:"groupFgColor"` // Group labels (NAV, ACTION)
}

// StatusStyle defines colors for status lines.
type StatusStyle struct {
	FgColor Color `yaml:"fgColor"`
	BgColor Color `yaml:"bgColor"`
}

// ModalStyle defines colors for modal dialogs.
type ModalStyle struct {
	DimmedFgColor Color `yaml:"dimmedFgColor"` // Dimmed background when modal visible
	BorderFgColor Color `yaml:"borderFgColor"` // Modal border
	AccentFgColor Color `yaml:"accentFgColor"` // Accent color for modal
}

// BorderStyle defines colors for borders.
type BorderStyle struct {
	FgColor       Color `yaml:"fgColor"`       // Default border color
	ActiveFgColor Color `yaml:"activeFgColor"` // Active/focused border
}

// Styles holds all the theme colors.
type Styles struct {
	Table  TableStyle  `yaml:"table"`
	Header HeaderStyle `yaml:"header"`
	Footer FooterStyle `yaml:"footer"`
	Status StatusStyle `yaml:"status"`
	Modal  ModalStyle  `yaml:"modal"`
	Border BorderStyle `yaml:"border"`
}

// Theme is the top-level theme configuration.
type Theme struct {
	Name   string `yaml:"name"`
	Styles Styles `yaml:"styles"`
}

// DefaultTheme returns the built-in Industrial theme.
func DefaultTheme() *Theme {
	return &Theme{
		Name: "industrial",
		Styles: Styles{
			Table: TableStyle{
				FgColor:        "#e6edf3",
				BgColor:        "#0d1117",
				CursorFgColor:  "#ffffff",
				CursorBgColor:  "#58a6ff",
				HeaderFgColor:  "#58a6ff",
				HeaderBgColor:  "#0d1117",
				SortIndicator:  "#3fb950",
				SelectedColumn: "#e6edf3",
				AddedFgColor:   "#3fb950",
				RemovedFgColor: "#f85149",
			},
			Header: HeaderStyle{
				FgColor: "#e6edf3",
				BgColor: "#0d1117",
				TitleFg: "#58a6ff",
				LiveFg:  "#3fb950",
				WarnFg:  "#d29922",
				StatsFg: "#7d8590",
			},
			Footer: FooterStyle{
				FgColor:      "#e6edf3",
				BgColor:      "#0d1117",
				KeyFgColor:   "#58a6ff",
				DescFgColor:  "#7d8590",
				GroupFgColor: "#e6edf3",
			},
			Status: StatusStyle{
				FgColor: "#7d8590",
				BgColor: "#0d1117",
			},
			Modal: ModalStyle{
				DimmedFgColor: "#7d8590",
				BorderFgColor: "#30363d",
				AccentFgColor: "#58a6ff",
			},
			Border: BorderStyle{
				FgColor:       "#30363d",
				ActiveFgColor: "#58a6ff",
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

	// Fall back to embedded default (industrial theme)
	data, err := defaultSkin.ReadFile("skins/industrial.yaml")
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
