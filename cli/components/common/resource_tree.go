// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package common

import (
	"fmt"
	"strings"
)

// ResourceNode represents a node in the hierarchical resource tree
type ResourceNode struct {
	// Identity
	ID           string
	Name         string
	Type         string // "subscription", "resource_type", "resource_instance", "sub_resource_type", "sub_resource_instance"
	ResourceType string // e.g., "azure_sql_server", "azure_sql_database"

	// Hierarchy
	Level    int
	Path     []string // Breadcrumb path for display
	Parent   *ResourceNode
	Children []*ResourceNode

	// Sub-resources (for resource types that have them)
	// Map of sub-resource type name -> list of instances
	SubResources map[string][]*ResourceNode

	// Discovery data
	ResourceCount     int            // Number of resources of this type
	SubResourceCounts map[string]int // Sub-resource type -> count

	// Scan state
	State         AssetState
	ChecksPassed  int
	ChecksFailed  int
	ChecksSkipped int
	Error         error

	// Check results
	Checks []CheckResult

	// Additional metadata
	Metadata map[string]interface{}
}

// ResourceTree manages the entire hierarchical resource structure
type ResourceTree struct {
	Root     *ResourceNode
	Current  *ResourceNode
	AllNodes map[string]*ResourceNode // ID -> Node for quick lookup
}

// NewResourceTree creates a new resource tree with a root node
func NewResourceTree(rootName, rootType string) *ResourceTree {
	root := &ResourceNode{
		ID:                rootName,
		Name:              rootName,
		Type:              rootType,
		Level:             0,
		Path:              []string{rootName},
		Children:          []*ResourceNode{},
		SubResources:      make(map[string][]*ResourceNode),
		SubResourceCounts: make(map[string]int),
		State:             AssetStatePending,
		Metadata:          make(map[string]interface{}),
	}

	return &ResourceTree{
		Root:     root,
		Current:  root,
		AllNodes: map[string]*ResourceNode{root.ID: root},
	}
}

// AddNode adds a new node to the tree as a child of the specified parent
func (t *ResourceTree) AddNode(parentID string, node *ResourceNode) error {
	parent, exists := t.AllNodes[parentID]
	if !exists {
		return fmt.Errorf("parent node with ID %s not found", parentID)
	}

	// Set parent and level
	node.Parent = parent
	node.Level = parent.Level + 1

	// Build path
	node.Path = append([]string{}, parent.Path...)
	node.Path = append(node.Path, node.Name)

	// Add to parent's children
	parent.Children = append(parent.Children, node)

	// Add to lookup map
	t.AllNodes[node.ID] = node

	return nil
}

// AddSubResource adds a sub-resource instance to a parent node
func (t *ResourceTree) AddSubResource(parentID, subResourceType string, node *ResourceNode) error {
	parent, exists := t.AllNodes[parentID]
	if !exists {
		return fmt.Errorf("parent node with ID %s not found", parentID)
	}

	// Set parent and level
	node.Parent = parent
	node.Level = parent.Level + 1

	// Build path
	node.Path = append([]string{}, parent.Path...)
	node.Path = append(node.Path, subResourceType, node.Name)

	// Initialize sub-resources map if needed
	if parent.SubResources == nil {
		parent.SubResources = make(map[string][]*ResourceNode)
	}

	// Add to parent's sub-resources
	parent.SubResources[subResourceType] = append(parent.SubResources[subResourceType], node)

	// Add to lookup map
	t.AllNodes[node.ID] = node

	return nil
}

// NavigateDown navigates to a child node by ID
func (t *ResourceTree) NavigateDown(nodeID string) error {
	if t.Current == nil {
		return fmt.Errorf("current node is nil")
	}

	node, exists := t.AllNodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s not found", nodeID)
	}

	// Check if node is a child or sub-resource of current node
	isChild := false
	for _, child := range t.Current.Children {
		if child.ID == nodeID {
			isChild = true
			break
		}
	}

	if !isChild {
		// Check sub-resources
		for _, subResources := range t.Current.SubResources {
			for _, subRes := range subResources {
				if subRes.ID == nodeID {
					isChild = true
					break
				}
			}
			if isChild {
				break
			}
		}
	}

	if !isChild {
		return fmt.Errorf("node %s is not a child of current node", nodeID)
	}

	t.Current = node
	return nil
}

