// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package ratelimit provides rate limiting utilities for API calls.
//
// This package wraps golang.org/x/time/rate to provide a simple rate limiter
// interface. Rate limiting helps prevent API throttling when scanning large
// numbers of resources.
//
// # Usage
//
//	limiter := ratelimit.New("aws", ratelimit.Config{RequestsPerSecond: 10, Burst: 20})
//	if err := limiter.Wait(ctx); err != nil {
//	    return err
//	}
//	// Make API call
//
// Provider-specific defaults are defined in the provider package.
package ratelimit
