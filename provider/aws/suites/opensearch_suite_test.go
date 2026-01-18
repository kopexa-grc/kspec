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
	suites.Register(&AWSOpenSearchSuite{})
}

// AWSOpenSearchSuite tests OpenSearch domain policies against fixtures.
type AWSOpenSearchSuite struct{}

func (s *AWSOpenSearchSuite) Name() string     { return "aws-opensearch" }
func (s *AWSOpenSearchSuite) Provider() string { return "aws" }

func (s *AWSOpenSearchSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSOpenSearchSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSOpenSearchSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure OpenSearch domain passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_opensearch_domain",
			Fixtures:     []core.Resource{awssuites.SecureOpenSearchDomain.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-opensearch-encryption-at-rest":   {Pass: true},
				"aws-opensearch-encryption-in-transit": {Pass: true},
				"aws-opensearch-not-public":           {Pass: true},
			},
		},
		{
			Name:         "insecure OpenSearch domain fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_opensearch_domain",
			Fixtures:     []core.Resource{awssuites.InsecureOpenSearchDomain.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-opensearch-encryption-at-rest":   {Pass: false},
				"aws-opensearch-encryption-in-transit": {Pass: false},
				"aws-opensearch-not-public":           {Pass: false},
			},
		},
	}
}

func TestAWSOpenSearchSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-opensearch")
	if !ok {
		t.Fatal("Suite aws-opensearch not found")
	}

	runner.RunSuite(t, suite)
}
