// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package suites

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/kopexa-grc/kspec/policy"
	"github.com/kopexa-grc/kspec/provider/scanner"
)

// Runner executes policy integration test suites.
type Runner struct {
	policiesDir string
}

// NewRunner creates a new test runner.
// policiesDir is the base directory for policy files.
func NewRunner(policiesDir string) (*Runner, error) {
	return &Runner{
		policiesDir: policiesDir,
	}, nil
}

// RunSuite executes all test cases in a suite.
func (r *Runner) RunSuite(t *testing.T, suite Suite) {
	t.Helper()

	if err := suite.Setup(t); err != nil {
		t.Fatalf("Suite setup failed: %v", err)
	}

	t.Cleanup(func() {
		if err := suite.Teardown(t); err != nil {
			t.Errorf("Suite teardown failed: %v", err)
		}
	})

	for _, tc := range suite.TestCases() {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Skip != "" {
				t.Skip(tc.Skip)
			}

			r.runTestCase(t, tc)
		})
	}
}

// runTestCase executes a single test case.
func (r *Runner) runTestCase(t *testing.T, tc TestCase) {
	t.Helper()

	// Load policies
	policyPath := tc.PolicyFile
	if !filepath.IsAbs(policyPath) {
		policyPath = filepath.Join(r.policiesDir, tc.PolicyFile)
	}

	policies, err := policy.Load(policyPath, "")
	if err != nil {
		t.Fatalf("Failed to load policies from %s: %v", tc.PolicyFile, err)
	}

	// Filter to specific policy if requested
	if tc.PolicyName != "" {
		var filtered []policy.Policy
		for _, p := range policies {
			if p.Metadata.Name == tc.PolicyName {
				filtered = append(filtered, p)
			}
		}
		policies = filtered
	}

	if len(policies) == 0 {
		t.Fatal("No policies found to test")
	}

	if len(tc.Fixtures) == 0 {
		t.Fatal("No fixtures provided to test")
	}

	// Test each fixture
	for fixtureIdx, fixture := range tc.Fixtures {
		results := scanner.EvaluatePolicies(
			fixture,
			tc.ResourceType,
			policies,
			nil,
			tc.Asset,
		)

		// Build a map of results by check ID for easier lookup
		resultsByID := make(map[string]scanner.CheckResult)
		for _, result := range results {
			resultsByID[result.ID] = result
		}

		// Verify expected results
		for checkID, expected := range tc.ExpectedResults {
			result, found := resultsByID[checkID]

			// Create subtest for each check
			checkName := checkID
			if len(tc.Fixtures) > 1 {
				checkName = strings.ReplaceAll(checkID, "/", "_") + "_fixture_" + strconv.Itoa(fixtureIdx)
			}

			t.Run(checkName, func(t *testing.T) {
				if expected.SkipReason != "" {
					t.Skip(expected.SkipReason)
				}

				if !found {
					t.Errorf("Check %q was not evaluated (resource type mismatch or check not found in policy)", checkID)
					return
				}

				actualPass := result.Status == scanner.StatusPassed

				if actualPass != expected.Pass {
					if expected.Pass {
						t.Errorf("Check %q: expected PASS, got %s", checkID, result.Status)
						if result.Details != "" {
							t.Logf("  Details: %s", result.Details)
						}
					} else {
						t.Errorf("Check %q: expected FAIL, got %s", checkID, result.Status)
					}
				}

				if expected.ErrorContains != "" && !strings.Contains(result.Details, expected.ErrorContains) {
					t.Errorf("Check %q: expected error containing %q, got %q", checkID, expected.ErrorContains, result.Details)
				}
			})
		}
	}
}

// RunAll executes all registered suites.
func (r *Runner) RunAll(t *testing.T) {
	t.Helper()

	for _, suite := range All() {
		suite := suite
		t.Run(suite.Name(), func(t *testing.T) {
			r.RunSuite(t, suite)
		})
	}
}

// RunProvider executes all suites for a specific provider.
func (r *Runner) RunProvider(t *testing.T, provider string) {
	t.Helper()

	suites := ForProvider(provider)
	if len(suites) == 0 {
		t.Skipf("No suites registered for provider %q", provider)
	}

	for _, suite := range suites {
		suite := suite
		t.Run(suite.Name(), func(t *testing.T) {
			r.RunSuite(t, suite)
		})
	}
}

// QuickTest is a helper function for running a quick test case
// without needing to implement the full Suite interface.
func QuickTest(t *testing.T, policiesDir string, tc TestCase) {
	t.Helper()

	runner, err := NewRunner(policiesDir)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	runner.runTestCase(t, tc)
}
