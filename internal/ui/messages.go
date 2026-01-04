package ui

import (
	"time"

	"github.com/kostyay/netmon/internal/model"
)

// TickMsg is sent on each refresh interval.
type TickMsg time.Time

// DataMsg contains updated network data.
type DataMsg struct {
	Snapshot *model.NetworkSnapshot
	Err      error
}

// NetIOMsg contains network I/O statistics from background collection.
type NetIOMsg struct {
	Stats map[int32]*model.NetIOStats // Keyed by PID
	Err   error
}
