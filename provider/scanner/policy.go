// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package scanner provides policy evaluation capabilities for security scanning.
package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/kopexa-grc/kspec/core"
)

// Check status constants.
const (
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// LoadPolicies loads policies from a file or directory
func LoadPolicies(policyFile, policyDir string) ([]core.Policy, error) {
	var policies []core.Policy

	loadPolicy := func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var p core.Policy
		if err := yaml.Unmarshal(data, &p); err != nil {
			return err
		}
		policies = append(policies, p)
		return nil
	}

	if policyDir != "" {
		files, err := os.ReadDir(policyDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read policy directory: %w", err)
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
				path := filepath.Join(policyDir, file.Name())
				if err := loadPolicy(path); err != nil {
					log.Printf("Warning: failed to load policy %s: %v", file.Name(), err)
				}
			}
		}
	} else if policyFile != "" {
		if err := loadPolicy(policyFile); err != nil {
			return nil, fmt.Errorf("failed to read policy file: %w", err)
		}
	}

	return policies, nil
}

// FilterPoliciesByProvider filters policies to only include those matching the provider
func FilterPoliciesByProvider(policies []core.Policy, providerName, providerAlias string) []core.Policy {
	var filtered []core.Policy

	for _, policy := range policies {
		if len(policy.Require) == 0 {
			// No requirements - policy applies to all providers
			filtered = append(filtered, policy)
			continue
		}

		// Check if any requirement matches the current provider
		for _, req := range policy.Require {
			if req.Provider == providerName || req.Provider == providerAlias {
				filtered = append(filtered, policy)
				break
			}
		}
	}

	return filtered
}

// evaluateGroupFilter performs a simple evaluation of group filters
// This handles common patterns like: asset.type == "github-org"
func evaluateGroupFilter(filter string, asset core.Asset) bool {
	filter = strings.TrimSpace(filter)

	// Pattern: asset.type == "value" or asset.type == 'value'
	if strings.Contains(filter, "asset.type") {
		for _, sep := range []string{`== "`, `== '`, `=="`, `=='`} {
			if idx := strings.Index(filter, sep); idx != -1 {
				start := idx + len(sep)
				end := strings.IndexAny(filter[start:], `"'`)
				if end != -1 {
					expectedType := filter[start : start+end]
					return asset.Type == expectedType
				}
			}
		}
	}

	// If we can't parse the filter, default to true (include the group)
	return true
}

// EvaluatePolicies runs policy checks for a given resource
func EvaluatePolicies(
	resource core.Resource,
	resourceType string,
	policies []core.Policy,
	registry map[string]core.ResourceSpec,
	asset core.Asset,
) []CheckResult {
	var results []CheckResult

	// Create evaluator
	evaluator, err := core.NewEvaluator(registry)
	if err != nil {
		return results
	}

	// Build definitions index from all policies
	definitions := make(map[string]core.Check)
	for _, policy := range policies {
		for _, q := range policy.Queries {
			if q.UID != "" {
				definitions[q.UID] = q
			}
			if q.ID != "" {
				definitions[q.ID] = q
			}
		}
	}

	// Process each policy
	for _, policy := range policies {
		for _, group := range policy.Groups {
			// Evaluate group filter
			if group.Filter != "" {
				if !evaluateGroupFilter(group.Filter, asset) {
					continue
				}
			}

			// Process checks in this group
			for _, checkRef := range group.Checks {
				check := checkRef

				// Resolve check definition if needed
				if check.Query == "" {
					var def core.Check
					var found bool

					if check.UID != "" {
						def, found = definitions[check.UID]
					}
					if !found && check.ID != "" {
						def, found = definitions[check.ID]
					}

					if !found {
						continue
					}

					// Merge definition into check
					check.Query = def.Query
					check.Resource = def.Resource
					check.Remediation = def.Remediation
					check.Severity = def.Severity
					check.Docs = def.Docs
					check.Audit = def.Audit
					if check.Title == "" {
						check.Title = def.Title
					}

					// Merge config
					config := make(map[string]string)
					for k, v := range def.Config {
						config[k] = v
					}
					for k, v := range checkRef.Config {
						config[k] = v
					}
					check.Config = config

					// Merge props
					if len(def.Props) > 0 && len(check.Props) == 0 {
						check.Props = def.Props
					}
				}

				// Check if this check applies to this resource type
				if check.Resource != resourceType {
					continue
				}

				// Build props map
				propsMap := make(map[string]interface{})
				for _, prop := range check.Props {
					propsMap[prop.UID] = prop.MQL
				}

				// Merge check config with asset config
				evalConfig := make(map[string]string)
				for k, v := range asset.Config {
					evalConfig[k] = v
				}
				for k, v := range check.Config {
					evalConfig[k] = v
				}

				// Evaluate the query
				passed, err := evaluator.Evaluate(
					check.Query,
					resource,
					evalConfig,
					propsMap,
					asset,
				)

				var status, details string

				switch {
				case err != nil:
					errStr := err.Error()
					if strings.Contains(errStr, "compile error") ||
						strings.Contains(errStr, "Syntax error") ||
						strings.Contains(errStr, "undeclared reference") {
						status = StatusSkipped
						details = "Query uses syntax not supported by CEL evaluator"
					} else {
						status = StatusFailed
						details = fmt.Sprintf("Evaluation error: %s", errStr)
					}
				case passed:
					status = StatusPassed
				default:
					status = StatusFailed
					// Details left empty - remediation is in separate field
				}

				// Use severity from check (default to medium if not specified)
				severity := check.Severity
				if severity == "" {
					severity = "medium"
				}

				// Get check ID
				checkID := check.UID
				if checkID == "" {
					checkID = check.ID
				}

				results = append(results, CheckResult{
					ID:          checkID,
					Group:       group.Title,
					Name:        check.Title,
					Status:      status,
					Severity:    severity,
					Details:     details,
					Remediation: check.Remediation,
					Docs:        check.Docs,
					Audit:       check.Audit,
				})
			}
		}
	}

	return results
}
