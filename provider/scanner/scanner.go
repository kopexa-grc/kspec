// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scanner

import (
	"context"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/discovery"
)

// Scanner handles resource discovery and policy evaluation.
// It uses Discoverer for graph traversal with integrated policy evaluation.
type Scanner struct {
	config     ScanConfig
	discoverer *discovery.Discoverer
	handler    ScanEventHandler
	errors     []*ScanError
}

// NewScanner creates a new Scanner instance.
func NewScanner(config ScanConfig) *Scanner {
	// Create discoverer config
	discConfig := discovery.Config{
		ProviderName:     config.ProviderName,
		ProviderConfig:   config.ProviderConfig,
		Asset:            config.Asset,
		Concurrency:      config.Concurrency.MaxWorkers,
		IncludeInstances: true, // Always need instances for policy evaluation
	}

	return &Scanner{
		config:     config,
		discoverer: discovery.NewDiscoverer(discConfig),
	}
}

// OnEvent sets the event handler for scan events.
func (s *Scanner) OnEvent(handler ScanEventHandler) {
	s.handler = handler

	// Forward discovery events to our handler
	s.discoverer.OnEvent(func(event discovery.Event) {
		var eventType ScanEventType
		switch event.Type {
		case discovery.EventDiscoveryStarted:
			eventType = EventDiscoveryStarted
		case discovery.EventEvaluateStarted:
			eventType = EventScanStarted
		case discovery.EventTreeCreated:
			eventType = EventTreeCreated
		case discovery.EventTreeUpdated:
			eventType = EventTreeUpdated
		case discovery.EventDiscoveryComplete:
			eventType = EventDiscoveryComplete
		case discovery.EventEvaluateComplete:
			eventType = EventScanComplete
		case discovery.EventNodeScanning, discovery.EventResourceDiscovering:
			eventType = EventResourceScanning
		case discovery.EventNodeComplete, discovery.EventResourceDiscovered:
			eventType = EventResourceComplete
		case discovery.EventNodeError, discovery.EventDiscoveryError:
			eventType = EventError
		default:
			return
		}

		// Use the tree from the event (always provided by discoverer)
		tree := event.Tree

		s.emit(ScanEvent{
			Type:         eventType,
			Tree:         tree,
			ResourceType: event.ResourceType,
			Error:        event.Error,
			Message:      event.Message,
		})
	})
}

// emit sends an event to the handler if one is registered.
func (s *Scanner) emit(event ScanEvent) {
	if s.handler != nil {
		s.handler(event)
	}
}

// recordError adds an error to the scanner's error list.
func (s *Scanner) recordError(phase, resourceType, message string, err error) {
	s.errors = append(s.errors, &ScanError{
		Phase:        phase,
		ResourceType: resourceType,
		Message:      message,
		Err:          err,
	})
}

// Initialize connects to the provider and sets up the registry.
func (s *Scanner) Initialize(ctx context.Context) error {
	return s.discoverer.Initialize(ctx)
}

// Registry returns the resource registry.
func (s *Scanner) Registry() map[string]core.ResourceSpec {
	return s.discoverer.Registry()
}

// Run executes the full scan process using graph traversal with integrated policy evaluation.
func (s *Scanner) Run(ctx context.Context) *ScanResult {
	policies := s.config.Policies
	asset := s.config.Asset

	// Initialize errors slice
	s.errors = make([]*ScanError, 0)

	// Use Evaluate() which does graph traversal with integrated policy evaluation
	result, err := s.discoverer.Evaluate(ctx, asset, policies)
	if err != nil {
		s.recordError("evaluation", "", err.Error(), err)
		return &ScanResult{
			Tree:   nil,
			Errors: s.errors,
		}
	}

	// Convert discovery errors to scan errors
	for _, discErr := range result.Errors {
		s.errors = append(s.errors, &ScanError{
			Phase:        "evaluation",
			ResourceType: discErr.ResourceType,
			Message:      discErr.Message,
			Err:          discErr.Err,
		})
	}

	// Build tree from result
	tree := &common.ResourceTree{
		Root:    result.Root,
		Current: result.Root,
	}

	return &ScanResult{
		Tree:   tree,
		Errors: s.errors,
	}
}
