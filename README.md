# netmon

A fast, interactive network monitor for macOS. See live connections grouped by process in your terminal, or pipe JSON to scripts and LLMs.

## Installation

```bash
go install github.com/yourusername/netmon/cmd/netmon@latest
```

Or build from source:

```bash
make build
./bin/netmon
```

## Usage

### TUI Mode (Interactive)

```bash
netmon              # Launch interactive monitor
netmon 443          # Filter to port 443
netmon --pid 1234   # Monitor specific process
```

### CLI Mode (JSON Output)

```bash
netmon --json                    # JSON to stdout
netmon --json | jq '.apps[0]'    # Pipe to jq
netmon --json | llm "summarize"  # Feed to LLM
```

JSON output is automatic when stdout is not a TTY:

```bash
netmon > connections.json        # Redirected = JSON
netmon | grep ESTABLISHED        # Piped = JSON
```

## JSON Schema

```json
{
  "apps": [
    {
      "name": "curl",
      "pids": [1234],
      "executable": "/usr/bin/curl",
      "connection_count": 2,
      "tx_bytes": 1024,
      "rx_bytes": 4096,
      "connections": [
        {
          "protocol": "tcp",
          "local_addr": "192.168.1.10",
          "local_port": 54321,
          "remote_addr": "93.184.216.34",
          "remote_port": 443,
          "state": "ESTABLISHED",
          "pid": 1234
        }
      ]
    }
  ],
  "timestamp": "2025-01-20T10:30:00Z"
}
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑` `k` | Move up |
| `↓` `j` | Move down |
| `PageUp` | Page up |
| `PageDown` | Page down |
| `Enter` `Space` | Drill down into process |
| `Esc` `Backspace` | Go back |
| `q` `Ctrl+C` | Quit |

### Views

| Key | Action |
|-----|--------|
| `v` | Toggle grouped/flat view |
| `/` | Search/filter |
| `s` | Sort mode (arrows to select column, Enter to confirm) |
| `?` | Help |
| `S` | Settings |

### Actions

| Key | Action |
|-----|--------|
| `x` | Kill process (SIGTERM) |
| `X` | Force kill (SIGKILL) |
| `+` `=` | Faster refresh (min 500ms) |
| `-` `_` | Slower refresh (max 10s) |

### Sort Mode

| Key | Action |
|-----|--------|
| `←` `h` / `→` `l` | Select column |
| `Enter` | Confirm (toggle direction on same column) |
| `Esc` | Cancel |

## Views

### 1. Process List (Default)

Shows all processes with network activity:

```
┌─ Processes ─────────────────────────────────────────────────┐
│ PID     Process          Conns  Est  Listen    TX       RX │
│ 1234    chrome              15   12       0  2.1MB   45.2MB │
│  892    Slack                8    6       0  128KB    1.2MB │
│  445    postgres             3    0       3     0B      0B │
└─────────────────────────────────────────────────────────────┘
```

### 2. Process Connections

Press `Enter` on a process to see its connections:

```
┌─ chrome ────────────────────────────────────────────────────┐
│ Proto  Local              Remote                     State  │
│ tcp    192.168.1.10:54321 142.250.80.46:443    ESTABLISHED │
│ tcp    192.168.1.10:54322 151.101.1.140:443    ESTABLISHED │
│ udp    192.168.1.10:5353  224.0.0.251:5353              -  │
└─────────────────────────────────────────────────────────────┘
```

### 3. All Connections

Press `v` to see all connections in a flat list:

```
┌─ All Connections ───────────────────────────────────────────┐
│ PID   Process   Proto  Local          Remote          State │
│ 1234  chrome    tcp    :54321   142.250.80.46:443     EST  │
│  892  Slack     tcp    :54400   52.6.142.34:443       EST  │
└─────────────────────────────────────────────────────────────┘
```

## Settings

Press `S` to configure (persisted to `~/.config/netmon/settings.yaml`):

- **DNS Resolution** — Resolve IPs to hostnames
- **Service Names** — Show port names (443 → https)
- **Highlight Changes** — Flash new/removed connections

## Search & Filter

Press `/` to filter. Matches against:
- Process name
- PID
- IP addresses
- Port numbers
- Protocol (tcp/udp)
- State (ESTABLISHED, LISTEN, etc.)

Case-insensitive substring match. Press `Esc` to clear.

## Use Cases

**Debug network issues:**
```bash
netmon --pid $(pgrep myapp)
```

**Find what's using a port:**
```bash
netmon 8080
```

**Export for analysis:**
```bash
netmon --json > snapshot.json
```

**Feed to LLM:**
```bash
netmon --json | llm "what processes are making external connections?"
```

**Monitor in scripts:**
```bash
while true; do
  netmon --json | jq '.apps | length'
  sleep 5
done
```

## Theming

Custom theme at `~/.config/netmon/skin.yaml`. See `skins/dracula.yaml` for format.

## Requirements

- macOS (Darwin)
- Go 1.25+

## License

MIT
