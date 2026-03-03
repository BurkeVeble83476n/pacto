package oci

import (
	"github.com/google/go-containerregistry/pkg/authn"
)

// CredentialOptions holds explicit credentials provided via CLI flags or env vars.
type CredentialOptions struct {
	Username string
	Password string
	Token    string
}

// NewKeychain builds a keychain that tries explicit credentials first,
// then falls back to Docker config, credential helpers, and cloud auto-detection.
func NewKeychain(opts CredentialOptions) authn.Keychain {
	if opts.Token != "" {
		return staticKeychain{auth: &authn.AuthConfig{RegistryToken: opts.Token}}
	}
	if opts.Username != "" && opts.Password != "" {
		return staticKeychain{auth: &authn.AuthConfig{Username: opts.Username, Password: opts.Password}}
	}
	return authn.DefaultKeychain
}

// staticKeychain returns the same credentials for any registry.
type staticKeychain struct {
	auth *authn.AuthConfig
}

func (k staticKeychain) Resolve(_ authn.Resource) (authn.Authenticator, error) {
	return authn.FromConfig(*k.auth), nil
}
