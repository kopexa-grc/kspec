// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package discovery provides resource discovery for kspec providers.
//
// Discovery allows you to inventory cloud resources without evaluating policies.
// This is useful for:
//   - Understanding what resources exist in an environment
//   - Building resource graphs for visualization
//   - Integration with other tools (via JSON export)
//   - Pre-scan inventory audits
//
// # Quick Start
//
// For simple discovery:
//
//	result, err := discovery.Discover(ctx, "github",
//	    map[string]string{"token": "ghp_xxx"},
//	    "github-org", "my-org")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d resources\n", result.TotalCount)
//
// # With Graph Data
//
// To include resource instances and graph data for visualization:
//
//	result, err := discovery.DiscoverWithInstances(ctx, "github",
//	    map[string]string{"token": "ghp_xxx"},
//	    "github-org", "my-org")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// result.Graph contains nodes and edges
//	discovery.ExportGraph(result.Graph, os.Stdout)
//
// # Quick Inventory
//
// For just resource counts:
//
//	inventory, err := discovery.QuickInventory(ctx, "aws",
//	    map[string]string{"region": "us-east-1"},
//	    "aws-account", "production")
//	// inventory is map[string]int: {"aws_s3_bucket": 42, "aws_iam_user": 15, ...}
//
// # Output Formats
//
// Results can be exported in multiple formats:
//
//	// JSON
//	discovery.Export(result, discovery.FormatJSON, os.Stdout)
//
//	// Table
//	discovery.Export(result, discovery.FormatTable, os.Stdout)
//
//	// Tree
//	discovery.Export(result, discovery.FormatTree, os.Stdout)
//
// # CLI Usage
//
// The discovery functionality is also available via CLI:
//
//	kspec discover github org my-org
//	kspec discover aws account --region us-east-1 -o json
//	kspec discover azure subscription my-sub -o tree
package discovery
