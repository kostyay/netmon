# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build binary to bin/netmon
make test           # Run all tests
make lint           # Run golangci-lint
make fmt            # Format code with gofmt
make run            # Run directly via go run
make security       # Run govulncheck, gosec, trufflehog, gitleaks

# Run single test
go test ./internal/ui -run TestName -v

# Run with coverage
go test ./... -coverprofile=coverage.out
```

## Architecture

**TUI Network Monitor** - displays live network connections grouped by process, built with Bubble Tea.

### Core Flow
```
cmd/netmon/root.go    → Entry point, CLI flags (--json, port filter)
                      → TUI mode: ui.NewModel() → tea.NewProgram()
                      → JSON mode: collector.CollectOnce() → output.RenderJSON()
```

### Key Packages

- **internal/ui/** - Bubble Tea model (Model/Update/View pattern)
  - `model.go` - State: navigation stack, snapshot, caches, modes (search/kill/settings/help)
  - `update.go` - Message handlers: key events, tick, data fetch, DNS resolution
  - `view.go` - Render: header, table, footer, modals
  - Navigation: 3 levels (ProcessList → Connections → AllConnections) via stack

- **internal/collector/** - Platform-specific data collection
  - `collector.go` - Interfaces: `Collector`, `NetIOCollector`
  - `darwin.go` - macOS impl using gopsutil (net.Connections, process info)
  - Groups connections by process name, caches process info per cycle

- **internal/model/** - Domain types
  - `NetworkSnapshot` → `[]Application` → `[]Connection`
  - `SelectionID` for stable cursor across data refreshes

- **internal/config/** - Theme/settings
  - `styles.go` - Dracula theme, user skin override (~/.config/netmon/skin.yaml)
  - `settings.go` - Persistent settings (DNS, service names)

### UI Patterns

- **ViewState stack** - Push/pop for drill-down navigation
- **SortMode** - Activated with 's', arrow keys select column, enter confirms
- **Change highlighting** - Diff connections between snapshots, highlight added/removed
- **DNS resolution** - Async lookups queued via tea.Cmd, cached in model
- **Kill mode** - Sends SIGTERM/SIGKILL to selected process

### Keybindings (defined in internal/ui/keys.go)
- j/k or ↑/↓: Navigate
- Enter/Space: Drill down / confirm sort
- Esc/Backspace: Go back
- s: Enter sort mode
- /: Search filter
- v: Toggle grouped/ungrouped view
- x: SIGTERM, X: SIGKILL
- ?: Help, ,: Settings

## Code Style

- Go 1.25+
- Conventional Commits
- Keep files <500 LOC
- Platform code via build tags (darwin.go, etc.)
