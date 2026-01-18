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
	suites.Register(&GitHubBranchSuite{})
}

// GitHubBranchSuite tests GitHub branch protection policies against fixtures.
type GitHubBranchSuite struct{}

func (s *GitHubBranchSuite) Name() string     { return "github-branch" }
func (s *GitHubBranchSuite) Provider() string { return "github" }

func (s *GitHubBranchSuite) Setup(_ *testing.T) error    { return nil }
func (s *GitHubBranchSuite) Teardown(_ *testing.T) error { return nil }

func (s *GitHubBranchSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure default branch passes all checks",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_branch",
			Fixtures:     []core.Resource{ghsuites.SecureDefaultBranch.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-default-branch-protection":    {Pass: true},
				"kopexa-github-repository-security-prevent-force-pushes-default": {Pass: true},
				"kopexa-github-repository-security-conversation-resolution":      {Pass: true},
				"kopexa-github-repository-security-status-checks":                {Pass: true},
				"kopexa-github-repository-security-signed-commits":               {Pass: true},
				"kopexa-github-repository-security-enforce-admins":               {Pass: true},
			},
		},
		{
			Name:         "insecure default branch fails protection check",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_branch",
			Fixtures:     []core.Resource{ghsuites.InsecureDefaultBranch.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-default-branch-protection": {Pass: false},
				// Note: These checks pass because their policy queries use short-circuit evaluation.
				// The queries check "!protected || <feature_enabled>" - when protected is false,
				// the entire expression evaluates to true without checking the feature flag.
				// This is intentional: unprotected branches can't have protection features enabled.
				"kopexa-github-repository-security-prevent-force-pushes-default": {Pass: true},
				"kopexa-github-repository-security-conversation-resolution":      {Pass: true},
				"kopexa-github-repository-security-status-checks":                {Pass: true},
				"kopexa-github-repository-security-signed-commits":               {Pass: true},
				"kopexa-github-repository-security-enforce-admins":               {Pass: true},
			},
		},
		{
			Name:         "partially protected default branch fails security feature checks",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_branch",
			Fixtures:     []core.Resource{ghsuites.PartiallyProtectedDefaultBranch.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-default-branch-protection":    {Pass: true},
				"kopexa-github-repository-security-prevent-force-pushes-default": {Pass: false},
				"kopexa-github-repository-security-conversation-resolution":      {Pass: false},
				"kopexa-github-repository-security-status-checks":                {Pass: false},
				"kopexa-github-repository-security-signed-commits":               {Pass: false},
				"kopexa-github-repository-security-enforce-admins":               {Pass: false},
			},
		},
		{
			Name:         "secure release branch passes release checks",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_branch",
			Fixtures:     []core.Resource{ghsuites.SecureReleaseBranch.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-release-branch-protection":    {Pass: true},
				"kopexa-github-repository-security-prevent-force-pushes-release": {Pass: true},
			},
		},
		{
			Name:         "insecure release branch fails release checks",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_branch",
			Fixtures:     []core.Resource{ghsuites.InsecureReleaseBranch.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-release-branch-protection": {Pass: false},
				// Force push check passes due to short-circuit evaluation in policy query:
				// "!protected || allow_force_pushes == false" evaluates to true when not protected.
				"kopexa-github-repository-security-prevent-force-pushes-release": {Pass: true},
			},
		},
		{
			Name:         "feature branch passes all checks (not default or release)",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_branch",
			Fixtures:     []core.Resource{ghsuites.NonDefaultNonReleaseBranch.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-repository-security-default-branch-protection":    {Pass: true}, // Not default
				"kopexa-github-repository-security-release-branch-protection":    {Pass: true}, // Not release
				"kopexa-github-repository-security-prevent-force-pushes-default": {Pass: true}, // Not default
				"kopexa-github-repository-security-prevent-force-pushes-release": {Pass: true}, // Not release
			},
		},
	}
}

func TestGitHubBranchSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("github-branch")
	if !ok {
		t.Fatal("Suite github-branch not found")
	}

	runner.RunSuite(t, suite)
}
