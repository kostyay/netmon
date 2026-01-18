package output

import (
	"encoding/json"
	"io"
	"time"

	"github.com/kostyay/netmon/internal/model"
)

// JSONConnection represents a connection in JSON output.
type JSONConnection struct {
	PID        int32  `json:"pid"`
	Protocol   string `json:"protocol"`
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	State      string `json:"state"`
}

// JSONApplication represents an application in JSON output.
type JSONApplication struct {
	Name             string           `json:"name"`
	PIDs             []int32          `json:"pids"`
	ConnectionCount  int              `json:"connection_count"`
	EstablishedCount int              `json:"established_count"`
	ListenCount      int              `json:"listen_count"`
	BytesSent        uint64           `json:"bytes_sent"`
	BytesRecv        uint64           `json:"bytes_recv"`
	Connections      []JSONConnection `json:"connections"`
}

// JSONOutput is the root JSON output structure.
type JSONOutput struct {
	Timestamp    time.Time         `json:"timestamp"`
	Applications []JSONApplication `json:"applications"`
	SkippedCount int               `json:"skipped_count"`
}

// RenderJSON writes the network snapshot as JSON to the writer.
func RenderJSON(w io.Writer, snapshot *model.NetworkSnapshot, ioStats map[int32]*model.NetIOStats) error {
	output := JSONOutput{
		Timestamp:    snapshot.Timestamp,
		Applications: make([]JSONApplication, 0, len(snapshot.Applications)),
		SkippedCount: snapshot.SkippedCount,
	}

	for _, app := range snapshot.Applications {
		jApp := JSONApplication{
			Name:             app.Name,
			PIDs:             app.PIDs,
			ConnectionCount:  len(app.Connections),
			EstablishedCount: app.EstablishedCount,
			ListenCount:      app.ListenCount,
			Connections:      make([]JSONConnection, 0, len(app.Connections)),
		}

		// Aggregate I/O stats across all PIDs
		if ioStats != nil {
			for _, pid := range app.PIDs {
				if stats, ok := ioStats[pid]; ok {
					jApp.BytesSent += stats.BytesSent
					jApp.BytesRecv += stats.BytesRecv
				}
			}
		}

		for _, conn := range app.Connections {
			jApp.Connections = append(jApp.Connections, JSONConnection{
				PID:        conn.PID,
				Protocol:   string(conn.Protocol),
				LocalAddr:  conn.LocalAddr,
				RemoteAddr: conn.RemoteAddr,
				State:      string(conn.State),
			})
		}

		output.Applications = append(output.Applications, jApp)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}
