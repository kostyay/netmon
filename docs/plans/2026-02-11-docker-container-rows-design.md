# Docker Container Virtual Rows

## Problem
Docker connections appear under generic daemon processes (com.docker.backend, docker-proxy). Users must drill into these to see which container owns which port. Too many clicks.

## Solution
Add virtual container rows to the process list. Each running Docker container with published ports gets its own row, showing port bindings at a glance.

## Setting
- `DockerContainers` (bool, default: `true`) in `config/Settings`
- Toggle via Settings modal (`S` key)
- Persisted to `~/.config/netmon/settings.yaml`

## Data Model

New type in `model/network.go`:
```go
type VirtualContainer struct {
    Info         ContainerInfo
    PortMappings []PortMapping
}
```

`NetworkSnapshot` gains `VirtualContainers []VirtualContainer`.

Docker resolver expanded: in addition to `map[int]*ContainerPort`, return `[]VirtualContainer` listing all containers with published ports.

## Process List Display

Virtual rows appended after real processes when setting is on.

| Column | Value |
|--------|-------|
| PID | Short container ID (12 chars) |
| Process | `🐳 name (image)` |
| Conns | Count of connections matching host ports |
| ESTAB | Counted from matched connections |
| LISTEN | Counted from matched connections |
| TX | `—` (unavailable) |
| RX | `—` (unavailable) |

## Drill-Down

Enter on virtual row pushes `LevelConnections`:
- Collects connections from all Docker daemon apps where `ExtractPort(localAddr)` matches container's host port bindings
- Shows Docker columns (Proto, Local, Remote, State, Container)
- Header: container name, image, ID, port mappings

## Kill / Stop

`x` on virtual row: `docker stop` (10s timeout) with confirmation.
`X` on virtual row: `docker kill` with confirmation.

New functions in `internal/docker/`:
- `StopContainer(ctx, containerID, timeout)`
- `KillContainer(ctx, containerID)`

## Search/Filter

Virtual rows match on: container name, image, container ID.

## Sort

Virtual rows sort alongside real processes normally. Container ID sorts lexicographically in PID column.

## Implementation Order

1. `config/settings.go` — add `DockerContainers` field
2. `model/network.go` — add `VirtualContainer`, expand `NetworkSnapshot`
3. `docker/resolver.go` — return `[]VirtualContainer` alongside port map
4. `docker/actions.go` — `StopContainer`, `KillContainer`
5. `ui/model.go` — store virtual containers, add container setting
6. `ui/view_table.go` — no changes (reuse existing columns)
7. `ui/view.go` — render virtual rows in process list, handle drill-down header
8. `ui/update.go` — handle drill-down into virtual row, kill/stop logic
9. `ui/view.go` (settings modal) — add DockerContainers toggle
10. Tests for each layer
