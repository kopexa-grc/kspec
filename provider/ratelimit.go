// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package provider

import "github.com/kopexa-grc/kspec/pkg/ratelimit"

// RateLimitDefaults contains default rate limit configurations for each provider.
// These values are conservative to prevent API throttling.
var RateLimitDefaults = map[string]ratelimit.Config{
	"aws":        {RequestsPerSecond: 10, Burst: 20},  // Conservative for IAM operations
	"github":     {RequestsPerSecond: 5, Burst: 10},   // 5000/hr = ~1.4/s, using 5/s for safety
	"azure":      {RequestsPerSecond: 10, Burst: 20},  // ARM has varying limits per resource
	"cloudflare": {RequestsPerSecond: 4, Burst: 8},    // 1200/5min = 4/s
	"ms365":      {RequestsPerSecond: 10, Burst: 20},  // Graph API limits vary
	"hetzner":    {RequestsPerSecond: 10, Burst: 20},  // Conservative default
	"factorial":  {RequestsPerSecond: 5, Burst: 10},   // Conservative default
	"atlassian":  {RequestsPerSecond: 10, Burst: 20},  // Atlassian API limits
	"network":    {RequestsPerSecond: 20, Burst: 40},  // Network scans can be faster
	"os":         {RequestsPerSecond: 50, Burst: 100}, // Local OS operations
	"sbom":       {RequestsPerSecond: 50, Burst: 100}, // Local file parsing
}

// DefaultRateLimitConfig is used for providers without specific defaults.
var DefaultRateLimitConfig = ratelimit.Config{RequestsPerSecond: 5, Burst: 10}

// NewRateLimiter creates a rate limiter for the given provider using provider defaults.
func NewRateLimiter(providerName string) *ratelimit.Limiter {
	cfg, ok := RateLimitDefaults[providerName]
	if !ok {
		cfg = DefaultRateLimitConfig
	}
	return ratelimit.New(providerName, cfg)
}
