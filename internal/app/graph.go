package app

import (
	"context"

	"github.com/trianalab/pacto/internal/graph"
	"github.com/trianalab/pacto/pkg/contract"
)

// GraphOptions holds options for the graph command.
type GraphOptions struct {
	Path string
}

// GraphResult holds the result of the graph command.
type GraphResult struct {
	Root      *graph.Node      `json:"root"`
	Cycles    [][]string       `json:"cycles,omitempty"`
	Conflicts []graph.Conflict `json:"conflicts,omitempty"`
}

// Graph resolves the dependency graph for a contract.
func (s *Service) Graph(ctx context.Context, opts GraphOptions) (*GraphResult, error) {
	ref := defaultPath(opts.Path)

	bundle, err := s.resolveBundle(ctx, ref)
	if err != nil {
		return nil, err
	}

	// Use BundleStore as a ContractFetcher if available.
	var fetcher graph.ContractFetcher
	if s.BundleStore != nil {
		fetcher = &bundleStoreFetcher{store: s.BundleStore}
	}

	result := graph.Resolve(ctx, bundle.Contract, fetcher)

	return &GraphResult{
		Root:      result.Root,
		Cycles:    result.Cycles,
		Conflicts: result.Conflicts,
	}, nil
}

// bundleStoreFetcher adapts oci.BundleStore to graph.ContractFetcher.
type bundleStoreFetcher struct {
	store BundlePuller
}

// BundlePuller is the subset of oci.BundleStore needed by the fetcher.
// Defined here to avoid importing internal/oci from internal/graph.
type BundlePuller interface {
	Pull(ctx context.Context, ref string) (*contract.Bundle, error)
}

func (f *bundleStoreFetcher) Fetch(ctx context.Context, ref string) (*contract.Contract, error) {
	bundle, err := f.store.Pull(ctx, ref)
	if err != nil {
		return nil, err
	}
	return bundle.Contract, nil
}
