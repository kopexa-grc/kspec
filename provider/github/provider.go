package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v62/github"
	"github.com/juliankoehn/kspec/core"
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
	token := config["token"] // extract token from config
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
	}
}
