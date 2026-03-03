package app

import (
	"context"
	"fmt"

	"github.com/trianalab/pacto/pkg/contract"
)

// PushOptions holds options for the push command.
type PushOptions struct {
	Ref  string
	Path string
}

// PushResult holds the result of the push command.
type PushResult struct {
	Ref     string
	Digest  string
	Name    string
	Version string
}

// Push validates a contract bundle, builds an OCI image, and pushes it to a registry.
func (s *Service) Push(ctx context.Context, opts PushOptions) (*PushResult, error) {
	if s.BundleStore == nil {
		return nil, fmt.Errorf("OCI registry client not configured")
	}

	path := defaultPath(opts.Path)

	c, _, bundleFS, err := loadAndValidateLocal(path)
	if err != nil {
		return nil, err
	}

	bundle := &contract.Bundle{Contract: c, FS: bundleFS}

	digest, err := s.BundleStore.Push(ctx, opts.Ref, bundle)
	if err != nil {
		return nil, err
	}

	return &PushResult{
		Ref:     opts.Ref,
		Digest:  digest,
		Name:    c.Service.Name,
		Version: c.Service.Version,
	}, nil
}
