package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kostyay/netmon/internal/config"
	"github.com/kostyay/netmon/internal/docker"
	"github.com/kostyay/netmon/internal/model"
)

// testModelWithDockerContainers creates a Model pre-loaded with snapshot and virtual containers.
func testModelWithDockerContainers() Model {
	snapshot := &model.NetworkSnapshot{
		Applications: []model.Application{
			{
				Name: "com.docker.backend",
				PIDs: []int32{100},
				Connections: []model.Connection{
					{PID: 100, Protocol: model.ProtocolTCP, LocalAddr: "0.0.0.0:8080", RemoteAddr: "*:*", State: model.StateListen},
					{PID: 100, Protocol: model.ProtocolTCP, LocalAddr: "0.0.0.0:3000", RemoteAddr: "1.2.3.4:5678", State: model.StateEstablished},
				},
				EstablishedCount: 1,
				ListenCount:      1,
			},
			{
				Name: "Chrome",
				PIDs: []int32{200},
				Connections: []model.Connection{
					{PID: 200, Protocol: model.ProtocolTCP, LocalAddr: "127.0.0.1:52341", RemoteAddr: "142.250.80.46:443", State: model.StateEstablished},
				},
				EstablishedCount: 1,
			},
		},
	}

	vcs := []model.VirtualContainer{
		{
			Info: model.ContainerInfo{Name: "nginx-proxy", Image: "nginx:latest", ID: "abc123def456"},
			PortMappings: []model.PortMapping{
				{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
			},
		},
		{
			Info: model.ContainerInfo{Name: "redis-cache", Image: "redis:7", ID: "def789abc012"},
			PortMappings: []model.PortMapping{
				{HostPort: 3000, ContainerPort: 6379, Protocol: "tcp"},
			},
		},
	}

	m := Model{
		collector:        newMockCollector(snapshot),
		netIOCollector:   newMockNetIOCollector(nil),
		refreshInterval:  DefaultRefreshInterval,
		netIOCache:       make(map[int32]*model.NetIOStats),
		changes:          make(map[ConnectionKey]Change),
		dnsCache:         make(map[string]string),
		dockerResolver:   newMockDockerResolver(nil),
		dockerCache:      make(map[int]*docker.ContainerPort),
		dockerContainers: true,
		virtualContainers: vcs,
		snapshot:         snapshot,
		width:            120,
		height:           40,
		stack: []ViewState{{
			Level:          LevelProcessList,
			SortColumn:     SortProcess,
			SortAscending:  true,
			SelectedColumn: SortProcess,
		}},
	}
	return m
}

func TestDockerContainers_SettingDefaultTrue(t *testing.T) {
	s := config.DefaultSettings()
	if !s.DockerContainers {
		t.Error("DockerContainers should default to true")
	}
}

func TestDockerContainers_NewModelInitFromSettings(t *testing.T) {
	orig := config.CurrentSettings.DockerContainers
	defer func() { config.CurrentSettings.DockerContainers = orig }()

	config.CurrentSettings.DockerContainers = false
	m := NewModel()
	if m.dockerContainers {
		t.Error("dockerContainers should be false when setting is false")
	}

	config.CurrentSettings.DockerContainers = true
	m = NewModel()
	if !m.dockerContainers {
		t.Error("dockerContainers should be true when setting is true")
	}
}

func TestContainerDisplayName(t *testing.T) {
	vc := model.VirtualContainer{
		Info: model.ContainerInfo{Name: "nginx", Image: "nginx:latest"},
	}
	got := containerDisplayName(vc)
	want := "🐳 nginx (nginx:latest)"
	if got != want {
		t.Errorf("containerDisplayName = %q, want %q", got, want)
	}
}

func TestIsVirtualContainerName(t *testing.T) {
	if !isVirtualContainerName("🐳 nginx (nginx:latest)") {
		t.Error("should detect virtual container name")
	}
	if isVirtualContainerName("Chrome") {
		t.Error("should not detect regular process name")
	}
	if isVirtualContainerName("") {
		t.Error("should not detect empty string")
	}
}

func TestFilteredVirtualContainers_WhenEnabled(t *testing.T) {
	m := testModelWithDockerContainers()

	vcs := m.filteredVirtualContainers()
	if len(vcs) != 2 {
		t.Fatalf("expected 2 virtual containers, got %d", len(vcs))
	}
}

func TestFilteredVirtualContainers_WhenDisabled(t *testing.T) {
	m := testModelWithDockerContainers()
	m.dockerContainers = false

	vcs := m.filteredVirtualContainers()
	if len(vcs) != 0 {
		t.Errorf("expected 0 virtual containers when disabled, got %d", len(vcs))
	}
}

func TestFilteredVirtualContainers_WithFilter(t *testing.T) {
	m := testModelWithDockerContainers()
	m.activeFilter = "nginx"

	vcs := m.filteredVirtualContainers()
	if len(vcs) != 1 {
		t.Fatalf("expected 1 filtered virtual container, got %d", len(vcs))
	}
	if vcs[0].Info.Name != "nginx-proxy" {
		t.Errorf("filtered container = %q, want 'nginx-proxy'", vcs[0].Info.Name)
	}
}

func TestFilteredVirtualContainers_FilterByImage(t *testing.T) {
	m := testModelWithDockerContainers()
	m.activeFilter = "redis:7"

	vcs := m.filteredVirtualContainers()
	if len(vcs) != 1 {
		t.Fatalf("expected 1 filtered virtual container, got %d", len(vcs))
	}
	if vcs[0].Info.Name != "redis-cache" {
		t.Errorf("filtered container = %q, want 'redis-cache'", vcs[0].Info.Name)
	}
}

func TestFilteredVirtualContainers_FilterByID(t *testing.T) {
	m := testModelWithDockerContainers()
	m.activeFilter = "abc123"

	vcs := m.filteredVirtualContainers()
	if len(vcs) != 1 {
		t.Fatalf("expected 1 filtered virtual container, got %d", len(vcs))
	}
	if vcs[0].Info.Name != "nginx-proxy" {
		t.Errorf("filtered container = %q, want 'nginx-proxy'", vcs[0].Info.Name)
	}
}

func TestFilteredCount_IncludesVirtualContainers(t *testing.T) {
	m := testModelWithDockerContainers()

	count := m.filteredCount()
	// 2 real apps + 2 virtual containers = 4
	if count != 4 {
		t.Errorf("filteredCount = %d, want 4", count)
	}
}

func TestFilteredCount_VirtualContainersExcludedWhenDisabled(t *testing.T) {
	m := testModelWithDockerContainers()
	m.dockerContainers = false

	count := m.filteredCount()
	// 2 real apps, no virtual containers
	if count != 2 {
		t.Errorf("filteredCount = %d, want 2", count)
	}
}

func TestVirtualContainerApp_BuildsSyntheticApp(t *testing.T) {
	m := testModelWithDockerContainers()

	vc := m.virtualContainers[0] // nginx-proxy, port 8080
	name := containerDisplayName(vc)
	app := m.virtualContainerApp(name)

	if app == nil {
		t.Fatal("virtualContainerApp returned nil")
	}
	if app.Name != name {
		t.Errorf("Name = %q, want %q", app.Name, name)
	}
	if app.Exe != "nginx:latest" {
		t.Errorf("Exe = %q, want 'nginx:latest'", app.Exe)
	}
	// Should have the connection on port 8080 from com.docker.backend
	if len(app.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(app.Connections))
	}
	if app.Connections[0].LocalAddr != "0.0.0.0:8080" {
		t.Errorf("LocalAddr = %q, want '0.0.0.0:8080'", app.Connections[0].LocalAddr)
	}
}

