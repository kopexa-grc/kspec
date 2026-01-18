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
	suites.Register(&AWSEC2InstanceSuite{})
	suites.Register(&AWSEC2SecurityGroupSuite{})
	suites.Register(&AWSEC2VolumeSuite{})
}

// AWSEC2InstanceSuite tests EC2 instance policies against fixtures.
type AWSEC2InstanceSuite struct{}

func (s *AWSEC2InstanceSuite) Name() string     { return "aws-ec2-instance" }
func (s *AWSEC2InstanceSuite) Provider() string { return "aws" }

func (s *AWSEC2InstanceSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSEC2InstanceSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSEC2InstanceSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure EC2 instance passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_instance",
			Fixtures:     []core.Resource{awssuites.SecureEC2Instance.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-imdsv2-required":     {Pass: true},
				"aws-ec2-no-public-ip":        {Pass: true},
				"aws-ec2-detailed-monitoring": {Pass: true},
			},
		},
		{
			Name:         "insecure EC2 instance fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_instance",
			Fixtures:     []core.Resource{awssuites.InsecureEC2Instance.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-imdsv2-required":     {Pass: false},
				"aws-ec2-no-public-ip":        {Pass: false},
				"aws-ec2-detailed-monitoring": {Pass: false},
			},
		},
	}
}

// AWSEC2SecurityGroupSuite tests EC2 security group policies against fixtures.
type AWSEC2SecurityGroupSuite struct{}

func (s *AWSEC2SecurityGroupSuite) Name() string     { return "aws-ec2-securitygroup" }
func (s *AWSEC2SecurityGroupSuite) Provider() string { return "aws" }

func (s *AWSEC2SecurityGroupSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSEC2SecurityGroupSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSEC2SecurityGroupSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure security group passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_securitygroup",
			Fixtures:     []core.Resource{awssuites.SecureSecurityGroup.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-security-group-no-ssh-from-internet":     {Pass: true},
				"aws-ec2-security-group-no-rdp-from-internet":     {Pass: true},
				"aws-ec2-security-group-no-unrestricted-ingress": {Pass: true},
			},
		},
		{
			Name:         "insecure security group fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_securitygroup",
			Fixtures:     []core.Resource{awssuites.InsecureSecurityGroup.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-security-group-no-ssh-from-internet":     {Pass: false},
				"aws-ec2-security-group-no-rdp-from-internet":     {Pass: false},
				"aws-ec2-security-group-no-unrestricted-ingress": {Pass: false},
			},
		},
	}
}

// AWSEC2VolumeSuite tests EBS volume policies against fixtures.
type AWSEC2VolumeSuite struct{}

func (s *AWSEC2VolumeSuite) Name() string     { return "aws-ec2-volume" }
func (s *AWSEC2VolumeSuite) Provider() string { return "aws" }

func (s *AWSEC2VolumeSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSEC2VolumeSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSEC2VolumeSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "encrypted EBS volume passes",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_volume",
			Fixtures:     []core.Resource{awssuites.SecureEBSVolume.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-ebs-encryption-enabled": {Pass: true},
			},
		},
		{
			Name:         "unencrypted EBS volume fails",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_ec2_volume",
			Fixtures:     []core.Resource{awssuites.InsecureEBSVolume.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-ec2-ebs-encryption-enabled": {Pass: false},
			},
		},
	}
}

func TestAWSEC2InstanceSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-ec2-instance")
	if !ok {
		t.Fatal("Suite aws-ec2-instance not found")
	}

	runner.RunSuite(t, suite)
}

func TestAWSEC2SecurityGroupSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-ec2-securitygroup")
	if !ok {
		t.Fatal("Suite aws-ec2-securitygroup not found")
	}

	runner.RunSuite(t, suite)
}

func TestAWSEC2VolumeSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-ec2-volume")
	if !ok {
		t.Fatal("Suite aws-ec2-volume not found")
	}

	runner.RunSuite(t, suite)
}
