// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package ratelimit

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// Config holds rate limit configuration.
type Config struct {
	// RequestsPerSecond is the rate of requests allowed per second.
	RequestsPerSecond float64
	// Burst is the maximum number of requests that can be made at once.
	Burst int
}

// Limiter wraps rate.Limiter with provider context.
type Limiter struct {
	limiter *rate.Limiter
	name    string
	config  Config
}

// New creates a new rate limiter with the specified configuration.
func New(name string, cfg Config) *Limiter {
	return &Limiter{
		limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst),
		name:    name,
		config:  cfg,
	}
}

// Wait blocks until the limiter permits one event.
// It returns an error if the context is canceled.
func (l *Limiter) Wait(ctx context.Context) error {
	return l.limiter.Wait(ctx)
}

// WaitN blocks until the limiter permits n events.
// It returns an error if the context is canceled.
func (l *Limiter) WaitN(ctx context.Context, n int) error {
	return l.limiter.WaitN(ctx, n)
}

// Allow reports whether one event may happen now.
func (l *Limiter) Allow() bool {
	return l.limiter.Allow()
}

// AllowN reports whether n events may happen now.
func (l *Limiter) AllowN(n int) bool {
	return l.limiter.AllowN(time.Now(), n)
}

// Name returns the name of the rate limiter (typically the provider name).
func (l *Limiter) Name() string {
	return l.name
}

// Config returns the configuration of the rate limiter.
func (l *Limiter) Config() Config {
	return l.config
}

// SetLimit adjusts the rate limit at runtime.
func (l *Limiter) SetLimit(rps float64) {
	l.limiter.SetLimit(rate.Limit(rps))
	l.config.RequestsPerSecond = rps
}

// SetBurst adjusts the burst limit at runtime.
func (l *Limiter) SetBurst(burst int) {
	l.limiter.SetBurst(burst)
	l.config.Burst = burst
}

// NopLimiter returns a limiter that never blocks.
// Useful for testing or when rate limiting should be disabled.
func NopLimiter() *Limiter {
	return &Limiter{
		limiter: rate.NewLimiter(rate.Inf, 0),
		name:    "nop",
		config:  Config{RequestsPerSecond: 0, Burst: 0},
	}
}
