# Connection Navigation Bug Fix

**Date:** 2026-01-18
**Status:** Resolved

## Problem

Cursor navigation stuck in connections view when there are duplicate connections (e.g., multiple UDP *:5353 rows). User presses Down but cursor snaps back to first row on every data refresh.

## Root Cause Analysis

**ConnectionKey** = (ProcessName, LocalAddr, RemoteAddr) is **not unique** for duplicate connections.

Example: 20+ UDP sockets all have:
- ProcessName: "Google Chrome Helper"
- LocalAddr: "*:5353"
- RemoteAddr: "*"

Three locations called `resolveSelectionIndex()` which finds FIRST matching item:

1. **Keyboard handlers** (update.go:155-180)
   - `idx := m.resolveSelectionIndex()` then `view.Cursor = idx + 1`
   - Bug: Even if cursor at row 10, idx returns 0 (first match), cursor goes to row 1

2. **validateSelection()** (selection.go:119-179)
   - On every DataMsg, tried to resolve SelectedID to cursor position
   - Bug: Always found first duplicate, snapped cursor there

3. **View rendering** (view.go:334, 447, 526)
   - `cursorIdx := m.resolveSelectionIndex()` for row highlighting
   - Bug: Highlighted wrong row even when cursor was correct

## Solution

Use array index (cursor) directly instead of ID-based resolution for connections.

### Changes

**update.go** - Keyboard handlers:
```go
// Before
idx := m.resolveSelectionIndex()
view.Cursor = idx + 1

// After
view.Cursor++
```

**selection.go** - validateSelection():
```go
// For connection views, just clamp cursor - no ID resolution
if view.Level == LevelConnections || view.Level == LevelAllConnections {
    if view.Cursor >= itemCount {
        view.Cursor = itemCount - 1
    }
    return
}
```

**view.go** - All three render functions:
```go
// Before
cursorIdx := m.resolveSelectionIndex()

// After
cursorIdx := view.Cursor
```

## Design Decision

**ID-based tracking** remains for process list (names are unique).
**Cursor-based tracking** for connection views (duplicates possible).

This is correct because:
- Processes have unique names - can track across re-sorts
- Connections can be content-identical - no way to distinguish, cursor is only stable reference
- OS allows duplicate sockets (SO_REUSEPORT) - this is valid, not a data bug

## Files Modified

- `internal/ui/update.go` - KeyUp/KeyDown handlers, drill-down
- `internal/ui/selection.go` - validateSelection(), added cursorMatchesSelectedID()
- `internal/ui/view.go` - renderProcessList(), renderConnectionsList(), renderAllConnections()

## Lessons Learned

1. **Trace full data flow** - Bug appeared in navigation but root cause was in multiple places
2. **ID uniqueness assumption** - Don't assume IDs are unique without verifying data model
3. **Rendering vs state** - View() should reflect state, not compute state
