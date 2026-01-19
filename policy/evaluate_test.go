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
		wantStatus   string
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
			wantStatus:   "", // No results expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Evaluate(tt.resource, tt.resourceType, policies, nil, core.Asset{})

			if tt.wantStatus == "" {
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
