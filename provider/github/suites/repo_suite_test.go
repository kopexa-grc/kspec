// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

//go:build fast

package suites_test

import (
	"testing"

	"github.com/kopexa-grc/kspec/core"
	ghsuites "github.com/kopexa-grc/kspec/provider/github/suites"
	"github.com/kopexa-grc/kspec/testing/suites"
)

func init() {
	suites.Register(&GitHubRepoSuite{})
}

// GitHubRepoSuite tests GitHub repository policies against fixtures.
type GitHubRepoSuite struct{}

func (s *GitHubRepoSuite) Name() string     { return "github-repo" }
func (s *GitHubRepoSuite) Provider() string { return "github" }

func (s *GitHubRepoSuite) Setup(_ *testing.T) error    { return nil }
func (s *GitHubRepoSuite) Teardown(_ *testing.T) error { return nil }

func (s *GitHubRepoSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "repository with Dependabot passes",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_repo",
			Fixtures:     []core.Resource{ghsuites.RepoWithDependabot.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-dependabot": {Pass: true},
			},
		},
		{
			Name:         "repository without Dependabot fails",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_repo",
			Fixtures:     []core.Resource{ghsuites.RepoWithoutDependabot.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-dependabot": {Pass: false},
			},
		},
	}
}

func TestGitHubRepoSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("github-repo")
	if !ok {
		t.Fatal("Suite github-repo not found")
	}

	runner.RunSuite(t, suite)
}
