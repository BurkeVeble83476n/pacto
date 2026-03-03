package app

import (
	"context"
	"testing"
)

func TestDiff_LocalFiles(t *testing.T) {
	oldPath := writeTestBundle(t)
	newPath := writeTestBundle(t)
	svc := NewService(nil, nil)
	result, err := svc.Diff(context.Background(), DiffOptions{OldPath: oldPath, NewPath: newPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OldPath != oldPath {
		t.Errorf("expected OldPath=%s, got %s", oldPath, result.OldPath)
	}
	if result.NewPath != newPath {
		t.Errorf("expected NewPath=%s, got %s", newPath, result.NewPath)
	}
	if result.Classification == "" {
		t.Error("expected non-empty classification")
	}
}

func TestDiff_OldPathError(t *testing.T) {
	newPath := writeTestBundle(t)
	svc := NewService(nil, nil)
	_, err := svc.Diff(context.Background(), DiffOptions{OldPath: "/nonexistent/pacto.yaml", NewPath: newPath})
	if err == nil {
		t.Error("expected error for nonexistent old path")
	}
}

func TestDiff_NewPathError(t *testing.T) {
	oldPath := writeTestBundle(t)
	svc := NewService(nil, nil)
	_, err := svc.Diff(context.Background(), DiffOptions{OldPath: oldPath, NewPath: "/nonexistent/pacto.yaml"})
	if err == nil {
		t.Error("expected error for nonexistent new path")
	}
}

func TestDiff_OCIRef(t *testing.T) {
	store := &mockBundleStore{}
	svc := NewService(store, nil)
	result, err := svc.Diff(context.Background(), DiffOptions{
		OldPath: "oci://ghcr.io/acme/svc:1.0.0",
		NewPath: "oci://ghcr.io/acme/svc:2.0.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Classification == "" {
		t.Error("expected non-empty classification")
	}
}
