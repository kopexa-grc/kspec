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
	suites.Register(&AWSRDSInstanceSuite{})
	suites.Register(&AWSRDSClusterSuite{})
}

// AWSRDSInstanceSuite tests RDS instance policies against fixtures.
type AWSRDSInstanceSuite struct{}

func (s *AWSRDSInstanceSuite) Name() string     { return "aws-rds-instance" }
func (s *AWSRDSInstanceSuite) Provider() string { return "aws" }

func (s *AWSRDSInstanceSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSRDSInstanceSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSRDSInstanceSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure RDS instance passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_rds_instance",
			Fixtures:     []core.Resource{awssuites.SecureRDSInstance.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-rds-encryption-enabled":      {Pass: true},
				"aws-rds-not-public":              {Pass: true},
				"aws-rds-multi-az-enabled":        {Pass: true},
				"aws-rds-deletion-protection":     {Pass: true},
				"aws-rds-iam-authentication":      {Pass: true},
				"aws-rds-backup-enabled":          {Pass: true},
				"aws-rds-auto-minor-upgrade":      {Pass: true},
				"aws-rds-no-pending-maintenance":  {Pass: true},
			},
		},
		{
			Name:         "insecure RDS instance fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_rds_instance",
			Fixtures:     []core.Resource{awssuites.InsecureRDSInstance.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-rds-encryption-enabled":      {Pass: false},
				"aws-rds-not-public":              {Pass: false},
				"aws-rds-multi-az-enabled":        {Pass: false},
				"aws-rds-deletion-protection":     {Pass: false},
				"aws-rds-iam-authentication":      {Pass: false},
				"aws-rds-backup-enabled":          {Pass: false},
				"aws-rds-auto-minor-upgrade":      {Pass: false},
				"aws-rds-no-pending-maintenance":  {Pass: false},
			},
		},
	}
}

// AWSRDSClusterSuite tests RDS Aurora cluster policies against fixtures.
type AWSRDSClusterSuite struct{}

func (s *AWSRDSClusterSuite) Name() string     { return "aws-rds-cluster" }
func (s *AWSRDSClusterSuite) Provider() string { return "aws" }

func (s *AWSRDSClusterSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSRDSClusterSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSRDSClusterSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure RDS cluster passes SSL check",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_rds_cluster",
			Fixtures:     []core.Resource{awssuites.SecureRDSCluster.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-rds-cluster-ssl-enforced": {Pass: true},
			},
		},
		{
			Name:         "insecure RDS cluster fails SSL check",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_rds_cluster",
			Fixtures:     []core.Resource{awssuites.InsecureRDSCluster.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-rds-cluster-ssl-enforced": {Pass: false},
			},
		},
	}
}

func TestAWSRDSInstanceSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-rds-instance")
	if !ok {
		t.Fatal("Suite aws-rds-instance not found")
	}

	runner.RunSuite(t, suite)
}

func TestAWSRDSClusterSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-rds-cluster")
	if !ok {
		t.Fatal("Suite aws-rds-cluster not found")
	}

	runner.RunSuite(t, suite)
}
