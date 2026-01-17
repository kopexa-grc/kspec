// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package ratelimit provides rate limiting utilities for API calls.
//
// This package wraps golang.org/x/time/rate to provide provider-specific
// rate limiters with sensible defaults. Rate limiting helps prevent
// API throttling when scanning large numbers of resources.
//
// # Usage
//
//	limiter := ratelimit.NewFromDefaults("aws")
//	if err := limiter.Wait(ctx); err != nil {
//	    return err
//	}
//	// Make API call
//
// # Provider Defaults
//
// Each provider has default rate limits configured based on their API limits:
//   - AWS: 10 req/s, burst 20
//   - GitHub: 5 req/s, burst 10
//   - Azure: 10 req/s, burst 20
//   - Cloudflare: 4 req/s, burst 8
//   - MS365: 10 req/s, burst 20
//   - Hetzner: 10 req/s, burst 20
//   - Factorial: 5 req/s, burst 10
package ratelimit
