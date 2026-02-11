# Docker Container Column

## Summary

When drilling into a Docker process (`com.docker.backend`, `dockerd`, `docker-proxy`, `containerd`), show an extra "Container" column in the connections table. Column displays: `name (image) hostPort→containerPort`.

## Decisions

- **Display**: Extra column in connections table (not a sub-grouping or separate view)
- **Visibility**: Docker processes only; non-Docker views unchanged
- **Data source**: Docker Engine API via `github.com/docker/docker/client`
- **Column content**: `containerName (image:tag) hostPort→containerPort`
- **Scope**: Drill-down only; containers do NOT appear as top-level process entries

## Architecture

### New package: `internal/docker/`

**resolver.go**:
- `ContainerResolver` interface: `Resolve(ctx) → map[int]ContainerInfo, error`
- `ContainerInfo`: `Name`, `Image`, `ID` (short), `Ports []PortMapping`
- `PortMapping`: `HostPort int`, `ContainerPort int`, `Protocol string`
- Impl connects via default Docker socket
- Calls `client.ContainerList(ctx, types.ContainerListOptions{})` for running containers
- Iterates `NetworkSettings.Ports` → builds `hostPort → ContainerInfo` map
- Graceful degradation: Docker unavailable → empty map, no error

### Model changes (`internal/model/network.go`)

```go
type ContainerInfo struct {
    Name  string
    Image string
    ID    string
}

type PortMapping struct {
    HostPort      int
    ContainerPort int
    Protocol      string
}
```

Add to `Connection`:
```go
Container    *ContainerInfo // nil for non-Docker
PortMapping  *PortMapping   // nil if no mapping found
```

### UI changes

**Detection** (`internal/ui/update.go`):
- `isDockerProcess(name string) bool` — matches known Docker process names
- On drill-down into Docker process, set flag `m.dockerView = true`

**Messages** (`internal/ui/messages.go`):
- `DockerResolveMsg` — triggers async Docker API call
- `DockerResolvedMsg` — carries `map[int]ContainerInfo`

**Model** (`internal/ui/model.go`):
- Add `dockerCache map[int]docker.ContainerInfo`
- Add `dockerView bool` (true when viewing Docker process connections)

**View** (`internal/ui/view_table.go`):
- When `dockerView`, add "Container" column after "State"
- Format: `name (image) hostPort→containerPort`
- Max width ~35 chars, truncate with `...`

**Sort** (`internal/ui/view_sort.go`):
- Add `SortContainer` to `SortColumn` enum
- Sortable when `dockerView` is true

### Data flow

```
Tick → Collect connections → snapshot
     → if Docker process in snapshot:
         fire DockerResolveMsg (async cmd)
     → DockerResolvedMsg received:
         store in m.dockerCache
     → View renders:
         for each connection, lookup localPort in dockerCache
         populate Container column
```

Cache refreshed every tick. Single async call, same pattern as NetIO.

## Files to create/modify

| File | Action |
|------|--------|
| `internal/docker/resolver.go` | **Create** — ContainerResolver interface + impl |
| `internal/docker/resolver_test.go` | **Create** — tests with mock Docker client |
| `internal/model/network.go` | **Modify** — add ContainerInfo, PortMapping types |
| `internal/ui/model.go` | **Modify** — add dockerCache, dockerView fields |
| `internal/ui/messages.go` | **Modify** — add Docker messages |
| `internal/ui/update.go` | **Modify** — handle Docker messages, detect Docker process |
| `internal/ui/view_table.go` | **Modify** — render Container column |
| `internal/ui/view_sort.go` | **Modify** — add SortContainer |
| `internal/ui/keys.go` | No change (sort mode keys already generic) |
| `go.mod` | **Modify** — add `github.com/docker/docker` dep |

## Tests

### `internal/docker/resolver_test.go` (new)

**Mock strategy**: Define `ContainerResolver` interface; tests use a mock impl (same pattern as `mockCollector` in `internal/ui/mock_collector_test.go`).

