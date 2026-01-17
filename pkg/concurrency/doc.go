// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package concurrency provides utilities for adaptive concurrency control.
//
// This package calculates optimal worker counts based on system resources
// and workload characteristics. It supports both automatic scaling and
// manual configuration.
//
// # Usage
//
//	cfg := concurrency.Config{Auto: true, MaxWorkers: 20}
//	workers := concurrency.OptimalWorkers(cfg, resourceCount)
//
// # Adaptive Concurrency
//
// When Auto is true, the number of workers is calculated based on:
//   - Available CPU cores (runtime.GOMAXPROCS)
//   - Number of resources to process
//   - Maximum worker limit
//
// For I/O-bound work like API calls, the default multiplier is 2x CPU cores.
package concurrency
