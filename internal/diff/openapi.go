package diff

import (
	"io/fs"

	"gopkg.in/yaml.v3"
)

// diffOpenAPI compares two OpenAPI spec files and returns changes for
// added/removed paths. This is the integration boundary — a richer
// implementation can be swapped in (e.g., oasdiff) without changing
// the engine interface.
func diffOpenAPI(contractPath string, oldFS, newFS fs.FS) []Change {
	if oldFS == nil || newFS == nil || contractPath == "" {
		return nil
	}

	oldPaths, oldErr := readOpenAPIPaths(oldFS, contractPath)
	newPaths, newErr := readOpenAPIPaths(newFS, contractPath)

	if oldErr != nil && newErr != nil {
		return nil
	}
	if oldErr != nil {
		// File didn't exist before, now it does — non-breaking addition.
		return nil
	}
	if newErr != nil {
		// File existed before, now it's gone — handled by interface removal.
		return nil
	}

	return diffStringSet(oldPaths, newPaths, "openapi.paths", "API path")
}

// readOpenAPIPaths parses an OpenAPI file and extracts the top-level path keys.
func readOpenAPIPaths(fsys fs.FS, path string) (map[string]bool, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	var spec struct {
		Paths map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	paths := make(map[string]bool, len(spec.Paths))
	for p := range spec.Paths {
		paths[p] = true
	}
	return paths, nil
}
