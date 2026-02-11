// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Example: Discovery with graph export using the kspec Go API.
//
// DiscoverWithInstances includes individual resource instances in the result,
// which enables graph generation. The graph contains nodes (resources) and
// edges (relationships) suitable for visualization tools.
//
// Usage:
//
//	go run ./_examples/discovery/graph/ example.com
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kopexa-grc/kspec/discovery"

	// IMPORTANT: Register all providers via blank import.
	_ "github.com/kopexa-grc/kspec/provider/all"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <domain>\n", os.Args[0])
		os.Exit(1)
	}

	target := os.Args[1]
	ctx := context.Background()

	// DiscoverWithInstances populates result.Graph with nodes and edges.
	// This uses more memory than Discover() but enables graph export.
	result, err := discovery.DiscoverWithInstances(ctx,
		"network",
		map[string]string{"target": target},
		"host",
		target,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Discovery failed: %v\n", err)
		os.Exit(1)
	}

	// Print the tree view first for a quick overview.
	fmt.Fprintln(os.Stderr, "--- Resource Tree ---")
	if err := discovery.Export(result, discovery.FormatTree, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Tree export failed: %v\n", err)
	}

	// Export the graph as JSON to stdout.
	if result.Graph != nil {
		fmt.Fprintln(os.Stderr, "\n--- Graph JSON (stdout) ---")
		if err := discovery.ExportGraph(result.Graph, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Graph export failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stderr, "No graph data available.")
	}
}
