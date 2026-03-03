package app

import (
	"os"

	"github.com/trianalab/pacto/internal/oci"
)

// Package-level function variables for filesystem operations,
// overridable in tests.
var (
	writeFileFn = os.WriteFile
	mkdirTempFn = os.MkdirTemp
)

// Service is the application service container. It holds injected dependencies
// and provides methods for each CLI command.
type Service struct {
	BundleStore  oci.BundleStore
	PluginRunner PluginRunner
}

// NewService creates a new application service with the given dependencies.
func NewService(store oci.BundleStore, pluginRunner PluginRunner) *Service {
	return &Service{
		BundleStore:  store,
		PluginRunner: pluginRunner,
	}
}
