// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package scanner

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/sourcegraph/conc/pool"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/concurrency"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider"
)

// discoveryResult holds the result of a single resource discovery.
type discoveryResult struct {
	resourceType string
	counts       map[string]int
	err          error
}

// discoverConcurrent runs resource discovery in parallel using conc.
func (s *Scanner) discoverConcurrent(ctx context.Context, tree *common.ResourceTree) (map[string]int, bool) {
	asset := s.config.Asset
	discoveredResources := make(map[string]int)
	hasDiscovery := false

	s.emit(ScanEvent{Type: EventDiscoveryStarted, Tree: tree})

	// Count discoverable resources
	var discoverers []struct {
		resType    string
		discoverer core.DiscoveryResource
	}
	for resType, resSpec := range s.registry {
		if discoverer, ok := resSpec.(core.DiscoveryResource); ok {
			discoverers = append(discoverers, struct {
				resType    string
				discoverer core.DiscoveryResource
			}{resType, discoverer})
			hasDiscovery = true
		}
	}

	if !hasDiscovery {
		s.emit(ScanEvent{Type: EventDiscoveryComplete, Tree: tree})
		return discoveredResources, false
	}

	// Calculate optimal workers
	workers := concurrency.DiscoveryWorkers(s.config.Concurrency, len(discoverers))

	// Create rate limiter for the provider
	limiter := provider.NewRateLimiter(s.config.ProviderName)

	// Use conc pool for parallel discovery
	p := pool.NewWithResults[discoveryResult]().
		WithContext(ctx).
		WithMaxGoroutines(workers)

	for _, d := range discoverers {
		resType := d.resType
		discoverer := d.discoverer
		p.Go(func(ctx context.Context) (discoveryResult, error) {
			// Rate limit the API call
			if err := limiter.Wait(ctx); err != nil {
				return discoveryResult{resourceType: resType, err: err}, nil //nolint:nilerr // Error recorded in result, don't fail pool
			}

			counts, err := discoverer.Discover(ctx, asset)
			return discoveryResult{
				resourceType: resType,
				counts:       counts,
				err:          err,
			}, nil
		})
	}

	results, _ := p.Wait() //nolint:errcheck // Errors collected in result struct

	// Merge results (thread-safe since we're single-threaded here)
	for _, result := range results {
		if result.err != nil {
			s.recordError("discovery", result.resourceType, result.err.Error(), result.err)
			continue
		}
		for resourceType, count := range result.counts {
			discoveredResources[resourceType] = count
		}
	}

	s.emit(ScanEvent{Type: EventDiscoveryComplete, Tree: tree})

	return discoveredResources, hasDiscovery
}

