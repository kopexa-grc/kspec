// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/concurrency"
	"github.com/kopexa-grc/kspec/policy"
	"github.com/kopexa-grc/kspec/provider"
)

// Discoverer performs resource discovery and evaluation using graph traversal.
type Discoverer struct {
	config     Config
	connection core.Connection
	registry   map[string]core.ResourceSpec
	handler    EventHandler
	errors     []Error
	tree       *common.ResourceTree // Current tree for event emissions
}

// NewDiscoverer creates a new Discoverer with the given configuration.
func NewDiscoverer(config Config) *Discoverer {
	// Set defaults
	if config.Concurrency == 0 {
		config.Concurrency = concurrency.DefaultMaxWorkers
	}

	return &Discoverer{
		config: config,
	}
}

// OnEvent sets the event handler for discovery events.
func (d *Discoverer) OnEvent(handler EventHandler) {
	d.handler = handler
}

// emit sends an event to the handler if one is registered.
// It automatically includes the tree reference if not already set.
func (d *Discoverer) emit(event Event) {
	if d.handler != nil {
		// Always include the tree reference for UI updates
		if event.Tree == nil && d.tree != nil {
			event.Tree = d.tree
		}
		d.handler(event)
	}
}

// recordError adds an error to the discoverer's error list.
func (d *Discoverer) recordError(resourceType, message string, err error) {
	d.errors = append(d.errors, Error{
		ResourceType: resourceType,
		Message:      message,
		Err:          err,
	})
}

// Initialize connects to the provider and sets up the registry.
func (d *Discoverer) Initialize(ctx context.Context) error {
	conn, reg, err := provider.InitProvider(ctx, d.config.ProviderName, d.config.ProviderConfig)
	if err != nil {
		return fmt.Errorf("failed to init provider %s: %w", d.config.ProviderName, err)
	}
	d.connection = conn
	d.registry = reg
	return nil
}

// Registry returns the resource registry.
func (d *Discoverer) Registry() map[string]core.ResourceSpec {
	return d.registry
}

// traversalSetup contains common setup data for discovery/evaluation.
type traversalSetup struct {
	startTime time.Time
	root      *common.ResourceNode
}

// setupTraversal performs common setup for both Discover and Evaluate.
func (d *Discoverer) setupTraversal(asset core.Asset) *traversalSetup {
	d.config.Asset = asset
	d.errors = make([]Error, 0)

	root := d.createRootNode(asset)
	d.tree = &common.ResourceTree{
		Root:     root,
		Current:  root,
		AllNodes: map[string]*common.ResourceNode{root.ID: root},
	}

	d.emit(Event{Type: EventTreeCreated, Node: root, ResourceType: root.ResourceType})

	return &traversalSetup{
		startTime: time.Now(),
		root:      root,
	}
}

// finalizeTraversal marks traversal as complete.
func (d *Discoverer) finalizeTraversal(root *common.ResourceNode) {
	root.State = common.AssetStateComplete
}

// Discover performs resource discovery without policy evaluation.
// It traverses the resource graph starting from the asset entry point.
func (d *Discoverer) Discover(ctx context.Context, asset core.Asset) (*Result, error) {
	setup := d.setupTraversal(asset)

	d.emit(Event{Type: EventDiscoveryStarted, Node: setup.root, ResourceType: setup.root.ResourceType})
	d.emit(Event{Type: EventNodeScanning, Node: setup.root, ResourceType: setup.root.ResourceType})

	d.traverse(ctx, setup.root, nil)
	d.finalizeTraversal(setup.root)

	d.emit(Event{Type: EventDiscoveryComplete, Node: setup.root, ResourceType: setup.root.ResourceType})

	return d.buildDiscoverResult(setup.root, setup.startTime), nil
}

// Evaluate performs resource discovery with policy evaluation.
// It traverses the resource graph and evaluates policies on each node.
func (d *Discoverer) Evaluate(ctx context.Context, asset core.Asset, policies []policy.Policy) (*EvaluateResult, error) {
	setup := d.setupTraversal(asset)

	d.emit(Event{Type: EventEvaluateStarted, Node: setup.root, ResourceType: setup.root.ResourceType})
	d.emit(Event{Type: EventNodeScanning, Node: setup.root, ResourceType: setup.root.ResourceType})

	d.traverse(ctx, setup.root, policies)
	d.finalizeTraversal(setup.root)

	d.emit(Event{Type: EventEvaluateComplete, Node: setup.root, ResourceType: setup.root.ResourceType})

	return d.buildEvaluateResult(setup.root, setup.startTime), nil
}