func TestVirtualContainerApp_NoMatchReturnsNil(t *testing.T) {
	m := testModelWithDockerContainers()

	app := m.virtualContainerApp("🐳 nonexistent (fake:latest)")
	if app != nil {
		t.Error("expected nil for nonexistent virtual container")
	}
}

func TestFindSelectedApp_RegularProcess(t *testing.T) {
	m := testModelWithDockerContainers()

	app := m.findSelectedApp("Chrome")
	if app == nil {
		t.Fatal("findSelectedApp returned nil for Chrome")
	}
	if app.Name != "Chrome" {
		t.Errorf("Name = %q, want 'Chrome'", app.Name)
	}
}

func TestFindSelectedApp_VirtualContainer(t *testing.T) {
	m := testModelWithDockerContainers()

	vc := m.virtualContainers[0]
	name := containerDisplayName(vc)
	app := m.findSelectedApp(name)

	if app == nil {
		t.Fatal("findSelectedApp returned nil for virtual container")
	}
	if app.Exe != "nginx:latest" {
		t.Errorf("Exe = %q, want 'nginx:latest'", app.Exe)
	}
}

func TestFindSelectedApp_NotFound(t *testing.T) {
	m := testModelWithDockerContainers()

	app := m.findSelectedApp("NonExistent")
	if app != nil {
		t.Error("expected nil for non-existent process")
	}
}