// scanResourcesConcurrent fetches and scans resources in parallel.
func (s *Scanner) scanResourcesConcurrent(ctx context.Context, tree *common.ResourceTree, discoveredResources map[string]int, policies []core.Policy) {
	asset := s.config.Asset

	// Build resource order from tree children
	resourceOrder := make([]string, 0, len(tree.Root.Children))
	for _, child := range tree.Root.Children {
		resourceOrder = append(resourceOrder, child.ResourceType)
	}

	// Calculate optimal workers for resource type fetching
	workers := concurrency.FetchWorkers(s.config.Concurrency, len(resourceOrder))

	// Create rate limiter for the provider
	limiter := provider.NewRateLimiter(s.config.ProviderName)

	// Mutex for tree updates (tree is not thread-safe)
	var treeMu sync.Mutex

	// Use conc pool for parallel resource fetching
	p := pool.New().
		WithContext(ctx).
		WithMaxGoroutines(workers)

	for _, resourceType := range resourceOrder {
		resType := resourceType
		resSpec, ok := s.registry[resType]
		if !ok {
			continue
		}

		nodeID := fmt.Sprintf("%s-%s", tree.Root.ID, resType)

		treeMu.Lock()
		resourceNode, exists := tree.GetNode(nodeID)
		treeMu.Unlock()

		if !exists {
			continue
		}

		// Skip if discovered with 0 count
		if count, discovered := discoveredResources[resType]; discovered && count == 0 {
			treeMu.Lock()
			resourceNode.State = common.AssetStateComplete
			treeMu.Unlock()
			s.emit(ScanEvent{Type: EventTreeUpdated, Tree: tree})
			continue
		}

		p.Go(func(ctx context.Context) error {
			// Mark as scanning
			treeMu.Lock()
			resourceNode.State = common.AssetStateScanning
			treeMu.Unlock()
			s.emit(ScanEvent{Type: EventResourceScanning, Tree: tree, ResourceType: resType})

			// Rate limit the fetch
			if err := limiter.Wait(ctx); err != nil {
				treeMu.Lock()
				resourceNode.State = common.AssetStateError
				resourceNode.Error = err
				treeMu.Unlock()
				s.emit(ScanEvent{Type: EventError, Tree: tree, Error: err, ResourceType: resType})
				return nil //nolint:nilerr // Error recorded in tree node, don't fail pool
			}

			// Fetch resources
			resources, err := resSpec.Fetch(ctx, asset)
			if err != nil {
				treeMu.Lock()
				resourceNode.State = common.AssetStateError
				resourceNode.Error = err
				treeMu.Unlock()
				s.emit(ScanEvent{Type: EventError, Tree: tree, Error: err, ResourceType: resType})
				return nil //nolint:nilerr // Error recorded in tree node, don't fail pool
			}

			treeMu.Lock()
			resourceNode.ResourceCount = len(resources)
			treeMu.Unlock()

			// Process instances concurrently within this resource type
			s.processInstancesConcurrent(ctx, tree, resourceNode, nodeID, resources, resType, resSpec, policies, &treeMu, limiter)

			// Mark resource type as complete
			treeMu.Lock()
			resourceNode.State = common.AssetStateComplete
			treeMu.Unlock()
			s.emit(ScanEvent{Type: EventResourceComplete, Tree: tree, ResourceType: resType})

			return nil
		})
	}

	// Wait for all resource types to complete
	_ = p.Wait() //nolint:errcheck // Errors recorded in tree nodes
}

// processInstancesConcurrent processes resource instances in parallel.
func (s *Scanner) processInstancesConcurrent(
	ctx context.Context,
	tree *common.ResourceTree,
	resourceNode *common.ResourceNode,
	nodeID string,
	resources []core.Resource,
	resourceType string,
	resSpec core.ResourceSpec,
	policies []core.Policy,
	treeMu *sync.Mutex,
	limiter *ratelimit.Limiter,
) {
	if len(resources) == 0 {
		return
	}

	// Calculate workers for instance processing
	workers := concurrency.FetchWorkers(s.config.Concurrency, len(resources))

	// Counters for aggregating results
	var (
		totalPassed  int
		totalFailed  int
		totalSkipped int
		counterMu    sync.Mutex
	)

	p := pool.New().
		WithContext(ctx).
		WithMaxGoroutines(workers)

	for i, resource := range resources {
		idx := i
		res := resource

		p.Go(func(ctx context.Context) error {
			// Rate limit sub-resource fetches (not the evaluation)
			instanceNode := s.processResourceInstanceConcurrent(ctx, tree, nodeID, res, resourceType, idx, resSpec, policies, treeMu, limiter)

			if instanceNode != nil {
				// Update counters atomically
				counterMu.Lock()
				totalPassed += instanceNode.ChecksPassed
				totalFailed += instanceNode.ChecksFailed
				totalSkipped += instanceNode.ChecksSkipped
				counterMu.Unlock()
			}

			return nil
		})
	}

	_ = p.Wait() //nolint:errcheck // Task completion always succeeds

	// Update parent counts
	treeMu.Lock()
	resourceNode.ChecksPassed += totalPassed
	resourceNode.ChecksFailed += totalFailed
	resourceNode.ChecksSkipped += totalSkipped

	tree.Root.ChecksPassed += totalPassed
	tree.Root.ChecksFailed += totalFailed
	tree.Root.ChecksSkipped += totalSkipped
	treeMu.Unlock()
}