| Test | Description |
|------|-------------|
| `TestResolve_RunningContainers` | Mock returns 2 containers with port bindings → verify port→ContainerInfo map has correct entries |
| `TestResolve_NoContainers` | Mock returns empty list → verify empty map, no error |
| `TestResolve_ContainerWithoutPorts` | Container running but no published ports → excluded from map |
| `TestResolve_MultiplePortsOneContainer` | Container publishes ports 80,443 → both map to same ContainerInfo |
| `TestResolve_OverlappingPorts` | Two containers claim same host port → last-write-wins (document behavior) |
| `TestResolve_DockerUnavailable` | Mock returns connection error → empty map, nil error (graceful degradation) |
| `TestResolve_ContextCancelled` | Cancelled context → returns context error |
| `TestContainerInfo_Format` | `ContainerInfo.Format()` returns `"name (image) hostPort→containerPort"` |
| `TestContainerInfo_FormatTruncation` | Long names truncated to max width with `…` |

### `internal/ui/update_test.go` (additions)

| Test | Description |
|------|-------------|
| `TestIsDockerProcess_KnownNames` | `"com.docker.backend"`, `"dockerd"`, `"docker-proxy"`, `"containerd"` → true |
| `TestIsDockerProcess_NonDocker` | `"Chrome"`, `"nginx"`, `"docker-cli"` → false |
| `TestIsDockerProcess_CaseInsensitive` | `"Docker-Proxy"` → true (if we decide case-insensitive) |
| `TestDrillIntoDocker_SetsDockerView` | Enter on Docker process → `m.dockerView == true` |
| `TestDrillIntoNonDocker_DockerViewFalse` | Enter on regular process → `m.dockerView == false` |
| `TestPopFromDockerView_ClearsDockerView` | Esc from Docker connections → `m.dockerView == false` |
| `TestDockerResolvedMsg_PopulatesCache` | Receive `DockerResolvedMsg` with data → `m.dockerCache` populated |
| `TestDockerResolvedMsg_EmptyResult` | Receive empty `DockerResolvedMsg` → cache empty, no error |
| `TestDockerResolvedMsg_ReplacesOldCache` | Second `DockerResolvedMsg` overwrites previous cache |

### `internal/ui/view_table_test.go` (additions or new)

| Test | Description |
|------|-------------|
| `TestRenderConnections_DockerView_HasContainerColumn` | Docker view renders "Container" header in table |
| `TestRenderConnections_NonDockerView_NoContainerColumn` | Regular view does NOT render "Container" header |
| `TestRenderConnections_DockerView_MatchedPort` | Connection on port 8080 + cache has port 8080 → shows `"nginx (nginx:latest) 8080→80"` |
| `TestRenderConnections_DockerView_UnmatchedPort` | Connection on port 9999 + cache has no entry → Container column empty |
| `TestRenderConnections_DockerView_EmptyCache` | Docker view with empty cache → all Container columns empty |

### `internal/ui/view_sort_test.go` (additions)

| Test | Description |
|------|-------------|
| `TestSortContainer_Ascending` | Sort by Container column ascending → alphabetical by container name |
| `TestSortContainer_Descending` | Sort descending → reverse alphabetical |
| `TestSortContainer_EmptyValues` | Connections without container info sort to bottom (ascending) or top (descending) |
| `TestSortContainer_OnlyInDockerView` | `SortContainer` not in column list for non-Docker views |

### `internal/model/network_test.go` (additions)

| Test | Description |
|------|-------------|
| `TestContainerInfo_Struct` | Verify struct fields populate correctly |
| `TestPortMapping_Struct` | Verify PortMapping fields |
| `TestConnection_WithContainer` | Connection with non-nil Container field |
| `TestConnection_WithoutContainer` | Connection with nil Container (backwards compat) |

### `internal/ui/mock_collector_test.go` (additions)

| Test | Description |
|------|-------------|
| `mockDockerResolver` | New mock implementing `ContainerResolver` interface — returns configurable `map[int]ContainerInfo` |

### Integration-style tests

| Test | Description |
|------|-------------|
| `TestDockerDrillDownFlow` | Full flow: create model with Docker snapshot → drill in → receive DockerResolvedMsg → verify view renders container column with data |
| `TestDockerDrillDownFlow_NoDocker` | Same flow but resolver returns empty → container column shows empty values, no errors |

## Open questions

None — all design decisions resolved.
