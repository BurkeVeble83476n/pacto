package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeAuth(t *testing.T) {
	encoded := encodeAuth("user", "pass")
	// base64("user:pass") = "dXNlcjpwYXNz"
	if encoded != "dXNlcjpwYXNz" {
		t.Errorf("expected dXNlcjpwYXNz, got %s", encoded)
	}
}

func TestWriteDockerConfig_NewFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if err := writeDockerConfig("ghcr.io", "user", "pass"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configPath := filepath.Join(dir, ".docker", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config file: %v", err)
	}

	var cfg dockerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	auth, ok := cfg.Auths["ghcr.io"]
	if !ok {
		t.Fatal("expected ghcr.io auth entry")
	}
	if auth.Auth != encodeAuth("user", "pass") {
		t.Errorf("expected encoded auth, got %s", auth.Auth)
	}
}

func TestWriteDockerConfig_MergeExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	dockerDir := filepath.Join(dir, ".docker")
	if err := os.MkdirAll(dockerDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Write initial config with one registry
	initial := dockerConfig{
		Auths: map[string]dockerAuth{
			"docker.io": {Auth: "existing"},
		},
	}
	data, _ := json.MarshalIndent(initial, "", "  ")
	if err := os.WriteFile(filepath.Join(dockerDir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	// Add a second registry
	if err := writeDockerConfig("ghcr.io", "user", "pass"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read and verify both exist
	result, err := os.ReadFile(filepath.Join(dockerDir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}

	var cfg dockerConfig
	if err := json.Unmarshal(result, &cfg); err != nil {
		t.Fatal(err)
	}

	if _, ok := cfg.Auths["docker.io"]; !ok {
		t.Error("expected docker.io to still exist")
	}
	if _, ok := cfg.Auths["ghcr.io"]; !ok {
		t.Error("expected ghcr.io to be added")
	}
}

func TestWriteDockerConfig_ReadOnlyHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Make home read-only so .docker dir cannot be created
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	err := writeDockerConfig("ghcr.io", "user", "pass")
	if err == nil {
		t.Error("expected error when home directory is read-only")
	}
}

func TestWriteDockerConfig_WriteFileError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Pre-create .docker dir, then make it read-only.
	// MkdirAll succeeds (dir exists), but WriteFile fails (dir not writable).
	dockerDir := filepath.Join(dir, ".docker")
	if err := os.MkdirAll(dockerDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dockerDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dockerDir, 0755) })

	err := writeDockerConfig("ghcr.io", "user", "pass")
	if err == nil {
		t.Error("expected error when WriteFile fails on read-only .docker dir")
	}
}

func TestWriteDockerConfig_InvalidExistingJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	dockerDir := filepath.Join(dir, ".docker")
	if err := os.MkdirAll(dockerDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(dockerDir, "config.json"), []byte("{invalid"), 0600); err != nil {
		t.Fatal(err)
	}

	err := writeDockerConfig("ghcr.io", "user", "pass")
	if err == nil {
		t.Error("expected error for invalid existing JSON")
	}
}

func TestLoginCommand_ReadPasswordError(t *testing.T) {
	old := readPasswordFn
	readPasswordFn = func(int) ([]byte, error) { return nil, fmt.Errorf("read failed") }
	defer func() { readPasswordFn = old }()

	cmd := newLoginCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"ghcr.io", "--username", "user"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when ReadPassword fails")
	}
}

func TestLoginCommand_ReadPasswordSuccess(t *testing.T) {
	old := readPasswordFn
	readPasswordFn = func(int) ([]byte, error) { return []byte("secret"), nil }
	defer func() { readPasswordFn = old }()

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cmd := newLoginCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"ghcr.io", "--username", "user"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteDockerConfig_UserHomeDirError(t *testing.T) {
	old := userHomeDirFn
	userHomeDirFn = func() (string, error) { return "", fmt.Errorf("no home") }
	defer func() { userHomeDirFn = old }()

	err := writeDockerConfig("ghcr.io", "user", "pass")
	if err == nil {
		t.Error("expected error when UserHomeDir fails")
	}
}

func TestReadPasswordFn_Default(t *testing.T) {
	// Exercise the default readPasswordFn (which wraps term.ReadPassword).
	// Using an invalid fd ensures it returns an error without needing a real terminal.
	_, err := readPasswordFn(-1)
	if err == nil {
		t.Error("expected error from readPasswordFn with invalid fd")
	}
}

func TestWriteDockerConfig_MarshalError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	old := jsonMarshalIndentFn
	jsonMarshalIndentFn = func(any, string, string) ([]byte, error) {
		return nil, fmt.Errorf("marshal failed")
	}
	defer func() { jsonMarshalIndentFn = old }()

	err := writeDockerConfig("ghcr.io", "user", "pass")
	if err == nil {
		t.Error("expected error when MarshalIndent fails")
	}
}
