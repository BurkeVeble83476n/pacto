package app

import (
	"context"
	"testing"

	"github.com/trianalab/pacto/pkg/contract"
)

func TestExplain_Local(t *testing.T) {
	path := writeTestBundle(t)
	svc := NewService(nil, nil)
	result, err := svc.Explain(context.Background(), ExplainOptions{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test-svc" {
		t.Errorf("expected Name=test-svc, got %s", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected Version=1.0.0, got %s", result.Version)
	}
	if result.PactoVersion != "1.0" {
		t.Errorf("expected PactoVersion=1.0, got %s", result.PactoVersion)
	}
	if result.Runtime.WorkloadType != "service" {
		t.Errorf("expected WorkloadType=service, got %s", result.Runtime.WorkloadType)
	}
}

func TestExplain_WithInterfaces(t *testing.T) {
	path := writeTestBundle(t)
	svc := NewService(nil, nil)
	result, err := svc.Explain(context.Background(), ExplainOptions{Path: path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(result.Interfaces))
	}
	iface := result.Interfaces[0]
	if iface.Name != "api" {
		t.Errorf("expected interface Name=api, got %s", iface.Name)
	}
	if iface.Type != "http" {
		t.Errorf("expected interface Type=http, got %s", iface.Type)
	}
	if iface.Port == nil || *iface.Port != 8080 {
		t.Errorf("expected interface Port=8080, got %v", iface.Port)
	}
}

func TestExplain_OCIRef(t *testing.T) {
	store := &mockBundleStore{}
	svc := NewService(store, nil)
	result, err := svc.Explain(context.Background(), ExplainOptions{Path: "oci://ghcr.io/acme/svc:1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test-svc" {
		t.Errorf("expected Name=test-svc, got %s", result.Name)
	}
}

func TestExplain_WithDependencies(t *testing.T) {
	store := &mockBundleStore{
		PullFn: func(_ context.Context, _ string) (*contract.Bundle, error) {
			b := testBundle()
			b.Contract.Dependencies = []contract.Dependency{
				{Ref: "ghcr.io/acme/dep:1.0.0", Required: true, Compatibility: "^1.0.0"},
			}
			return b, nil
		},
	}
	svc := NewService(store, nil)
	result, err := svc.Explain(context.Background(), ExplainOptions{Path: "oci://ghcr.io/acme/svc:1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(result.Dependencies))
	}
	dep := result.Dependencies[0]
	if dep.Ref != "ghcr.io/acme/dep:1.0.0" {
		t.Errorf("expected Ref=ghcr.io/acme/dep:1.0.0, got %s", dep.Ref)
	}
	if !dep.Required {
		t.Error("expected Required=true")
	}
}

func TestExplain_NotFound(t *testing.T) {
	svc := NewService(nil, nil)
	_, err := svc.Explain(context.Background(), ExplainOptions{Path: "/nonexistent/pacto.yaml"})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}
