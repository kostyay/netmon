package docker

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/kostyay/netmon/internal/model"
)

// mockDockerAPI implements dockerAPI for testing.
type mockDockerAPI struct {
	containers []container.Summary
	err        error
}

func (m *mockDockerAPI) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.containers, nil
}

func (m *mockDockerAPI) Close() error { return nil }

func newTestResolver(mock *mockDockerAPI) *dockerResolver {
	return &dockerResolver{
		newClient: func() (dockerAPI, error) {
			return mock, nil
		},
	}
}

func newFailingResolver(err error) *dockerResolver {
	return &dockerResolver{
		newClient: func() (dockerAPI, error) {
			return nil, err
		},
	}
}

func TestResolve_RunningContainers(t *testing.T) {
	mock := &mockDockerAPI{
		containers: []container.Summary{
			{
				ID:    "abc123def456789012",
				Names: []string{"/nginx-proxy"},
				Image: "nginx:latest",
				Ports: []container.Port{
					{PublicPort: 8080, PrivatePort: 80, Type: "tcp"},
					{PublicPort: 8443, PrivatePort: 443, Type: "tcp"},
				},
			},
			{
				ID:    "def789abc012345678",
				Names: []string{"/redis-cache"},
				Image: "redis:7",
				Ports: []container.Port{
					{PublicPort: 6379, PrivatePort: 6379, Type: "tcp"},
				},
			},
		},
	}

	r := newTestResolver(mock)
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 port mappings, got %d", len(result))
	}

	cp := result[8080]
	if cp == nil {
		t.Fatal("expected mapping for port 8080")
	}
	if cp.Container.Name != "nginx-proxy" {
		t.Errorf("Name = %q, want 'nginx-proxy'", cp.Container.Name)
	}
	if cp.Container.Image != "nginx:latest" {
		t.Errorf("Image = %q, want 'nginx:latest'", cp.Container.Image)
	}
	if cp.ContainerPort != 80 {
		t.Errorf("ContainerPort = %d, want 80", cp.ContainerPort)
	}
	if cp.HostPort != 8080 {
		t.Errorf("HostPort = %d, want 8080", cp.HostPort)
	}

	cp6379 := result[6379]
	if cp6379 == nil {
		t.Fatal("expected mapping for port 6379")
	}
	if cp6379.Container.Name != "redis-cache" {
		t.Errorf("Name = %q, want 'redis-cache'", cp6379.Container.Name)
	}
}

func TestResolve_NoContainers(t *testing.T) {
	mock := &mockDockerAPI{containers: []container.Summary{}}
	r := newTestResolver(mock)
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestResolve_ContainerWithoutPorts(t *testing.T) {
	mock := &mockDockerAPI{
		containers: []container.Summary{
			{
				ID:    "abc123def456",
				Names: []string{"/worker"},
				Image: "myapp:latest",
				Ports: []container.Port{}, // no published ports
			},
		},
	}
	r := newTestResolver(mock)
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map for container without ports, got %d", len(result))
	}
}

func TestResolve_ContainerWithUnpublishedPort(t *testing.T) {
	mock := &mockDockerAPI{
		containers: []container.Summary{
			{
				ID:    "abc123def456",
				Names: []string{"/worker"},
				Image: "myapp:latest",
				Ports: []container.Port{
					{PublicPort: 0, PrivatePort: 3000, Type: "tcp"}, // exposed but not published
				},
			},
		},
	}
	r := newTestResolver(mock)
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map for unpublished port, got %d", len(result))
	}
}

func TestResolve_MultiplePortsOneContainer(t *testing.T) {
	mock := &mockDockerAPI{
		containers: []container.Summary{
			{
				ID:    "abc123def456",
				Names: []string{"/web"},
				Image: "nginx:latest",
				Ports: []container.Port{
					{PublicPort: 80, PrivatePort: 80, Type: "tcp"},
					{PublicPort: 443, PrivatePort: 443, Type: "tcp"},
					{PublicPort: 8080, PrivatePort: 80, Type: "tcp"},
				},
			},
		},
	}
	r := newTestResolver(mock)
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(result))
	}
	for _, port := range []int{80, 443, 8080} {
		if result[port] == nil {
			t.Errorf("missing mapping for port %d", port)
		}
		if result[port] != nil && result[port].Container.Name != "web" {
			t.Errorf("port %d: Name = %q, want 'web'", port, result[port].Container.Name)
		}
	}
}

