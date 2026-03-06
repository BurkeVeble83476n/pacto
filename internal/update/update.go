package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// UpdateResult holds the outcome of a self-update.
type UpdateResult struct {
	PreviousVersion string
	NewVersion      string
}

// Testability hooks.
var (
	osExecutable  = os.Executable
	runtimeGOOS   = runtime.GOOS
	runtimeGOARCH = runtime.GOARCH
	osChmod       = os.Chmod
	osRename      = os.Rename
)

// Update downloads and installs the specified version of pacto.
// If targetVersion is empty, it fetches the latest release.
func Update(currentVersion, targetVersion string) (*UpdateResult, error) {
	if targetVersion == "" {
		latest, err := fetchLatestVersion()
		if err != nil {
			return nil, fmt.Errorf("failed to determine latest version: %w", err)
		}
		targetVersion = latest
	}

	// Normalize version prefix
	if !strings.HasPrefix(targetVersion, "v") {
		targetVersion = "v" + targetVersion
	}

	// Validate the release exists
	if err := validateRelease(targetVersion); err != nil {
		return nil, err
	}

	// Download and replace the binary
	if err := downloadAndReplace(buildDownloadURL(targetVersion)); err != nil {
		return nil, err
	}

	// Update cache so notification isn't shown
	WriteCacheAfterUpdate(targetVersion)

	return &UpdateResult{
		PreviousVersion: currentVersion,
		NewVersion:      targetVersion,
	}, nil
}

// downloadAndReplace downloads the binary and atomically replaces the current executable.
func downloadAndReplace(downloadURL string) error {
	execPath, err := osExecutable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Download to temp file in the same directory (ensures same filesystem for atomic rename)
	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "pacto-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath) // Clean up temp file on any error
	}()

	if err := downloadBinary(downloadURL, tmpFile); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Close()

	if err := osChmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	if err := osRename(tmpPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// validateRelease checks that a release with the given tag exists on GitHub.
func validateRelease(tag string) error {
	url := fmt.Sprintf("%s/repos/TrianaLab/pacto/releases/tags/%s", githubAPIBaseURL, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body) // drain body

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("release %s not found", tag)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	return nil
}

// buildDownloadURL constructs the download URL for the platform binary.
func buildDownloadURL(tag string) string {
	ext := ""
	if runtimeGOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf(
		"%s/TrianaLab/pacto/releases/download/%s/pacto_%s_%s%s",
		githubDownloadURL, tag, runtimeGOOS, runtimeGOARCH, ext,
	)
}

// downloadBinary downloads the binary from the given URL into the writer.
func downloadBinary(downloadURL string, w io.Writer) error {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}
	return nil
}
