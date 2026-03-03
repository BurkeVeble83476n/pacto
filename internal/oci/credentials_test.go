package oci_test

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/trianalab/pacto/internal/oci"
)

func TestNewKeychain_WithToken(t *testing.T) {
	kc := oci.NewKeychain(oci.CredentialOptions{Token: "my-token"})

	reg, err := name.NewRegistry("example.com", name.Insecure)
	if err != nil {
		t.Fatalf("NewRegistry() error: %v", err)
	}

	auth, err := kc.Resolve(reg)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	cfg, err := auth.Authorization()
	if err != nil {
		t.Fatalf("Authorization() error: %v", err)
	}

	if cfg.RegistryToken != "my-token" {
		t.Errorf("RegistryToken = %q, want %q", cfg.RegistryToken, "my-token")
	}
}

func TestNewKeychain_WithUsernamePassword(t *testing.T) {
	kc := oci.NewKeychain(oci.CredentialOptions{Username: "user", Password: "pass"})

	reg, err := name.NewRegistry("example.com", name.Insecure)
	if err != nil {
		t.Fatalf("NewRegistry() error: %v", err)
	}

	auth, err := kc.Resolve(reg)
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}

	cfg, err := auth.Authorization()
	if err != nil {
		t.Fatalf("Authorization() error: %v", err)
	}

	if cfg.Username != "user" {
		t.Errorf("Username = %q, want %q", cfg.Username, "user")
	}
	if cfg.Password != "pass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "pass")
	}
}

func TestNewKeychain_Default(t *testing.T) {
	kc := oci.NewKeychain(oci.CredentialOptions{})

	// When no credentials are provided, the default keychain is returned.
	if kc != authn.DefaultKeychain {
		t.Errorf("expected DefaultKeychain, got %T", kc)
	}
}
