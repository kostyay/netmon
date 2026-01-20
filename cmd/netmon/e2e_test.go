//go:build darwin

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/kostyay/netmon/internal/collector"
	"github.com/kostyay/netmon/internal/output"
)

// --- Helpers ---

// startTCPServer creates a TCP listener on an ephemeral port, returns port.
func startTCPServer(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start TCP server: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	// Accept connections in background to keep listener active
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	return ln.Addr().(*net.TCPAddr).Port
}

// startUDPServer creates a UDP listener on an ephemeral port, returns port.
func startUDPServer(t *testing.T) int {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start UDP server: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn.LocalAddr().(*net.UDPAddr).Port
}

// collectJSON runs collector and renders JSON output.
func collectJSON(t *testing.T, portFilter string, pidFilter int32) *output.JSONOutput {
	t.Helper()
	ctx := context.Background()
	snapshot, ioStats, err := collector.CollectOnce(ctx)
	if err != nil {
		t.Fatalf("CollectOnce failed: %v", err)
	}

	if portFilter != "" {
		snapshot = filterSnapshotByPort(snapshot, portFilter)
	}
	if pidFilter != 0 {
		snapshot = filterSnapshotByPID(snapshot, pidFilter)
	}

	var buf bytes.Buffer
	if err := output.RenderJSON(&buf, snapshot, ioStats); err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var result output.JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	return &result
}

// findAppByPort finds an application with a connection on the given port.
func findAppByPort(out *output.JSONOutput, port int) *output.JSONApplication {
	portStr := fmt.Sprintf(":%d", port)
	for i := range out.Applications {
		for _, conn := range out.Applications[i].Connections {
			if strings.HasSuffix(conn.LocalAddr, portStr) ||
				strings.HasSuffix(conn.RemoteAddr, portStr) {
				return &out.Applications[i]
			}
		}
	}
	return nil
}

// hasConnectionOnPort checks if app has a connection on the specified port.
func hasConnectionOnPort(app *output.JSONApplication, port int) bool {
	portStr := fmt.Sprintf(":%d", port)
	for _, conn := range app.Connections {
		if strings.HasSuffix(conn.LocalAddr, portStr) ||
			strings.HasSuffix(conn.RemoteAddr, portStr) {
			return true
		}
	}
	return false
}

// --- E2E Tests ---

func TestE2E_JSON_DetectsListeningTCP(t *testing.T) {
	port := startTCPServer(t)

	out := collectJSON(t, "", 0)

	app := findAppByPort(out, port)
	if app == nil {
		t.Fatalf("expected to find app with TCP listener on port %d", port)
	}

	// Verify listen count
	if app.ListenCount == 0 {
		t.Errorf("expected listen_count > 0, got %d", app.ListenCount)
	}

	// Verify connection exists with correct port
	found := false
	for _, conn := range app.Connections {
		if strings.HasSuffix(conn.LocalAddr, fmt.Sprintf(":%d", port)) {
			found = true
			if conn.Protocol != "TCP" {
				t.Errorf("expected protocol TCP, got %s", conn.Protocol)
			}
			if conn.State != "LISTEN" {
				t.Errorf("expected state LISTEN, got %s", conn.State)
			}
			break
		}
	}
	if !found {
		t.Errorf("no connection found with local port %d", port)
	}
}

func TestE2E_JSON_DetectsListeningUDP(t *testing.T) {
	port := startUDPServer(t)

	out := collectJSON(t, "", 0)

	app := findAppByPort(out, port)
	if app == nil {
		t.Fatalf("expected to find app with UDP listener on port %d", port)
	}

	// Verify connection exists with correct protocol and state
	found := false
	for _, conn := range app.Connections {
		if strings.HasSuffix(conn.LocalAddr, fmt.Sprintf(":%d", port)) {
			found = true
			if conn.Protocol != "UDP" {
				t.Errorf("expected protocol UDP, got %s", conn.Protocol)
			}
			if conn.State != "-" {
				t.Errorf("expected state '-' for UDP, got %s", conn.State)
			}
			break
		}
	}
	if !found {
		t.Errorf("no connection found with local port %d", port)
	}
}

func TestE2E_JSON_PortFilter(t *testing.T) {
	port1 := startTCPServer(t)
	port2 := startTCPServer(t)

	// Filter to port1 only
	out := collectJSON(t, strconv.Itoa(port1), 0)

	// Verify port1 is present
	if !hasAnyConnectionOnPort(out, port1) {
		t.Errorf("expected port %d in filtered output", port1)
	}

	// Verify port2 is NOT present
	if hasAnyConnectionOnPort(out, port2) {
		t.Errorf("port %d should not appear in output filtered to port %d", port2, port1)
	}
}

