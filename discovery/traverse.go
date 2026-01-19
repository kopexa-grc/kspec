// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"context"
	"fmt"
	"sync"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/policy"
)

// workItem represents a unit of work to be processed.
type workItem struct {
	node *common.ResourceNode
	spec core.ResourceSpec
}

// traverseContext holds the context for a graph traversal.
type traverseContext struct {
	ctx      context.Context
	d        *Discoverer
	policies []policy.Policy

	// Work queue and synchronization
	workQueue chan workItem
	wg        sync.WaitGroup
	mu        sync.Mutex // Protects node modifications

	// Semaphore for concurrency control
	sem chan struct{}
}

// traverse performs a graph traversal starting from the given node.
// If policies is not nil, policy evaluation is performed on each node.
func (d *Discoverer) traverse(ctx context.Context, root *common.ResourceNode, policies []policy.Policy) {
	tc := &traverseContext{
		ctx:       ctx,
		d:         d,
		policies:  policies,
		workQueue: make(chan workItem, 1000), // Buffered queue for work items
		sem:       make(chan struct{}, d.config.Concurrency),
	}

	// Start workers
	numWorkers := d.config.Concurrency
	for range numWorkers {
		go tc.worker()
	}

	// Get root spec
	rootSpec, ok := d.registry[root.ResourceType]
	if !ok {
		tc.setNodeError(root, fmt.Errorf("unknown resource type: %s", root.ResourceType))
		close(tc.workQueue)
		return
	}

	// Submit root node
	tc.wg.Add(1)
	tc.workQueue <- workItem{node: root, spec: rootSpec}

	// Wait for all work to complete
	tc.wg.Wait()

	// Close work queue to stop workers
	close(tc.workQueue)
}

// worker processes work items from the queue.
func (tc *traverseContext) worker() {
	for item := range tc.workQueue {
		tc.processWorkItem(item)
	}
}

// processWorkItem processes a single work item.
func (tc *traverseContext) processWorkItem(item workItem) {
	defer tc.wg.Done()

	// Acquire semaphore with context cancellation support
	select {
	case tc.sem <- struct{}{}:
		defer func() { <-tc.sem }()
	case <-tc.ctx.Done():
		return
	}

	// Check context again after acquiring semaphore
	if tc.ctx.Err() != nil {
		return
	}

	tc.traverseNode(item.node, item.spec)
}

// traverseNode processes a single node and discovers its children.
func (tc *traverseContext) traverseNode(node *common.ResourceNode, spec core.ResourceSpec) {
	// Instance nodes already have metadata - don't fetch again
	if node.Type == NodeTypeResourceInstance {
		tc.setNodeState(node, common.AssetStateScanning)
		tc.d.emit(Event{Type: EventNodeScanning, Node: node, ResourceType: node.ResourceType})
		tc.handleInstanceNode(node, spec)
		return
	}

	// Mark as scanning
	tc.setNodeState(node, common.AssetStateScanning)
	tc.d.emit(Event{Type: EventNodeScanning, Node: node, ResourceType: node.ResourceType})

	// Build asset for fetching
	asset := tc.d.buildAssetForNode(node)

	// Fetch resources
	resources, err := spec.Fetch(tc.ctx, asset)
	if err != nil {
		tc.setNodeError(node, err)
		tc.d.emit(Event{Type: EventNodeError, Node: node, ResourceType: node.ResourceType, Error: err})
		return
	}

	// Handle the fetched resources based on node type
	tc.handleFetchedResources(node, spec, resources)
}

// handleFetchedResources processes fetched resources based on node type.
func (tc *traverseContext) handleFetchedResources(node *common.ResourceNode, spec core.ResourceSpec, resources []core.Resource) {
	switch node.Type {
	case NodeTypeRoot, NodeTypeResourceType:
		// This node represents a resource type - resources are the instances
		tc.handleResourceTypeNode(node, spec, resources)

	case NodeTypeResourceInstance:
		// This node is an instance - resource is already set, discover sub-resources
		tc.handleInstanceNode(node, spec)

	default:
		// Single resource node (e.g., organization)
		if len(resources) > 0 {
			tc.mu.Lock()
			node.Metadata = resources[0]
			tc.mu.Unlock()
		}
		tc.processNodeComplete(node, spec)
	}
}

