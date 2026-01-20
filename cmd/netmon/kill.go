package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kostyay/netmon/internal/collector"
	"github.com/kostyay/netmon/internal/model"
	"github.com/kostyay/netmon/internal/process"
)

var (
	killPorts  []int
	killSignal string
	killYes    bool
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill processes listening on specified ports",
	Long: `Kill processes that are listening on the specified ports.

Examples:
  netmon kill --port 8080
  netmon kill --port 8080,3000,5432
  netmon kill -p 8080 -p 3000
  netmon kill --port 8080 --signal SIGKILL --yes`,
	RunE: runKill,
}

func init() {
	killCmd.Flags().IntSliceVarP(&killPorts, "port", "p", nil, "Port(s) to kill processes on (required, can specify multiple)")
	killCmd.Flags().StringVarP(&killSignal, "signal", "s", "SIGTERM", "Signal to send (SIGTERM, SIGKILL, SIGHUP, SIGINT, SIGQUIT or numeric)")
	killCmd.Flags().BoolVarP(&killYes, "yes", "y", false, "Skip confirmation prompt")
	_ = killCmd.MarkFlagRequired("port")
	rootCmd.AddCommand(killCmd)
}

type processInfo struct {
	pid  int32
	name string
	port int
}

func runKill(cmd *cobra.Command, args []string) error {
	// Validate signal
	sig, ok := process.SignalMap[strings.ToUpper(killSignal)]
	if !ok {
		return fmt.Errorf("unknown signal: %s", killSignal)
	}

	// Collect network data
	c := collector.New()
	snapshot, err := c.Collect(context.Background())
	if err != nil {
		return fmt.Errorf("failed to collect network data: %w", err)
	}

	// Find processes on requested ports
	portSet := make(map[int]bool, len(killPorts))
	for _, p := range killPorts {
		portSet[p] = true
	}

	// Use map for O(1) duplicate detection
	seen := make(map[string]bool)
	var targets []processInfo

	for _, app := range snapshot.Applications {
		for _, conn := range app.Connections {
			port := model.ExtractPort(conn.LocalAddr)
			if port <= 0 || !portSet[port] {
				continue
			}
			key := fmt.Sprintf("%d:%d", conn.PID, port)
			if seen[key] {
				continue
			}
			seen[key] = true
			targets = append(targets, processInfo{
				pid:  conn.PID,
				name: app.Name,
				port: port,
			})
		}
	}

	if len(targets) == 0 {
		fmt.Println("No processes found on specified port(s)")
		return nil
	}

	// Show what will be killed
	fmt.Println("Processes to kill:")
	for _, t := range targets {
		fmt.Printf("  PID %d (%s) on port %d\n", t.pid, t.name, t.port)
	}
	fmt.Printf("Signal: %s\n", killSignal)

	// Confirm unless --yes
	if !killYes {
		fmt.Print("\nProceed? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted")
			return nil
		}
	}

	// Kill processes
	var killed, failed int
	for _, t := range targets {
		if err := syscall.Kill(int(t.pid), sig); err != nil {
			fmt.Printf("Failed to kill PID %d (%s): %v\n", t.pid, t.name, err)
			failed++
		} else {
			fmt.Printf("Killed PID %d (%s)\n", t.pid, t.name)
			killed++
		}
	}

	fmt.Printf("\nKilled: %d, Failed: %d\n", killed, failed)
	if failed > 0 {
		return fmt.Errorf("%d process(es) could not be killed", failed)
	}
	return nil
}
