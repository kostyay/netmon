package ui

import (
	"sort"

	"github.com/kostyay/netmon/internal/model"
)

// sortAllConnections sorts connections with process names based on current view state.
// Uses (process name, local addr, remote addr) as secondary keys for stable ordering.
func (m Model) sortAllConnections(conns []connectionWithProcess) []connectionWithProcess {
	view := m.CurrentView()
	if view == nil {
		return conns
	}

	sorted := make([]connectionWithProcess, len(conns))
	copy(sorted, conns)

	sort.Slice(sorted, func(i, j int) bool {
		var cmp int
		switch view.SortColumn {
		case SortPID:
			cmp = compareInt32(sorted[i].PID, sorted[j].PID)
		case SortProcess:
			cmp = compareString(sorted[i].ProcessName, sorted[j].ProcessName)
		case SortProtocol:
			cmp = compareString(string(sorted[i].Protocol), string(sorted[j].Protocol))
		case SortLocal:
			cmp = compareString(sorted[i].LocalAddr, sorted[j].LocalAddr)
		case SortRemote:
			cmp = compareString(sorted[i].RemoteAddr, sorted[j].RemoteAddr)
		case SortState:
			cmp = compareString(string(sorted[i].State), string(sorted[j].State))
		default:
			cmp = compareInt32(sorted[i].PID, sorted[j].PID)
		}

		// Secondary sort for stable ordering when primary keys are equal
		if cmp == 0 {
			cmp = compareString(sorted[i].ProcessName, sorted[j].ProcessName)
		}
		if cmp == 0 {
			cmp = compareString(sorted[i].LocalAddr, sorted[j].LocalAddr)
		}
		if cmp == 0 {
			cmp = compareString(sorted[i].RemoteAddr, sorted[j].RemoteAddr)
		}

		if view.SortAscending {
			return cmp < 0
		}
		return cmp > 0
	})

	return sorted
}

// sortProcessList sorts applications based on current view state.
// Uses process name as secondary key to ensure stable ordering when primary keys are equal.
func (m Model) sortProcessList(apps []model.Application) []model.Application {
	view := m.CurrentView()
	if view == nil {
		return apps
	}

	sorted := make([]model.Application, len(apps))
	copy(sorted, apps)

	sort.Slice(sorted, func(i, j int) bool {
		var cmp int // -1: i<j, 0: equal, 1: i>j
		switch view.SortColumn {
		case SortPID:
			pidI := int32(0)
			pidJ := int32(0)
			if len(sorted[i].PIDs) > 0 {
				pidI = sorted[i].PIDs[0]
			}
			if len(sorted[j].PIDs) > 0 {
				pidJ = sorted[j].PIDs[0]
			}
			cmp = compareInt32(pidI, pidJ)
		case SortProcess:
			cmp = compareString(sorted[i].Name, sorted[j].Name)
		case SortConns:
			cmp = compareInt(len(sorted[i].Connections), len(sorted[j].Connections))
		case SortEstablished:
			cmp = compareInt(sorted[i].EstablishedCount, sorted[j].EstablishedCount)
		case SortListen:
			cmp = compareInt(sorted[i].ListenCount, sorted[j].ListenCount)
		case SortTX:
			cmp = compareUint64(m.getAggregatedBytes(sorted[i].PIDs, true), m.getAggregatedBytes(sorted[j].PIDs, true))
		case SortRX:
			cmp = compareUint64(m.getAggregatedBytes(sorted[i].PIDs, false), m.getAggregatedBytes(sorted[j].PIDs, false))
		default:
			cmp = compareString(sorted[i].Name, sorted[j].Name)
		}

		// Secondary sort by name when primary keys are equal (ensures stable ordering)
		if cmp == 0 && view.SortColumn != SortProcess {
			cmp = compareString(sorted[i].Name, sorted[j].Name)
		}

		if view.SortAscending {
			return cmp < 0
		}
		return cmp > 0
	})

	return sorted
}

// Compare helpers return -1, 0, or 1
func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareInt32(a, b int32) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareUint64(a, b uint64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareString(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// getAggregatedBytes returns total bytes (TX if isSent, RX otherwise) for all PIDs.
func (m Model) getAggregatedBytes(pids []int32, isSent bool) uint64 {
	var total uint64
	for _, pid := range pids {
		if stats, ok := m.netIOCache[pid]; ok {
			if isSent {
				total += stats.BytesSent
			} else {
				total += stats.BytesRecv
			}
		}
	}
	return total
}

// sortConnectionsForView sorts connections based on current view state.
// Uses (local addr, remote addr) as secondary keys for stable ordering.
func (m Model) sortConnectionsForView(conns []model.Connection) []model.Connection {
	view := m.CurrentView()
	if view == nil {
		return conns
	}

	sorted := make([]model.Connection, len(conns))
	copy(sorted, conns)

	sort.Slice(sorted, func(i, j int) bool {
		var cmp int
		switch view.SortColumn {
		case SortPID:
			cmp = compareInt32(sorted[i].PID, sorted[j].PID)
		case SortProtocol:
			cmp = compareString(string(sorted[i].Protocol), string(sorted[j].Protocol))
		case SortLocal:
			cmp = compareString(sorted[i].LocalAddr, sorted[j].LocalAddr)
		case SortRemote:
			cmp = compareString(sorted[i].RemoteAddr, sorted[j].RemoteAddr)
		case SortState:
			cmp = compareString(string(sorted[i].State), string(sorted[j].State))
		default:
			cmp = compareString(sorted[i].LocalAddr, sorted[j].LocalAddr)
		}

		// Secondary sort for stable ordering when primary keys are equal
		if cmp == 0 {
			cmp = compareString(sorted[i].LocalAddr, sorted[j].LocalAddr)
		}
		if cmp == 0 {
			cmp = compareString(sorted[i].RemoteAddr, sorted[j].RemoteAddr)
		}

		if view.SortAscending {
			return cmp < 0
		}
		return cmp > 0
	})

	return sorted
}
