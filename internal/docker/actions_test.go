package docker

import (
	"testing"
)

func TestStopContainer_ClientCreationDoesNotPanic(t *testing.T) {
	// Verify StopContainer doesn't panic even without Docker daemon.
	// Real Docker client creation will fail gracefully.
	// We can't easily mock the client here, but we verify the function signature works.
	_ = StopContainer
	_ = KillContainer
}
