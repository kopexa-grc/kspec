// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Example: Basic resource discovery using the kspec Go API.
//
// This example discovers resources for a network host and prints
// the results as a formatted table.
//
// Usage:
//
//	go run ./_examples/discovery/basic/ example.com
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kopexa-grc/kspec/discovery"

	// IMPORTANT: Import the provider/all package to register all providers.
	// Without this import, no providers are available and discovery will fail.
	_ "github.com/kopexa-grc/kspec/provider/all"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <domain>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s example.com\n", os.Args[0])
		os.Exit(1)
	}

	target := os.Args[1]
	ctx := context.Background()

	// Discover resources for a network host.
	// The network provider requires no credentials — just a target domain.
	result, err := discovery.Discover(ctx,
		"network",                          // provider name
		map[string]string{"target": target}, // provider config
		"host",  // asset type
		target,  // asset name
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Discovery failed: %v\n", err)
		os.Exit(1)
	}

	// Export results as a formatted table to stdout.
	if err := discovery.Export(result, discovery.FormatTable, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
		os.Exit(1)
	}
}
