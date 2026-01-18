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
	suites.Register(&AWSCloudTrailSuite{})
}

// AWSCloudTrailSuite tests CloudTrail policies against fixtures.
type AWSCloudTrailSuite struct{}

func (s *AWSCloudTrailSuite) Name() string     { return "aws-cloudtrail" }
func (s *AWSCloudTrailSuite) Provider() string { return "aws" }

func (s *AWSCloudTrailSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSCloudTrailSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSCloudTrailSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure CloudTrail passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_cloudtrail",
			Fixtures:     []core.Resource{awssuites.SecureCloudTrail.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-cloudtrail-enabled":            {Pass: true},
				"aws-cloudtrail-multi-region":       {Pass: true},
				"aws-cloudtrail-log-validation":     {Pass: true},
				"aws-cloudtrail-encryption-enabled": {Pass: true},
				"aws-cloudtrail-cloudwatch-logs":    {Pass: true},
			},
		},
		{
			Name:         "insecure CloudTrail fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_cloudtrail",
			Fixtures:     []core.Resource{awssuites.InsecureCloudTrail.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-cloudtrail-enabled":            {Pass: false},
				"aws-cloudtrail-multi-region":       {Pass: false},
				"aws-cloudtrail-log-validation":     {Pass: false},
				"aws-cloudtrail-encryption-enabled": {Pass: false},
				"aws-cloudtrail-cloudwatch-logs":    {Pass: false},
			},
		},
	}
}

func TestAWSCloudTrailSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-cloudtrail")
	if !ok {
		t.Fatal("Suite aws-cloudtrail not found")
	}

	runner.RunSuite(t, suite)
}
