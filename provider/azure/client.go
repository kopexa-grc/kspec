// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

// Client wraps Azure credentials and subscription info with rate limiting support.
type Client struct {
	credential     azcore.TokenCredential
	subscriptionID string
	limiter        *ratelimit.Limiter
}

// ClientConfig holds configuration for creating an Azure client.
type ClientConfig struct {
	// Credential is the Azure token credential for authentication.
	Credential azcore.TokenCredential
	// SubscriptionID is the Azure subscription to scan.
	SubscriptionID string
	// Limiter is an optional rate limiter. If nil, no rate limiting is applied.
	// Azure SDK has built-in retry, but rate limiting adds an extra safety layer.
	Limiter *ratelimit.Limiter
}

// NewClient creates a new Azure client with the given configuration.
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		credential:     cfg.Credential,
		subscriptionID: cfg.SubscriptionID,
		limiter:        cfg.Limiter,
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

// Credential returns the Azure token credential.
func (c *Client) Credential() azcore.TokenCredential {
	return c.credential
}

// SubscriptionID returns the Azure subscription ID.
func (c *Client) SubscriptionID() string {
	return c.subscriptionID
}

// Limiter returns the rate limiter, or nil if not configured.
func (c *Client) Limiter() *ratelimit.Limiter {
	return c.limiter
}
