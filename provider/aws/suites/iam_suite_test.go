// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

//go:build fast

package suites_test

import (
	"testing"

	"github.com/kopexa-grc/kspec/core"
	awssuites "github.com/kopexa-grc/kspec/provider/aws/suites"
	"github.com/kopexa-grc/kspec/testing/suites"
)

func init() {
	suites.Register(&AWSIAMUserSuite{})
	suites.Register(&AWSIAMPolicySuite{})
	suites.Register(&AWSIAMPasswordPolicySuite{})
	suites.Register(&AWSIAMGroupSuite{})
}

// AWSIAMUserSuite tests IAM user policies against fixtures.
type AWSIAMUserSuite struct{}

func (s *AWSIAMUserSuite) Name() string     { return "aws-iam-user" }
func (s *AWSIAMUserSuite) Provider() string { return "aws" }

func (s *AWSIAMUserSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSIAMUserSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSIAMUserSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure IAM user passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_user",
			Fixtures:     []core.Resource{awssuites.SecureIAMUser.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-users-mfa-enabled":           {Pass: true},
				"aws-iam-users-no-inline-policies":    {Pass: true},
				"aws-iam-access-keys-rotated":         {Pass: true},
				"aws-iam-unused-credentials-disabled": {Pass: true},
				"aws-iam-user-single-access-key":      {Pass: true},
			},
		},
		{
			Name:         "insecure IAM user fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_user",
			Fixtures:     []core.Resource{awssuites.InsecureIAMUser.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-users-mfa-enabled":           {Pass: false},
				"aws-iam-users-no-inline-policies":    {Pass: false},
				"aws-iam-access-keys-rotated":         {Pass: false},
				"aws-iam-unused-credentials-disabled": {Pass: false},
				"aws-iam-user-single-access-key":      {Pass: false},
			},
		},
		{
			Name:         "IAM user without console access passes MFA check",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_user",
			Fixtures:     []core.Resource{awssuites.SecureIAMUser.With("has_console_access", false).With("has_mfa", false).Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-users-mfa-enabled": {Pass: true}, // MFA not required if no console access
			},
		},
		{
			Name:         "IAM user without access keys passes rotation check",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_user",
			Fixtures:     []core.Resource{awssuites.SecureIAMUser.With("has_access_keys", false).Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-access-keys-rotated": {Pass: true}, // No keys to rotate
			},
		},
	}
}

// AWSIAMPolicySuite tests IAM policy checks against fixtures.
type AWSIAMPolicySuite struct{}

func (s *AWSIAMPolicySuite) Name() string     { return "aws-iam-policy" }
func (s *AWSIAMPolicySuite) Provider() string { return "aws" }

func (s *AWSIAMPolicySuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSIAMPolicySuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSIAMPolicySuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "non-admin IAM policy passes",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_policy",
			Fixtures:     []core.Resource{awssuites.SecureIAMPolicy.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-no-admin-privileges":   {Pass: true},
				"aws-iam-no-wildcard-policies": {Pass: true},
			},
		},
		{
			Name:         "admin IAM policy fails",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_policy",
			Fixtures:     []core.Resource{awssuites.InsecureIAMPolicy.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-no-admin-privileges":   {Pass: false},
				"aws-iam-no-wildcard-policies": {Pass: false},
			},
		},
	}
}

// AWSIAMPasswordPolicySuite tests IAM password policy checks against fixtures.
type AWSIAMPasswordPolicySuite struct{}

func (s *AWSIAMPasswordPolicySuite) Name() string     { return "aws-iam-password-policy" }
func (s *AWSIAMPasswordPolicySuite) Provider() string { return "aws" }

func (s *AWSIAMPasswordPolicySuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSIAMPasswordPolicySuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSIAMPasswordPolicySuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure password policy passes",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_password_policy",
			Fixtures:     []core.Resource{awssuites.SecurePasswordPolicy.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-password-policy": {Pass: true},
			},
		},
		{
			Name:         "insecure password policy fails",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_password_policy",
			Fixtures:     []core.Resource{awssuites.InsecurePasswordPolicy.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-password-policy": {Pass: false},
			},
		},
	}
}

// AWSIAMGroupSuite tests IAM group checks against fixtures.
type AWSIAMGroupSuite struct{}

func (s *AWSIAMGroupSuite) Name() string     { return "aws-iam-group" }
func (s *AWSIAMGroupSuite) Provider() string { return "aws" }

func (s *AWSIAMGroupSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSIAMGroupSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSIAMGroupSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "IAM group with users passes",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_group",
			Fixtures:     []core.Resource{awssuites.IAMGroupWithUsers.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-group-has-users": {Pass: true},
			},
		},
		{
			Name:         "empty IAM group fails",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_iam_group",
			Fixtures:     []core.Resource{awssuites.IAMGroupEmpty.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-iam-group-has-users": {Pass: false},
			},
		},
	}
}

func TestAWSIAMUserSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-iam-user")
	if !ok {
		t.Fatal("Suite aws-iam-user not found")
	}

	runner.RunSuite(t, suite)
}

func TestAWSIAMPolicySuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-iam-policy")
	if !ok {
		t.Fatal("Suite aws-iam-policy not found")
	}

	runner.RunSuite(t, suite)
}

func TestAWSIAMPasswordPolicySuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-iam-password-policy")
	if !ok {
		t.Fatal("Suite aws-iam-password-policy not found")
	}

	runner.RunSuite(t, suite)
}

func TestAWSIAMGroupSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-iam-group")
	if !ok {
		t.Fatal("Suite aws-iam-group not found")
	}

	runner.RunSuite(t, suite)
}
