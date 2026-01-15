package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/kostyay/netmon/internal/collector"
	"github.com/kostyay/netmon/internal/output"
	"github.com/kostyay/netmon/internal/ui"
)

var jsonOutput bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (for scripting/agent consumption)")
}

var rootCmd = &cobra.Command{
	Use:   "netmon",
	Short: "Network monitor - view and manage network connections",
	Long:  `netmon is a TUI application for monitoring network connections and managing processes.`,
	Run: func(cmd *cobra.Command, args []string) {
		// JSON mode: explicit flag or non-TTY stdout
		if jsonOutput || !term.IsTerminal(int(os.Stdout.Fd())) {
			runJSONMode()
			return
		}

		// Default behavior: launch TUI
		p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runJSONMode() {
	ctx := context.Background()
	snapshot, ioStats, err := collector.CollectOnce(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting data: %v\n", err)
		os.Exit(1)
	}

	if err := output.RenderJSON(os.Stdout, snapshot, ioStats); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering JSON: %v\n", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
