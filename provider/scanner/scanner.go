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
func (s *Scanner) Run(ctx context.Context) (*common.ResourceTree, error) {
	asset := s.config.Asset
	policies := s.config.Policies

	// Create resource tree with root node
	tree := common.NewResourceTree(asset.Name, asset.Type)
	tree.Root.State = common.AssetStateDiscovery

	s.emit(ScanEvent{Type: EventTreeCreated, Tree: tree})

	// Phase 1: Discovery
	discoveredResources, hasDiscovery := s.discover(ctx, tree)

	// Create resource type nodes
	s.createResourceNodes(ctx, tree, discoveredResources, hasDiscovery)

	// Mark root as scanning
	tree.Root.State = common.AssetStateScanning
	s.emit(ScanEvent{Type: EventScanStarted, Tree: tree})

	// Phase 2: Fetch and scan resources
	s.scanResources(ctx, tree, discoveredResources, policies)

	// Mark root as complete
	tree.Root.State = common.AssetStateComplete
	s.emit(ScanEvent{Type: EventScanComplete, Tree: tree})

	return tree, nil
}

// discover runs the discovery phase
func (s *Scanner) discover(ctx context.Context, tree *common.ResourceTree) (map[string]int, bool) {
	asset := s.config.Asset
	discoveredResources := make(map[string]int)
	hasDiscovery := false

	s.emit(ScanEvent{Type: EventDiscoveryStarted, Tree: tree})

	for _, resSpec := range s.registry {
		if discoverer, ok := resSpec.(core.DiscoveryResource); ok {
			hasDiscovery = true
			discovered, err := discoverer.Discover(ctx, asset)
			if err != nil {
				continue
			}

			for resourceType, count := range discovered {
				discoveredResources[resourceType] = count
			}
		}
	}

	s.emit(ScanEvent{Type: EventDiscoveryComplete, Tree: tree})

	return discoveredResources, hasDiscovery
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

// scanResources fetches and scans all resources
func (s *Scanner) scanResources(ctx context.Context, tree *common.ResourceTree, discoveredResources map[string]int, policies []core.Policy) {
	asset := s.config.Asset

	// Build resource order from tree children
	resourceOrder := make([]string, 0, len(tree.Root.Children))
	for _, child := range tree.Root.Children {
		resourceOrder = append(resourceOrder, child.ResourceType)
	}

	for _, resourceType := range resourceOrder {
		resSpec, ok := s.registry[resourceType]
		if !ok {
			continue
		}

		nodeID := fmt.Sprintf("%s-%s", tree.Root.ID, resourceType)
		resourceNode, exists := tree.GetNode(nodeID)
		if !exists {
			continue
		}

		// Skip if discovered with 0 count
		if count, discovered := discoveredResources[resourceType]; discovered && count == 0 {
			resourceNode.State = common.AssetStateComplete
			s.emit(ScanEvent{Type: EventTreeUpdated, Tree: tree})
			continue
		}

		// Mark as scanning
		resourceNode.State = common.AssetStateScanning
		s.emit(ScanEvent{Type: EventResourceScanning, Tree: tree, ResourceType: resourceType})

		// Fetch resources
		resources, err := resSpec.Fetch(ctx, asset)
		if err != nil {
			resourceNode.State = common.AssetStateError
			resourceNode.Error = err
			s.emit(ScanEvent{Type: EventError, Tree: tree, Error: err, ResourceType: resourceType})
			continue
		}

		resourceNode.ResourceCount = len(resources)

		// Process each resource instance
		for i, resource := range resources {
			s.processResourceInstance(ctx, tree, resourceNode, nodeID, resource, resourceType, i, resSpec, policies)
		}

		// Mark resource type as complete
		resourceNode.State = common.AssetStateComplete
		s.emit(ScanEvent{Type: EventResourceComplete, Tree: tree, ResourceType: resourceType})
	}
}

// processResourceInstance processes a single resource instance
func (s *Scanner) processResourceInstance(
	ctx context.Context,
	tree *common.ResourceTree,
	resourceNode *common.ResourceNode,
	nodeID string,
	resource core.Resource,
	resourceType string,
	index int,
	resSpec core.ResourceSpec,
	policies []core.Policy,
) {
	asset := s.config.Asset

	// Get resource name for display
	resourceName := getResourceName(resource, resourceType, index)

	instanceID := fmt.Sprintf("%s-%d", nodeID, index)
	instanceNode := &common.ResourceNode{
		ID:           instanceID,
		Name:         resourceName,
		Type:         "resource_instance",
		ResourceType: resourceType,
		State:        common.AssetStateScanning,
		Checks:       []CheckResult{},
		Metadata:     resource,
	}

	if err := tree.AddNode(nodeID, instanceNode); err != nil {
		log.Printf("Error adding instance node: %v", err)
		return
	}

	// Run policy checks
	checkResults := EvaluatePolicies(resource, resourceType, policies, s.registry, asset)

	// Add check results to instance node
	for _, checkResult := range checkResults {
		instanceNode.Checks = append(instanceNode.Checks, checkResult)

		switch checkResult.Status {
		case "passed":
			instanceNode.ChecksPassed++
		case "failed":
			instanceNode.ChecksFailed++
		case "skipped":
			instanceNode.ChecksSkipped++
		}
	}

	instanceNode.State = common.AssetStateComplete

	// Update parent counts
	resourceNode.ChecksPassed += instanceNode.ChecksPassed
	resourceNode.ChecksFailed += instanceNode.ChecksFailed
	resourceNode.ChecksSkipped += instanceNode.ChecksSkipped

	// Update root counts
	tree.Root.ChecksPassed += instanceNode.ChecksPassed
	tree.Root.ChecksFailed += instanceNode.ChecksFailed
	tree.Root.ChecksSkipped += instanceNode.ChecksSkipped

	// Process sub-resources
	s.processSubResources(ctx, tree, instanceNode, instanceID, resource, resourceType, resSpec, policies)
}

// processSubResources processes sub-resources for a resource instance
func (s *Scanner) processSubResources(
	ctx context.Context,
	tree *common.ResourceTree,
	instanceNode *common.ResourceNode,
	instanceID string,
	resource core.Resource,
	resourceType string,
	resSpec core.ResourceSpec,
	policies []core.Policy,
) {
	asset := s.config.Asset

	subProvider, ok := resSpec.(core.SubResourceProvider)
	if !ok {
		return
	}

	subSpecs := subProvider.SubResources()
	for _, subSpec := range subSpecs {
		subResType := subSpec.Name()

		// Create instance-specific asset config
		instanceAsset := core.Asset{
			Type:   asset.Type,
			Name:   asset.Name,
			Config: make(map[string]string),
		}
		for k, v := range asset.Config {
			instanceAsset.Config[k] = v
		}

		// Add instance-specific config for sub-resource fetching
		if resourceType == "github_repo" {
			if repoName, ok := resource["name"]; ok {
				instanceAsset.Config["repo"] = fmt.Sprintf("%v", repoName)
			}
		}

		// Fetch sub-resources
		subResources, err := subSpec.Fetch(ctx, instanceAsset)
		if err != nil {
			continue
		}

		// Process each sub-resource
		for j, subResource := range subResources {
			subResName := fmt.Sprintf("%s-%d", subResType, j)
			if name, ok := subResource["name"]; ok {
				subResName = fmt.Sprintf("%v", name)
			}

			subInstanceID := fmt.Sprintf("%s-sub-%s-%d", instanceID, subResType, j)

			// Run checks for sub-resource
			subCheckResults := EvaluatePolicies(subResource, subResType, policies, s.registry, asset)

			subInstanceNode := &common.ResourceNode{
				ID:           subInstanceID,
				Name:         subResName,
				Type:         "sub_resource_instance",
				ResourceType: subResType,
				State:        common.AssetStateComplete,
				Checks:       subCheckResults,
				Metadata:     subResource,
			}

			// Count check results
			for _, cr := range subCheckResults {
				switch cr.Status {
				case "passed":
					subInstanceNode.ChecksPassed++
				case "failed":
					subInstanceNode.ChecksFailed++
				case "skipped":
					subInstanceNode.ChecksSkipped++
				}
			}

			// Add to parent instance's sub-resources
			if err := tree.AddSubResource(instanceID, subResType, subInstanceNode); err != nil {
				log.Printf("Error adding sub-resource: %v", err)
				continue
			}

			// Update counts up the tree
			instanceNode.ChecksPassed += subInstanceNode.ChecksPassed
			instanceNode.ChecksFailed += subInstanceNode.ChecksFailed
			instanceNode.ChecksSkipped += subInstanceNode.ChecksSkipped

			// Find parent resource node to update its counts
			if parent := instanceNode.Parent; parent != nil {
				parent.ChecksPassed += subInstanceNode.ChecksPassed
				parent.ChecksFailed += subInstanceNode.ChecksFailed
				parent.ChecksSkipped += subInstanceNode.ChecksSkipped
			}

			tree.Root.ChecksPassed += subInstanceNode.ChecksPassed
			tree.Root.ChecksFailed += subInstanceNode.ChecksFailed
			tree.Root.ChecksSkipped += subInstanceNode.ChecksSkipped
		}
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
