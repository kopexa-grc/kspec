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
	suites.Register(&AWSKMSSuite{})
}

// AWSKMSSuite tests KMS key policies against fixtures.
type AWSKMSSuite struct{}

func (s *AWSKMSSuite) Name() string     { return "aws-kms" }
func (s *AWSKMSSuite) Provider() string { return "aws" }

func (s *AWSKMSSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSKMSSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSKMSSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure KMS key passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_kms_key",
			Fixtures:     []core.Resource{awssuites.SecureKMSKey.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-kms-key-rotation-enabled": {Pass: true},
				"aws-kms-key-not-public":       {Pass: true},
			},
		},
		{
			Name:         "insecure KMS key fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_kms_key",
			Fixtures:     []core.Resource{awssuites.InsecureKMSKey.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-kms-key-rotation-enabled": {Pass: false},
				"aws-kms-key-not-public":       {Pass: false},
			},
		},
		{
			Name:         "AWS-managed KMS key passes rotation check",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_kms_key",
			Fixtures: []core.Resource{
				awssuites.SecureKMSKey.Clone().
					Set("is_aws_managed", true).
					Set("key_rotation_enabled", false).
					Build(),
			},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-kms-key-rotation-enabled": {Pass: true},
				"aws-kms-key-not-public":       {Pass: true},
			},
		},
	}
}

func TestAWSKMSSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-kms")
	if !ok {
		t.Fatal("Suite aws-kms not found")
	}

	runner.RunSuite(t, suite)
}
