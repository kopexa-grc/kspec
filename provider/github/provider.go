package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v62/github"
	"github.com/kopexa-grc/kspec/core"
	"golang.org/x/oauth2"
)

// GithubProvider is now a factory
type GithubProvider struct{}

func NewGithubProvider() *GithubProvider {
	return &GithubProvider{}
}

func (p *GithubProvider) Name() string {
	return "github"
}

func (p *GithubProvider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
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
	return &GithubConnection{client: client}, nil
}

// GithubConnection holds the session
type GithubConnection struct {
	client *github.Client
}

func (c *GithubConnection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&RepoResource{client: c.client},
		&TeamResource{client: c.client},
		&OrganizationResource{client: c.client},
		&BranchResource{client: c.client},
	}
}