// processResourceInstanceConcurrent processes a single resource instance (thread-safe).
func (s *Scanner) processResourceInstanceConcurrent(
	ctx context.Context,
	tree *common.ResourceTree,
	nodeID string,
	resource core.Resource,
	resourceType string,
	index int,
	resSpec core.ResourceSpec,
	policies []core.Policy,
	treeMu *sync.Mutex,
	limiter *ratelimit.Limiter,
) *common.ResourceNode {
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

	treeMu.Lock()
	if err := tree.AddNode(nodeID, instanceNode); err != nil {
		treeMu.Unlock()
		log.Printf("Error adding instance node: %v", err)
		return nil
	}
	treeMu.Unlock()

	// Run policy checks (CPU-bound, no rate limiting needed)
	checkResults := EvaluatePolicies(resource, resourceType, policies, s.registry, asset)

	// Add check results to instance node
	for _, checkResult := range checkResults {
		instanceNode.Checks = append(instanceNode.Checks, checkResult)

		switch checkResult.Status {
		case StatusPassed:
			instanceNode.ChecksPassed++
		case StatusFailed:
			instanceNode.ChecksFailed++
		case StatusSkipped:
			instanceNode.ChecksSkipped++
		}
	}

	instanceNode.State = common.AssetStateComplete

	// Process sub-resources
	s.processSubResourcesConcurrent(ctx, tree, instanceNode, instanceID, resource, resourceType, resSpec, policies, treeMu, limiter)

	return instanceNode
}

// processSubResourcesConcurrent processes sub-resources in parallel.
func (s *Scanner) processSubResourcesConcurrent(
	ctx context.Context,
	tree *common.ResourceTree,
	instanceNode *common.ResourceNode,
	instanceID string,
	resource core.Resource,
	resourceType string,
	resSpec core.ResourceSpec,
	policies []core.Policy,
	treeMu *sync.Mutex,
	limiter *ratelimit.Limiter,
) {
	asset := s.config.Asset

	subProvider, ok := resSpec.(core.SubResourceProvider)
	if !ok {
		return
	}

	subSpecs := subProvider.SubResources()
	if len(subSpecs) == 0 {
		return
	}

	// Calculate workers for sub-resource fetching
	workers := concurrency.FetchWorkers(s.config.Concurrency, len(subSpecs))

	var (
		totalPassed  int
		totalFailed  int
		totalSkipped int
		counterMu    sync.Mutex
	)

	p := pool.New().
		WithContext(ctx).
		WithMaxGoroutines(workers)

	for _, subSpec := range subSpecs {
		spec := subSpec

		p.Go(func(ctx context.Context) error {
			subResType := spec.Name()

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

			// Rate limit the sub-resource fetch
			if err := limiter.Wait(ctx); err != nil {
				return nil //nolint:nilerr // Context canceled, continue with other sub-resources
			}

			// Fetch sub-resources
			subResources, err := spec.Fetch(ctx, instanceAsset)
			if err != nil {
				s.recordError("fetch", subResType, err.Error(), err)
				return nil
			}

			// Process each sub-resource
			for j, subResource := range subResources {
				subResName := fmt.Sprintf("%s-%d", subResType, j)
				if name, ok := subResource["name"]; ok {
					subResName = fmt.Sprintf("%v", name)
				}

				subInstanceID := fmt.Sprintf("%s-sub-%s-%d", instanceID, subResType, j)

				// Run checks for sub-resource (CPU-bound)
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
					case StatusPassed:
						subInstanceNode.ChecksPassed++
					case StatusFailed:
						subInstanceNode.ChecksFailed++
					case StatusSkipped:
						subInstanceNode.ChecksSkipped++
					}
				}

				// Add to parent instance's sub-resources
				treeMu.Lock()
				if err := tree.AddSubResource(instanceID, subResType, subInstanceNode); err != nil {
					treeMu.Unlock()
					log.Printf("Error adding sub-resource: %v", err)
					continue
				}
				treeMu.Unlock()

				// Update local counters
				counterMu.Lock()
				totalPassed += subInstanceNode.ChecksPassed
				totalFailed += subInstanceNode.ChecksFailed
				totalSkipped += subInstanceNode.ChecksSkipped
				counterMu.Unlock()
			}

			return nil
		})
	}

	_ = p.Wait() //nolint:errcheck // Task completion always succeeds

	// Update instance node counts
	treeMu.Lock()
	instanceNode.ChecksPassed += totalPassed
	instanceNode.ChecksFailed += totalFailed
	instanceNode.ChecksSkipped += totalSkipped

	// Update parent resource node counts
	if parent := instanceNode.Parent; parent != nil {
		parent.ChecksPassed += totalPassed
		parent.ChecksFailed += totalFailed
		parent.ChecksSkipped += totalSkipped
	}

	tree.Root.ChecksPassed += totalPassed
	tree.Root.ChecksFailed += totalFailed
	tree.Root.ChecksSkipped += totalSkipped
	treeMu.Unlock()
}
