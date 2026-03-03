package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/trianalab/pacto/pkg/contract"
)

func TestPush_Success(t *testing.T) {
	path := writeTestBundle(t)
	store := &mockBundleStore{}
	svc := NewService(store, nil)
	result, err := svc.Push(context.Background(), PushOptions{Ref: "ghcr.io/acme/svc:1.0.0", Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Ref != "ghcr.io/acme/svc:1.0.0" {
		t.Errorf("expected Ref=ghcr.io/acme/svc:1.0.0, got %s", result.Ref)
	}
	if result.Digest != "sha256:abc123" {
		t.Errorf("expected Digest=sha256:abc123, got %s", result.Digest)
	}
	if result.Name != "test-svc" {
		t.Errorf("expected Name=test-svc, got %s", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected Version=1.0.0, got %s", result.Version)
	}
}

func TestPush_NilStore(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.Push(context.Background(), PushOptions{Ref: "ghcr.io/acme/svc:1.0.0", Path: "pacto.yaml"})
	if err == nil {
		t.Error("expected error for nil store")
	}
}

func TestPush_InvalidContract(t *testing.T) {
	path := writeInvalidBundle(t)
	store := &mockBundleStore{}
	svc := NewService(store, nil)
	_, err := svc.Push(context.Background(), PushOptions{Ref: "ghcr.io/acme/svc:1.0.0", Path: path})
	if err == nil {
		t.Error("expected error for invalid contract")
	}
}

func TestPush_FileNotFound(t *testing.T) {
	store := &mockBundleStore{}
	svc := NewService(store, nil)
	_, err := svc.Push(context.Background(), PushOptions{Ref: "ghcr.io/acme/svc:1.0.0", Path: "/nonexistent/pacto.yaml"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestPush_StoreError(t *testing.T) {
	path := writeTestBundle(t)
	store := &mockBundleStore{
		PushFn: func(_ context.Context, _ string, _ *contract.Bundle) (string, error) {
			return "", fmt.Errorf("push failed")
		},
	}
	svc := NewService(store, nil)
	_, err := svc.Push(context.Background(), PushOptions{Ref: "ghcr.io/acme/svc:1.0.0", Path: path})
	if err == nil {
		t.Error("expected error from store")
	}
}