// NavigateUp navigates to the parent node
func (t *ResourceTree) NavigateUp() error {
	if t.Current == nil {
		return fmt.Errorf("current node is nil")
	}
	if t.Current.Parent == nil {
		return fmt.Errorf("already at root node")
	}

	t.Current = t.Current.Parent
	return nil
}

// NavigateToRoot navigates back to the root node
func (t *ResourceTree) NavigateToRoot() {
	t.Current = t.Root
}

// GetBreadcrumb returns the breadcrumb path for the current node
func (t *ResourceTree) GetBreadcrumb() []string {
	if t.Current == nil {
		return nil
	}
	return t.Current.Path
}

// GetBreadcrumbString returns the breadcrumb as a formatted string
func (t *ResourceTree) GetBreadcrumbString() string {
	if t.Current == nil {
		return ""
	}
	return strings.Join(t.Current.Path, " > ")
}

// UpdateNodeState updates the state and check counts for a node
func (t *ResourceTree) UpdateNodeState(nodeID string, state AssetState, passed, failed, skipped int) error {
	node, exists := t.AllNodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s not found", nodeID)
	}

	node.State = state
	node.ChecksPassed = passed
	node.ChecksFailed = failed
	node.ChecksSkipped = skipped

	return nil
}

// AddCheckResult adds a check result to a node
func (t *ResourceTree) AddCheckResult(nodeID string, check CheckResult) error {
	node, exists := t.AllNodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s not found", nodeID)
	}

	node.Checks = append(node.Checks, check)

	// Update check counts
	switch check.Status {
	case "passed":
		node.ChecksPassed++
	case "failed":
		node.ChecksFailed++
	case "skipped":
		node.ChecksSkipped++
	}

	return nil
}

// GetNode retrieves a node by ID
func (t *ResourceTree) GetNode(nodeID string) (*ResourceNode, bool) {
	node, exists := t.AllNodes[nodeID]
	return node, exists
}

// GetCurrentChildren returns all navigable items from the current node
// This includes both direct children and sub-resource types
func (t *ResourceTree) GetCurrentChildren() []*ResourceNode {
	children := make([]*ResourceNode, 0)

	// Safety check for nil Current
	if t.Current == nil {
		return children
	}

	// Add direct children
	children = append(children, t.Current.Children...)

	// Add sub-resource type nodes (one representative node per sub-resource type)
	for subResType, subResInstances := range t.Current.SubResources {
		if len(subResInstances) > 0 {
			// Create a virtual node for the sub-resource type
			typeNode := &ResourceNode{
				ID:            fmt.Sprintf("%s-subtype-%s", t.Current.ID, subResType),
				Name:          subResType,
				Type:          "sub_resource_type_group",
				ResourceType:  subResType,
				Level:         t.Current.Level + 1,
				Path:          append(t.Current.Path, subResType),
				ResourceCount: len(subResInstances),
				Parent:        t.Current,
			}
			children = append(children, typeNode)
		}
	}

	return children
}

// HasChildren returns true if the current node has navigable children
func (t *ResourceTree) HasChildren() bool {
	if t.Current == nil {
		return false
	}
	return len(t.Current.Children) > 0 || len(t.Current.SubResources) > 0
}

// GetSubResourceInstances returns all instances of a specific sub-resource type
func (t *ResourceTree) GetSubResourceInstances(subResourceType string) []*ResourceNode {
	if t.Current == nil {
		return []*ResourceNode{}
	}
	if instances, exists := t.Current.SubResources[subResourceType]; exists {
		return instances
	}
	return []*ResourceNode{}
}

// TotalChecks returns the total number of checks for a node
func (n *ResourceNode) TotalChecks() int {
	return n.ChecksPassed + n.ChecksFailed + n.ChecksSkipped
}

// HasChecks returns true if the node has any check results
func (n *ResourceNode) HasChecks() bool {
	return len(n.Checks) > 0
}

// StatusString returns a formatted status string for display
func (n *ResourceNode) StatusString() string {
	return string(n.State)
}

// ChecksString returns a formatted check results string
func (n *ResourceNode) ChecksString() string {
	return fmt.Sprintf("✓%d ✗%d ⊘%d", n.ChecksPassed, n.ChecksFailed, n.ChecksSkipped)
}
