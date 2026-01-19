// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithImports(t *testing.T) {
	testdata := testdataDir()
	mainPath := filepath.Join(testdata, "main-policy.yaml")

	policies, err := Load(mainPath, "")
	require.NoError(t, err)
	require.Len(t, policies, 1)

	p := policies[0]
	assert.Equal(t, "main-policy", p.Metadata.Name)

	// Should have 3 queries: 2 imported + 1 local
	assert.Len(t, p.Queries, 3)
	assert.Equal(t, "base-check-1", p.Queries[0].UID)
	assert.Equal(t, "base-check-2", p.Queries[1].UID)
	assert.Equal(t, "local-check", p.Queries[2].UID)
}

func TestLoad_WithGlobImports(t *testing.T) {
	testdata := testdataDir()
	globPath := filepath.Join(testdata, "glob-policy.yaml")

	policies, err := Load(globPath, "")
	require.NoError(t, err)
	require.Len(t, policies, 1)

	p := policies[0]
	assert.Equal(t, "glob-import-policy", p.Metadata.Name)
	assert.Len(t, p.Queries, 3)

	uids := make(map[string]bool)
	for _, q := range p.Queries {
		uids[q.UID] = true
	}
	assert.True(t, uids["storage-encryption"])
	assert.True(t, uids["network-firewall"])
	assert.True(t, uids["network-encryption"])
}

func TestLoad_WithNestedImports(t *testing.T) {
	testdata := filepath.Join(testdataDir(), "nested")
	level1Path := filepath.Join(testdata, "level1.yaml")

	policies, err := Load(level1Path, "")
	require.NoError(t, err)
	require.Len(t, policies, 1)

	p := policies[0]
	assert.Equal(t, "level1-policy", p.Metadata.Name)
	assert.Len(t, p.Queries, 3)
	assert.Equal(t, "level3-check", p.Queries[0].UID)
	assert.Equal(t, "level2-check", p.Queries[1].UID)
	assert.Equal(t, "level1-check", p.Queries[2].UID)
}

func TestLoad_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	policyContent := `
apiVersion: kspec/v1
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
    resource: test_resource
    query: "true"
`
	policyPath := filepath.Join(tmpDir, "test-policy.yaml")
	err := os.WriteFile(policyPath, []byte(policyContent), 0o644)
	require.NoError(t, err)

	policies, err := Load(policyPath, "")
	require.NoError(t, err)
	require.Len(t, policies, 1)
	assert.Equal(t, "test-policy", policies[0].Metadata.Name)
}

func TestLoad_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	policy1 := `
apiVersion: kspec/v1
kind: Policy
metadata:
  name: policy-one
queries:
  - uid: check-one
`
	policy2 := `
apiVersion: kspec/v1
kind: Policy
metadata:
  name: policy-two
queries:
  - uid: check-two
`
	err := os.WriteFile(filepath.Join(tmpDir, "policy1.yaml"), []byte(policy1), 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "policy2.yaml"), []byte(policy2), 0o644)
	require.NoError(t, err)

	policies, err := Load("", tmpDir)
	require.NoError(t, err)
	assert.Len(t, policies, 2)
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/policy.yaml", "")
	assert.Error(t, err)
}

func TestLoad_NonExistentDirectory(t *testing.T) {
	_, err := Load("", "/nonexistent/directory")
	assert.Error(t, err)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(policyPath, []byte("{{invalid yaml"), 0o644)
	require.NoError(t, err)

	_, err = Load(policyPath, "")
	assert.Error(t, err)
}

func TestLoadFromBytes(t *testing.T) {
	data := []byte(`
apiVersion: kspec/v1
kind: Policy
metadata:
  name: bytes-policy
queries:
  - uid: test-check
    title: Test
`)

	p, err := LoadFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, "bytes-policy", p.Metadata.Name)
	assert.Len(t, p.Queries, 1)
}

func TestLoadFromFile(t *testing.T) {
	testdata := testdataDir()
	p, err := LoadFromFile(filepath.Join(testdata, "base-checks.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "base-security-checks", p.Metadata.Name)
	assert.Len(t, p.Queries, 2)
}
