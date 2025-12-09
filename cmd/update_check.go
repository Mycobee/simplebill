package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"simplebill/internal/config"
)

const (
	githubReleaseURL = "https://api.github.com/repos/mycobee/simplebill/releases/latest"
	checkInterval    = 24 * time.Hour
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckForUpdate checks GitHub for a newer version and prints a warning if found.
// Checks at most once per 24 hours. Skip with skip_update_check: true in config.yml.
func CheckForUpdate(currentVersion string) {
	cfg, err := config.Load()
	if err == nil && cfg.SkipUpdateCheck {
		return
	}

	dir, err := config.Dir()
	if err != nil {
		return
	}

	checkFile := filepath.Join(dir, ".last-version-check")

	// Check if we've checked recently
	if info, err := os.Stat(checkFile); err == nil {
		if time.Since(info.ModTime()) < checkInterval {
			// Read cached version
			data, err := os.ReadFile(checkFile)
			if err == nil {
				cached := strings.TrimSpace(string(data))
				if cached != "" && cached != "v"+currentVersion && cached != currentVersion {
					printUpdateWarning(currentVersion, cached)
				}
			}
			return
		}
	}

	// Fetch latest version from GitHub
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(githubReleaseURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	// Cache the result
	os.WriteFile(checkFile, []byte(release.TagName), 0644)

	// Compare versions
	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest != current {
		printUpdateWarning(currentVersion, release.TagName)
	}
}

func printUpdateWarning(current, latest string) {
	fmt.Fprintf(os.Stderr, "simplebill %s available (you have %s) - https://github.com/mycobee/simplebill/releases\n", latest, current)
	fmt.Fprintf(os.Stderr, "(suppress with skip_update_check: true in config.yml)\n\n")
}
