// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
)

// mockResource is a simple resource spec for testing.
type mockResource struct {
	name           string
	childResources []string
	fetchFunc      func(ctx context.Context, asset core.Asset) ([]core.Resource, error)
}

func (m *mockResource) Name() string { return m.name }

func (m *mockResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, asset)
	}
	return []core.Resource{{"id": m.name + "-1", "name": m.name + " Instance"}}, nil
}

func (m *mockResource) ChildResources() []string {
	return m.childResources
}

// mockConnection implements core.Connection.
type mockConnection struct {
	resources  []core.ResourceSpec
	entryTypes map[string]string
}

func (c *mockConnection) Resources() []core.ResourceSpec { return c.resources }

func (c *mockConnection) EntryResourceType(assetType string) string {
	return c.entryTypes[assetType]
}

func TestTraverseWithChildResources(t *testing.T) {
	// Create mock resources with parent-child relationship
	orgResource := &mockResource{
		name:           "github_organization",
		childResources: []string{"github_repo", "github_team"},
		fetchFunc: func(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
			return []core.Resource{{"id": "org-1", "name": "Test Org"}}, nil
		},
	}

	repoResource := &mockResource{
		name: "github_repo",
		fetchFunc: func(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
			return []core.Resource{
				{"id": "repo-1", "name": "Repo 1"},
				{"id": "repo-2", "name": "Repo 2"},
			}, nil
		},
	}

	teamResource := &mockResource{
		name: "github_team",
		fetchFunc: func(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
			return []core.Resource{
				{"id": "team-1", "name": "Team 1"},
			}, nil
		},
	}

	// Build registry
	registry := map[string]core.ResourceSpec{
		"github_organization": orgResource,
		"github_repo":         repoResource,
		"github_team":         teamResource,
	}

	// Create discoverer
	d := &Discoverer{
		config: Config{
			ProviderName: "github",
			Asset: core.Asset{
				Type:   "github-org",
				Name:   "test-org",
				Config: map[string]string{},
			},
			Concurrency: 4,
		},
		registry: registry,
		connection: &mockConnection{
			resources:  []core.ResourceSpec{orgResource, repoResource, teamResource},
			entryTypes: map[string]string{"github-org": "github_organization"},
		},
	}

	// Create root node
	root := &common.ResourceNode{
		ID:           "test-org",
		Name:         "test-org",
		Type:         NodeTypeRoot,
		ResourceType: "github_organization",
		State:        common.AssetStatePending,
		Children:     []*common.ResourceNode{},
		SubResources: make(map[string][]*common.ResourceNode),
		Checks:       []common.CheckResult{},
	}

	// Run traversal
	ctx := context.Background()
	d.traverse(ctx, root, nil)

	// Verify results
	t.Logf("Root node: %s, Type: %s, Children: %d", root.Name, root.Type, len(root.Children))

	// Root should have 1 child (the org instance)
	require.GreaterOrEqual(t, len(root.Children), 1, "Root should have at least 1 child (org instance)")

	orgInstance := root.Children[0]
	t.Logf("Org instance: %s, Type: %s, Children: %d", orgInstance.Name, orgInstance.Type, len(orgInstance.Children))

	// Org instance should have 2 children (github_repo type, github_team type)
	assert.Equal(t, 2, len(orgInstance.Children), "Org instance should have 2 children (repo type + team type)")

	// Count total resource instances
	resourceCounts := make(map[string]int)
	countResourcesByType(root, resourceCounts)

	t.Logf("Resource counts: %+v", resourceCounts)

	// Should have 1 org + 2 repos + 1 team = 4 total instances
	assert.Equal(t, 1, resourceCounts["github_organization"], "Should have 1 organization")
	assert.Equal(t, 2, resourceCounts["github_repo"], "Should have 2 repos")
	assert.Equal(t, 1, resourceCounts["github_team"], "Should have 1 team")
}

func TestTraverseCountsNodes(t *testing.T) {
	// Simple test to verify countNodes works correctly
	root := &common.ResourceNode{
		ID:   "root",
		Type: NodeTypeRoot,
		Children: []*common.ResourceNode{
			{
				ID:   "child-1",
				Type: NodeTypeResourceInstance,
				Children: []*common.ResourceNode{
					{ID: "grandchild-1", Type: NodeTypeResourceInstance},
					{ID: "grandchild-2", Type: NodeTypeResourceInstance},
				},
			},
		},
		SubResources: map[string][]*common.ResourceNode{
			"sub_type": {
				{ID: "sub-1", Type: NodeTypeSubResourceInstance},
			},
		},
	}

	count := countNodes(root)
	assert.Equal(t, 5, count, "Should count all nodes including sub-resources")
}
