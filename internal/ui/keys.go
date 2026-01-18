package ui

// Keybinding represents a keyboard shortcut with its display name.
type Keybinding struct {
	Key  string // actual key(s) to match
	Desc string // description for help display
}

// Global keybindings (always available)
var (
	KeyQuit        = Keybinding{Key: "q", Desc: "Quit"}
	KeyQuitAlt     = Keybinding{Key: "ctrl+c", Desc: "Quit"}
	KeyHelp        = Keybinding{Key: "?", Desc: "Show help"}
	KeySettings    = Keybinding{Key: "S", Desc: "Settings"}
	KeySearch      = Keybinding{Key: "/", Desc: "Search/filter"}
	KeyToggleView  = Keybinding{Key: "v", Desc: "Toggle grouped/flat view"}
	KeySortMode    = Keybinding{Key: "s", Desc: "Enter sort mode"}
	KeyRefreshUp   = Keybinding{Key: "+", Desc: "Increase refresh rate"}
	KeyRefreshDown = Keybinding{Key: "-", Desc: "Decrease refresh rate"}
)

// Navigation keybindings
var (
	KeyUp       = Keybinding{Key: "up", Desc: "Move up"}
	KeyUpAlt    = Keybinding{Key: "k", Desc: "Move up"}
	KeyDown     = Keybinding{Key: "down", Desc: "Move down"}
	KeyDownAlt  = Keybinding{Key: "j", Desc: "Move down"}
	KeyLeft     = Keybinding{Key: "left", Desc: "Move left (sort mode)"}
	KeyLeftAlt  = Keybinding{Key: "h", Desc: "Move left (sort mode)"}
	KeyRight    = Keybinding{Key: "right", Desc: "Move right (sort mode)"}
	KeyRightAlt = Keybinding{Key: "l", Desc: "Move right (sort mode)"}
	KeyEnter    = Keybinding{Key: "enter", Desc: "Select/drill-down"}
	KeySpace    = Keybinding{Key: " ", Desc: "Select/drill-down"}
	KeyEsc      = Keybinding{Key: "esc", Desc: "Back/cancel"}
	KeyBack     = Keybinding{Key: "backspace", Desc: "Back/cancel"}
)

// Kill keybindings
var (
	KeyKillTerm  = Keybinding{Key: "x", Desc: "Kill process (SIGTERM)"}
	KeyKillForce = Keybinding{Key: "X", Desc: "Force kill (SIGKILL)"}
)

// Confirm/cancel keybindings
var (
	KeyConfirmYes = Keybinding{Key: "y", Desc: "Confirm"}
	KeyConfirmNo  = Keybinding{Key: "n", Desc: "Cancel"}
)

// matchKey checks if the input matches the keybinding.
func matchKey(input string, keys ...Keybinding) bool {
	for _, k := range keys {
		if input == k.Key {
			return true
		}
	}
	return false
}
