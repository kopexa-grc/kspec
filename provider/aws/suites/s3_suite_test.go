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
	suites.Register(&AWSS3Suite{})
}

// AWSS3Suite tests S3 bucket policies against fixtures.
type AWSS3Suite struct{}

func (s *AWSS3Suite) Name() string     { return "aws-s3" }
func (s *AWSS3Suite) Provider() string { return "aws" }

func (s *AWSS3Suite) Setup(_ *testing.T) error    { return nil }
func (s *AWSS3Suite) Teardown(_ *testing.T) error { return nil }

func (s *AWSS3Suite) TestCases() []suites.TestCase {
	return []suites.TestCase{
		{
			Name:         "secure S3 bucket passes all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_s3_bucket",
			Fixtures:     []core.Resource{awssuites.SecureS3Bucket.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-s3-bucket-no-public-access":    {Pass: true},
				"aws-s3-bucket-encryption-enabled":  {Pass: true},
				"aws-s3-bucket-versioning-enabled":  {Pass: true},
				"aws-s3-bucket-logging-enabled":     {Pass: true},
				"aws-s3-bucket-ssl-requests-only":   {Pass: true},
				"aws-s3-bucket-public-access-block": {Pass: true},
				"aws-s3-bucket-no-public-acl":       {Pass: true},
			},
		},
		{
			Name:         "insecure S3 bucket fails all checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_s3_bucket",
			Fixtures:     []core.Resource{awssuites.InsecureS3Bucket.Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-s3-bucket-no-public-access":    {Pass: false},
				"aws-s3-bucket-encryption-enabled":  {Pass: false},
				"aws-s3-bucket-versioning-enabled":  {Pass: false},
				"aws-s3-bucket-logging-enabled":     {Pass: false},
				"aws-s3-bucket-ssl-requests-only":   {Pass: false},
				"aws-s3-bucket-public-access-block": {Pass: false},
				"aws-s3-bucket-no-public-acl":       {Pass: false},
			},
		},
		{
			Name:         "S3 bucket without versioning fails versioning check only",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_s3_bucket",
			Fixtures:     []core.Resource{awssuites.SecureS3Bucket.With("has_versioning", false).Build()},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-s3-bucket-no-public-access":    {Pass: true},
				"aws-s3-bucket-encryption-enabled":  {Pass: true},
				"aws-s3-bucket-versioning-enabled":  {Pass: false},
				"aws-s3-bucket-logging-enabled":     {Pass: true},
				"aws-s3-bucket-ssl-requests-only":   {Pass: true},
				"aws-s3-bucket-public-access-block": {Pass: true},
				"aws-s3-bucket-no-public-acl":       {Pass: true},
			},
		},
		{
			Name:         "public S3 bucket fails public access checks",
			PolicyFile:   "aws-security.yml",
			ResourceType: "aws_s3_bucket",
			Fixtures: []core.Resource{
				awssuites.SecureS3Bucket.Clone().
					Set("is_public", true).
					Set("has_public_acl", true).
					Set("public_access_block_enabled", false).
					Build(),
			},
			ExpectedResults: map[string]suites.ExpectedResult{
				"aws-s3-bucket-no-public-access":    {Pass: false},
				"aws-s3-bucket-encryption-enabled":  {Pass: true},
				"aws-s3-bucket-versioning-enabled":  {Pass: true},
				"aws-s3-bucket-logging-enabled":     {Pass: true},
				"aws-s3-bucket-ssl-requests-only":   {Pass: true},
				"aws-s3-bucket-public-access-block": {Pass: false},
				"aws-s3-bucket-no-public-acl":       {Pass: false},
			},
		},
	}
}

func TestAWSS3Suite(t *testing.T) {
	runner, err := suites.NewRunner("../../../policies")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	suite, ok := suites.Get("aws-s3")
	if !ok {
		t.Fatal("Suite aws-s3 not found")
	}

	runner.RunSuite(t, suite)
}
