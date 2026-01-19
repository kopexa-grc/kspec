// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package scanner

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/kopexa-grc/kspec/core"
)

// testdataDir returns the path to the testdata directory
func testdataDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get caller information")
	}
	return filepath.Join(filepath.Dir(filename), "testdata", "imports")
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com/policy.yaml", true},
		{"http://example.com/policy.yaml", true},
		{"./local/policy.yaml", false},
		{"/absolute/path/policy.yaml", false},
		{"relative/path.yaml", false},
		{"ftp://example.com/file", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGlob(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"*.yaml", true},
		{"policies/*.yml", true},
		{"checks/[a-z]*.yaml", true},
		{"specific-file.yaml", false},
		{"/path/to/file.yaml", false},
		{"https://example.com/*.yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isGlob(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImportResolver_ResolveFileImport(t *testing.T) {
	testdata := testdataDir()

	resolver := NewImportResolver(testdata)
	queries, err := resolver.resolveFileImport("base-checks.yaml", "")
	require.NoError(t, err)
	assert.Len(t, queries, 2)
	assert.Equal(t, "base-check-1", queries[0].UID)
	assert.Equal(t, "base-check-2", queries[1].UID)
	assert.Equal(t, "high", queries[0].Severity)
	assert.Equal(t, "medium", queries[1].Severity)
}

func TestImportResolver_ResolveGlobImport(t *testing.T) {
	testdata := testdataDir()

	resolver := NewImportResolver(testdata)
	queries, err := resolver.resolveGlobImport("checks/*.yaml", "")
	require.NoError(t, err)
	assert.Len(t, queries, 3) // 1 from storage-checks + 2 from network-checks

	// Check that all UIDs are present
	uids := make(map[string]bool)
	for _, q := range queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["storage-encryption"], "should have storage-encryption check")
	assert.True(t, uids["network-firewall"], "should have network-firewall check")
	assert.True(t, uids["network-encryption"], "should have network-encryption check")
}

