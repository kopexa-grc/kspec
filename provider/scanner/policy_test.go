package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

func TestLoadPolicies_SingleFile(t *testing.T) {
	// Create a temporary policy file
	tmpDir := t.TempDir()
	policyContent := `
apiVersion: kopexa.io/v1alpha1
kind: Policy
metadata:
  name: test-policy
  version: "1.0.0"
require:
  - provider: github
groups:
  - title: Test Group
    checks:
      - uid: test-check
queries:
  - uid: test-check
    title: Test Check
    resource: github_repo
    query: "true"
`
	policyFile := filepath.Join(tmpDir, "test-policy.yml")
	if err := os.WriteFile(policyFile, []byte(policyContent), 0o644); err != nil {
		t.Fatalf("Failed to write policy file: %v", err)
	}

	policies, err := LoadPolicies(policyFile, "")
	if err != nil {
		t.Fatalf("LoadPolicies() error = %v", err)
	}

	if len(policies) != 1 {
		t.Errorf("LoadPolicies() returned %d policies, want 1", len(policies))
	}

	if policies[0].Metadata.Name != "test-policy" {
		t.Errorf("Policy name = %q, want %q", policies[0].Metadata.Name, "test-policy")
	}

	if len(policies[0].Require) != 1 || policies[0].Require[0].Provider != "github" {
		t.Errorf("Policy require = %v, want [{Provider: github}]", policies[0].Require)
	}
}

func TestLoadPolicies_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple policy files
	policies := []struct {
		name    string
		content string
	}{
		{
			name: "github-policy.yml",
			content: `
apiVersion: kopexa.io/v1alpha1
kind: Policy
metadata:
  name: github-policy
require:
  - provider: github
groups: []
queries: []
`,
		},
		{
			name: "azure-policy.yaml",
			content: `
apiVersion: kopexa.io/v1alpha1
kind: Policy
metadata:
  name: azure-policy
require:
  - provider: azure
groups: []
queries: []
`,
		},
		{
			name:    "not-a-policy.txt",
			content: "This is not a policy file",
		},
	}

	for _, p := range policies {
		path := filepath.Join(tmpDir, p.name)
		if err := os.WriteFile(path, []byte(p.content), 0o644); err != nil {
			t.Fatalf("Failed to write file %s: %v", p.name, err)
		}
	}

	result, err := LoadPolicies("", tmpDir)
	if err != nil {
		t.Fatalf("LoadPolicies() error = %v", err)
	}

	// Should only load .yml and .yaml files
	if len(result) != 2 {
		t.Errorf("LoadPolicies() returned %d policies, want 2", len(result))
	}
}

func TestLoadPolicies_NonExistentFile(t *testing.T) {
	_, err := LoadPolicies("/nonexistent/path/policy.yml", "")
	if err == nil {
		t.Error("LoadPolicies() expected error for nonexistent file")
	}
}

func TestLoadPolicies_NonExistentDirectory(t *testing.T) {
	_, err := LoadPolicies("", "/nonexistent/directory")
	if err == nil {
		t.Error("LoadPolicies() expected error for nonexistent directory")
	}
}

func TestLoadPolicies_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	invalidContent := `
this is not: valid yaml: [
`
	policyFile := filepath.Join(tmpDir, "invalid.yml")
	if err := os.WriteFile(policyFile, []byte(invalidContent), 0o644); err != nil {
		t.Fatalf("Failed to write policy file: %v", err)
	}

	_, err := LoadPolicies(policyFile, "")
	if err == nil {
		t.Error("LoadPolicies() expected error for invalid YAML")
	}
}

func TestFilterPoliciesByProvider(t *testing.T) {
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "github-policy"},
			Require:  []core.Requirement{{Provider: "github"}},
		},
		{
			Metadata: core.Metadata{Name: "azure-policy"},
			Require:  []core.Requirement{{Provider: "azure"}},
		},
		{
			Metadata: core.Metadata{Name: "multi-policy"},
			Require: []core.Requirement{
				{Provider: "github"},
				{Provider: "azure"},
			},
		},
		{
			Metadata: core.Metadata{Name: "universal-policy"},
			Require:  []core.Requirement{}, // No requirements - applies to all
		},
	}

	tests := []struct {
		name          string
		providerName  string
		providerAlias string
		wantCount     int
		wantNames     []string
	}{
		{
			name:          "Filter for github",
			providerName:  "github",
			providerAlias: "",
			wantCount:     3, // github-policy, multi-policy, universal-policy
			wantNames:     []string{"github-policy", "multi-policy", "universal-policy"},
		},
		{
			name:          "Filter for azure",
			providerName:  "azure",
			providerAlias: "",
			wantCount:     3, // azure-policy, multi-policy, universal-policy
			wantNames:     []string{"azure-policy", "multi-policy", "universal-policy"},
		},
		{
			name:          "Filter for unknown provider",
			providerName:  "unknown",
			providerAlias: "",
			wantCount:     1, // only universal-policy
			wantNames:     []string{"universal-policy"},
		},
		{
			name:          "Filter with alias",
			providerName:  "hetzner",
			providerAlias: "hcloud",
			wantCount:     1, // only universal-policy (no hetzner policy)
			wantNames:     []string{"universal-policy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterPoliciesByProvider(policies, tt.providerName, tt.providerAlias)
			if len(result) != tt.wantCount {
				t.Errorf("FilterPoliciesByProvider() returned %d policies, want %d", len(result), tt.wantCount)
			}

			// Verify expected policy names
			names := make(map[string]bool)
			for _, p := range result {
				names[p.Metadata.Name] = true
			}
			for _, wantName := range tt.wantNames {
				if !names[wantName] {
					t.Errorf("FilterPoliciesByProvider() missing policy %q", wantName)
				}
			}
		})
	}
}

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

