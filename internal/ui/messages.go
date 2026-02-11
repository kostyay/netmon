package ui

import (
	"time"

	"github.com/kostyay/netmon/internal/docker"
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

// DNSResolvedMsg contains a DNS resolution result.
type DNSResolvedMsg struct {
	IP       string
	Hostname string
	Err      error
}

// VersionCheckMsg contains result of GitHub release check.
type VersionCheckMsg struct {
	LatestVersion string // empty if up-to-date
	Err           error  // nil on success (even if up-to-date)
}

// DockerResolvedMsg contains Docker container resolution results.
type DockerResolvedMsg struct {
	Containers map[int]*docker.ContainerPort // host port → container info
	Err        error
}

// AnimationTickMsg is sent for UI animation updates (e.g., live indicator pulse).
type AnimationTickMsg time.Time
