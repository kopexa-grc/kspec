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
	suites.Register(&AWSAPIGatewayStageSuite{})
	suites.Register(&AWSAPIGatewayMethodSuite{})
}

// AWSAPIGatewayStageSuite tests API Gateway stage policies against fixtures.
type AWSAPIGatewayStageSuite struct{}

func (s *AWSAPIGatewayStageSuite) Name() string     { return "aws-apigateway-stage" }
func (s *AWSAPIGatewayStageSuite) Provider() string { return "aws" }

func (s *AWSAPIGatewayStageSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSAPIGatewayStageSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSAPIGatewayStageSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure API Gateway stage passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_apigateway_stage",
			Fixtures:     []core.Resource{awssuites.SecureAPIGatewayStage.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-apigateway-logging-enabled":  {Pass: true},
				"aws-apigateway-waf-enabled":      {Pass: true},
				"aws-apigateway-cache-encrypted":  {Pass: true},
				"aws-apigateway-tls-version":      {Pass: true},
				"aws-apigateway-xray-enabled":     {Pass: true},
			},
		},
		{
			Name:         "insecure API Gateway stage fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_apigateway_stage",
			Fixtures:     []core.Resource{awssuites.InsecureAPIGatewayStage.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-apigateway-logging-enabled":  {Pass: false},
				"aws-apigateway-waf-enabled":      {Pass: false},
				"aws-apigateway-cache-encrypted":  {Pass: false},
				"aws-apigateway-tls-version":      {Pass: false},
				"aws-apigateway-xray-enabled":     {Pass: false},
			},
		},
	}
}

// AWSAPIGatewayMethodSuite tests API Gateway method policies against fixtures.
type AWSAPIGatewayMethodSuite struct{}

func (s *AWSAPIGatewayMethodSuite) Name() string     { return "aws-apigateway-method" }
func (s *AWSAPIGatewayMethodSuite) Provider() string { return "aws" }

func (s *AWSAPIGatewayMethodSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSAPIGatewayMethodSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSAPIGatewayMethodSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure API Gateway method passes authorization check",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_apigateway_method",
			Fixtures:     []core.Resource{awssuites.SecureAPIGatewayMethod.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-apigateway-require-authorization": {Pass: true},
			},
		},
		{
			Name:         "insecure API Gateway method fails authorization check",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_apigateway_method",
			Fixtures:     []core.Resource{awssuites.InsecureAPIGatewayMethod.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-apigateway-require-authorization": {Pass: false},
			},
		},
	}
}

func TestAWSAPIGatewayStageSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-apigateway-stage")
	if !ok {
		t.Fatal("Suite aws-apigateway-stage not found")
	}

	runner.RunSuite(t, suite)
}

func TestAWSAPIGatewayMethodSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-apigateway-method")
	if !ok {
		t.Fatal("Suite aws-apigateway-method not found")
	}

	runner.RunSuite(t, suite)
}
