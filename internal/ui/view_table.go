package ui

import (
	"strings"
)

// columnDef defines a table column with sizing properties.
type columnDef struct {
	label      string
	id         SortColumn // column identifier for selection/sorting
	minWidth   int        // minimum width
	flex       int        // flex weight for extra space distribution (0 = fixed)
	rightAlign bool       // true for right-aligned columns (numbers)
}

// renderRow renders a table row with selection styling.
func renderRow(content string, isSelected bool) string {
	row := "  " + content
	if isSelected {
		return SelectedConnStyle().Render(row) + "\n"
	}
	return ConnStyle().Render(row) + "\n"
}

// renderRowWithHighlight renders a table row with selection and change highlight styling.
// changeType: nil=no change, ChangeAdded=green, ChangeRemoved=red
func renderRowWithHighlight(content string, isSelected bool, change *Change) string {
	row := "  " + content

	// Selection takes priority for foreground
	if isSelected {
		return SelectedConnStyle().Render(row) + "\n"
	}

	// Apply highlight based on change type
	if change != nil {
		switch change.Type {
		case ChangeAdded:
			return AddedConnStyle().Render(row) + "\n"
		case ChangeRemoved:
			return RemovedConnStyle().Render(row) + "\n"
		}
	}

	return ConnStyle().Render(row) + "\n"
}

// renderTableHeader renders a table header with optional sort indicators.
// If showSort is false, sort indicators are not displayed (for process list).
func renderTableHeader(columns []columnDef, widths []int, selectedCol, sortCol SortColumn, sortAsc, showSort bool) string {
	var b strings.Builder

	// Add 2-space prefix to align with data rows (which have "  " prefix from renderRow)
	b.WriteString("  ")

	headerStyle := TableHeaderStyle()
	selectedStyle := TableHeaderSelectedStyle()
	sortStyle := SortIndicatorStyle()

	for i, col := range columns {
		if i > 0 {
			b.WriteString(" ")
		}

		isSelected := selectedCol == col.id
		isSorted := showSort && sortCol == col.id

		header := col.label

		var sortIndicator string
		if isSorted {
			if sortAsc {
				sortIndicator = "△"
			} else {
				sortIndicator = "▽"
			}
		}

		padWidth := widths[i] - len(header)
		if isSorted {
			padWidth -= 1
		}
		if padWidth < 0 {
			padWidth = 0
		}
		var paddedHeader string
		if col.rightAlign {
			paddedHeader = strings.Repeat(" ", padWidth) + header
		} else {
			paddedHeader = header + strings.Repeat(" ", padWidth)
		}

		if isSelected {
			b.WriteString(selectedStyle.Render(paddedHeader))
		} else {
			b.WriteString(headerStyle.Render(paddedHeader))
		}

		if isSorted {
			b.WriteString(sortStyle.Render(sortIndicator))
		}
	}

	return b.String()
}

// calculateColumnWidths distributes available width among columns.
// Fixed columns (flex=0) get their minWidth, remaining space goes to flex columns.
func calculateColumnWidths(columns []columnDef, availableWidth int) []int {
	widths := make([]int, len(columns))

	// Account for spaces between columns and selection marker
	separators := len(columns) - 1
	selectionMarker := 2 // "  " prefix for all rows
	availableWidth -= separators + selectionMarker

	// First pass: assign minimum widths and calculate total flex
	totalMinWidth := 0
	totalFlex := 0
	for i, col := range columns {
		widths[i] = col.minWidth
		totalMinWidth += col.minWidth
		totalFlex += col.flex
	}

	// Distribute remaining space to flex columns
	extraSpace := availableWidth - totalMinWidth
	if extraSpace > 0 && totalFlex > 0 {
		for i, col := range columns {
			if col.flex > 0 {
				extra := (extraSpace * col.flex) / totalFlex
				widths[i] += extra
			}
		}
	}

	return widths
}

// processListColumns returns the column definitions for the process list.
func processListColumns() []columnDef {
	return []columnDef{
		{label: "PID", id: SortPID, minWidth: 6, flex: 0, rightAlign: true},
		{label: "Process", id: SortProcess, minWidth: 15, flex: 3, rightAlign: false},
		{label: "Conns", id: SortConns, minWidth: 6, flex: 0, rightAlign: true},
		{label: "ESTAB", id: SortEstablished, minWidth: 6, flex: 0, rightAlign: true},
		{label: "LISTEN", id: SortListen, minWidth: 7, flex: 0, rightAlign: true},
		{label: "TX", id: SortTX, minWidth: 8, flex: 1, rightAlign: true},
		{label: "RX", id: SortRX, minWidth: 8, flex: 1, rightAlign: true},
	}
}

// connectionsColumns returns the column definitions for the connections list.
func connectionsColumns() []columnDef {
	return []columnDef{
		{label: "Proto", id: SortProtocol, minWidth: 6, flex: 0},
		{label: "Local", id: SortLocal, minWidth: 20, flex: 2},
		{label: "Remote", id: SortRemote, minWidth: 20, flex: 2},
		{label: "State", id: SortState, minWidth: 11, flex: 1},
	}
}

// allConnectionsColumns returns the column definitions for the all-connections list.
func allConnectionsColumns() []columnDef {
	return []columnDef{
		{label: "PID", id: SortPID, minWidth: 6, flex: 0, rightAlign: true},
		{label: "Process", id: SortProcess, minWidth: 12, flex: 2},
		{label: "Proto", id: SortProtocol, minWidth: 6, flex: 0},
		{label: "Local", id: SortLocal, minWidth: 18, flex: 2},
		{label: "Remote", id: SortRemote, minWidth: 18, flex: 2},
		{label: "State", id: SortState, minWidth: 11, flex: 1},
	}
}
