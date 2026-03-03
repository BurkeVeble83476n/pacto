package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	readPasswordFn      = func(fd int) ([]byte, error) { return term.ReadPassword(fd) }
	userHomeDirFn       = os.UserHomeDir
	jsonMarshalIndentFn = json.MarshalIndent
)

func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login <registry>",
		Short: "Log in to an OCI registry",
		Long:  "Stores credentials for an OCI registry in ~/.docker/config.json using Docker's standard format.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := args[0]
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")

			if username == "" {
				return fmt.Errorf("--username is required")
			}

			if password == "" {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), "Password: ")
				pw, err := readPasswordFn(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
				password = string(pw)
			}

			if err := writeDockerConfig(registry, username, password); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Login succeeded for %s\n", registry)
			return nil
		},
	}

	cmd.Flags().StringP("username", "u", "", "registry username")
	cmd.Flags().StringP("password", "p", "", "registry password")

	return cmd
}

// dockerConfig represents the relevant subset of ~/.docker/config.json.
type dockerConfig struct {
	Auths map[string]dockerAuth `json:"auths"`
}

type dockerAuth struct {
	Auth string `json:"auth"`
}

// writeDockerConfig writes credentials to ~/.docker/config.json.
func writeDockerConfig(registry, username, password string) error {
	home, err := userHomeDirFn()
	if err != nil {
		return fmt.Errorf("failed to find home directory: %w", err)
	}

	configDir := filepath.Join(home, ".docker")
	configPath := filepath.Join(configDir, "config.json")

	var cfg dockerConfig

	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse existing %s: %w", configPath, err)
		}
	}

	if cfg.Auths == nil {
		cfg.Auths = make(map[string]dockerAuth)
	}

	// Base64-encode "username:password" per Docker convention.
	encoded := encodeAuth(username, password)
	cfg.Auths[registry] = dockerAuth{Auth: encoded}

	out, err := jsonMarshalIndentFn(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create %s: %w", configDir, err)
	}

	if err := os.WriteFile(configPath, out, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}

	return nil
}

func encodeAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
