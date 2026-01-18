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
	suites.Register(&AWSVPCSuite{})
}

// AWSVPCSuite tests VPC policies against fixtures.
type AWSVPCSuite struct{}

func (s *AWSVPCSuite) Name() string     { return "aws-vpc" }
func (s *AWSVPCSuite) Provider() string { return "aws" }

func (s *AWSVPCSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSVPCSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSVPCSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure VPC passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_vpc",
			Fixtures:     []core.Resource{awssuites.SecureVPC.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-vpc-flow-logs-enabled":    {Pass: true},
				"aws-vpc-no-default-vpc":       {Pass: true},
				"aws-vpc-endpoint-enabled":     {Pass: true},
				"aws-vpc-block-public-access":  {Pass: true},
			},
		},
		{
			Name:         "insecure VPC (default) fails checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_vpc",
			Fixtures:     []core.Resource{awssuites.InsecureVPC.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-vpc-flow-logs-enabled":    {Pass: false},
				"aws-vpc-no-default-vpc":       {Pass: false},
				"aws-vpc-endpoint-enabled":     {Pass: false},
				"aws-vpc-block-public-access":  {Pass: false},
			},
		},
	}
}

func TestAWSVPCSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-vpc")
	if !ok {
		t.Fatal("Suite aws-vpc not found")
	}

	runner.RunSuite(t, suite)
}
