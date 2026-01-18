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
	suites.Register(&GitHubOrganizationSuite{})
}

// GitHubOrganizationSuite tests GitHub organization policies against fixtures.
type GitHubOrganizationSuite struct{}

func (s *GitHubOrganizationSuite) Name() string     { return "github-organization" }
func (s *GitHubOrganizationSuite) Provider() string { return "github" }

func (s *GitHubOrganizationSuite) Setup(_ *testing.T) error    { return nil }
func (s *GitHubOrganizationSuite) Teardown(_ *testing.T) error { return nil }

func (s *GitHubOrganizationSuite) TestCases() []suites.TestCase {
	// Organization checks require an Asset with type "github-org" to match the group filter
	orgAsset := core.Asset{Type: "github-org"}

	return []suites.TestCase{
		{
			Name:         "secure organization passes all checks",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_organization",
			Fixtures:     []core.Resource{ghsuites.SecureOrganization.Build()},
			Asset:        orgAsset,
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-organization-security-two-factor-auth":          {Pass: true},
				"kopexa-github-organization-security-verified-domain":          {Pass: true},
				"kopexa-github-organization-security-default-permission-level": {Pass: true},
			},
		},
		{
			Name:         "insecure organization fails all checks",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_organization",
			Fixtures:     []core.Resource{ghsuites.InsecureOrganization.Build()},
			Asset:        orgAsset,
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-organization-security-two-factor-auth":          {Pass: false},
				"kopexa-github-organization-security-verified-domain":          {Pass: false},
				"kopexa-github-organization-security-default-permission-level": {Pass: false},
			},
		},
		{
			Name:         "organization with only 2FA enabled",
			PolicyFile:   "github-security.yml",
			ResourceType: "github_organization",
			Fixtures: []core.Resource{
				ghsuites.InsecureOrganization.Clone().
					Set("two_factor_requirement_enabled", true).
					Build(),
			},
			Asset: orgAsset,
			ExpectedResults: map[string]suites.ExpectedResult{
				"kopexa-github-organization-security-two-factor-auth":          {Pass: true},
				"kopexa-github-organization-security-verified-domain":          {Pass: false},
				"kopexa-github-organization-security-default-permission-level": {Pass: false},
			},
		},
	}
}

func TestGitHubOrganizationSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("github-organization")
	if !ok {
		t.Fatal("Suite github-organization not found")
	}

	runner.RunSuite(t, suite)
}
