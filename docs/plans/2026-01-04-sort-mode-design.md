# Sort Mode Design

## Problem

Left/right keys silently change the selected column. Users accidentally trigger sorts when pressing Enter because they didn't realize the column selection changed.

## Solution

Add a dedicated sort mode. Press `s` to enter sort mode, then use left/right to select a column, Enter to confirm, or Escape to cancel.

## Behavior

### Entering Sort Mode

- Press `s` to enter sort mode from any view with sortable columns
- Footer shows: `[SORT MODE] ←/→: select column | Enter: sort | Esc: cancel`
- `SelectedColumn` initializes to current `SortColumn`

### In Sort Mode

- Left/right (`h`/`l`): move selected column highlight
- Enter: sort by selected column (toggle direction if same column), exit sort mode
- Escape: cancel without changing sort, exit sort mode

### Outside Sort Mode

- Left/right keys do nothing
- Enter/Space drills into selected process (process list view)
- Current sort column remains visually indicated in header

## Key Bindings

| Key | Normal Mode | Sort Mode |
|-----|-------------|-----------|
| `s` | Enter sort mode | No-op |
| `←`/`h` | Nothing | Select prev column |
| `→`/`l` | Nothing | Select next column |
| Enter | Drill down | Apply sort, exit |
| Esc | Go back | Cancel, exit |

## Visual Design

- Header: sort column shows `▲`/`▼` indicator (unchanged)
- Header: selected column highlighted only in sort mode
- Footer: shows `[SORT MODE]` message when active

## Implementation

### model.go

Add `SortMode bool` to `ViewState` struct.

### update.go

1. Add `s` handler: set `SortMode = true`, set `SelectedColumn = SortColumn`
2. Left/right handlers: only process when `SortMode == true`
3. Enter handler: if `SortMode`, apply sort and exit; otherwise drill down
4. Escape handler: if `SortMode`, exit without popping view

### view.go

Modify footer to show sort mode indicator when `SortMode == true`.

## Testing

- Verify `s` enters sort mode
- Verify left/right work only in sort mode
- Verify Enter applies sort and exits
- Verify Escape cancels and exits
- Verify left/right do nothing outside sort mode
