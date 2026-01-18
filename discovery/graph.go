// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"fmt"
)

// ToGraph converts a Result to a Graph.
// This is useful when you have a result from native/registry mode
// and want to add graph data after fetching instances.
func (r *Result) ToGraph() *Graph {
	if r.Graph != nil {
		return r.Graph
	}

	graph := &Graph{
		Nodes: make([]GraphNode, 0),
		Edges: make([]GraphEdge, 0),
	}

	// Add root node
	rootType := GetRootType(r.Provider)
	rootID := fmt.Sprintf("%s-%s", rootType, r.AssetName)

	rootNode := GraphNode{
		ID:    rootID,
		Label: r.AssetName,
		Type:  rootType,
		Group: GetResourceGroup(r.Provider, rootType),
	}
	graph.Nodes = append(graph.Nodes, rootNode)

	// Add resource type nodes (summary nodes when no instances available)
	if !r.hasInstances() {
		for _, resInfo := range r.Resources {
			if resInfo.Count <= 0 {
				continue
			}

			node := GraphNode{
				ID:    fmt.Sprintf("%s-summary", resInfo.Type),
				Label: fmt.Sprintf("%s (%d)", resInfo.Type, resInfo.Count),
				Type:  resInfo.Type,
				Group: GetResourceGroup(r.Provider, resInfo.Type),
				Metadata: map[string]interface{}{
					"count":      resInfo.Count,
					"is_summary": true,
				},
			}
			graph.Nodes = append(graph.Nodes, node)

			// Connect to root
			edge := GraphEdge{
				Source:   rootID,
				Target:   node.ID,
				Relation: GetRelationType(rootType, resInfo.Type),
			}
			graph.Edges = append(graph.Edges, edge)
		}
	} else {
		// Add actual instance nodes
		nodeIDs := make(map[string]bool)
		nodeIDs[rootID] = true

		for _, resInfo := range r.Resources {
			for _, inst := range resInfo.Instances {
				node := GraphNode{
					ID:    inst.ID,
					Label: inst.Name,
					Type:  inst.Type,
					Group: GetResourceGroup(r.Provider, inst.Type),
				}
				graph.Nodes = append(graph.Nodes, node)
				nodeIDs[inst.ID] = true

				// Create edge
				parentID := inst.ParentID
				if parentID == "" {
					parentID = rootID
				}

				parentType := GetParentType(r.Provider, inst.Type)
				if parentType == "" {
					parentType = rootType
				}

				edge := GraphEdge{
					Source:   parentID,
					Target:   inst.ID,
					Relation: GetRelationType(parentType, inst.Type),
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}

	r.Graph = graph
	return graph
}

// hasInstances checks if the result contains any resource instances.
func (r *Result) hasInstances() bool {
	for _, resInfo := range r.Resources {
		if len(resInfo.Instances) > 0 {
			return true
		}
	}
	return false
}

// GetResourcesByType returns all resources of a specific type.
func (r *Result) GetResourcesByType(resourceType string) []ResourceInstance {
	for _, resInfo := range r.Resources {
		if resInfo.Type == resourceType {
			return resInfo.Instances
		}
	}
	return nil
}

// GetResourceCount returns the count of a specific resource type.
// Returns -1 if the resource type is not found.
func (r *Result) GetResourceCount(resourceType string) int {
	for _, resInfo := range r.Resources {
		if resInfo.Type == resourceType {
			return resInfo.Count
		}
	}
	return -1
}

// HasErrors returns true if there were any errors during discovery.
func (r *Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// GraphStats returns statistics about the discovery graph.
type GraphStats struct {
	TotalNodes      int            `json:"total_nodes"`
	TotalEdges      int            `json:"total_edges"`
	NodesByType     map[string]int `json:"nodes_by_type"`
	NodesByGroup    map[string]int `json:"nodes_by_group"`
	EdgesByRelation map[string]int `json:"edges_by_relation"`
}

// Stats returns statistics about the graph.
func (g *Graph) Stats() GraphStats {
	stats := GraphStats{
		TotalNodes:      len(g.Nodes),
		TotalEdges:      len(g.Edges),
		NodesByType:     make(map[string]int),
		NodesByGroup:    make(map[string]int),
		EdgesByRelation: make(map[string]int),
	}

	for _, node := range g.Nodes {
		stats.NodesByType[node.Type]++
		stats.NodesByGroup[node.Group]++
	}

	for _, edge := range g.Edges {
		stats.EdgesByRelation[edge.Relation]++
	}

	return stats
}

// FilterByGroup returns a new graph containing only nodes in the specified group.
func (g *Graph) FilterByGroup(group string) *Graph {
	filtered := &Graph{
		Nodes: make([]GraphNode, 0),
		Edges: make([]GraphEdge, 0),
	}

	// Collect node IDs in the group
	nodeIDs := make(map[string]bool)
	for _, node := range g.Nodes {
		if node.Group == group {
			filtered.Nodes = append(filtered.Nodes, node)
			nodeIDs[node.ID] = true
		}
	}

	// Include edges where both source and target are in the filtered nodes
	for _, edge := range g.Edges {
		if nodeIDs[edge.Source] && nodeIDs[edge.Target] {
			filtered.Edges = append(filtered.Edges, edge)
		}
	}

	return filtered
}

// FilterByType returns a new graph containing only nodes of the specified types.
func (g *Graph) FilterByType(types ...string) *Graph {
	filtered := &Graph{
		Nodes: make([]GraphNode, 0),
		Edges: make([]GraphEdge, 0),
	}

	// Create type set for fast lookup
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	// Collect node IDs of matching types
	nodeIDs := make(map[string]bool)
	for _, node := range g.Nodes {
		if typeSet[node.Type] {
			filtered.Nodes = append(filtered.Nodes, node)
			nodeIDs[node.ID] = true
		}
	}

	// Include edges where both source and target are in the filtered nodes
	for _, edge := range g.Edges {
		if nodeIDs[edge.Source] && nodeIDs[edge.Target] {
			filtered.Edges = append(filtered.Edges, edge)
		}
	}

	return filtered
}

// GetNode returns a node by ID, or nil if not found.
func (g *Graph) GetNode(id string) *GraphNode {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

// GetChildren returns all nodes that have an edge from the given node ID.
func (g *Graph) GetChildren(nodeID string) []GraphNode {
	children := make([]GraphNode, 0)
	childIDs := make(map[string]bool)

	for _, edge := range g.Edges {
		if edge.Source == nodeID {
			childIDs[edge.Target] = true
		}
	}

	for _, node := range g.Nodes {
		if childIDs[node.ID] {
			children = append(children, node)
		}
	}

	return children
}

// GetParent returns the parent node of the given node ID, or nil if none.
func (g *Graph) GetParent(nodeID string) *GraphNode {
	for _, edge := range g.Edges {
		if edge.Target == nodeID {
			return g.GetNode(edge.Source)
		}
	}
	return nil
}