// handleResourceTypeNode handles nodes that represent a resource type (e.g., "github_repo").
// The fetched resources become child instances.
func (tc *traverseContext) handleResourceTypeNode(node *common.ResourceNode, spec core.ResourceSpec, resources []core.Resource) {
	tc.mu.Lock()
	node.ResourceCount = len(resources)
	tc.mu.Unlock()

	// Create instance nodes for each resource and submit them
	for i, res := range resources {
		instanceNode := tc.createInstanceNode(node, res, i)

		tc.mu.Lock()
		node.Children = append(node.Children, instanceNode)
		tc.mu.Unlock()

		// Register node for navigation
		tc.registerNode(instanceNode)

		// Submit instance for processing
		tc.submitWork(instanceNode, spec)
	}

	// Mark resource type node as complete
	tc.setNodeState(node, common.AssetStateComplete)
	tc.d.emit(Event{Type: EventNodeComplete, Node: node, ResourceType: node.ResourceType})
}

// handleInstanceNode handles a resource instance node.
func (tc *traverseContext) handleInstanceNode(node *common.ResourceNode, spec core.ResourceSpec) {
	// Policy evaluation (synchronous - fast operation)
	if tc.policies != nil && node.Metadata != nil {
		tc.evaluateNode(node)
	}

	// Discover and traverse children
	tc.discoverChildren(node, spec)

	// Mark as complete
	tc.setNodeState(node, common.AssetStateComplete)
	tc.d.emit(Event{Type: EventNodeComplete, Node: node, ResourceType: node.ResourceType})
}

// processNodeComplete processes a node that represents a single resource.
func (tc *traverseContext) processNodeComplete(node *common.ResourceNode, spec core.ResourceSpec) {
	// Policy evaluation (synchronous - fast operation)
	if tc.policies != nil && node.Metadata != nil {
		tc.evaluateNode(node)
	}

	// Discover and traverse children
	tc.discoverChildren(node, spec)

	// Mark as complete
	tc.setNodeState(node, common.AssetStateComplete)
	tc.d.emit(Event{Type: EventNodeComplete, Node: node, ResourceType: node.ResourceType})
}

// submitWork adds a work item to the queue.
func (tc *traverseContext) submitWork(node *common.ResourceNode, spec core.ResourceSpec) {
	tc.wg.Add(1)
	tc.workQueue <- workItem{node: node, spec: spec}
}

// evaluateNode evaluates policies on a node.
func (tc *traverseContext) evaluateNode(node *common.ResourceNode) {
	if node.Metadata == nil {
		return
	}

	results := policy.Evaluate(
		node.Metadata,
		node.ResourceType,
		tc.policies,
		tc.d.registry,
		tc.d.config.Asset,
	)

	tc.mu.Lock()
	node.Checks = results
	for _, r := range results {
		switch r.Status {
		case policy.StatusPassed:
			node.ChecksPassed++
		case policy.StatusFailed:
			node.ChecksFailed++
		case policy.StatusSkipped:
			node.ChecksSkipped++
		}
	}
	tc.mu.Unlock()
}

// discoverChildren discovers child resources and submits them for traversal.
func (tc *traverseContext) discoverChildren(node *common.ResourceNode, spec core.ResourceSpec) {
	// Check for ChildResourceProvider (top-level children like Org → Teams, Repos)
	if childProvider, ok := spec.(core.ChildResourceProvider); ok {
		childTypes := childProvider.ChildResources()
		for _, childType := range childTypes {
			childSpec, exists := tc.d.registry[childType]
			if !exists {
				continue
			}

			childNode := &common.ResourceNode{
				ID:           fmt.Sprintf("%s-%s", node.ID, childType),
				Name:         childType,
				Type:         NodeTypeResourceType,
				ResourceType: childType,
				State:        common.AssetStatePending,
				Parent:       node,
			}

			tc.mu.Lock()
			node.Children = append(node.Children, childNode)
			tc.mu.Unlock()

			// Register node for navigation
			tc.registerNode(childNode)

			// Submit child resource type for traversal
			tc.submitWork(childNode, childSpec)
		}
	}

	// Check for SubResourceProvider (sub-resources like Repo → Branches)
	if subProvider, ok := spec.(core.SubResourceProvider); ok {
		subSpecs := subProvider.SubResources()
		for _, subSpec := range subSpecs {
			tc.traverseSubResource(node, subSpec)
		}
	}
}

