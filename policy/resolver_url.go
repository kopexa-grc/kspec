// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
)

// resolveURL fetches a policy from a URL.
func (r *Resolver) resolveURL(urlStr string) ([]Check, error) {
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

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse policy: %w", err)
	}

	// Recursively resolve imports in the imported policy
	if err := r.ResolveImports(&p, urlStr); err != nil {
		return nil, err
	}

	return p.Queries, nil
}
