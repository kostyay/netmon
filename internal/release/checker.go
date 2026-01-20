package release

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// githubRelease represents the minimal response from GitHub releases API.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckLatest fetches the latest release from GitHub API.
// Returns the latest tag if newer than currentVersion, empty string if current.
func CheckLatest(owner, repo, currentVersion string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", nil
	}

	// Compare versions (strip leading 'v' for comparison)
	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if isNewer(latest, current) {
		return release.TagName, nil
	}
	return "", nil
}

// isNewer returns true if a > b using simple string comparison.
// Works for semver-like versions (e.g., "1.2.3" > "1.2.2").
func isNewer(a, b string) bool {
	// Handle dev/empty versions - always show latest release
	if b == "" || b == "dev" {
		return true
	}
	return a > b
}