func TestEvaluatePolicies_BasicCheck(t *testing.T) {
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "test-policy"},
			Groups: []core.Group{
				{
					Title: "Test Group",
					Checks: []core.Check{
						{UID: "check-1"},
					},
				},
			},
			Queries: []core.Check{
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
			results := EvaluatePolicies(tt.resource, tt.resourceType, policies, nil, core.Asset{})

			if tt.wantStatus == "" {
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
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "test-policy"},
			Groups: []core.Group{
				{
					Title:  "GitHub Org Group",
					Filter: `asset.type == "github-org"`,
					Checks: []core.Check{
						{UID: "org-check"},
					},
				},
				{
					Title:  "GitHub Repo Group",
					Filter: `asset.type == "github-repo"`,
					Checks: []core.Check{
						{UID: "repo-check"},
					},
				},
			},
			Queries: []core.Check{
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
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "test-policy"},
			Groups: []core.Group{
				{
					Title: "Test Group",
					Checks: []core.Check{
						{UID: "no-severity-check"},
					},
				},
			},
			Queries: []core.Check{
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
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "test-policy"},
			Groups: []core.Group{
				{
					Title: "Test Group",
					Checks: []core.Check{
						{UID: "invalid-check"},
					},
				},
			},
			Queries: []core.Check{
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

func TestEvaluatePolicies_ConfigMerge(t *testing.T) {
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "test-policy"},
			Groups: []core.Group{
				{
					Title: "Test Group",
					Checks: []core.Check{
						{
							UID: "config-check",
							Config: map[string]string{
								"override": "group-value",
								"group":    "group-only",
							},
						},
					},
				},
			},
			Queries: []core.Check{
				{
					UID:      "config-check",
					Title:    "Config Check",
					Resource: "test_resource",
					Query:    "true",
					Config: map[string]string{
						"override": "query-value",
						"query":    "query-only",
					},
				},
			},
		},
	}

	// The check should execute (we can't easily verify config merging without
	// modifying the evaluator, but we can verify the check runs)
	results := EvaluatePolicies(
		core.Resource{"name": "test"},
		"test_resource",
		policies,
		nil,
		core.Asset{Config: map[string]string{"asset": "asset-value"}},
	)

	if len(results) != 1 {
		t.Fatalf("EvaluatePolicies() returned %d results, want 1", len(results))
	}

	if results[0].Status != StatusPassed {
		t.Errorf("Status = %q, want %q", results[0].Status, StatusPassed)
	}
}

func TestEvaluatePolicies_CheckResolution(t *testing.T) {
	// Test that check references are properly resolved from queries
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "test-policy"},
			Groups: []core.Group{
				{
					Title: "Test Group",
					Checks: []core.Check{
						{UID: "ref-by-uid"},
						{ID: "ref-by-id"},
					},
				},
			},
			Queries: []core.Check{
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

// FuzzEvaluateGroupFilter tests the group filter parsing with random input.
func FuzzEvaluateGroupFilter(f *testing.F) {
	// Seed corpus with valid and edge-case filters
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
		`asset.type == "test" && other`,
		`multiple == "quotes" == "here"`,
		"asset.type == \"unclosed",
		"asset.type == 'unclosed",
		`== "value"`,
		`"value" == asset.type`,
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

// FuzzYAMLParsing tests YAML parsing with random input.
func FuzzYAMLParsing(f *testing.F) {
	// Seed corpus with valid and malformed YAML
	seeds := []string{
		`apiVersion: kopexa.io/v1alpha1
kind: Policy
metadata:
  name: test
`,
		`{}`,
		`[]`,
		`null`,
		``,
		`invalid: yaml: [`,
		`- item1
- item2`,
		`key: value
another: 123`,
		`deeply:
  nested:
    structure:
      here: true`,
		`"string only"`,
		`123`,
		`true`,
		`key: |
  multiline
  string`,
		`key: >
  folded
  string`,
		`---
document1: true
---
document2: true`,
		`&anchor
key: *anchor`,
		string(make([]byte, 1000)), // zeros
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, content string) {
		// Create a temp file with the fuzzed content
		tmpDir := t.TempDir()
		policyFile := tmpDir + "/fuzz-policy.yml"
		if err := os.WriteFile(policyFile, []byte(content), 0o644); err != nil {
			t.Skip("Failed to write temp file")
		}

		// Should never panic, errors are expected for invalid YAML
		_, _ = LoadPolicies(policyFile, "")
	})
}

// TestEvaluateGroupFilter_EdgeCases tests edge cases in filter parsing.
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
		{
			name:   "Newline in filter",
			filter: "asset.type ==\n\"test\"",
			asset:  core.Asset{Type: "any"},
			want:   true, // Falls back to true
		},
		{
			name:   "Tab in filter",
			filter: "asset.type ==\t\"test\"",
			asset:  core.Asset{Type: "any"},
			want:   true, // Falls back to true
		},
		{
			name:   "Only whitespace after separator",
			filter: `asset.type == "   "`,
			asset:  core.Asset{Type: "   "},
			want:   true,
		},
		{
			name:   "Very long value",
			filter: `asset.type == "` + string(make([]byte, 1000)) + `"`,
			asset:  core.Asset{Type: string(make([]byte, 1000))},
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

func TestEvaluatePolicies_UnresolvedCheck(t *testing.T) {
	// Test that unresolved check references are skipped
	policies := []core.Policy{
		{
			Metadata: core.Metadata{Name: "test-policy"},
			Groups: []core.Group{
				{
					Title: "Test Group",
					Checks: []core.Check{
						{UID: "nonexistent-check"},
					},
				},
			},
			Queries: []core.Check{
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