func TestResolve_OverlappingPorts(t *testing.T) {
	mock := &mockDockerAPI{
		containers: []container.Summary{
			{
				ID:    "first111111111111",
				Names: []string{"/first"},
				Image: "app1:latest",
				Ports: []container.Port{{PublicPort: 8080, PrivatePort: 80, Type: "tcp"}},
			},
			{
				ID:    "second2222222222",
				Names: []string{"/second"},
				Image: "app2:latest",
				Ports: []container.Port{{PublicPort: 8080, PrivatePort: 3000, Type: "tcp"}},
			},
		},
	}
	r := newTestResolver(mock)
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Last-write-wins
	cp := result[8080]
	if cp == nil {
		t.Fatal("expected mapping for port 8080")
	}
	if cp.Container.Name != "second" {
		t.Errorf("Name = %q, want 'second' (last-write-wins)", cp.Container.Name)
	}
}

func TestResolve_DockerUnavailable(t *testing.T) {
	mock := &mockDockerAPI{err: errors.New("connection refused")}
	r := newTestResolver(mock)
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for unavailable Docker, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestResolve_ClientCreationFails(t *testing.T) {
	r := newFailingResolver(errors.New("no docker socket"))
	result, err := r.Resolve(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for client creation failure, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestResolve_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mock := &mockDockerAPI{err: ctx.Err()}
	r := newTestResolver(mock)
	_, err := r.Resolve(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

// Tests for helper functions

func TestCleanContainerName(t *testing.T) {
	tests := []struct {
		names []string
		want  string
	}{
		{[]string{"/nginx-proxy"}, "nginx-proxy"},
		{[]string{"/web"}, "web"},
		{[]string{"no-slash"}, "no-slash"},
		{[]string{}, ""},
		{nil, ""},
	}
	for _, tt := range tests {
		got := cleanContainerName(tt.names)
		if got != tt.want {
			t.Errorf("cleanContainerName(%v) = %q, want %q", tt.names, got, tt.want)
		}
	}
}

func TestShortID(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"abc123def456789012345678", "abc123def456"},
		{"short", "short"},
		{"exactly12ch", "exactly12ch"},
		{"", ""},
	}
	for _, tt := range tests {
		got := shortID(tt.id)
		if got != tt.want {
			t.Errorf("shortID(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestIsDockerProcess(t *testing.T) {
	dockerNames := []string{
		"com.docker.backend", "dockerd", "docker-proxy",
		"containerd", "docker", "com.docker.vpnkit", "vpnkit-bridge",
	}
	for _, name := range dockerNames {
		if !IsDockerProcess(name) {
			t.Errorf("IsDockerProcess(%q) = false, want true", name)
		}
	}

	nonDockerNames := []string{
		"Chrome", "nginx", "docker-cli", "mydockertool",
		"containerd-shim", "Firefox", "",
	}
	for _, name := range nonDockerNames {
		if IsDockerProcess(name) {
			t.Errorf("IsDockerProcess(%q) = true, want false", name)
		}
	}
}

func TestIsDockerProcess_CaseInsensitive(t *testing.T) {
	if !IsDockerProcess("Docker-Proxy") {
		t.Error("IsDockerProcess should be case-insensitive")
	}
	if !IsDockerProcess("DOCKERD") {
		t.Error("IsDockerProcess should be case-insensitive")
	}
}

func TestFormatColumn(t *testing.T) {
	cp := &ContainerPort{
		Container:     model.ContainerInfo{Name: "nginx", Image: "nginx:latest", ID: "abc123"},
		HostPort:      8080,
		ContainerPort: 80,
		Protocol:      "tcp",
	}
	got := FormatColumn(cp, 0)
	want := "nginx (nginx:latest) 8080→80"
	if got != want {
		t.Errorf("FormatColumn = %q, want %q", got, want)
	}
}

func TestFormatColumn_Nil(t *testing.T) {
	got := FormatColumn(nil, 0)
	if got != "" {
		t.Errorf("FormatColumn(nil) = %q, want empty", got)
	}
}

func TestFormatColumn_Truncation(t *testing.T) {
	cp := &ContainerPort{
		Container:     model.ContainerInfo{Name: "very-long-container", Image: "registry.io/org/image:v1.2.3"},
		HostPort:      8080,
		ContainerPort: 80,
	}
	got := FormatColumn(cp, 25)
	runes := []rune(got)
	if len(runes) > 25 {
		t.Errorf("rune len = %d, want <= 25", len(runes))
	}
}
