package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/kostyay/netmon/internal/collector"
	"github.com/kostyay/netmon/internal/model"
	"github.com/kostyay/netmon/internal/output"
	"github.com/kostyay/netmon/internal/ui"
)

var jsonOutput bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (for scripting/agent consumption)")
}

var rootCmd = &cobra.Command{
	Use:   "netmon [port]",
	Short: "Network monitor - view and manage network connections",
	Long: `netmon is a TUI application for monitoring network connections and managing processes.

Optionally pass a port number to filter connections:
  netmon 8080        # TUI filtered to port 8080
  netmon 8080 --json # JSON output filtered to port 8080`,
	Args: cobra.MaximumNArgs(1),
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

		// JSON mode: explicit flag or non-TTY stdout
		if jsonOutput || !term.IsTerminal(int(os.Stdout.Fd())) {
			runJSONMode(portFilter)
			return
		}

		// Default behavior: launch TUI
		m := ui.NewModel()
		if portFilter != "" {
			m = m.WithFilter(portFilter)
		}
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runJSONMode(portFilter string) {
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
		for _, conn := range app.Connections {
			// Check if port appears in local or remote address
			if strings.HasSuffix(conn.LocalAddr, ":"+port) ||
				strings.HasSuffix(conn.RemoteAddr, ":"+port) {
				matchingConns = append(matchingConns, conn)
			}
		}
		if len(matchingConns) > 0 {
			filteredApp := model.Application{
				Name:        app.Name,
				PIDs:        app.PIDs,
				Connections: matchingConns,
			}
			// Recount established/listen for filtered connections
			for _, conn := range matchingConns {
				if conn.State == model.StateEstablished {
					filteredApp.EstablishedCount++
				} else if conn.State == model.StateListen {
					filteredApp.ListenCount++
				}
			}
			filtered.Applications = append(filtered.Applications, filteredApp)
		}
	}

	return filtered
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
