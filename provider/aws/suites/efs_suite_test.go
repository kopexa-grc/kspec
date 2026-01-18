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
	suites.Register(&AWSEFSSuite{})
}

// AWSEFSSuite tests EFS filesystem policies against fixtures.
type AWSEFSSuite struct{}

func (s *AWSEFSSuite) Name() string     { return "aws-efs" }
func (s *AWSEFSSuite) Provider() string { return "aws" }

func (s *AWSEFSSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSEFSSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSEFSSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure EFS filesystem passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_efs_filesystem",
			Fixtures:     []core.Resource{awssuites.SecureEFSFilesystem.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-efs-encryption-enabled":     {Pass: true},
				"aws-efs-encryption-in-transit":  {Pass: true},
			},
		},
		{
			Name:         "insecure EFS filesystem fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_efs_filesystem",
			Fixtures:     []core.Resource{awssuites.InsecureEFSFilesystem.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-efs-encryption-enabled":     {Pass: false},
				"aws-efs-encryption-in-transit":  {Pass: false},
			},
		},
	}
}

func TestAWSEFSSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-efs")
	if !ok {
		t.Fatal("Suite aws-efs not found")
	}

	runner.RunSuite(t, suite)
}
