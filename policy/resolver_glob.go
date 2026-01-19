// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"fmt"
	"path/filepath"
)

// resolveGlob resolves a glob pattern and imports all matching files.
func (r *Resolver) resolveGlob(pattern, sourcePath string) ([]Check, error) {
	// Make pattern relative to source file
	baseDir := r.basePath
	if sourcePath != "" && !isURL(sourcePath) {
		baseDir = filepath.Dir(sourcePath)
	}

	// If pattern is relative, make it absolute
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(baseDir, pattern)
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no files match pattern %q", pattern)
	}

	var allQueries []Check
	for _, match := range matches {
		queries, err := r.resolveFile(match, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to import %q: %w", match, err)
		}
		allQueries = append(allQueries, queries...)
	}

	return allQueries, nil
}
