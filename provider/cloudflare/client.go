// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package cloudflare

import (
	"github.com/cloudflare/cloudflare-go/v4"
)

// Client wraps the Cloudflare API client.
// The Cloudflare SDK has built-in rate limiting and automatic retry,
// so no additional rate limiting is needed.
type Client struct {
	cf        *cloudflare.Client
	accountID string
}

// ClientConfig holds configuration for creating a Cloudflare client.
type ClientConfig struct {
	// CF is the underlying Cloudflare client.
	CF *cloudflare.Client
	// AccountID is the Cloudflare account ID (optional).
	AccountID string
}

// NewClient creates a new Cloudflare client wrapper.
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		cf:        cfg.CF,
		accountID: cfg.AccountID,
	}
}

// Cloudflare returns the underlying cloudflare-go client for direct API access.
func (c *Client) Cloudflare() *cloudflare.Client {
	return c.cf
}

// AccountID returns the configured account ID.
func (c *Client) AccountID() string {
	return c.accountID
}
