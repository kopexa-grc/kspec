// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"encoding/json"
	"time"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
)

// Mode determines how discovery is performed.
type Mode string

const (
	// ModeNative uses the DiscoveryResource interface (fast, lightweight).
	// Only resources implementing core.DiscoveryResource will be discovered.
	ModeNative Mode = "native"

	// ModeFetch fetches actual resources and counts them (accurate, slower).
	// This calls Fetch() on all resources and collects instance data.
	ModeFetch Mode = "fetch"

	// ModeRegistry lists registered resources without API calls (fastest, no counts).
	// Only returns what resource types are available, not actual counts.
	ModeRegistry Mode = "registry"
)

// Node type constants for graph traversal.
const (
	// NodeTypeRoot is the root node type.
	NodeTypeRoot = "root"
	// NodeTypeResourceType is a node representing a resource type.
	NodeTypeResourceType = "resource_type"
	// NodeTypeResourceInstance is a node representing a resource instance.
	NodeTypeResourceInstance = "resource_instance"
	// NodeTypeSubResourceInstance is a node representing a sub-resource instance.
	NodeTypeSubResourceInstance = "sub_resource_instance"
)

// Config holds the configuration for a discovery operation.
type Config struct {
	// ProviderName is the name of the provider to discover (e.g., "github", "aws").
	ProviderName string

	// ProviderConfig contains provider-specific configuration (credentials, region, etc.).
	ProviderConfig map[string]string

	// Asset defines the asset to discover resources for.
	Asset core.Asset

	// Mode determines how discovery is performed.
	// Default is ModeNative.
	Mode Mode

	// Concurrency sets the maximum number of concurrent workers.
	// 0 means use default.
	Concurrency int

	// IncludeInstances determines whether to include individual resource instances.
	// When true, the Instances field in ResourceInfo will be populated.
	// This is required for graph generation but increases memory usage.
	IncludeInstances bool
}

// Result contains the results of a discovery operation.
type Result struct {
	// Provider is the name of the provider that was discovered.
	Provider string `json:"provider"`

	// AssetType is the type of asset that was discovered (e.g., "github-org").
	AssetType string `json:"asset_type"`

	// AssetName is the name of the asset that was discovered.
	AssetName string `json:"asset_name"`

	// DiscoveredAt is the timestamp when discovery started.
	DiscoveredAt time.Time `json:"discovered_at"`

	// Mode indicates which discovery mode was used.
	Mode Mode `json:"mode"`

	// Duration is how long the discovery took.
	Duration time.Duration `json:"duration_ms"`

	// Resources contains information about each discovered resource type.
	Resources []ResourceInfo `json:"resources"`

	// TotalCount is the sum of all resource counts.
	TotalCount int `json:"total_count"`

	// Graph contains the graph representation for visualization.
	// Only populated when IncludeInstances is true.
	Graph *Graph `json:"graph,omitempty"`

	// Errors contains non-fatal errors that occurred during discovery.
	Errors []Error `json:"errors,omitempty"`
}

// MarshalJSON implements json.Marshaler to output duration in milliseconds.
func (r Result) MarshalJSON() ([]byte, error) {
	type Alias Result
	return json.Marshal(&struct {
		Alias
		Duration int64 `json:"duration_ms"`
	}{
		Alias:    Alias(r),
		Duration: r.Duration.Milliseconds(),
	})
}

// ResourceInfo contains information about a discovered resource type.
type ResourceInfo struct {
	// Type is the resource type name (e.g., "github_repo", "aws_s3_bucket").
	Type string `json:"type"`

	// Count is the number of instances of this resource type.
	// -1 indicates unknown (registry mode).
	Count int `json:"count"`

	// SupportsNative indicates whether this resource type supports native discovery.
	SupportsNative bool `json:"supports_native_discovery"`

	// Instances contains individual resource instances.
	// Only populated when IncludeInstances is true in Config.
	Instances []ResourceInstance `json:"instances,omitempty"`
}

