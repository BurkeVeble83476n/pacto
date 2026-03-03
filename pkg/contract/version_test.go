package contract_test

import (
	"testing"

	"github.com/trianalab/pacto/pkg/contract"
)

func TestIsValidSpecVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0", true},
		{"2.0", false},
		{"", false},
		{"1.1", false},
		{"v1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := contract.IsValidSpecVersion(tt.version); got != tt.valid {
				t.Errorf("IsValidSpecVersion(%q) = %v, want %v", tt.version, got, tt.valid)
			}
		})
	}
}
