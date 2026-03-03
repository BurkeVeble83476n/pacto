package validation

import (
	"encoding/json"
	"io/fs"

	"github.com/trianalab/pacto/pkg/contract"
	"gopkg.in/yaml.v3"
)

// Validate runs all three validation layers in order on the given contract.
// If structural validation fails, subsequent layers are skipped.
// The rawYAML parameter is the original YAML bytes for JSON Schema validation.
// The bundleFS parameter provides access to bundle files for cross-field validation.
func Validate(c *contract.Contract, rawYAML []byte, bundleFS fs.FS) ValidationResult {
	var result ValidationResult

	// Layer 1: Structural validation via JSON Schema.
	// Convert YAML to a generic interface{} for JSON Schema validation.
	structuralData, err := yamlToGeneric(rawYAML)
	if err != nil {
		result.AddError("", "YAML_PARSE_ERROR", err.Error())
		return result
	}

	structuralResult := ValidateStructural(structuralData)
	result.Merge(structuralResult)
	if !result.IsValid() {
		return result
	}

	// Layer 2: Cross-field validation.
	crossFieldResult := ValidateCrossField(c, bundleFS)
	result.Merge(crossFieldResult)
	if !result.IsValid() {
		return result
	}

	// Layer 3: Semantic validation.
	semanticResult := ValidateSemantic(c)
	result.Merge(semanticResult)

	return result
}

// yamlToGeneric converts YAML bytes to a generic interface{} suitable for
// JSON Schema validation. It goes through JSON to ensure type compatibility
// with the JSON Schema library.
func yamlToGeneric(data []byte) (interface{}, error) {
	var yamlObj interface{}
	if err := yaml.Unmarshal(data, &yamlObj); err != nil {
		return nil, err
	}

	// Convert map[string]interface{} (yaml uses map[interface{}]interface{} for nested)
	converted := convertYAMLToJSON(yamlObj)

	// Round-trip through JSON to ensure types match JSON Schema expectations
	jsonBytes, err := json.Marshal(converted)
	if err != nil {
		return nil, err
	}

	var result interface{}
	_ = json.Unmarshal(jsonBytes, &result)

	return result, nil
}

// convertYAMLToJSON recursively converts YAML-style maps to JSON-compatible maps.
func convertYAMLToJSON(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			result[key] = convertYAMLToJSON(val)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			result[key.(string)] = convertYAMLToJSON(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = convertYAMLToJSON(val)
		}
		return result
	default:
		return v
	}
}
