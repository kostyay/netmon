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
cmd/netmon/root.go    → Entry point, CLI flags (--json, --pid, port filter)
                      → TUI mode: ui.NewModel() → tea.NewProgram()
                      → JSON mode: collector.CollectOnce() → output.RenderJSON()
                      → Auto-JSON if stdout non-TTY
```

### Key Packages

- **internal/ui/** - Bubble Tea model (Model/Update/View pattern)
  - `model.go` - State: navigation stack, snapshot, caches, modes (search/kill/settings/help)
  - `update.go` - Message handlers: key events, tick, data fetch, DNS resolution
  - `view.go` - Render: header, table, footer, modals
  - `keys.go` - Keybinding definitions
  - Navigation: 3 levels (ProcessList → Connections → AllConnections) via stack

- **internal/collector/** - Platform-specific data collection
  - `collector.go` - Interfaces: `Collector`, `NetIOCollector`
  - `darwin.go` - macOS impl using gopsutil (net.Connections, process info)
  - Groups connections by process name, caches process info per cycle
  - TX/RX bytes stats per process

- **internal/model/** - Domain types
  - `NetworkSnapshot` → `[]Application` → `[]Connection`
  - `SelectionID` for stable cursor across data refreshes

- **internal/config/** - Theme/settings
  - `styles.go` - Dracula theme, user skin override (~/.config/netmon/skin.yaml)
  - `settings.go` - Persistent settings (~/.config/netmon/settings.yaml)

## Features

### CLI Modes
- `--json` - Machine-readable JSON output for scripting
- `--pid <PID>` - Filter to specific process, auto-drill into connections
- `[port]` - Filter connections by port number (positional arg)
- Auto-detect: JSON if non-TTY, otherwise TUI

### Views (3-Level Navigation Stack)
1. **Process List** - All processes with network activity
   - Columns: PID, Process, Conns, Established, Listen, TX, RX
2. **Connections** - Connections for selected process
   - Header: process name, executable path, PIDs, aggregated stats
   - Columns: Protocol, Local, Remote, State
3. **All Connections** - Flat list of all connections (toggle with `v`)
   - Columns: PID, Process, Protocol, Local, Remote, State

### Keybindings (internal/ui/keys.go)
| Key | Action |
|-----|--------|
| `↑/k`, `↓/j` | Navigate up/down |
| `PageUp/PageDown` | Page navigation |
| `Enter/Space` | Drill down / confirm |
| `Esc/Backspace` | Back / cancel |
| `s` | Enter sort mode |
| `←/h`, `→/l` | Select column (sort mode) |
| `/` | Search filter |
| `v` | Toggle grouped/flat view |
| `x` | Kill (SIGTERM) with confirm |
| `X` | Force kill (SIGKILL) with confirm |
| `+/=` | Increase refresh rate (min 500ms) |
| `-/_` | Decrease refresh rate (max 10s) |
| `S` | Settings modal |
| `?` | Help modal |
| `q`, `Ctrl+c` | Quit |

### Search Filter (`/`)
- Substring match (case-insensitive) on: process, PID, addresses, protocol, state
- CLI filters use exact port match; interactive uses substring
- Filter clears when drilling down

### Sort Mode (`s`)
- Arrow keys select column, Enter confirms, Esc cancels
- Toggle same column to reverse direction
- Per-view column sets

### Process Kill (`x`/`X`)
- Confirmation required (y/n)
- SIGTERM (graceful) or SIGKILL (force)
- Works on process list (all PIDs) or single connection
- Result displayed for 2s

### Settings Modal (`S`)
Persisted to `~/.config/netmon/settings.yaml`:
- **DNS Resolution** - Reverse DNS lookup for IPs (async, cached, 2s timeout)
- **Service Names** - Port → service name (80→http, 443→https, etc.)
- **Highlight Changes** - Visual diff added/removed connections (3s expiry)

### UI Features
- Frozen column headers while scrolling
- Breadcrumbs: `📍 Processes > ProcessName | Refresh: X.Xs`
- Connection count in frame title
- UTF-8 box drawing (╭ ╮ ╰ ╯)
- Dynamic viewport/column sizing
- Error display inline with header

### Data Collection
- Live connection capture via gopsutil
- Per-process TX/RX bytes (formatted: B, KB, MB, GB)
- Connection states: ESTABLISHED, LISTEN, TIME_WAIT, CLOSE_WAIT, "-" (UDP)
- Change diffing between snapshots
- DNS caching (max 10 concurrent lookups)

### Theme System
- Embedded Dracula theme (skins/dracula.yaml)
- User override: ~/.config/netmon/skin.yaml
- Themeable: table, header, footer, status, selected, connection states

## Code Style

- Go 1.25+
- Conventional Commits
- Keep files <500 LOC
- Platform code via build tags (darwin.go, etc.)
