package network

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/juliankoehn/kspec/core"
)

type HTTPResource struct{}

func (r *HTTPResource) Name() string {
	return "http"
}

func (r *HTTPResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	domainName := asset.FQDN
	if domainName == "" {
		domainName = asset.Config["domain"]
	}

	if domainName == "" {
		// Fallback: try to see if "host" is passed or similar?
		// Usually 'scan host google.com' puts 'google.com' in domain via our main.go injection logic.
		return nil, fmt.Errorf("missing 'domain' in asset for http resource")
	}

	// Simple heuristic: if no protocol, try https first, then http?
	// Or just default to https.
	targetURL := domainName
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
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
