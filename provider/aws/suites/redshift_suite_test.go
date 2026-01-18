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
	suites.Register(&AWSRedshiftSuite{})
}

// AWSRedshiftSuite tests Redshift cluster policies against fixtures.
type AWSRedshiftSuite struct{}

func (s *AWSRedshiftSuite) Name() string     { return "aws-redshift" }
func (s *AWSRedshiftSuite) Provider() string { return "aws" }

func (s *AWSRedshiftSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSRedshiftSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSRedshiftSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure Redshift cluster passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_redshift_cluster",
			Fixtures:     []core.Resource{awssuites.SecureRedshiftCluster.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-redshift-not-public":        {Pass: true},
				"aws-redshift-encryption-enabled": {Pass: true},
				"aws-redshift-audit-logging":      {Pass: true},
			},
		},
		{
			Name:         "insecure Redshift cluster fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_redshift_cluster",
			Fixtures:     []core.Resource{awssuites.InsecureRedshiftCluster.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-redshift-not-public":        {Pass: false},
				"aws-redshift-encryption-enabled": {Pass: false},
				"aws-redshift-audit-logging":      {Pass: false},
			},
		},
	}
}

func TestAWSRedshiftSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-redshift")
	if !ok {
		t.Fatal("Suite aws-redshift not found")
	}

	runner.RunSuite(t, suite)
}
