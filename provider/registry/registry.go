// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu        sync.RWMutex
	providers = make(map[string]*ProviderDefinition)
	aliases   = make(map[string]string) // alias -> primary name
)

// Register adds a provider definition to the global registry.
// This is typically called from init() functions in provider packages.
func Register(def *ProviderDefinition) {
	mu.Lock()
	defer mu.Unlock()

	if def.Name == "" {
		panic("provider definition must have a name")
	}

	if _, exists := providers[def.Name]; exists {
		panic(fmt.Sprintf("provider %q already registered", def.Name))
	}

	providers[def.Name] = def

	// Register aliases
	for _, alias := range def.Aliases {
		if existing, exists := aliases[alias]; exists {
			panic(fmt.Sprintf("alias %q already registered for provider %q", alias, existing))
		}
		aliases[alias] = def.Name
	}
}

// Get returns the provider definition for the given name or alias.
func Get(name string) (*ProviderDefinition, bool) {
	mu.RLock()
	defer mu.RUnlock()

	// Check if it's a direct provider name
	if def, ok := providers[name]; ok {
		return def, true
	}

	// Check if it's an alias
	if primaryName, ok := aliases[name]; ok {
		return providers[primaryName], true
	}

	return nil, false
}

// All returns all registered provider definitions.
func All() []*ProviderDefinition {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]*ProviderDefinition, 0, len(providers))
	for _, def := range providers {
		result = append(result, def)
	}

	// Sort by name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// Names returns all registered provider names (not aliases).
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]string, 0, len(providers))
	for name := range providers {
		result = append(result, name)
	}

	sort.Strings(result)
	return result
}

// Exists checks if a provider with the given name or alias exists.
func Exists(name string) bool {
	_, ok := Get(name)
	return ok
}

// MustGet returns the provider definition or panics if not found.
func MustGet(name string) *ProviderDefinition {
	def, ok := Get(name)
	if !ok {
		panic(fmt.Sprintf("provider %q not found", name))
	}
	return def
}

// AllFlags returns all unique flags across all providers.
// Flags are deduplicated by name.
func AllFlags() []FlagDefinition {
	mu.RLock()
	defer mu.RUnlock()

	seen := make(map[string]bool)
	var result []FlagDefinition

	for _, def := range providers {
		for _, flag := range def.Flags {
			if !seen[flag.Name] {
				seen[flag.Name] = true
				result = append(result, flag)
			}
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}
