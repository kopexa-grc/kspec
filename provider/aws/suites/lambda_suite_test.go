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
	suites.Register(&AWSLambdaSuite{})
}

// AWSLambdaSuite tests Lambda function policies against fixtures.
type AWSLambdaSuite struct{}

func (s *AWSLambdaSuite) Name() string     { return "aws-lambda" }
func (s *AWSLambdaSuite) Provider() string { return "aws" }

func (s *AWSLambdaSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSLambdaSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSLambdaSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure Lambda function passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_lambda_function",
			Fixtures:     []core.Resource{awssuites.SecureLambdaFunction.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-lambda-in-vpc":                 {Pass: true},
				"aws-lambda-no-public-url":          {Pass: true},
				"aws-lambda-no-deprecated-runtime":  {Pass: true},
				"aws-lambda-dlq-configured":         {Pass: true},
				"aws-lambda-concurrency-configured": {Pass: true},
			},
		},
		{
			Name:         "insecure Lambda function fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_lambda_function",
			Fixtures:     []core.Resource{awssuites.InsecureLambdaFunction.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-lambda-in-vpc":                 {Pass: false},
				"aws-lambda-no-public-url":          {Pass: false},
				"aws-lambda-no-deprecated-runtime":  {Pass: false},
				"aws-lambda-dlq-configured":         {Pass: false},
				"aws-lambda-concurrency-configured": {Pass: false},
			},
		},
	}
}

func TestAWSLambdaSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-lambda")
	if !ok {
		t.Fatal("Suite aws-lambda not found")
	}

	runner.RunSuite(t, suite)
}