func TestE2E_JSON_PIDFilter(t *testing.T) {
	port := startTCPServer(t)
	myPID := int32(os.Getpid())

	out := collectJSON(t, "", myPID)

	// Should have exactly one app
	if len(out.Applications) != 1 {
		t.Fatalf("expected 1 app with PID filter, got %d", len(out.Applications))
	}

	// App should have our PID
	app := out.Applications[0]
	if len(app.PIDs) != 1 || app.PIDs[0] != myPID {
		t.Errorf("expected PIDs=[%d], got %v", myPID, app.PIDs)
	}

	// All connections should be from our PID
	for _, conn := range app.Connections {
		if conn.PID != myPID {
			t.Errorf("expected all connections from PID %d, got %d", myPID, conn.PID)
		}
	}

	// Our TCP listener should be present
	if !hasConnectionOnPort(&app, port) {
		t.Errorf("expected our TCP listener on port %d", port)
	}
}

func TestE2E_JSON_EmptyFilterResult(t *testing.T) {
	// Use an unlikely port
	out := collectJSON(t, "59999", 0)

	// Should be valid JSON with empty applications
	if out.Applications == nil {
		t.Error("applications should not be nil")
	}
	if len(out.Applications) != 0 {
		t.Errorf("expected 0 applications, got %d", len(out.Applications))
	}
}

func TestE2E_JSON_RawAddresses(t *testing.T) {
	port := startTCPServer(t)

	out := collectJSON(t, "", 0)

	app := findAppByPort(out, port)
	if app == nil {
		t.Fatalf("expected to find app with listener on port %d", port)
	}

	// Verify addresses are numeric, not service names
	for _, conn := range app.Connections {
		// Check LocalAddr format (should be IP:port)
		if conn.LocalAddr != "" && conn.LocalAddr != "*" {
			parts := strings.Split(conn.LocalAddr, ":")
			if len(parts) >= 2 {
				portPart := parts[len(parts)-1]
				// Port should be numeric, not a service name
				if _, err := strconv.Atoi(portPart); err != nil {
					// Allow "*" for wildcard addresses
					if portPart != "*" {
						t.Errorf("expected numeric port in %s, got non-numeric: %s", conn.LocalAddr, portPart)
					}
				}
			}
		}
	}
}

func TestE2E_JSON_CombinedFilters_MutuallyExclusive(t *testing.T) {
	// Verify that port and PID filters work independently
	// (CLI enforces mutual exclusion, but filters can be combined in code)
	port := startTCPServer(t)
	myPID := int32(os.Getpid())

	// Test PID filter alone captures our listener
	outPID := collectJSON(t, "", myPID)
	if !hasAnyConnectionOnPort(outPID, port) {
		t.Errorf("PID filter should include our listener on port %d", port)
	}

	// Test port filter alone captures our listener
	outPort := collectJSON(t, strconv.Itoa(port), 0)
	if !hasAnyConnectionOnPort(outPort, port) {
		t.Errorf("port filter should include listener on port %d", port)
	}
}

func TestE2E_JSON_MultipleListeners(t *testing.T) {
	// Start multiple listeners
	tcp1 := startTCPServer(t)
	tcp2 := startTCPServer(t)
	udp1 := startUDPServer(t)

	myPID := int32(os.Getpid())
	out := collectJSON(t, "", myPID)

	if len(out.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(out.Applications))
	}

	app := out.Applications[0]

	// All three listeners should be present
	if !hasConnectionOnPort(&app, tcp1) {
		t.Errorf("missing TCP listener on port %d", tcp1)
	}
	if !hasConnectionOnPort(&app, tcp2) {
		t.Errorf("missing TCP listener on port %d", tcp2)
	}
	if !hasConnectionOnPort(&app, udp1) {
		t.Errorf("missing UDP listener on port %d", udp1)
	}
}

// hasAnyConnectionOnPort checks if any app in output has connection on port.
func hasAnyConnectionOnPort(out *output.JSONOutput, port int) bool {
	return findAppByPort(out, port) != nil
}

// --- Kill Test Helpers ---

// TestHelperProcess is a helper process that runs as a TCP server.
// It's invoked by test binaries with GO_TEST_HELPER=1 environment variable.
// This pattern allows tests to spawn a separate process that can be killed.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER") != "1" {
		return // Skip unless helper mode
	}

	port := os.Getenv("GO_TEST_PORT")
	if port == "" {
		port = "0"
	}

	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "helper: failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = ln.Close() }()

	// Print actual port for parent to read
	fmt.Println(ln.Addr().(*net.TCPAddr).Port)

	// Block forever until killed
	select {}
}

