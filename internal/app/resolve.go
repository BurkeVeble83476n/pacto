package app

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/trianalab/pacto/internal/validation"
	"github.com/trianalab/pacto/pkg/contract"
)

const ociPrefix = "oci://"

// DefaultContractPath is the default filename looked up when no path is given.
const DefaultContractPath = "pacto.yaml"

// defaultPath returns the given path if non-empty, otherwise DefaultContractPath.
func defaultPath(path string) string {
	if path == "" {
		return DefaultContractPath
	}
	return path
}

// resolveBundle loads a contract bundle from either a local path or an OCI
// reference (prefixed with "oci://"). For local paths it reads the file from
// disk and uses the parent directory as the bundle FS. For OCI references it
// delegates to the configured BundleStore.
func (s *Service) resolveBundle(ctx context.Context, ref string) (*contract.Bundle, error) {
	if ociRef, ok := strings.CutPrefix(ref, ociPrefix); ok {
		if s.BundleStore == nil {
			return nil, fmt.Errorf("OCI registry not configured")
		}
		return s.BundleStore.Pull(ctx, ociRef)
	}

	rawYAML, err := os.ReadFile(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", ref, err)
	}

	c, err := contract.Parse(bytes.NewReader(rawYAML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", ref, err)
	}

	return &contract.Bundle{
		Contract: c,
		RawYAML:  rawYAML,
		FS:       os.DirFS(filepath.Dir(ref)),
	}, nil
}

// loadAndValidateLocal reads a local contract file, parses it, validates it,
// and returns the parsed contract and bundle FS. This is the shared helper for
// pack and push commands that must validate before proceeding.
func loadAndValidateLocal(path string) (*contract.Contract, []byte, fs.FS, error) {
	rawYAML, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	c, err := contract.Parse(bytes.NewReader(rawYAML))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid contract: %w", err)
	}

	bundleFS := os.DirFS(filepath.Dir(path))

	result := validation.Validate(c, rawYAML, bundleFS)
	if !result.IsValid() {
		return nil, nil, nil, fmt.Errorf("contract validation failed with %d error(s)", len(result.Errors))
	}

	return c, rawYAML, bundleFS, nil
}

// isOCIRef reports whether ref uses the oci:// scheme.
func isOCIRef(ref string) bool {
	return strings.HasPrefix(ref, ociPrefix)
}

// extractBundleFS writes all files from a bundle FS to the given directory.
func extractBundleFS(fsys fs.FS, dir string) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		target := filepath.Join(dir, path)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		return os.WriteFile(target, data, 0644)
	})
}
