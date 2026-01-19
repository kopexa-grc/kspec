// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// resolveFile imports a local file.
func (r *Resolver) resolveFile(filePath, sourcePath string) ([]Check, error) {
	// Make path relative to source file if not absolute
	if !filepath.IsAbs(filePath) {
		baseDir := r.basePath
		if sourcePath != "" && !isURL(sourcePath) {
			baseDir = filepath.Dir(sourcePath)
		}
		filePath = filepath.Join(baseDir, filePath)
	}

	// Get absolute path for cycle detection
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if already visited
	if r.visited[absPath] {
		return nil, nil // Already imported, skip
	}
	r.visited[absPath] = true

	// Read and parse file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	// Recursively resolve imports in the imported policy
	if err := r.ResolveImports(&p, absPath); err != nil {
		return nil, err
	}

	return p.Queries, nil
}
