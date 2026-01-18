// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

// HTTP provides access to HTTP response information including headers and status.
type HTTP struct{}

// NewHTTP creates a new HTTP resource instance.
func NewHTTP() *HTTP {
	return &HTTP{}
}

// Name returns the resource type name.
func (r *HTTP) Name() string {
	return "http"
}

// Fetch retrieves HTTP response data from the target URL specified in the asset.
func (r *HTTP) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	domainName := asset.FQDN
	if domainName == "" {
		domainName = asset.Config["domain"]
	}
	if domainName == "" {
		domainName = asset.Config["target"]
	}

	if domainName == "" {
		return nil, fmt.Errorf("missing domain configuration: provide 'domain' or 'target' in asset config, or set asset FQDN")
	}

	// Simple heuristic: if no protocol, try https first, then http?
	// Or just default to https.
	targetURL := domainName
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		// We might want to follow redirects? Default client follows 10 redirects.
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse Headers into a map friendly for CEL
	// http.Header is map[string][]string.
	// CEL works best with map[string]interface{} where values are lists or strings.
	// We'll expose headers as map[string []string] to allow .exists, .all, etc.
	headers := make(map[string]interface{})
	for k, v := range resp.Header {
		// Normalize key? The policy uses "X-Content-Type-Options" (canonical).
		// Go's http.Header canonicalizes keys.
		headers[k] = v
	}

	res := core.Resource{
		"name":    domainName,
		"url":     targetURL,
		"status":  resp.StatusCode,
		"proto":   resp.Proto,
		"headers": headers,
	}

	return []core.Resource{res}, nil
}
