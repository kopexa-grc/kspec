// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

func TestEvaluateGroupFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		asset  core.Asset
		want   bool
	}{
		{
			name:   "Empty filter returns true",
			filter: "",
			asset:  core.Asset{Type: "github-org"},
			want:   true,
		},
		{
			name:   "Match asset type with double quotes",
			filter: `asset.type == "github-org"`,
			asset:  core.Asset{Type: "github-org"},
			want:   true,
		},
		{
			name:   "Match asset type with single quotes",
			filter: `asset.type == 'github-repo'`,
			asset:  core.Asset{Type: "github-repo"},
			want:   true,
		},
		{
			name:   "Non-matching asset type",
			filter: `asset.type == "azure-subscription"`,
			asset:  core.Asset{Type: "github-org"},
			want:   false,
		},
		{
			name:   "Match without space after ==",
			filter: `asset.type =="hetzner-project"`,
			asset:  core.Asset{Type: "hetzner-project"},
			want:   true,
		},
		{
			name:   "Unparseable filter returns true (permissive)",
			filter: `some.unknown.expression`,
			asset:  core.Asset{Type: "any"},
			want:   true,
		},
		{
			name:   "Filter with extra whitespace",
			filter: `  asset.type == "factorial-company"  `,
			asset:  core.Asset{Type: "factorial-company"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateGroupFilter(tt.filter, tt.asset)
			if result != tt.want {
				t.Errorf("evaluateGroupFilter(%q, %v) = %v, want %v", tt.filter, tt.asset.Type, result, tt.want)
			}
		})
	}
}

func TestEvaluateGroupFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		asset  core.Asset
		want   bool
	}{
		{
			name:   "Empty value in quotes",
			filter: `asset.type == ""`,
			asset:  core.Asset{Type: ""},
			want:   true,
		},
		{
			name:   "Empty value single quotes",
			filter: `asset.type == ''`,
			asset:  core.Asset{Type: ""},
			want:   true,
		},
		{
			name:   "Unclosed double quote",
			filter: `asset.type == "unclosed`,
			asset:  core.Asset{Type: "any"},
			want:   true, // Falls back to true
		},
		{
			name:   "Unclosed single quote",
			filter: `asset.type == 'unclosed`,
			asset:  core.Asset{Type: "any"},
			want:   true, // Falls back to true
		},
		{
			name:   "Value before operator",
			filter: `"value" == asset.type`,
			asset:  core.Asset{Type: "value"},
			want:   true, // Falls back to true (not parsed)
		},
		{
			name:   "Multiple separators",
			filter: `asset.type == "first" == "second"`,
			asset:  core.Asset{Type: "first"},
			want:   true, // Should match first value
		},
		{
			name:   "Unicode in filter",
			filter: `asset.type == "日本語"`,
			asset:  core.Asset{Type: "日本語"},
			want:   true,
		},
		{
			name:   "Special characters in value",
			filter: `asset.type == "test-with-dashes_and_underscores"`,
			asset:  core.Asset{Type: "test-with-dashes_and_underscores"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateGroupFilter(tt.filter, tt.asset)
			if result != tt.want {
				t.Errorf("evaluateGroupFilter(%q) = %v, want %v", tt.filter, result, tt.want)
			}
		})
	}
}