func TestImportResolver_ResolveURLImport(t *testing.T) {
	// Create test server serving a policy
	policy := `
apiVersion: kspec/v1
kind: Policy
queries:
  - uid: remote-check
    title: Remote Check
    resource: test_resource
    query: "resource.remote == true"
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(policy))
	}))
	defer server.Close()

	resolver := NewImportResolver(".")
	queries, err := resolver.resolveURLImport(server.URL + "/policy.yaml")
	require.NoError(t, err)
	assert.Len(t, queries, 1)
	assert.Equal(t, "remote-check", queries[0].UID)
}

func TestImportResolver_ResolveImports_MainPolicy(t *testing.T) {
	testdata := testdataDir()

	// Load main policy from testdata
	data, err := os.ReadFile(filepath.Join(testdata, "main-policy.yaml"))
	require.NoError(t, err)

	var policy core.Policy
	err = yaml.Unmarshal(data, &policy)
	require.NoError(t, err)

	// Verify it has imports defined
	assert.Equal(t, []string{"base-checks.yaml"}, policy.Imports)

	// Resolve imports
	resolver := NewImportResolver(testdata)
	err = resolver.ResolveImports(&policy, filepath.Join(testdata, "main-policy.yaml"))
	require.NoError(t, err)

	// Should have 3 queries: 2 imported + 1 local
	assert.Len(t, policy.Queries, 3)

	// Imported checks are prepended
	assert.Equal(t, "base-check-1", policy.Queries[0].UID)
	assert.Equal(t, "base-check-2", policy.Queries[1].UID)
	assert.Equal(t, "local-check", policy.Queries[2].UID)
}

func TestImportResolver_ResolveImports_GlobPolicy(t *testing.T) {
	testdata := testdataDir()

	// Load glob policy from testdata
	data, err := os.ReadFile(filepath.Join(testdata, "glob-policy.yaml"))
	require.NoError(t, err)

	var policy core.Policy
	err = yaml.Unmarshal(data, &policy)
	require.NoError(t, err)

	// Verify it has glob imports defined
	assert.Equal(t, []string{"checks/*.yaml"}, policy.Imports)

	// Resolve imports
	resolver := NewImportResolver(testdata)
	err = resolver.ResolveImports(&policy, filepath.Join(testdata, "glob-policy.yaml"))
	require.NoError(t, err)

	// Should have 3 queries from glob imports (no local queries in this policy)
	assert.Len(t, policy.Queries, 3)

	// Verify all imported checks are present
	uids := make(map[string]bool)
	for _, q := range policy.Queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["storage-encryption"])
	assert.True(t, uids["network-firewall"])
	assert.True(t, uids["network-encryption"])
}

func TestImportResolver_CircularImport(t *testing.T) {
	testdata := filepath.Join(testdataDir(), "circular")

	// Load policy A which imports policy B (which imports policy A back)
	data, err := os.ReadFile(filepath.Join(testdata, "policy-a.yaml"))
	require.NoError(t, err)

	var policy core.Policy
	err = yaml.Unmarshal(data, &policy)
	require.NoError(t, err)

	// Resolve imports - should not infinite loop due to cycle detection
	resolver := NewImportResolver(testdata)
	err = resolver.ResolveImports(&policy, filepath.Join(testdata, "policy-a.yaml"))
	require.NoError(t, err)

	// Should have check-b (from policy-b) and check-a (local)
	// policy-b's import of policy-a should be skipped (already visited)
	assert.Len(t, policy.Queries, 2)

	uids := make(map[string]bool)
	for _, q := range policy.Queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["check-a"], "should have check-a")
	assert.True(t, uids["check-b"], "should have check-b")
}

func TestImportResolver_NestedImports(t *testing.T) {
	testdata := filepath.Join(testdataDir(), "nested")

	// Load level 1 which imports level 2 which imports level 3
	data, err := os.ReadFile(filepath.Join(testdata, "level1.yaml"))
	require.NoError(t, err)

	var policy core.Policy
	err = yaml.Unmarshal(data, &policy)
	require.NoError(t, err)

	// Resolve imports
	resolver := NewImportResolver(testdata)
	err = resolver.ResolveImports(&policy, filepath.Join(testdata, "level1.yaml"))
	require.NoError(t, err)

	// Should have all 3 checks in order: level3 -> level2 -> level1
	assert.Len(t, policy.Queries, 3)
	assert.Equal(t, "level3-check", policy.Queries[0].UID)
	assert.Equal(t, "level2-check", policy.Queries[1].UID)
	assert.Equal(t, "level1-check", policy.Queries[2].UID)
}

func TestImportResolver_NoMatches(t *testing.T) {
	testdata := testdataDir()

	resolver := NewImportResolver(testdata)
	_, err := resolver.resolveGlobImport("nonexistent-*.yaml", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no files match pattern")
}

func TestImportResolver_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid YAML file
	err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte("{{invalid yaml"), 0o644)
	require.NoError(t, err)

	resolver := NewImportResolver(tmpDir)
	_, err = resolver.resolveFileImport("invalid.yaml", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse policy")
}

func TestImportResolver_FileNotFound(t *testing.T) {
	testdata := testdataDir()

	resolver := NewImportResolver(testdata)
	_, err := resolver.resolveFileImport("nonexistent.yaml", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestImportResolver_HTTPError(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resolver := NewImportResolver(".")
	_, err := resolver.resolveURLImport(server.URL + "/policy.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestLoadPolicies_WithImports(t *testing.T) {
	testdata := testdataDir()
	mainPath := filepath.Join(testdata, "main-policy.yaml")

	// Load policies
	policies, err := LoadPolicies(mainPath, "")
	require.NoError(t, err)
	require.Len(t, policies, 1)

	policy := policies[0]
	assert.Equal(t, "main-policy", policy.Metadata.Name)

	// Should have 3 queries: 2 imported + 1 local
	assert.Len(t, policy.Queries, 3)
	assert.Equal(t, "base-check-1", policy.Queries[0].UID)
	assert.Equal(t, "base-check-2", policy.Queries[1].UID)
	assert.Equal(t, "local-check", policy.Queries[2].UID)
}

func TestLoadPolicies_WithGlobImports(t *testing.T) {
	testdata := testdataDir()
	globPath := filepath.Join(testdata, "glob-policy.yaml")

	// Load policies
	policies, err := LoadPolicies(globPath, "")
	require.NoError(t, err)
	require.Len(t, policies, 1)

	policy := policies[0]
	assert.Equal(t, "glob-import-policy", policy.Metadata.Name)

	// Should have 3 queries from glob imports
	assert.Len(t, policy.Queries, 3)

	uids := make(map[string]bool)
	for _, q := range policy.Queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["storage-encryption"])
	assert.True(t, uids["network-firewall"])
	assert.True(t, uids["network-encryption"])
}

func TestLoadPolicies_WithNestedImports(t *testing.T) {
	testdata := filepath.Join(testdataDir(), "nested")
	level1Path := filepath.Join(testdata, "level1.yaml")

	// Load policies
	policies, err := LoadPolicies(level1Path, "")
	require.NoError(t, err)
	require.Len(t, policies, 1)

	policy := policies[0]
	assert.Equal(t, "level1-policy", policy.Metadata.Name)

	// Should have all 3 checks from nested imports
	assert.Len(t, policy.Queries, 3)
	assert.Equal(t, "level3-check", policy.Queries[0].UID)
	assert.Equal(t, "level2-check", policy.Queries[1].UID)
	assert.Equal(t, "level1-check", policy.Queries[2].UID)
}

func TestLoadPolicies_DirectoryWithImports(t *testing.T) {
	// Create temp directory structure for this test
	// (uses temp because we need multiple policy files in a directory)
	tmpDir := t.TempDir()
	sharedDir := filepath.Join(tmpDir, "shared")
	err := os.MkdirAll(sharedDir, 0o755)
	require.NoError(t, err)

	// Create shared checks
	sharedChecks := `
apiVersion: kspec/v1
kind: Policy
queries:
  - uid: shared-check
    title: Shared Check
    resource: test_resource
    query: "resource.shared == true"
`
	err = os.WriteFile(filepath.Join(sharedDir, "shared.yaml"), []byte(sharedChecks), 0o644)
	require.NoError(t, err)

	// Create policy that imports shared checks
	policy1 := `
apiVersion: kspec/v1
kind: Policy
metadata:
  name: policy-one
imports:
  - shared/shared.yaml
queries:
  - uid: policy1-check
    title: Policy 1 Check
    resource: test_resource
    query: "resource.policy1 == true"
`
	err = os.WriteFile(filepath.Join(tmpDir, "policy1.yaml"), []byte(policy1), 0o644)
	require.NoError(t, err)

	// Create another policy without imports
	policy2 := `
apiVersion: kspec/v1
kind: Policy
metadata:
  name: policy-two
queries:
  - uid: policy2-check
    title: Policy 2 Check
    resource: test_resource
    query: "resource.policy2 == true"
`
	err = os.WriteFile(filepath.Join(tmpDir, "policy2.yaml"), []byte(policy2), 0o644)
	require.NoError(t, err)

	// Load policies from directory
	policies, err := LoadPolicies("", tmpDir)
	require.NoError(t, err)
	require.Len(t, policies, 2)

	// Find policy-one and verify imports were resolved
	var policyOne *core.Policy
	for i := range policies {
		if policies[i].Metadata.Name == "policy-one" {
			policyOne = &policies[i]
			break
		}
	}
	require.NotNil(t, policyOne)
	assert.Len(t, policyOne.Queries, 2)
}