// createRootNode creates the root node for traversal.
func (d *Discoverer) createRootNode(asset core.Asset) *common.ResourceNode {
	// Determine entry resource type from asset
	entryType := d.getEntryResourceType(asset)

	return &common.ResourceNode{
		ID:           asset.Name,
		Name:         asset.Name,
		Type:         NodeTypeRoot,
		ResourceType: entryType,
		State:        common.AssetStatePending,
		Children:     []*common.ResourceNode{},
		SubResources: make(map[string][]*common.ResourceNode),
		Checks:       []common.CheckResult{},
	}
}

// getEntryResourceType determines the entry point resource type for an asset.
func (d *Discoverer) getEntryResourceType(asset core.Asset) string {
	// Use the connection to get the entry resource type (provider-defined)
	if d.connection != nil {
		if entryType := d.connection.EntryResourceType(asset.Type); entryType != "" {
			return entryType
		}
	}

	// Fallback: try to find a matching resource type in registry
	for resType := range d.registry {
		if resType == asset.Type {
			return resType
		}
	}

	// Last resort: return asset type as-is
	return asset.Type
}

// buildDiscoverResult builds the result for a Discover operation.
func (d *Discoverer) buildDiscoverResult(root *common.ResourceNode, startTime time.Time) *Result {
	result := &Result{
		Provider:     d.config.ProviderName,
		AssetType:    d.config.Asset.Type,
		AssetName:    d.config.Asset.Name,
		DiscoveredAt: startTime,
		Mode:         ModeFetch,
		Duration:     time.Since(startTime),
		Resources:    make([]ResourceInfo, 0),
		Errors:       d.errors,
	}

	// Count resources by type
	resourceCounts := make(map[string]int)
	countResourcesByType(root, resourceCounts)

	for resType, count := range resourceCounts {
		info := ResourceInfo{
			Type:           resType,
			Count:          count,
			SupportsNative: true,
		}

		if d.config.IncludeInstances {
			info.Instances = d.extractInstancesFromNode(root, resType)
		}

		result.Resources = append(result.Resources, info)
		result.TotalCount += count
	}

	// Build graph if requested
	if d.config.IncludeInstances {
		result.Graph = d.buildGraphFromNode(root)
	}

	return result
}

// buildEvaluateResult builds the result for an Evaluate operation.
func (d *Discoverer) buildEvaluateResult(root *common.ResourceNode, startTime time.Time) *EvaluateResult {
	// Count checks
	passed, failed, skipped := countChecks(root)

	return &EvaluateResult{
		Provider:      d.config.ProviderName,
		Asset:         d.config.Asset.Name,
		AssetType:     d.config.Asset.Type,
		Root:          root,
		TotalCount:    countNodes(root),
		Duration:      time.Since(startTime),
		TotalChecks:   passed + failed + skipped,
		PassedChecks:  passed,
		FailedChecks:  failed,
		SkippedChecks: skipped,
		Errors:        d.errors,
	}
}

// countResourcesByType counts resources by type recursively.
func countResourcesByType(node *common.ResourceNode, counts map[string]int) {
	if node.Type == NodeTypeResourceInstance {
		counts[node.ResourceType]++
	}

	for _, child := range node.Children {
		countResourcesByType(child, counts)
	}

	for _, subNodes := range node.SubResources {
		for _, subNode := range subNodes {
			counts[subNode.ResourceType]++
		}
	}
}

// countNodes counts total nodes in the tree.
func countNodes(node *common.ResourceNode) int {
	count := 1

	for _, child := range node.Children {
		count += countNodes(child)
	}

	for _, subNodes := range node.SubResources {
		count += len(subNodes)
	}

	return count
}

