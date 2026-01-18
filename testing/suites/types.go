// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package suites provides a framework for integration testing policies
// against provider resources.
package suites

import (
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

// Suite defines a policy integration test suite.
// Each provider can implement one or more suites to test their resources
// against policy definitions.
type Suite interface {
	// Name returns the unique suite name (e.g., "aws-s3", "github-repo").
	Name() string

	// Provider returns the provider this suite tests (e.g., "aws", "github").
	Provider() string

	// Setup prepares the suite for testing.
	// Called once before any test cases are run.
	Setup(t *testing.T) error

	// Teardown cleans up after the suite.
	// Called once after all test cases have run.
	Teardown(t *testing.T) error

	// TestCases returns all test cases in this suite.
	TestCases() []TestCase
}

// TestCase represents a single policy integration test.
type TestCase struct {
	// Name is the test case name, displayed in test output.
	Name string

	// PolicyFile is the path to the policy YAML file (relative to policies dir).
	PolicyFile string

	// PolicyName filters to a specific policy by name (optional).
	// If empty, all policies from the file are tested.
	PolicyName string

	// ResourceType is the resource type being tested (e.g., "aws_s3_bucket").
	ResourceType string

	// Fixtures are the resource instances to evaluate against.
	Fixtures []core.Resource

	// ExpectedResults maps check UID to expected outcome.
	ExpectedResults map[string]ExpectedResult

	// Asset provides context for the evaluation.
	Asset core.Asset

	// Skip indicates this test should be skipped with the given reason.
	Skip string
}

// ExpectedResult defines the expected outcome of a check.
type ExpectedResult struct {
	// Pass indicates if the check should pass (true) or fail (false).
	Pass bool

	// SkipReason if set, indicates the check should be skipped.
	SkipReason string

	// ErrorContains if set, the evaluation error should contain this string.
	ErrorContains string
}

// FixtureBuilder provides a fluent API for building resource fixtures.
type FixtureBuilder struct {
	resource core.Resource
}

// NewFixture creates a new fixture builder with an empty resource.
func NewFixture() *FixtureBuilder {
	return &FixtureBuilder{
		resource: make(core.Resource),
	}
}

// Set adds a key-value pair to the fixture.
func (b *FixtureBuilder) Set(key string, value interface{}) *FixtureBuilder {
	b.resource[key] = value
	return b
}

// SetAll adds multiple key-value pairs from a map.
func (b *FixtureBuilder) SetAll(kv map[string]interface{}) *FixtureBuilder {
	for k, v := range kv {
		b.resource[k] = v
	}
	return b
}

// Merge combines another fixture's values into this one.
// Existing keys are overwritten.
func (b *FixtureBuilder) Merge(other *FixtureBuilder) *FixtureBuilder {
	if other != nil {
		for k, v := range other.resource {
			b.resource[k] = v
		}
	}
	return b
}

// Build returns the constructed resource.
func (b *FixtureBuilder) Build() core.Resource {
	// Return a copy to prevent mutation
	result := make(core.Resource, len(b.resource))
	for k, v := range b.resource {
		result[k] = v
	}
	return result
}

// Clone creates a deep copy of this builder for modifications.
func (b *FixtureBuilder) Clone() *FixtureBuilder {
	clone := &FixtureBuilder{
		resource: make(core.Resource, len(b.resource)),
	}
	for k, v := range b.resource {
		clone.resource[k] = v
	}
	return clone
}

// With returns a clone with the specified key-value pair set.
// Useful for creating variations: SecureBucket.With("is_public", true)
func (b *FixtureBuilder) With(key string, value interface{}) *FixtureBuilder {
	return b.Clone().Set(key, value)
}

// Without returns a clone with the specified key removed.
func (b *FixtureBuilder) Without(key string) *FixtureBuilder {
	clone := b.Clone()
	delete(clone.resource, key)
	return clone
}
