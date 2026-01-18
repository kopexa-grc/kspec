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
	suites.Register(&AWSSnapshotSuite{})
}

// AWSSnapshotSuite tests EBS snapshot policies against fixtures.
type AWSSnapshotSuite struct{}

func (s *AWSSnapshotSuite) Name() string     { return "aws-snapshot" }
func (s *AWSSnapshotSuite) Provider() string { return "aws" }

func (s *AWSSnapshotSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSSnapshotSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSSnapshotSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure EBS snapshot passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_snapshot",
			Fixtures:     []core.Resource{awssuites.SecureEBSSnapshot.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-ebs-snapshot-not-public": {Pass: true},
				"aws-ec2-ebs-snapshot-encrypted":  {Pass: true},
			},
		},
		{
			Name:         "insecure EBS snapshot fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_snapshot",
			Fixtures:     []core.Resource{awssuites.InsecureEBSSnapshot.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-ebs-snapshot-not-public": {Pass: false},
				"aws-ec2-ebs-snapshot-encrypted":  {Pass: false},
			},
		},
	}
}

func TestAWSSnapshotSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-snapshot")
	if !ok {
		t.Fatal("Suite aws-snapshot not found")
	}

	runner.RunSuite(t, suite)
}