// ResourceInstance represents a single discovered resource instance.
type ResourceInstance struct {
	// ID is the unique identifier for this resource.
	ID string `json:"id"`

	// Name is the human-readable name of this resource.
	Name string `json:"name"`

	// Type is the resource type (e.g., "github_repo").
	Type string `json:"type"`

	// ParentID is the ID of the parent resource in the hierarchy.
	// Empty for root resources.
	ParentID string `json:"parent_id,omitempty"`

	// Metadata contains additional resource-specific data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Graph represents a graph structure for visualization.
type Graph struct {
	// Nodes contains all graph nodes (resources).
	Nodes []GraphNode `json:"nodes"`

	// Edges contains all graph edges (relationships).
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a node in the discovery graph.
type GraphNode struct {
	// ID is the unique identifier for this node.
	ID string `json:"id"`

	// Label is the display label for this node.
	Label string `json:"label"`

	// Type is the resource type (e.g., "github_repo").
	Type string `json:"type"`

	// Group is used for visual grouping (e.g., "compute", "storage", "identity").
	Group string `json:"group"`

	// Metadata contains additional node data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GraphEdge represents an edge (relationship) in the discovery graph.
type GraphEdge struct {
	// Source is the ID of the source node.
	Source string `json:"source"`

	// Target is the ID of the target node.
	Target string `json:"target"`

	// Relation describes the relationship type (e.g., "contains", "owns", "member_of").
	Relation string `json:"relation"`
}

// Error represents a non-fatal error that occurred during discovery.
type Error struct {
	// ResourceType is the resource type that caused the error.
	ResourceType string `json:"resource_type"`

	// Message is the error message.
	Message string `json:"message"`

	// Err is the underlying error (not serialized).
	Err error `json:"-"`
}

// EventType represents the type of discovery event.
type EventType int

const (
	// EventDiscoveryStarted is sent when discovery begins.
	EventDiscoveryStarted EventType = iota

	// EventDiscoveryComplete is sent when discovery finishes.
	EventDiscoveryComplete

	// EventEvaluateStarted is sent when evaluation begins.
	EventEvaluateStarted

	// EventEvaluateComplete is sent when evaluation finishes.
	EventEvaluateComplete

	// EventNodeScanning is sent when a node starts being processed.
	EventNodeScanning

	// EventNodeComplete is sent when a node has been processed.
	EventNodeComplete

	// EventNodeError is sent when a node encounters an error.
	EventNodeError

	// EventTreeCreated is sent when the resource tree is created (legacy).
	EventTreeCreated

	// EventTreeUpdated is sent when the resource tree is updated (legacy).
	EventTreeUpdated

	// EventResourceDiscovering is sent when a resource type starts being discovered.
	EventResourceDiscovering

	// EventResourceDiscovered is sent when a resource type has been discovered.
	EventResourceDiscovered

	// EventDiscoveryError is sent when a non-fatal error occurs.
	EventDiscoveryError
)

// Event represents an event during discovery/evaluation.
type Event struct {
	// Type is the event type.
	Type EventType

	// Node is the resource node (for node-based events).
	Node *common.ResourceNode

	// Tree is the resource tree (for legacy events).
	Tree *common.ResourceTree

	// ResourceType is the resource type being discovered (if applicable).
	ResourceType string

	// Count is the number of resources discovered (for EventResourceDiscovered).
	Count int

	// Error is the error that occurred (for EventNodeError, EventDiscoveryError).
	Error error

	// Message is an optional message.
	Message string
}

// EventHandler is a callback for discovery events.
type EventHandler func(event Event)

// EvaluateResult contains the results of an evaluation operation.
type EvaluateResult struct {
	// Provider is the name of the provider.
	Provider string `json:"provider"`

	// Asset is the asset that was evaluated.
	Asset string `json:"asset"`

	// AssetType is the type of asset.
	AssetType string `json:"asset_type"`

	// Root is the root node of the resource graph.
	Root *common.ResourceNode `json:"root"`

	// TotalCount is the total number of resources evaluated.
	TotalCount int `json:"total_count"`

	// Duration is how long the evaluation took.
	Duration time.Duration `json:"duration_ms"`

	// Check aggregates
	TotalChecks   int `json:"total_checks"`
	PassedChecks  int `json:"passed_checks"`
	FailedChecks  int `json:"failed_checks"`
	SkippedChecks int `json:"skipped_checks"`

	// Errors contains non-fatal errors that occurred.
	Errors []Error `json:"errors,omitempty"`
}

// MarshalJSON implements json.Marshaler to output duration in milliseconds.
func (r EvaluateResult) MarshalJSON() ([]byte, error) {
	type Alias EvaluateResult
	return json.Marshal(&struct {
		Alias
		Duration int64 `json:"duration_ms"`
	}{
		Alias:    Alias(r),
		Duration: r.Duration.Milliseconds(),
	})
}
