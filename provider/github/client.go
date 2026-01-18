// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"

	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

// Client wraps the GitHub API client with rate limiting support.
type Client struct {
	gh      *github.Client
	limiter *ratelimit.Limiter
}

// ClientConfig holds configuration for creating a GitHub client.
type ClientConfig struct {
	// Token is the GitHub personal access token or OAuth token.
	Token string

	// Limiter is an optional rate limiter. If nil, no rate limiting is applied.
	// GitHub's API has built-in rate limit headers, but we use proactive
	// rate limiting to avoid hitting limits.
	Limiter *ratelimit.Limiter
}

// NewClient creates a new GitHub client with the given configuration.
func NewClient(ctx context.Context, cfg ClientConfig) *Client {
	var tc *http.Client
	if cfg.Token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.Token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	return &Client{
		gh:      github.NewClient(tc),
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

// GitHub returns the underlying go-github client for direct API access.
// Callers should use Do() for rate-limited operations.
func (c *Client) GitHub() *github.Client {
	return c.gh
}

// Organizations returns the Organizations service.
func (c *Client) Organizations() *github.OrganizationsService {
	return c.gh.Organizations
}

// Repositories returns the Repositories service.
func (c *Client) Repositories() *github.RepositoriesService {
	return c.gh.Repositories
}

// Teams returns the Teams service.
func (c *Client) Teams() *github.TeamsService {
	return c.gh.Teams
}

// Git returns the Git service.
func (c *Client) Git() *github.GitService {
	return c.gh.Git
}
