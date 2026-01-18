// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package hetzner provides a kspec provider for Hetzner Cloud infrastructure scanning.
package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

// Client wraps the Hetzner Cloud API client with rate limiting support.
type Client struct {
	hc      *hcloud.Client
	limiter *ratelimit.Limiter
}

// ClientConfig holds configuration for creating a Hetzner client.
type ClientConfig struct {
	// Token is the Hetzner Cloud API token.
	Token string

	// Limiter is an optional rate limiter. If nil, no rate limiting is applied.
	Limiter *ratelimit.Limiter
}

// NewClient creates a new Hetzner Cloud client with the given configuration.
func NewClient(_ context.Context, cfg ClientConfig) *Client {
	hc := hcloud.NewClient(hcloud.WithToken(cfg.Token))

	return &Client{
		hc:      hc,
		limiter: cfg.Limiter,
	}
}

// Do executes a function with rate limiting.
// If no rate limiter is configured, the function is executed immediately.
func (c *Client) Do(ctx context.Context, fn func() error) error {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}
	}
	return fn()
}

// HCloud returns the underlying hcloud client for direct API access.
// Callers should use Do() for rate-limited operations.
func (c *Client) HCloud() *hcloud.Client {
	return c.hc
}