// traverseSubResource fetches and processes sub-resources.
func (tc *traverseContext) traverseSubResource(parentNode *common.ResourceNode, subSpec core.ResourceSpec) {
	subResType := subSpec.Name()

	// Build asset for sub-resource fetching
	asset := tc.d.buildAssetForSubResource(parentNode, subResType)

	// Fetch sub-resources
	subResources, err := subSpec.Fetch(tc.ctx, asset)
	if err != nil {
		tc.d.recordError(subResType, err.Error(), err)
		return
	}

	// Process each sub-resource
	for i, subRes := range subResources {
		subNode := &common.ResourceNode{
			ID:           fmt.Sprintf("%s-sub-%s-%d", parentNode.ID, subResType, i),
			Name:         getResourceName(subRes, subResType, i),
			Type:         NodeTypeSubResourceInstance,
			ResourceType: subResType,
			State:        common.AssetStateScanning,
			Parent:       parentNode,
			Metadata:     subRes,
		}

		// Add to parent's sub-resources
		tc.mu.Lock()
		if parentNode.SubResources == nil {
			parentNode.SubResources = make(map[string][]*common.ResourceNode)
		}
		parentNode.SubResources[subResType] = append(parentNode.SubResources[subResType], subNode)
		tc.mu.Unlock()

		// Register node for navigation
		tc.registerNode(subNode)

		// Evaluate policies if needed
		if tc.policies != nil {
			tc.evaluateNode(subNode)
		}

		tc.setNodeState(subNode, common.AssetStateComplete)
	}
}

// createInstanceNode creates a new instance node from a resource.
func (tc *traverseContext) createInstanceNode(parent *common.ResourceNode, res core.Resource, index int) *common.ResourceNode {
	return &common.ResourceNode{
		ID:           fmt.Sprintf("%s-%d", parent.ID, index),
		Name:         getResourceName(res, parent.ResourceType, index),
		Type:         NodeTypeResourceInstance,
		ResourceType: parent.ResourceType,
		State:        common.AssetStatePending,
		Parent:       parent,
		Metadata:     res,
		Checks:       []common.CheckResult{},
	}
}

// setNodeState sets the state of a node (thread-safe).
func (tc *traverseContext) setNodeState(node *common.ResourceNode, state common.AssetState) {
	tc.mu.Lock()
	node.State = state
	tc.mu.Unlock()
}

// registerNode adds a node to the tree's AllNodes map for navigation.
func (tc *traverseContext) registerNode(node *common.ResourceNode) {
	if tc.d.tree != nil && tc.d.tree.AllNodes != nil {
		tc.mu.Lock()
		tc.d.tree.AllNodes[node.ID] = node
		tc.mu.Unlock()
	}
}

// setNodeError sets an error on a node (thread-safe).
func (tc *traverseContext) setNodeError(node *common.ResourceNode, err error) {
	tc.mu.Lock()
	node.State = common.AssetStateError
	node.Error = err
	tc.mu.Unlock()
	tc.d.recordError(node.ResourceType, err.Error(), err)
}

// buildAssetForNode builds an asset config for fetching a node's resources.
func (d *Discoverer) buildAssetForNode(node *common.ResourceNode) core.Asset {
	asset := core.Asset{
		Type:   d.config.Asset.Type,
		Name:   d.config.Asset.Name,
		Config: make(map[string]string),
	}

	// Copy base config
	for k, v := range d.config.Asset.Config {
		asset.Config[k] = v
	}

	return asset
}

// buildAssetForSubResource builds an asset config for fetching sub-resources.
func (d *Discoverer) buildAssetForSubResource(parent *common.ResourceNode, subResType string) core.Asset {
	baseAsset := d.buildAssetForNode(parent)

	// Check if the parent resource spec implements AssetContextProvider
	if parent.Metadata != nil {
		if spec, ok := d.registry[parent.ResourceType]; ok {
			if contextProvider, ok := spec.(core.AssetContextProvider); ok {
				return contextProvider.BuildChildAsset(parent.Metadata, baseAsset)
			}
		}
	}

	return baseAsset
}
