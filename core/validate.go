// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package core

import (
	"bytes"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// PolicyValidator validates policy files against a JSON schema.
type PolicyValidator struct {
	schema *jsonschema.Schema
}

// NewPolicyValidator creates a new validator from a schema file.
func NewPolicyValidator(schemaPath string) (*PolicyValidator, error) {
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	return NewPolicyValidatorFromBytes(schemaData)
}

// NewPolicyValidatorFromBytes creates a new validator from schema bytes.
func NewPolicyValidatorFromBytes(schemaData []byte) (*PolicyValidator, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("policy.schema.json", bytes.NewReader(schemaData)); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	schema, err := compiler.Compile("policy.schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &PolicyValidator{schema: schema}, nil
}

// Validate validates policy data (JSON or YAML bytes) against the JSON schema.
func (v *PolicyValidator) Validate(data []byte) error {
	// Parse as YAML (YAML is a superset of JSON)
	var doc interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse policy data: %w", err)
	}

	// Convert YAML types to JSON-compatible types
	doc = convertYAMLToJSON(doc)

	if err := v.schema.Validate(doc); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	return nil
}

// ValidateFile validates a policy file against the JSON schema.
func (v *PolicyValidator) ValidateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	return v.Validate(data)
}

// convertYAMLToJSON recursively converts YAML-specific types to JSON-compatible types.
// YAML uses map[string]interface{} while some parsers use map[interface{}]interface{}.
func convertYAMLToJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range val {
			m[fmt.Sprintf("%v", k)] = convertYAMLToJSON(v)
		}
		return m
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k, v := range val {
			m[k] = convertYAMLToJSON(v)
		}
		return m
	case []interface{}:
		for i, v := range val {
			val[i] = convertYAMLToJSON(v)
		}
		return val
	default:
		return v
	}
}
