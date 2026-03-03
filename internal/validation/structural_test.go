package validation

import (
	"fmt"
	"testing"
)

func TestCompileSchema_InvalidJSON(t *testing.T) {
	_, err := compileSchema([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCompileSchema_InvalidSchema(t *testing.T) {
	_, err := compileSchema([]byte(`{"type": 12345}`))
	if err == nil {
		t.Error("expected error for invalid schema")
	}
}

func TestMustCompileSchema_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid schema data")
		}
	}()
	mustCompileSchema([]byte("bad"))
}

func TestValidateStructural_NonValidationError(t *testing.T) {
	old := schemaValidateFn
	schemaValidateFn = func(interface{}) error { return fmt.Errorf("internal error") }
	defer func() { schemaValidateFn = old }()

	result := ValidateStructural(nil)
	if result.IsValid() {
		t.Error("expected invalid result")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}
	if result.Errors[0].Code != "SCHEMA_ERROR" {
		t.Errorf("expected SCHEMA_ERROR, got %s", result.Errors[0].Code)
	}
}
