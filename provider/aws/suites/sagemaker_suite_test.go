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
	suites.Register(&AWSSageMakerSuite{})
}

// AWSSageMakerSuite tests SageMaker notebook policies against fixtures.
type AWSSageMakerSuite struct{}

func (s *AWSSageMakerSuite) Name() string     { return "aws-sagemaker" }
func (s *AWSSageMakerSuite) Provider() string { return "aws" }

func (s *AWSSageMakerSuite) Setup(_ *testing.T) error    { return nil }
func (s *AWSSageMakerSuite) Teardown(_ *testing.T) error { return nil }

func (s *AWSSageMakerSuite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure SageMaker notebook passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_sagemaker_notebook",
			Fixtures:     []core.Resource{awssuites.SecureSageMakerNotebook.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-sagemaker-notebook-encryption": {Pass: true},
				"aws-sagemaker-notebook-not-public": {Pass: true},
			},
		},
		{
			Name:         "insecure SageMaker notebook fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_sagemaker_notebook",
			Fixtures:     []core.Resource{awssuites.InsecureSageMakerNotebook.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-sagemaker-notebook-encryption": {Pass: false},
				"aws-sagemaker-notebook-not-public": {Pass: false},
			},
		},
	}
}

func TestAWSSageMakerSuite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-sagemaker")
	if !ok {
		t.Fatal("Suite aws-sagemaker not found")
	}

	runner.RunSuite(t, suite)
}