func TestDrillDown_VirtualContainer(t *testing.T) {
	m := testModelWithDockerContainers()

	// Move cursor to first virtual container (index 2, after 2 real apps)
	view := m.CurrentView()
	view.Cursor = 2

	// Press Enter to drill down
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should have pushed a new view
	if len(newModel.stack) != 2 {
		t.Fatalf("stack length = %d, want 2", len(newModel.stack))
	}
	currentView := newModel.CurrentView()
	if currentView.Level != LevelConnections {
		t.Errorf("Level = %v, want LevelConnections", currentView.Level)
	}
	if !isVirtualContainerName(currentView.ProcessName) {
		t.Errorf("ProcessName = %q, expected virtual container name", currentView.ProcessName)
	}
	if !newModel.dockerView {
		t.Error("dockerView should be true after drilling into virtual container")
	}
}

func TestSettingsToggle_DockerContainers(t *testing.T) {
	m := testModelWithDockerContainers()
	m.settingsMode = true
	m.settingsCursor = 4 // Docker Containers is index 4

	// Press Enter/Space to toggle
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	// Should have toggled off
	if newModel.dockerContainers {
		t.Error("dockerContainers should be false after toggle")
	}
	if newModel.virtualContainers != nil {
		t.Error("virtualContainers should be nil after disabling")
	}
}

func TestDockerResolvedMsg_StoresVirtualContainers(t *testing.T) {
	m := testModelWithDockerContainers()
	m.virtualContainers = nil // start empty

	vcs := []model.VirtualContainer{
		{Info: model.ContainerInfo{Name: "test", Image: "test:1", ID: "aaa111bbb222"}},
	}
	msg := DockerResolvedMsg{
		Containers:        map[int]*docker.ContainerPort{},
		VirtualContainers: vcs,
	}

	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if len(newModel.virtualContainers) != 1 {
		t.Fatalf("expected 1 virtual container, got %d", len(newModel.virtualContainers))
	}
	if newModel.virtualContainers[0].Info.Name != "test" {
		t.Errorf("Name = %q, want 'test'", newModel.virtualContainers[0].Info.Name)
	}
}

func TestKillMode_VirtualContainer(t *testing.T) {
	m := testModelWithDockerContainers()

	// Move cursor to first virtual container
	view := m.CurrentView()
	view.Cursor = 2

	// Press x to enter kill mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updated, _ := m.Update(msg)
	newModel := updated.(Model)

	if !newModel.killMode {
		t.Error("should be in kill mode")
	}
	if newModel.killTarget == nil {
		t.Fatal("killTarget should not be nil")
	}
	if newModel.killTarget.ContainerID == "" {
		t.Error("ContainerID should be set for virtual container")
	}
	if newModel.killTarget.ContainerID != "abc123def456" {
		t.Errorf("ContainerID = %q, want 'abc123def456'", newModel.killTarget.ContainerID)
	}
}

func TestKillModalContent_Container(t *testing.T) {
	m := testModelWithDockerContainers()
	m.killMode = true
	m.killTarget = &killTargetInfo{
		ProcessName: "🐳 nginx (nginx:latest)",
		Exe:         "nginx:latest",
		Signal:      "SIGTERM",
		ContainerID: "abc123def456",
	}

	content := m.renderKillModalContent()
	if content == "" {
		t.Fatal("kill modal content should not be empty")
	}
	// Should contain container-specific text
	if !containsText(content, "Stop this container") {
		t.Error("kill modal should say 'Stop this container'")
	}
	if !containsText(content, "abc123def456") {
		t.Error("kill modal should show container ID")
	}
}

func TestRenderProcessListData_IncludesVirtualContainerRows(t *testing.T) {
	m := testModelWithDockerContainers()
	m.ready = true

	content := m.renderProcessListData()
	if content == "" {
		t.Fatal("renderProcessListData should not be empty")
	}
	// Should contain virtual container names
	if !containsText(content, "nginx-proxy") {
		t.Error("process list should include nginx-proxy virtual container")
	}
	if !containsText(content, "redis-cache") {
		t.Error("process list should include redis-cache virtual container")
	}
	// Should also contain regular processes
	if !containsText(content, "Chrome") {
		t.Error("process list should include Chrome")
	}
}

func TestRenderProcessListData_NoVirtualContainersWhenDisabled(t *testing.T) {
	m := testModelWithDockerContainers()
	m.dockerContainers = false
	m.ready = true

	content := m.renderProcessListData()
	if containsText(content, "nginx-proxy") {
		t.Error("process list should not include virtual containers when disabled")
	}
}

// containsText checks if content contains the substring after stripping ANSI codes.
func containsText(content, substr string) bool {
	return strings.Contains(stripAnsi(content), substr)
}
