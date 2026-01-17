// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package concurrency

import (
	"runtime"
)

// DefaultMaxWorkers is the default maximum number of workers when not configured.
const DefaultMaxWorkers = 20

// DefaultMultiplier is the multiplier applied to GOMAXPROCS for I/O-bound work.
const DefaultMultiplier = 2

// Config holds concurrency configuration.
type Config struct {
	// Auto enables adaptive concurrency based on system resources.
	// When true, worker count is calculated dynamically.
	// Default: true
	Auto bool `yaml:"auto"`

	// MaxWorkers is the hard cap on the number of concurrent workers.
	// Default: DefaultMaxWorkers (20)
	MaxWorkers int `yaml:"max_workers"`
}

// DefaultConfig returns the default concurrency configuration.
func DefaultConfig() Config {
	return Config{
		Auto:       true,
		MaxWorkers: DefaultMaxWorkers,
	}
}

// OptimalWorkers calculates the optimal number of workers based on configuration
// and the number of resources to process.
func OptimalWorkers(cfg Config, resourceCount int) int {
	maxWorkers := cfg.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = DefaultMaxWorkers
	}

	if !cfg.Auto {
		return maxWorkers
	}

	// Base worker count on CPU cores
	base := runtime.GOMAXPROCS(0)

	// For I/O-bound work (API calls), scale up from CPU count
	// but cap at the configured maximum
	optimal := min(base*DefaultMultiplier, maxWorkers)

	// Don't over-parallelize for small workloads
	if resourceCount > 0 && resourceCount < optimal {
		return max(1, resourceCount)
	}

	return max(1, optimal)
}

// DiscoveryWorkers returns the optimal worker count for resource discovery.
// Discovery is typically lighter weight, so we can use more workers.
func DiscoveryWorkers(cfg Config, resourceTypeCount int) int {
	return OptimalWorkers(cfg, resourceTypeCount)
}

// FetchWorkers returns the optimal worker count for fetching resource instances.
// Fetching involves more API calls, so we're more conservative.
func FetchWorkers(cfg Config, instanceCount int) int {
	return OptimalWorkers(cfg, instanceCount)
}

// EvaluationWorkers returns the optimal worker count for policy evaluation.
// Evaluation is CPU-bound, so we use exactly GOMAXPROCS.
func EvaluationWorkers(cfg Config) int {
	maxWorkers := cfg.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = DefaultMaxWorkers
	}

	if !cfg.Auto {
		return maxWorkers
	}

	// CPU-bound work: use exactly GOMAXPROCS
	return min(runtime.GOMAXPROCS(0), maxWorkers)
}