func TestEvaluate_BasicCheck(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "test-policy"},
			Groups: []Group{
				{
					Title: "Test Group",
					Checks: []Check{
						{UID: "check-1"},
					},
				},
			},
			Queries: []Check{
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
		wantStatus   Status
		wantNoResult bool
	}{
		{
			name:         "Passing check",
			resource:     core.Resource{"enabled": true, "name": "test"},
			resourceType: "test_resource",
			wantStatus:   StatusPassed,
		},
		{
			name:         "Failing check",
			resource:     core.Resource{"enabled": false, "name": "test"},
			resourceType: "test_resource",
			wantStatus:   StatusFailed,
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
			results := Evaluate(tt.resource, tt.resourceType, policies, nil, core.Asset{})

			if tt.wantNoResult {
				if len(results) != 0 {
					t.Errorf("Evaluate() returned %d results, want 0", len(results))
				}
				return
			}

			if len(results) != 1 {
				t.Fatalf("Evaluate() returned %d results, want 1", len(results))
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

func TestEvaluate_GroupFilter(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "filtered-policy"},
			Groups: []Group{
				{
					Title:  "GitHub Group",
					Filter: `asset.type == "github-org"`,
					Checks: []Check{
						{UID: "github-check"},
					},
				},
				{
					Title:  "Azure Group",
					Filter: `asset.type == "azure-subscription"`,
					Checks: []Check{
						{UID: "azure-check"},
					},
				},
			},
			Queries: []Check{
				{
					UID:      "github-check",
					Title:    "GitHub Check",
					Resource: "test_resource",
					Query:    "true",
				},
				{
					UID:      "azure-check",
					Title:    "Azure Check",
					Resource: "test_resource",
					Query:    "true",
				},
			},
		},
	}

	t.Run("Only matching group filter is evaluated", func(t *testing.T) {
		results := Evaluate(
			core.Resource{"name": "test"},
			"test_resource",
			policies,
			nil,
			core.Asset{Type: "github-org"},
		)

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if results[0].Group != "GitHub Group" {
			t.Errorf("Expected GitHub Group, got %s", results[0].Group)
		}
	})
}

func TestEvaluate_DefinitionByID(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "id-policy"},
			Groups: []Group{
				{
					Title: "Test Group",
					Checks: []Check{
						{ID: "check-by-id"},
					},
				},
			},
			Queries: []Check{
				{
					ID:       "check-by-id",
					Title:    "Check by ID",
					Resource: "test_resource",
					Query:    "resource.value == true",
					Severity: "medium",
				},
			},
		},
	}

	results := Evaluate(
		core.Resource{"value": true},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != StatusPassed {
		t.Errorf("Expected passed, got %s", results[0].Status)
	}
}

func TestEvaluate_EmptyPolicies(t *testing.T) {
	results := Evaluate(
		core.Resource{"name": "test"},
		"test_resource",
		[]Policy{},
		nil,
		core.Asset{},
	)

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty policies, got %d", len(results))
	}
}

func TestEvaluate_CheckWithProps(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "props-policy"},
			Groups: []Group{
				{
					Title: "Props Group",
					Checks: []Check{
						{UID: "props-check"},
					},
				},
			},
			Queries: []Check{
				{
					UID:      "props-check",
					Title:    "Props Check",
					Resource: "test_resource",
					Query:    "resource.enabled == true",
					Props: []Prop{
						{UID: "helper", MQL: "resource.name"},
					},
				},
			},
		},
	}

	results := Evaluate(
		core.Resource{"enabled": true, "name": "test-value"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != StatusPassed {
		t.Errorf("Expected passed, got %s", results[0].Status)
	}
}

func TestEvaluate_ConfigMerge(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "config-policy"},
			Groups: []Group{
				{
					Title: "Config Group",
					Checks: []Check{
						{
							UID: "config-check",
							Config: map[string]string{
								"check_override": "from_check",
							},
						},
					},
				},
			},
			Queries: []Check{
				{
					UID:      "config-check",
					Title:    "Config Check",
					Resource: "test_resource",
					Query:    "true",
					Config: map[string]string{
						"def_key":        "from_def",
						"check_override": "from_def",
					},
				},
			},
		},
	}

	results := Evaluate(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{
			Config: map[string]string{
				"asset_key": "from_asset",
			},
		},
	)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
}

func TestEvaluate_DefaultSeverity(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "no-severity-policy"},
			Groups: []Group{
				{
					Title: "Test Group",
					Checks: []Check{
						{UID: "no-severity-check"},
					},
				},
			},
			Queries: []Check{
				{
					UID:      "no-severity-check",
					Title:    "No Severity Check",
					Resource: "test_resource",
					Query:    "true",
					// No severity specified
				},
			},
		},
	}

	results := Evaluate(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Severity != "medium" {
		t.Errorf("Expected default severity 'medium', got %s", results[0].Severity)
	}
}

