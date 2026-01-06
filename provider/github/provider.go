// Package github provides GitHub repository and organization scanning
// capabilities for security policy evaluation.
package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"

	"github.com/kopexa-grc/kspec/core"
)

// Provider implements the core.Provider interface for GitHub.
type Provider struct{}

// NewGithubProvider creates a new GitHub provider.
func NewGithubProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "github"
}

// Connect establishes a connection to GitHub APIs.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Try to parse credentials from config
	var token string

	// Attempt to parse credential
	cred, err := core.ParseCredentialFromConfig(config)
	if err != nil {
		// Fall back to legacy token support if credential parsing fails
		token = config["token"]
	} else {
		// Resolve token based on credential type
		switch cred.Type {
		case core.CredentialTypeBearer:
			token = cred.Secret
		case core.CredentialTypeEnv:
			token, err = cred.ResolveSecret()
			if err != nil {
				return nil, fmt.Errorf("failed to resolve credential from environment: %w", err)
			}
		case core.CredentialTypePassword:
			// For GitHub, password auth typically means Personal Access Token in the secret field
			token = cred.Secret
		default:
			return nil, fmt.Errorf("unsupported credential type for GitHub provider: %s", cred.Type)
		}
	}

	var tc *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	client := github.NewClient(tc)
	return &Connection{client: client}, nil
}

// Connection represents an active connection to GitHub APIs.
type Connection struct {
	client *github.Client
}

// Resources returns all available GitHub resources.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&RepoResource{client: c.client},
		&TeamResource{client: c.client},
		&OrganizationResource{client: c.client},
		&BranchResource{client: c.client},
	}
}
