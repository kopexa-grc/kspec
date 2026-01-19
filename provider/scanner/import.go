// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package scanner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/kopexa-grc/kspec/core"
)

// ImportResolver handles resolving and loading policy imports.
type ImportResolver struct {
	// visited tracks already loaded policies to prevent circular imports
	visited map[string]bool
	// httpClient for fetching remote policies
	httpClient *http.Client
	// basePath is the base directory for resolving relative imports
	basePath string
}

// NewImportResolver creates a new import resolver.
func NewImportResolver(basePath string) *ImportResolver {
	return &ImportResolver{
		visited: make(map[string]bool),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		basePath: basePath,
	}
}

// ResolveImports recursively resolves all imports in a policy and merges them.
// The returned policy contains all queries from imported policies.
func (r *ImportResolver) ResolveImports(policy *core.Policy, sourcePath string) error {
	if len(policy.Imports) == 0 {
		return nil
	}

	// Mark current policy as visited
	if sourcePath != "" {
		absPath, err := filepath.Abs(sourcePath)
		if err == nil {
			r.visited[absPath] = true
		}
	}

	// Collect all imported queries
	var importedQueries []core.Check

	for _, importPath := range policy.Imports {
		queries, err := r.resolveImport(importPath, sourcePath)
		if err != nil {
			return fmt.Errorf("failed to resolve import %q: %w", importPath, err)
		}
		importedQueries = append(importedQueries, queries...)
	}

	// Prepend imported queries (local queries take precedence for same UIDs)
	policy.Queries = append(importedQueries, policy.Queries...)

	return nil
}

// resolveImport resolves a single import and returns the queries.
func (r *ImportResolver) resolveImport(importPath, sourcePath string) ([]core.Check, error) {
	// Determine import type
	switch {
	case isURL(importPath):
		return r.resolveURLImport(importPath)
	case isGlob(importPath):
		return r.resolveGlobImport(importPath, sourcePath)
	default:
		return r.resolveFileImport(importPath, sourcePath)
	}
}

// resolveURLImport fetches a policy from a URL.
func (r *ImportResolver) resolveURLImport(urlStr string) ([]core.Check, error) {
	// Check if already visited
	if r.visited[urlStr] {
		return nil, nil // Already imported, skip
	}
	r.visited[urlStr] = true

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var policy core.Policy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	// Recursively resolve imports in the imported policy
	if err := r.ResolveImports(&policy, urlStr); err != nil {
		return nil, err
	}

	return policy.Queries, nil
}

// resolveGlobImport resolves a glob pattern and imports all matching files.
func (r *ImportResolver) resolveGlobImport(pattern, sourcePath string) ([]core.Check, error) {
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

	var allQueries []core.Check
	for _, match := range matches {
		queries, err := r.resolveFileImport(match, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to import %q: %w", match, err)
		}
		allQueries = append(allQueries, queries...)
	}

	return allQueries, nil
}

// resolveFileImport imports a local file.
func (r *ImportResolver) resolveFileImport(filePath, sourcePath string) ([]core.Check, error) {
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

	var policy core.Policy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	// Recursively resolve imports in the imported policy
	if err := r.ResolveImports(&policy, absPath); err != nil {
		return nil, err
	}

	return policy.Queries, nil
}

// isURL checks if a string is a URL.
func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

// isGlob checks if a string contains glob characters.
func isGlob(s string) bool {
	return strings.ContainsAny(s, "*?[")
}
