// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package scanner

import (
	"testing"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/policy"
)

func TestEvaluatePolicies_BasicCheck(t *testing.T) {
	policies := []policy.Policy{
		{
			Metadata: policy.Metadata{Name: "test-policy"},
			Groups: []policy.Group{
				{
					Title: "Test Group",
					Checks: []policy.Check{
						{UID: "check-1"},
					},
				},
			},
			Queries: []policy.Check{
				{
					UID:      "check-1",
					Title:    "Test Check",
					Resource: "test_resource",
					Query:    "resource.enabled == true",
					Severity: "high",
				},
			},
		},
	}

	tests := []struct {
		name         string
		resource     core.Resource
		resourceType string
		wantStatus   policy.Status
		wantNoResult bool
	}{
		{
			name:         "Passing check",
			resource:     core.Resource{"enabled": true, "name": "test"},
			resourceType: "test_resource",
			wantStatus:   policy.StatusPassed,
		},
		{
			name:         "Failing check",
			resource:     core.Resource{"enabled": false, "name": "test"},
			resourceType: "test_resource",
			wantStatus:   policy.StatusFailed,
		},
		{
			name:         "Non-matching resource type",
			resource:     core.Resource{"enabled": true},
			resourceType: "other_resource",
			wantNoResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := EvaluatePolicies(tt.resource, tt.resourceType, policies, nil, core.Asset{})

			if tt.wantNoResult {
				if len(results) != 0 {
					t.Errorf("EvaluatePolicies() returned %d results, want 0", len(results))
				}
				return
			}

			if len(results) != 1 {
				t.Fatalf("EvaluatePolicies() returned %d results, want 1", len(results))
			}

			if results[0].Status != tt.wantStatus {
				t.Errorf("Check status = %q, want %q", results[0].Status, tt.wantStatus)
			}

			if results[0].Severity != "high" {
				t.Errorf("Check severity = %q, want %q", results[0].Severity, "high")
			}
		})
	}
}

func TestEvaluatePolicies_GroupFilter(t *testing.T) {
	policies := []policy.Policy{
		{
			Metadata: policy.Metadata{Name: "test-policy"},
			Groups: []policy.Group{
				{
					Title:  "GitHub Org Group",
					Filter: `asset.type == "github-org"`,
					Checks: []policy.Check{
						{UID: "org-check"},
					},
				},
				{
					Title:  "GitHub Repo Group",
					Filter: `asset.type == "github-repo"`,
					Checks: []policy.Check{
						{UID: "repo-check"},
					},
				},
			},
			Queries: []policy.Check{
				{
					UID:      "org-check",
					Title:    "Org Check",
					Resource: "test_resource",
					Query:    "true",
				},
				{
					UID:      "repo-check",
					Title:    "Repo Check",
					Resource: "test_resource",
					Query:    "true",
				},
			},
		},
	}

	tests := []struct {
		name      string
		asset     core.Asset
		wantCount int
		wantCheck string
	}{
		{
			name:      "Only org checks for org asset",
			asset:     core.Asset{Type: "github-org"},
			wantCount: 1,
			wantCheck: "org-check",
		},
		{
			name:      "Only repo checks for repo asset",
			asset:     core.Asset{Type: "github-repo"},
			wantCount: 1,
			wantCheck: "repo-check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := EvaluatePolicies(
				core.Resource{"name": "test"},
				"test_resource",
				policies,
				nil,
				tt.asset,
			)

			if len(results) != tt.wantCount {
				t.Errorf("EvaluatePolicies() returned %d results, want %d", len(results), tt.wantCount)
			}

			if tt.wantCount > 0 && results[0].ID != tt.wantCheck {
				t.Errorf("Check ID = %q, want %q", results[0].ID, tt.wantCheck)
			}
		})
	}
}

func TestEvaluatePolicies_DefaultSeverity(t *testing.T) {
	policies := []policy.Policy{
		{
			Metadata: policy.Metadata{Name: "test-policy"},
			Groups: []policy.Group{
				{
					Title: "Test Group",
					Checks: []policy.Check{
						{UID: "no-severity-check"},
					},
				},
			},
			Queries: []policy.Check{
				{
					UID:      "no-severity-check",
					Title:    "Check Without Severity",
					Resource: "test_resource",
					Query:    "true",
					// No severity specified
				},
			},
		},
	}

	results := EvaluatePolicies(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("EvaluatePolicies() returned %d results, want 1", len(results))
	}

	if results[0].Severity != "medium" {
		t.Errorf("Default severity = %q, want %q", results[0].Severity, "medium")
	}
}

func TestEvaluatePolicies_InvalidQuery(t *testing.T) {
	policies := []policy.Policy{
		{
			Metadata: policy.Metadata{Name: "test-policy"},
			Groups: []policy.Group{
				{
					Title: "Test Group",
					Checks: []policy.Check{
						{UID: "invalid-check"},
					},
				},
			},
			Queries: []policy.Check{
				{
					UID:      "invalid-check",
					Title:    "Invalid Query Check",
					Resource: "test_resource",
					Query:    "invalid.syntax[[[",
				},
			},
		},
	}

	results := EvaluatePolicies(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("EvaluatePolicies() returned %d results, want 1", len(results))
	}

	// Invalid syntax should result in skipped status
	if results[0].Status != StatusSkipped {
		t.Errorf("Status = %q, want %q for invalid query", results[0].Status, StatusSkipped)
	}
}

func TestEvaluatePolicies_CheckResolution(t *testing.T) {
	// Test that check references are properly resolved from queries
	policies := []policy.Policy{
		{
			Metadata: policy.Metadata{Name: "test-policy"},
			Groups: []policy.Group{
				{
					Title: "Test Group",
					Checks: []policy.Check{
						{UID: "ref-by-uid"},
						{ID: "ref-by-id"},
					},
				},
			},
			Queries: []policy.Check{
				{
					UID:      "ref-by-uid",
					Title:    "Referenced by UID",
					Resource: "test_resource",
					Query:    "true",
				},
				{
					ID:       "ref-by-id",
					Title:    "Referenced by ID",
					Resource: "test_resource",
					Query:    "true",
				},
			},
		},
	}

	results := EvaluatePolicies(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 2 {
		t.Fatalf("EvaluatePolicies() returned %d results, want 2", len(results))
	}

	// Both checks should pass
	for _, r := range results {
		if r.Status != StatusPassed {
			t.Errorf("Check %q status = %q, want %q", r.ID, r.Status, StatusPassed)
		}
	}
}

func TestEvaluatePolicies_UnresolvedCheck(t *testing.T) {
	// Test that unresolved check references are skipped
	policies := []policy.Policy{
		{
			Metadata: policy.Metadata{Name: "test-policy"},
			Groups: []policy.Group{
				{
					Title: "Test Group",
					Checks: []policy.Check{
						{UID: "nonexistent-check"},
					},
				},
			},
			Queries: []policy.Check{
				// No matching query for "nonexistent-check"
			},
		},
	}

	results := EvaluatePolicies(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	// Should return no results since the check reference couldn't be resolved
	if len(results) != 0 {
		t.Errorf("EvaluatePolicies() returned %d results, want 0 for unresolved check", len(results))
	}
}
