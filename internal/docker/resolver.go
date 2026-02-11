package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/kostyay/netmon/internal/model"
)

// Resolver resolves host ports to Docker container info.
type Resolver interface {
	Resolve(ctx context.Context) (map[int]*ContainerPort, error)
}

// ContainerPort maps a host port to its container and internal port.
type ContainerPort struct {
	Container     model.ContainerInfo
	HostPort      int
	ContainerPort int
	Protocol      string
}

// dockerResolver implements Resolver using the Docker Engine API.
type dockerResolver struct {
	newClient func() (dockerAPI, error)
}

// dockerAPI is the subset of Docker client we need (for testing).
type dockerAPI interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	Close() error
}

// NewResolver creates a Resolver that talks to the Docker daemon.
func NewResolver() Resolver {
	return &dockerResolver{
		newClient: func() (dockerAPI, error) {
			return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		},
	}
}

// Resolve queries Docker for running containers and builds a host-port → container map.
// Returns empty map (not error) if Docker is unavailable.
func (r *dockerResolver) Resolve(ctx context.Context) (map[int]*ContainerPort, error) {
	cli, err := r.newClient()
	if err != nil {
		return map[int]*ContainerPort{}, nil // graceful degradation
	}
	defer func() { _ = cli.Close() }()

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		// Context cancellation is a real error
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return map[int]*ContainerPort{}, nil // Docker unavailable
	}

	result := make(map[int]*ContainerPort)
	for _, c := range containers {
		ci := model.ContainerInfo{
			Name:  cleanContainerName(c.Names),
			Image: c.Image,
			ID:    shortID(c.ID),
		}
		for _, p := range c.Ports {
			if p.PublicPort == 0 {
				continue // no host binding
			}
			result[int(p.PublicPort)] = &ContainerPort{
				Container:     ci,
				HostPort:      int(p.PublicPort),
				ContainerPort: int(p.PrivatePort),
				Protocol:      p.Type,
			}
		}
	}
	return result, nil
}

// cleanContainerName strips the leading "/" from Docker container names.
func cleanContainerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

// shortID returns the first 12 chars of a container ID.
func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// IsDockerProcess returns true if the process name is a known Docker daemon process.
func IsDockerProcess(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "com.docker.backend", "dockerd", "docker-proxy", "containerd",
		"docker", "com.docker.vpnkit", "vpnkit-bridge":
		return true
	}
	return false
}

// FormatColumn formats a ContainerPort for display in the Container column.
func FormatColumn(cp *ContainerPort, maxWidth int) string {
	if cp == nil {
		return ""
	}
	s := fmt.Sprintf("%s (%s) %d→%d", cp.Container.Name, cp.Container.Image, cp.HostPort, cp.ContainerPort)
	runes := []rune(s)
	if maxWidth > 0 && len(runes) > maxWidth {
		return string(runes[:maxWidth-1]) + "…"
	}
	return s
}
