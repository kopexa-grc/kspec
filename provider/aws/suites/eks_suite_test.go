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
	suites.Register(&AWSEKSSuite{})
}

// AWSEKSSuite tests EKS cluster policies against fixtures.
type AWSEKSSuite struct{}

func (s *AWSEKSSuite) Name() string     { return "aws-eks" }
func (s *AWSEKSSuite) Provider() string { return "aws" }

func (s *AWSEKSSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSEKSSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSEKSSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure EKS cluster passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_eks_cluster",
			Fixtures:     []core.Resource{awssuites.SecureEKSCluster.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-eks-secrets-encryption":  {Pass: true},
				"aws-eks-logging-enabled":     {Pass: true},
				"aws-eks-no-public-endpoint":  {Pass: true},
			},
		},
		{
			Name:         "insecure EKS cluster fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_eks_cluster",
			Fixtures:     []core.Resource{awssuites.InsecureEKSCluster.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-eks-secrets-encryption":  {Pass: false},
				"aws-eks-logging-enabled":     {Pass: false},
				"aws-eks-no-public-endpoint":  {Pass: false},
			},
		},
	}
}

func TestAWSEKSSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-eks")
	if !ok {
		t.Fatal("Suite aws-eks not found")
	}

	runner.RunSuite(t, suite)
}