func TestEvaluate_CheckNotFound(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "missing-check-policy"},
			Groups: []Group{
				{
					Title: "Test Group",
					Checks: []Check{
						{UID: "nonexistent-check"},
					},
				},
			},
			Queries: []Check{}, // No definitions
		},
	}

	results := Evaluate(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	// Check reference not found, should be skipped
	if len(results) != 0 {
		t.Errorf("Expected 0 results for nonexistent check, got %d", len(results))
	}
}

func TestEvaluate_InlineCheck(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "inline-policy"},
			Groups: []Group{
				{
					Title: "Test Group",
					Checks: []Check{
						{
							UID:      "inline-check",
							Title:    "Inline Check",
							Resource: "test_resource",
							Query:    "resource.valid == true",
							Severity: "high",
						},
					},
				},
			},
			// No Queries section - check is inline
		},
	}

	results := Evaluate(
		core.Resource{"valid": true},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Status != StatusPassed {
		t.Errorf("Expected passed, got %s", results[0].Status)
	}
}

func TestEvaluate_ErrorInQuery(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "error-policy"},
			Groups: []Group{
				{
					Title: "Test Group",
					Checks: []Check{
						{
							UID:      "error-check",
							Title:    "Error Check",
							Resource: "test_resource",
							Query:    "resource.nonexistent.deeply.nested",
							Severity: "high",
						},
					},
				},
			},
		},
	}

	results := Evaluate(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Should be failed due to error
	if results[0].Status != StatusFailed {
		t.Errorf("Expected failed due to evaluation error, got %s", results[0].Status)
	}
}

func TestEvaluate_MetadataFields(t *testing.T) {
	policies := []Policy{
		{
			Metadata: Metadata{Name: "metadata-policy"},
			Groups: []Group{
				{
					Title: "Security Checks",
					Checks: []Check{
						{UID: "metadata-check"},
					},
				},
			},
			Queries: []Check{
				{
					UID:         "metadata-check",
					Title:       "Full Metadata Check",
					Resource:    "test_resource",
					Query:       "true",
					Severity:    "critical",
					Remediation: "Fix this by doing X",
					Docs:        "Documentation about this check",
					Audit:       "Manual audit instructions",
				},
			},
		},
	}

	results := Evaluate(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{},
	)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.ID != "metadata-check" {
		t.Errorf("Expected ID 'metadata-check', got %s", r.ID)
	}
	if r.Group != "Security Checks" {
		t.Errorf("Expected Group 'Security Checks', got %s", r.Group)
	}
	if r.Name != "Full Metadata Check" {
		t.Errorf("Expected Name 'Full Metadata Check', got %s", r.Name)
	}
	if r.Remediation != "Fix this by doing X" {
		t.Errorf("Expected Remediation set, got %s", r.Remediation)
	}
	if r.Docs != "Documentation about this check" {
		t.Errorf("Expected Docs set, got %s", r.Docs)
	}
	if r.Audit != "Manual audit instructions" {
		t.Errorf("Expected Audit set, got %s", r.Audit)
	}
}

// FuzzEvaluateGroupFilter tests the group filter parsing with random input.
func FuzzEvaluateGroupFilter(f *testing.F) {
	seeds := []string{
		"",
		`asset.type == "github-org"`,
		`asset.type == 'github-repo'`,
		`asset.type =="test"`,
		`asset.type=='test'`,
		"some.unknown.expression",
		"asset.type",
		`asset.type == "`,
		`asset.type == '`,
		`asset.type == ""`,
		`asset.type == ''`,
		"   ",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, filter string) {
		// Should never panic
		asset := core.Asset{Type: "test-type"}
		_ = evaluateGroupFilter(filter, asset)
	})
}
