// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load loads policies from a file or directory.
// It also resolves any imports defined in the policies.
func Load(policyFile, policyDir string) ([]Policy, error) {
	var policies []Policy

	// Determine base path for import resolution
	basePath := "."
	if policyDir != "" {
		basePath = policyDir
	} else if policyFile != "" {
		basePath = filepath.Dir(policyFile)
	}

	// Create import resolver
	resolver := NewResolver(basePath)

	loadPolicy := func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var p Policy
		if err := yaml.Unmarshal(data, &p); err != nil {
			return err
		}

		// Resolve imports if any
		if len(p.Imports) > 0 {
			if err := resolver.ResolveImports(&p, path); err != nil {
				return fmt.Errorf("failed to resolve imports: %w", err)
			}
		}

		policies = append(policies, p)
		return nil
	}

	if policyDir != "" {
		files, err := os.ReadDir(policyDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read policy directory: %w", err)
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
				path := filepath.Join(policyDir, file.Name())
				if err := loadPolicy(path); err != nil {
					log.Printf("Warning: failed to load policy %s: %v", file.Name(), err)
				}
			}
		}
	} else if policyFile != "" {
		if err := loadPolicy(policyFile); err != nil {
			return nil, fmt.Errorf("failed to read policy file: %w", err)
		}
	}

	return policies, nil
}

// LoadFromBytes loads a single policy from YAML bytes.
func LoadFromBytes(data []byte) (*Policy, error) {
	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}
	return &p, nil
}

// LoadFromFile loads a single policy from a file.
func LoadFromFile(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}
	return LoadFromBytes(data)
}
