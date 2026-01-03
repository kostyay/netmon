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
