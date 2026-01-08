// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// schema-gen generates JSON Schema from Go structs.
// Usage: go run ./cmd/schema-gen
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/kopexa-grc/kspec/core"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Create reflector with custom settings
	r := &jsonschema.Reflector{
		DoNotReference:             false,
		ExpandedStruct:             false,
		AllowAdditionalProperties:  false,
		RequiredFromJSONSchemaTags: true,
	}

	// Generate schema from Policy struct
	schema := r.Reflect(&core.Policy{})

	// Set schema metadata
	schema.ID = "https://kopexa.io/schemas/policy.schema.json"
	schema.Title = "Kopexa Policy"
	schema.Description = "Schema for Kopexa security policy files"

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Determine output path
	outputPath := "schema/policy.schema.json"
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	fmt.Printf("Schema generated: %s\n", outputPath)
	return nil
}
