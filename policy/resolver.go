// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// Resolver handles resolving and loading policy imports.
type Resolver struct {
	visited    map[string]bool
	httpClient *http.Client
	basePath   string
}

// NewResolver creates a new import resolver.
func NewResolver(basePath string) *Resolver {
	return &Resolver{
		visited: make(map[string]bool),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		basePath: basePath,
	}
}

// ResolveImports recursively resolves all imports in a policy and merges them.
// Imported queries are prepended to the policy's queries.
func (r *Resolver) ResolveImports(p *Policy, sourcePath string) error {
	if len(p.Imports) == 0 {
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
	var importedQueries []Check

	for _, importPath := range p.Imports {
		queries, err := r.resolveImport(importPath, sourcePath)
		if err != nil {
			return fmt.Errorf("failed to resolve import %q: %w", importPath, err)
		}
		importedQueries = append(importedQueries, queries...)
	}

	// Prepend imported queries (local queries take precedence for same UIDs)
	p.Queries = append(importedQueries, p.Queries...)

	return nil
}

// resolveImport resolves a single import and returns the queries.
func (r *Resolver) resolveImport(importPath, sourcePath string) ([]Check, error) {
	switch {
	case isURL(importPath):
		return r.resolveURL(importPath)
	case isGlob(importPath):
		return r.resolveGlob(importPath, sourcePath)
	default:
		return r.resolveFile(importPath, sourcePath)
	}
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
