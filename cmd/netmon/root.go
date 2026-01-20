package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/kostyay/netmon/internal/collector"
	"github.com/kostyay/netmon/internal/config"
	"github.com/kostyay/netmon/internal/model"
	"github.com/kostyay/netmon/internal/output"
	"github.com/kostyay/netmon/internal/ui"
)

// Version is set via ldflags at build time
var Version = "dev"

var (
	jsonOutput bool
	pidFilter  int
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (for scripting/agent consumption)")
	rootCmd.PersistentFlags().IntVar(&pidFilter, "pid", 0, "Filter to specific process ID and drill into its connections")
}

var rootCmd = &cobra.Command{
	Use:   "netmon [port]",
	Short: "Network monitor - view and manage network connections",
	Long: `netmon is a TUI application for monitoring network connections and managing processes.

Optionally pass a port number to filter connections:
  netmon 8080        # TUI filtered to port 8080
  netmon 8080 --json # JSON output filtered to port 8080`,
	Args: cobra.MaximumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load user settings and theme from config files
		if err := config.InitSettings(); err != nil {
			return fmt.Errorf("failed to load settings: %w", err)
		}
		if err := config.InitTheme(); err != nil {
			return fmt.Errorf("failed to load theme: %w", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var portFilter string
		if len(args) > 0 {
			// Validate it's a number
			if _, err := strconv.Atoi(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid port: %s\n", args[0])
				os.Exit(1)
			}
			portFilter = args[0]
		}

		// Validate --pid and port are mutually exclusive
		if pidFilter != 0 && portFilter != "" {
			fmt.Fprintf(os.Stderr, "Error: cannot specify both --pid and port filter\n")
			os.Exit(1)
		}

		// Validate PID exists if specified
		if pidFilter != 0 {
			if !pidExists(int32(pidFilter)) {
				fmt.Fprintf(os.Stderr, "Error: process %d not found\n", pidFilter)
				os.Exit(1)
			}
		}

		// JSON mode: explicit flag or non-TTY stdout
		if jsonOutput || !term.IsTerminal(int(os.Stdout.Fd())) {
			runJSONMode(portFilter, int32(pidFilter))
			return
		}

		// Default behavior: launch TUI
		m := ui.NewModel().WithVersion(Version)
		if portFilter != "" {
			m = m.WithFilter(portFilter)
		}
		if pidFilter != 0 {
			m = m.WithPID(int32(pidFilter))
		}
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runJSONMode(portFilter string, pidFilter int32) {
	ctx := context.Background()
	snapshot, ioStats, err := collector.CollectOnce(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting data: %v\n", err)
		os.Exit(1)
	}

	// Filter by port if specified
	if portFilter != "" {
		snapshot = filterSnapshotByPort(snapshot, portFilter)
	}

	// Filter by PID if specified
	if pidFilter != 0 {
		snapshot = filterSnapshotByPID(snapshot, pidFilter)
	}

	if err := output.RenderJSON(os.Stdout, snapshot, ioStats); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering JSON: %v\n", err)
		os.Exit(1)
	}
}

func filterSnapshotByPort(snapshot *model.NetworkSnapshot, port string) *model.NetworkSnapshot {
	filtered := &model.NetworkSnapshot{
		Timestamp:    snapshot.Timestamp,
		Applications: make([]model.Application, 0),
		SkippedCount: snapshot.SkippedCount,
	}

	for _, app := range snapshot.Applications {
		var matchingConns []model.Connection
		matchingPIDs := make(map[int32]bool)
		for _, conn := range app.Connections {
			// Check if port appears in local or remote address
			if strings.HasSuffix(conn.LocalAddr, ":"+port) ||
				strings.HasSuffix(conn.RemoteAddr, ":"+port) {
				matchingConns = append(matchingConns, conn)
				matchingPIDs[conn.PID] = true
			}
		}
		if len(matchingConns) > 0 {
			// Build PIDs list from matching connections only
			var pids []int32
			for _, pid := range app.PIDs {
				if matchingPIDs[pid] {
					pids = append(pids, pid)
				}
			}
			filteredApp := model.Application{
				Name:        app.Name,
				Exe:         app.Exe,
				PIDs:        pids,
				Connections: matchingConns,
			}
			// Recount established/listen for filtered connections
			for _, conn := range matchingConns {
				switch conn.State {
				case model.StateEstablished:
					filteredApp.EstablishedCount++
				case model.StateListen:
					filteredApp.ListenCount++
				}
			}
			filtered.Applications = append(filtered.Applications, filteredApp)
		}
	}

	return filtered
}

func filterSnapshotByPID(snapshot *model.NetworkSnapshot, pid int32) *model.NetworkSnapshot {
	filtered := &model.NetworkSnapshot{
		Timestamp:    snapshot.Timestamp,
		Applications: make([]model.Application, 0),
		SkippedCount: snapshot.SkippedCount,
	}

	for _, app := range snapshot.Applications {
		// Check if this app contains the target PID
		hasPID := false
		for _, p := range app.PIDs {
			if p == pid {
				hasPID = true
				break
			}
		}
		if !hasPID {
			continue
		}

		// Filter connections to only those from the target PID
		var matchingConns []model.Connection
		for _, conn := range app.Connections {
			if conn.PID == pid {
				matchingConns = append(matchingConns, conn)
			}
		}

		filteredApp := model.Application{
			Name:        app.Name,
			Exe:         app.Exe,
			PIDs:        []int32{pid},
			Connections: matchingConns,
		}
		for _, conn := range matchingConns {
			switch conn.State {
			case model.StateEstablished:
				filteredApp.EstablishedCount++
			case model.StateListen:
				filteredApp.ListenCount++
			}
		}
		filtered.Applications = append(filtered.Applications, filteredApp)
	}

	return filtered
}

func pidExists(pid int32) bool {
	exists, err := process.PidExists(pid)
	if err != nil {
		return false
	}
	return exists
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