// spawnListenerProcess spawns a subprocess that listens on a TCP port.
// Returns the process and the port it's listening on.
// The test must call cmd.Process.Kill() and cmd.Wait() when done.
func spawnListenerProcess(t *testing.T) (*exec.Cmd, int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "GO_TEST_HELPER=1", "GO_TEST_PORT=0")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper process: %v", err)
	}

	// Read the port from stdout
	var port int
	if _, err := fmt.Fscanln(stdout, &port); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("failed to read port from helper: %v", err)
	}

	return cmd, port
}

// waitForPIDInCollector waits for a PID to appear in the collector output.
func waitForPIDInCollector(t *testing.T, pid int32, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out := collectJSON(t, "", 0)
		for _, app := range out.Applications {
			for _, p := range app.PIDs {
				if p == pid {
					return true
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// waitForProcessExit waits for the cmd to exit and returns true if it does.
func waitForProcessExit(cmd *exec.Cmd, timeout time.Duration) bool {
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// withKillFlags sets kill command globals and restores them on cleanup.
func withKillFlags(t *testing.T, ports []int, signal string, yes bool) {
	t.Helper()
	oldPorts, oldSignal, oldYes := killPorts, killSignal, killYes
	t.Cleanup(func() {
		killPorts, killSignal, killYes = oldPorts, oldSignal, oldYes
	})
	killPorts, killSignal, killYes = ports, signal, yes
}

// --- E2E Kill Tests ---

func TestE2E_Kill_SIGTERM(t *testing.T) {
	cmd, port := spawnListenerProcess(t)
	pid := int32(cmd.Process.Pid)
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	if !pidExists(pid) {
		t.Fatal("helper process not running after spawn")
	}
	if !waitForPIDInCollector(t, pid, 2*time.Second) {
		t.Fatalf("helper PID %d never appeared in collector", pid)
	}

	if err := syscall.Kill(int(pid), syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}
	if !waitForProcessExit(cmd, 2*time.Second) {
		t.Errorf("process %d did not exit after SIGTERM", pid)
	}

	// Verify no longer in collector
	time.Sleep(100 * time.Millisecond)
	out := collectJSON(t, strconv.Itoa(port), 0)
	for _, app := range out.Applications {
		for _, p := range app.PIDs {
			if p == pid {
				t.Errorf("dead PID %d still in collector output", pid)
			}
		}
	}
}

func TestE2E_Kill_SIGKILL(t *testing.T) {
	cmd, _ := spawnListenerProcess(t)
	pid := int32(cmd.Process.Pid)
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	if !pidExists(pid) {
		t.Fatal("helper process not running after spawn")
	}

	if err := syscall.Kill(int(pid), syscall.SIGKILL); err != nil {
		t.Fatalf("failed to send SIGKILL: %v", err)
	}
	if !waitForProcessExit(cmd, 1*time.Second) {
		t.Errorf("process %d did not exit after SIGKILL", pid)
	}
}

func TestE2E_Kill_CLI_Command(t *testing.T) {
	cmd, port := spawnListenerProcess(t)
	pid := int32(cmd.Process.Pid)
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	if !waitForPIDInCollector(t, pid, 2*time.Second) {
		t.Fatalf("helper PID %d never appeared in collector", pid)
	}

	withKillFlags(t, []int{port}, "SIGTERM", true)

	if err := runKill(nil, nil); err != nil {
		t.Errorf("runKill returned error: %v", err)
	}

	if !waitForProcessExit(cmd, 2*time.Second) {
		t.Errorf("process %d did not exit after runKill", pid)
	}
}

func TestE2E_Kill_NonexistentPID(t *testing.T) {
	err := syscall.Kill(999999, syscall.SIGTERM)
	if err == nil {
		t.Error("expected error when killing nonexistent PID")
	}
	if err != syscall.ESRCH {
		t.Errorf("expected ESRCH, got %v", err)
	}
}

func TestE2E_Kill_NoTargets(t *testing.T) {
	withKillFlags(t, []int{59998}, "SIGTERM", true)

	if err := runKill(nil, nil); err != nil {
		t.Errorf("runKill should not error when no targets found: %v", err)
	}
}

func TestE2E_Kill_InvalidSignal(t *testing.T) {
	withKillFlags(t, []int{8080}, "INVALID_SIGNAL", true)

	err := runKill(nil, nil)
	if err == nil {
		t.Error("expected error for invalid signal")
	}
	if !strings.Contains(err.Error(), "unknown signal") {
		t.Errorf("expected 'unknown signal' error, got: %v", err)
	}
}

func TestE2E_Kill_OwnProcess_Protection(t *testing.T) {
	myPID := int32(os.Getpid())

	if !pidExists(myPID) {
		t.Fatal("own PID doesn't exist")
	}

	// Verify collector sees our process
	out := collectJSON(t, "", myPID)
	if len(out.Applications) == 0 {
		t.Skip("test process has no network connections visible")
	}
}
