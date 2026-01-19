// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

// FilterByProvider filters policies to only include those matching the provider.
func FilterByProvider(policies []Policy, providerName, providerAlias string) []Policy {
	var filtered []Policy

	for _, p := range policies {
		if len(p.Require) == 0 {
			// No requirements - policy applies to all providers
			filtered = append(filtered, p)
			continue
		}

		// Check if any requirement matches the current provider
		for _, req := range p.Require {
			if req.Provider == providerName || req.Provider == providerAlias {
				filtered = append(filtered, p)
				break
			}
		}
	}

	return filtered
}
