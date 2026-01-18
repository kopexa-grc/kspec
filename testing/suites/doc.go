// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package suites provides an integration testing framework for kspec policies.
//
// # Overview
//
// This package enables testing policies against provider resources using
// predefined fixtures. It supports two categories of tests:
//
//   - Fast tests (//go:build fast): Use mocked resources, run on every PR
//   - Integration tests (//go:build integration): Use containers (LocalStack, etc.), run nightly
//
// # Usage
//
// Create a suite by implementing the Suite interface:
//
//	type MyProviderSuite struct{}
//
//	func (s *MyProviderSuite) Name() string     { return "myprovider-resource" }
//	func (s *MyProviderSuite) Provider() string { return "myprovider" }
//	func (s *MyProviderSuite) Setup(t *testing.T) error    { return nil }
//	func (s *MyProviderSuite) Teardown(t *testing.T) error { return nil }
//
//	func (s *MyProviderSuite) TestCases() []suites.TestCase {
//	    return []suites.TestCase{
//	        {
//	            Name:         "resource should pass security check",
//	            PolicyFile:   "myprovider/security.policy.yaml",
//	            ResourceType: "myprovider_resource",
//	            Fixtures:     []core.Resource{SecureResourceFixture.Build()},
//	            ExpectedResults: map[string]suites.ExpectedResult{
//	                "check-id": {Pass: true},
//	            },
//	        },
//	    }
//	}
//
// Register the suite in an init function:
//
//	func init() {
//	    suites.Register(&MyProviderSuite{})
//	}
//
// # Fixtures
//
// Use FixtureBuilder to create resource fixtures:
//
//	var SecureResource = suites.NewFixture().
//	    Set("id", "secure-resource").
//	    Set("has_encryption", true).
//	    Set("is_public", false)
//
//	// Clone and modify for variations
//	var InsecureResource = SecureResource.Clone().
//	    Set("has_encryption", false).
//	    Set("is_public", true)
//
// # Running Tests
//
//	go test ./... -tags=fast          # Run fast tests
//	go test ./... -tags=integration   # Run integration tests
package suites
