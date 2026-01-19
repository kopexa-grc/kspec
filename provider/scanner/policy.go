// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package scanner provides policy evaluation capabilities for security scanning.
package scanner

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/policy"
)

// Check status constants (aliases to policy package for backwards compatibility).
const (
	StatusPassed  = policy.StatusPassed
	StatusFailed  = policy.StatusFailed
	StatusSkipped = policy.StatusSkipped
)

// EvaluatePolicies runs policy checks for a given resource.
// This is a convenience wrapper around policy.Evaluate.
func EvaluatePolicies(
	resource core.Resource,
	resourceType string,
	policies []policy.Policy,
	registry map[string]core.ResourceSpec,
	asset core.Asset,
) []CheckResult {
	return policy.Evaluate(resource, resourceType, policies, registry, asset)
}
