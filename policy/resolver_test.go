// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

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
)

// testdataDir returns the path to the testdata directory.
func testdataDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get caller information")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
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

func TestResolver_ResolveFile(t *testing.T) {
	testdata := testdataDir()

	resolver := NewResolver(testdata)
	queries, err := resolver.resolveFile("base-checks.yaml", "")
	require.NoError(t, err)
	assert.Len(t, queries, 2)
	assert.Equal(t, "base-check-1", queries[0].UID)
	assert.Equal(t, "base-check-2", queries[1].UID)
	assert.Equal(t, "high", queries[0].Severity)
	assert.Equal(t, "medium", queries[1].Severity)
}

func TestResolver_ResolveGlob(t *testing.T) {
	testdata := testdataDir()

	resolver := NewResolver(testdata)
	queries, err := resolver.resolveGlob("checks/*.yaml", "")
	require.NoError(t, err)
	assert.Len(t, queries, 3) // 1 from storage-checks + 2 from network-checks

	uids := make(map[string]bool)
	for _, q := range queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["storage-encryption"], "should have storage-encryption check")
	assert.True(t, uids["network-firewall"], "should have network-firewall check")
	assert.True(t, uids["network-encryption"], "should have network-encryption check")
}

func TestResolver_ResolveURL(t *testing.T) {
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

	resolver := NewResolver(".")
	queries, err := resolver.resolveURL(server.URL + "/policy.yaml")
	require.NoError(t, err)
	assert.Len(t, queries, 1)
	assert.Equal(t, "remote-check", queries[0].UID)
}

func TestResolver_ResolveImports_MainPolicy(t *testing.T) {
	testdata := testdataDir()

	data, err := os.ReadFile(filepath.Join(testdata, "main-policy.yaml"))
	require.NoError(t, err)

	var p Policy
	err = yaml.Unmarshal(data, &p)
	require.NoError(t, err)

	assert.Equal(t, []string{"base-checks.yaml"}, p.Imports)

	resolver := NewResolver(testdata)
	err = resolver.ResolveImports(&p, filepath.Join(testdata, "main-policy.yaml"))
	require.NoError(t, err)

	// Should have 3 queries: 2 imported + 1 local
	assert.Len(t, p.Queries, 3)
	assert.Equal(t, "base-check-1", p.Queries[0].UID)
	assert.Equal(t, "base-check-2", p.Queries[1].UID)
	assert.Equal(t, "local-check", p.Queries[2].UID)
}

func TestResolver_ResolveImports_GlobPolicy(t *testing.T) {
	testdata := testdataDir()

	data, err := os.ReadFile(filepath.Join(testdata, "glob-policy.yaml"))
	require.NoError(t, err)

	var p Policy
	err = yaml.Unmarshal(data, &p)
	require.NoError(t, err)

	assert.Equal(t, []string{"checks/*.yaml"}, p.Imports)

	resolver := NewResolver(testdata)
	err = resolver.ResolveImports(&p, filepath.Join(testdata, "glob-policy.yaml"))
	require.NoError(t, err)

	assert.Len(t, p.Queries, 3)

	uids := make(map[string]bool)
	for _, q := range p.Queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["storage-encryption"])
	assert.True(t, uids["network-firewall"])
	assert.True(t, uids["network-encryption"])
}

func TestResolver_CircularImport(t *testing.T) {
	testdata := filepath.Join(testdataDir(), "circular")

	data, err := os.ReadFile(filepath.Join(testdata, "policy-a.yaml"))
	require.NoError(t, err)

	var p Policy
	err = yaml.Unmarshal(data, &p)
	require.NoError(t, err)

	resolver := NewResolver(testdata)
	err = resolver.ResolveImports(&p, filepath.Join(testdata, "policy-a.yaml"))
	require.NoError(t, err)

	assert.Len(t, p.Queries, 2)

	uids := make(map[string]bool)
	for _, q := range p.Queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["check-a"], "should have check-a")
	assert.True(t, uids["check-b"], "should have check-b")
}

func TestResolver_NestedImports(t *testing.T) {
	testdata := filepath.Join(testdataDir(), "nested")

	data, err := os.ReadFile(filepath.Join(testdata, "level1.yaml"))
	require.NoError(t, err)

	var p Policy
	err = yaml.Unmarshal(data, &p)
	require.NoError(t, err)

	resolver := NewResolver(testdata)
	err = resolver.ResolveImports(&p, filepath.Join(testdata, "level1.yaml"))
	require.NoError(t, err)

	assert.Len(t, p.Queries, 3)
	assert.Equal(t, "level3-check", p.Queries[0].UID)
	assert.Equal(t, "level2-check", p.Queries[1].UID)
	assert.Equal(t, "level1-check", p.Queries[2].UID)
}

func TestResolver_NoMatches(t *testing.T) {
	testdata := testdataDir()

	resolver := NewResolver(testdata)
	_, err := resolver.resolveGlob("nonexistent-*.yaml", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no files match pattern")
}

func TestResolver_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte("{{invalid yaml"), 0o644)
	require.NoError(t, err)

	resolver := NewResolver(tmpDir)
	_, err = resolver.resolveFile("invalid.yaml", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse policy")
}

func TestResolver_FileNotFound(t *testing.T) {
	testdata := testdataDir()

	resolver := NewResolver(testdata)
	_, err := resolver.resolveFile("nonexistent.yaml", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestResolver_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	resolver := NewResolver(".")
	_, err := resolver.resolveURL(server.URL + "/policy.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}
