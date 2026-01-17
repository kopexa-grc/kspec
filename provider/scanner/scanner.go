package scanner

import (
	"context"
	"fmt"
	"log"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider"
)

// Scanner handles resource discovery and policy evaluation
type Scanner struct {
	config   ScanConfig
	registry map[string]core.ResourceSpec
	handler  ScanEventHandler
	errors   []*ScanError
}

// NewScanner creates a new Scanner instance
func NewScanner(config ScanConfig) *Scanner {
	return &Scanner{
		config: config,
	}
}

// OnEvent sets the event handler for scan events
func (s *Scanner) OnEvent(handler ScanEventHandler) {
	s.handler = handler
}

// emit sends an event to the handler if one is registered
func (s *Scanner) emit(event ScanEvent) {
	if s.handler != nil {
		s.handler(event)
	}
}

// recordError adds an error to the scanner's error list
func (s *Scanner) recordError(phase, resourceType, message string, err error) {
	s.errors = append(s.errors, &ScanError{
		Phase:        phase,
		ResourceType: resourceType,
		Message:      message,
		Err:          err,
	})
}

// Initialize connects to the provider and sets up the registry
func (s *Scanner) Initialize(ctx context.Context) error {
	registry, err := provider.InitProvider(ctx, s.config.ProviderName, s.config.ProviderConfig)
	if err != nil {
		return fmt.Errorf("failed to init provider %s: %w", s.config.ProviderName, err)
	}
	s.registry = registry
	return nil
}

// Registry returns the resource registry
func (s *Scanner) Registry() map[string]core.ResourceSpec {
	return s.registry
}

// Run executes the full scan process
func (s *Scanner) Run(ctx context.Context) *ScanResult {
	asset := s.config.Asset
	policies := s.config.Policies

	// Initialize errors slice
	s.errors = make([]*ScanError, 0)

	// Create resource tree with root node
	tree := common.NewResourceTree(asset.Name, asset.Type)
	tree.Root.State = common.AssetStateDiscovery

	s.emit(ScanEvent{Type: EventTreeCreated, Tree: tree})

	// Phase 1: Discovery (concurrent)
	discoveredResources, hasDiscovery := s.discoverConcurrent(ctx, tree)

	// Create resource type nodes
	s.createResourceNodes(ctx, tree, discoveredResources, hasDiscovery)

	// Mark root as scanning
	tree.Root.State = common.AssetStateScanning
	s.emit(ScanEvent{Type: EventScanStarted, Tree: tree})

	// Phase 2: Fetch and scan resources (concurrent)
	s.scanResourcesConcurrent(ctx, tree, discoveredResources, policies)

	// Mark root as complete
	tree.Root.State = common.AssetStateComplete
	s.emit(ScanEvent{Type: EventScanComplete, Tree: tree})

	return &ScanResult{
		Tree:   tree,
		Errors: s.errors,
	}
}

// createResourceNodes creates nodes in the tree for discovered resources
func (s *Scanner) createResourceNodes(ctx context.Context, tree *common.ResourceTree, discoveredResources map[string]int, hasDiscovery bool) {
	asset := s.config.Asset

	if hasDiscovery {
		// Get ordered types for this provider
		orderedTypes := ResourceOrder[s.config.ProviderName]
		if orderedTypes == nil {
			// For unknown providers, use discovered order
			for resourceType := range discoveredResources {
				orderedTypes = append(orderedTypes, resourceType)
			}
		}

		// Create nodes in order
		for _, resourceType := range orderedTypes {
			count, exists := discoveredResources[resourceType]
			if !exists || count == 0 {
				continue
			}

			resourceNode := &common.ResourceNode{
				ID:                fmt.Sprintf("%s-%s", tree.Root.ID, resourceType),
				Name:              resourceType,
				Type:              "resource_type",
				ResourceType:      resourceType,
				ResourceCount:     count,
				State:             common.AssetStatePending,
				Children:          []*common.ResourceNode{},
				SubResources:      make(map[string][]*common.ResourceNode),
				SubResourceCounts: make(map[string]int),
				Metadata:          make(map[string]interface{}),
			}

			if err := tree.AddNode(tree.Root.ID, resourceNode); err != nil {
				log.Printf("Error adding resource node: %v", err)
				continue
			}

			tree.Root.ResourceCount += count

			// Check for sub-resources
			if resSpec, ok := s.registry[resourceType]; ok {
				if subProvider, ok := resSpec.(core.SubResourceProvider); ok {
					subSpecs := subProvider.SubResources()
					for _, subSpec := range subSpecs {
						if subDiscoverer, ok := subSpec.(core.DiscoveryResource); ok {
							subDiscovered, err := subDiscoverer.Discover(ctx, asset)
							if err != nil {
								s.recordError("discovery", subSpec.Name(), err.Error(), err)
								continue
							}
							for subResType, subCount := range subDiscovered {
								if subCount > 0 {
									resourceNode.SubResourceCounts[subResType] = subCount
								}
							}
						}
					}
				}
			}

			s.emit(ScanEvent{Type: EventTreeUpdated, Tree: tree})
		}
	} else {
		// Non-discovery mode: create nodes based on registry
		for _, resSpec := range s.registry {
			resourceType := resSpec.Name()
			nodeID := fmt.Sprintf("%s-%s", tree.Root.ID, resourceType)

			if _, exists := tree.GetNode(nodeID); !exists {
				resourceNode := &common.ResourceNode{
					ID:                nodeID,
					Name:              resourceType,
					Type:              "resource_type",
					ResourceType:      resourceType,
					ResourceCount:     1,
					State:             common.AssetStatePending,
					Children:          []*common.ResourceNode{},
					SubResources:      make(map[string][]*common.ResourceNode),
					SubResourceCounts: make(map[string]int),
					Metadata:          make(map[string]interface{}),
				}

				if err := tree.AddNode(tree.Root.ID, resourceNode); err != nil {
					log.Printf("Error adding resource node: %v", err)
				}
			}
		}
		s.emit(ScanEvent{Type: EventTreeUpdated, Tree: tree})
	}
}

// getResourceName extracts a display name from a resource
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
