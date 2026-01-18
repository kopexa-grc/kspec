// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package ms365

import (
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

// Client wraps the Microsoft Graph API client with rate limiting support.
type Client struct {
	graph   *msgraphsdk.GraphServiceClient
	limiter *ratelimit.Limiter
}

// ClientConfig holds configuration for creating a Microsoft Graph client.
type ClientConfig struct {
	// Graph is the Microsoft Graph SDK client.
	Graph *msgraphsdk.GraphServiceClient

	// Limiter is an optional rate limiter. If nil, no rate limiting is applied.
	// Microsoft Graph SDK has built-in throttling handling, but we use proactive
	// rate limiting to avoid hitting limits.
	Limiter *ratelimit.Limiter
}

// NewClient creates a new Microsoft Graph client wrapper with the given configuration.
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		graph:   cfg.Graph,
		limiter: cfg.Limiter,
	}
}

// Graph returns the underlying Microsoft Graph SDK client for direct API access.
func (c *Client) Graph() *msgraphsdk.GraphServiceClient {
	return c.graph
}

// Limiter returns the rate limiter, if configured.
func (c *Client) Limiter() *ratelimit.Limiter {
	return c.limiter
}