// countChecks counts check results recursively.
func countChecks(node *common.ResourceNode) (passed, failed, skipped int) {
	passed += node.ChecksPassed
	failed += node.ChecksFailed
	skipped += node.ChecksSkipped

	for _, child := range node.Children {
		p, f, s := countChecks(child)
		passed += p
		failed += f
		skipped += s
	}

	for _, subNodes := range node.SubResources {
		for _, subNode := range subNodes {
			passed += subNode.ChecksPassed
			failed += subNode.ChecksFailed
			skipped += subNode.ChecksSkipped
		}
	}

	return passed, failed, skipped
}

// extractInstancesFromNode extracts resource instances of a specific type.
func (d *Discoverer) extractInstancesFromNode(node *common.ResourceNode, resType string) []ResourceInstance {
	instances := make([]ResourceInstance, 0)

	if node.Type == "resource_instance" && node.ResourceType == resType {
		inst := ResourceInstance{
			ID:   node.ID,
			Name: node.Name,
			Type: node.ResourceType,
		}
		if node.Metadata != nil {
			inst.Metadata = node.Metadata
			inst.ID = extractID(node.Metadata, node.ID)
			inst.Name = extractName(node.Metadata, node.Name)
		}
		instances = append(instances, inst)
	}

	for _, child := range node.Children {
		instances = append(instances, d.extractInstancesFromNode(child, resType)...)
	}

	return instances
}

// buildGraphFromNode builds a graph from a resource node.
func (d *Discoverer) buildGraphFromNode(root *common.ResourceNode) *Graph {
	graph := &Graph{
		Nodes: make([]GraphNode, 0),
		Edges: make([]GraphEdge, 0),
	}

	// Add root node
	rootGraphNode := GraphNode{
		ID:    root.ID,
		Label: root.Name,
		Type:  root.Type,
		Group: "root",
	}
	graph.Nodes = append(graph.Nodes, rootGraphNode)

	// Add all nodes recursively
	d.addNodeToGraph(root, graph)

	return graph
}

// addNodeToGraph adds a node and its children to the graph.
func (d *Discoverer) addNodeToGraph(parent *common.ResourceNode, graph *Graph) {
	for _, child := range parent.Children {
		node := GraphNode{
			ID:    child.ID,
			Label: child.Name,
			Type:  child.ResourceType,
			Group: GetResourceGroup(d.config.ProviderName, child.ResourceType),
		}

		if child.Metadata != nil {
			node.Metadata = make(map[string]interface{})
			for _, key := range []string{"id", "name", "arn"} {
				if val, ok := child.Metadata[key]; ok {
					node.Metadata[key] = val
				}
			}
		}

		graph.Nodes = append(graph.Nodes, node)
		graph.Edges = append(graph.Edges, GraphEdge{
			Source:   parent.ID,
			Target:   child.ID,
			Relation: "contains",
		})

		d.addNodeToGraph(child, graph)
	}

	for subType, subNodes := range parent.SubResources {
		for _, subNode := range subNodes {
			node := GraphNode{
				ID:    subNode.ID,
				Label: subNode.Name,
				Type:  subType,
				Group: GetResourceGroup(d.config.ProviderName, subType),
			}
			graph.Nodes = append(graph.Nodes, node)
			graph.Edges = append(graph.Edges, GraphEdge{
				Source:   parent.ID,
				Target:   subNode.ID,
				Relation: "contains",
			})
		}
	}
}

// Helper functions

func extractID(metadata map[string]interface{}, fallback string) string {
	for _, key := range []string{"id", "arn", "uid"} {
		if val, ok := metadata[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return fallback
}

func extractName(metadata map[string]interface{}, fallback string) string {
	for _, key := range []string{"name", "login", "full_name", "display_name"} {
		if val, ok := metadata[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return fallback
}

// getResourceName extracts a display name from a resource.
func getResourceName(resource core.Resource, resourceType string, index int) string {
	if name, ok := resource["name"]; ok {
		return fmt.Sprintf("%v", name)
	}
	if id, ok := resource["id"]; ok {
		return fmt.Sprintf("%v", id)
	}
	if login, ok := resource["login"]; ok {
		return fmt.Sprintf("%v", login)
	}
	if fullName, ok := resource["full_name"]; ok {
		return fmt.Sprintf("%v", fullName)
	}
	return fmt.Sprintf("%s-%d", resourceType, index)
}
