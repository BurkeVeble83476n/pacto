package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/trianalab/pacto/pkg/contract"
)

func TestGraph_Local(t *testing.T) {
	path := writeTestBundle(t)
	svc := NewService(nil, nil)
	result, err := svc.Graph(context.Background(), GraphOptions{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Root == nil {
		t.Fatal("expected non-nil root")
	}
	if result.Root.Name != "test-svc" {
		t.Errorf("expected root Name=test-svc, got %s", result.Root.Name)
	}
}

func TestGraph_WithStore(t *testing.T) {
	path := writeTestBundle(t)
	store := &mockBundleStore{}
	svc := NewService(store, nil)
	result, err := svc.Graph(context.Background(), GraphOptions{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Root == nil {
		t.Fatal("expected non-nil root")
	}
}

func TestGraph_OCIRef(t *testing.T) {
	store := &mockBundleStore{}
	svc := NewService(store, nil)
	result, err := svc.Graph(context.Background(), GraphOptions{Path: "oci://ghcr.io/acme/svc:1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Root == nil {
		t.Fatal("expected non-nil root")
	}
}

func TestGraph_ResolveError(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.Graph(context.Background(), GraphOptions{Path: "/nonexistent/pacto.yaml"})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestBundleStoreFetcher_Success(t *testing.T) {
	store := &mockBundleStore{}
	fetcher := &bundleStoreFetcher{store: store}
	c, err := fetcher.Fetch(context.Background(), "ghcr.io/acme/svc:1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Service.Name != "test-svc" {
		t.Errorf("expected test-svc, got %s", c.Service.Name)
	}
}

func TestBundleStoreFetcher_Error(t *testing.T) {
	store := &mockBundleStore{
		PullFn: func(_ context.Context, _ string) (*contract.Bundle, error) {
			return nil, fmt.Errorf("pull failed")
		},
	}
	fetcher := &bundleStoreFetcher{store: store}
	_, err := fetcher.Fetch(context.Background(), "ghcr.io/acme/svc:1.0.0")
	if err == nil {
		t.Error("expected error from store")
	}
}
