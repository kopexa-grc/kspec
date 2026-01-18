// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package suites

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = make(map[string]Suite)
)

// Register adds a suite to the global registry.
// Called from init() functions in suite packages.
// Panics if suite is nil, has no name, or name is already registered.
func Register(suite Suite) {
	mu.Lock()
	defer mu.Unlock()

	if suite == nil {
		panic("suites: suite must not be nil")
	}

	name := suite.Name()
	if name == "" {
		panic("suites: suite must have a name")
	}

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("suites: suite %q already registered", name))
	}

	registry[name] = suite
}

// Get returns the suite with the given name.
// Returns nil and false if not found.
func Get(name string) (Suite, bool) {
	mu.RLock()
	defer mu.RUnlock()

	suite, ok := registry[name]
	return suite, ok
}

// MustGet returns the suite with the given name.
// Panics if the suite is not found.
func MustGet(name string) Suite {
	suite, ok := Get(name)
	if !ok {
		panic(fmt.Sprintf("suites: suite %q not found", name))
	}
	return suite
}

// All returns all registered suites, sorted by name.
func All() []Suite {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]Suite, 0, len(registry))
	for _, s := range registry {
		result = append(result, s)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})

	return result
}

// ForProvider returns all suites for a specific provider, sorted by name.
func ForProvider(provider string) []Suite {
	mu.RLock()
	defer mu.RUnlock()

	var result []Suite
	for _, s := range registry {
		if s.Provider() == provider {
			result = append(result, s)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})

	return result
}

// Names returns all registered suite names, sorted alphabetically.
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Providers returns all unique provider names that have registered suites.
func Providers() []string {
	mu.RLock()
	defer mu.RUnlock()

	providerSet := make(map[string]struct{})
	for _, s := range registry {
		providerSet[s.Provider()] = struct{}{}
	}

	providers := make([]string, 0, len(providerSet))
	for p := range providerSet {
		providers = append(providers, p)
	}

	sort.Strings(providers)
	return providers
}

// Reset clears all registered suites.
// Only for use in tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]Suite)
}
