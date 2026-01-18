// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// OutputFormat represents the output format for discovery results.
type OutputFormat string

const (
	// FormatJSON outputs results as JSON.
	FormatJSON OutputFormat = "json"
	// FormatTable outputs results as a formatted table.
	FormatTable OutputFormat = "table"
	// FormatTree outputs results as a tree structure.
	FormatTree OutputFormat = "tree"
)

// Export writes the discovery result in the specified format.
func Export(result *Result, format OutputFormat, w io.Writer) error {
	switch format {
	case FormatJSON:
		return exportJSON(result, w)
	case FormatTable:
		return exportTable(result, w)
	case FormatTree:
		return exportTree(result, w)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func exportJSON(result *Result, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func exportTable(result *Result, w io.Writer) error {
	// Header
	fmt.Fprintf(w, "Provider: %s | Asset: %s (%s)\n", result.Provider, result.AssetName, result.AssetType)                                        //nolint:errcheck // Best effort output
	fmt.Fprintf(w, "Discovered at: %s | Duration: %s\n\n", result.DiscoveredAt.Format("2006-01-02 15:04:05"), result.Duration.Round(100*1000000)) //nolint:errcheck // Best effort output

	// Sort resources by type
	resources := make([]ResourceInfo, len(result.Resources))
	copy(resources, result.Resources)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Type < resources[j].Type
	})

	// Find max width for resource type column
	maxTypeLen := len("RESOURCE TYPE")
	for _, r := range resources {
		if len(r.Type) > maxTypeLen {
			maxTypeLen = len(r.Type)
		}
	}

	// Table header
	fmt.Fprintf(w, "%-*s %10s\n", maxTypeLen, "RESOURCE TYPE", "COUNT") //nolint:errcheck // Best effort output
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", maxTypeLen+12))          //nolint:errcheck // Best effort output

	// Rows
	for _, r := range resources {
		countStr := fmt.Sprintf("%d", r.Count)
		if r.Count < 0 {
			countStr = "unknown"
		}
		fmt.Fprintf(w, "%-*s %10s\n", maxTypeLen, r.Type, countStr) //nolint:errcheck // Best effort output
	}

	// Footer
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", maxTypeLen+12))                                 //nolint:errcheck // Best effort output
	fmt.Fprintf(w, "Total: %d resources across %d types\n", result.TotalCount, len(resources)) //nolint:errcheck // Best effort output

	// Errors
	if len(result.Errors) > 0 {
		fmt.Fprintf(w, "\nErrors (%d):\n", len(result.Errors)) //nolint:errcheck // Best effort output
		for _, e := range result.Errors {
			fmt.Fprintf(w, "  - %s: %s\n", e.ResourceType, e.Message) //nolint:errcheck // Best effort output
		}
	}

	return nil
}

func exportTree(result *Result, w io.Writer) error {
	// Root
	fmt.Fprintf(w, "%s (%s)\n", result.AssetName, result.AssetType) //nolint:errcheck // Best effort output

	// Group resources by prefix (e.g., aws_s3, aws_ec2, github_)
	groups := make(map[string][]ResourceInfo)
	for _, r := range result.Resources {
		parts := strings.SplitN(r.Type, "_", 3)
		prefix := ""
		if len(parts) >= 2 {
			prefix = parts[0] + "_" + parts[1]
		} else {
			prefix = parts[0]
		}
		groups[prefix] = append(groups[prefix], r)
	}

	// Sort group keys
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		resources := groups[key]
		isLast := i == len(keys)-1

		prefix := "├── "
		childPrefix := "│   "
		if isLast {
			prefix = "└── "
			childPrefix = "    "
		}

		fmt.Fprintf(w, "%s%s\n", prefix, key) //nolint:errcheck // Best effort output

		for j, r := range resources {
			isLastChild := j == len(resources)-1
			childPre := childPrefix + "├── "
			if isLastChild {
				childPre = childPrefix + "└── "
			}

			countStr := ""
			if r.Count >= 0 {
				countStr = fmt.Sprintf(" (%d)", r.Count)
			}

			fmt.Fprintf(w, "%s%s%s\n", childPre, r.Type, countStr) //nolint:errcheck // Best effort output
		}
	}

	return nil
}

// ExportGraph exports just the graph portion of the result.
func ExportGraph(graph *Graph, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(graph)
}
