package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// newDockerClient creates a Docker client configured from environment.
func newDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// StopContainer sends a stop signal to a Docker container.
// Timeout is the seconds to wait before force-killing.
func StopContainer(ctx context.Context, containerID string, timeoutSecs int) error {
	cli, err := newDockerClient()
	if err != nil {
		return err
	}
	defer func() { _ = cli.Close() }()

	opts := container.StopOptions{}
	if timeoutSecs > 0 {
		opts.Timeout = &timeoutSecs
	}
	return cli.ContainerStop(ctx, containerID, opts)
}

// KillContainer sends SIGKILL to a Docker container.
func KillContainer(ctx context.Context, containerID string) error {
	cli, err := newDockerClient()
	if err != nil {
		return err
	}
	defer func() { _ = cli.Close() }()

	return cli.ContainerKill(ctx, containerID, "SIGKILL")
}
