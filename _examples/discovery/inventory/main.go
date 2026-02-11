// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Example: Quick inventory using the kspec Go API.
//
// QuickInventory returns a map of resource type → count, which is the
// fastest way to get an overview of what resources a provider exposes.
// The output is printed as JSON for easy integration with other tools.
//
// Usage:
//
//	go run ./_examples/discovery/inventory/ example.com
package main

import (
	"context"
	"encoding/json"
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

	// QuickInventory returns map[string]int — resource type to count.
	inventory, err := discovery.QuickInventory(ctx,
		"network",
		map[string]string{"target": target},
		"host",
		target,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Inventory failed: %v\n", err)
		os.Exit(1)
	}

	// Output as JSON for scripting / piping.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(inventory); err != nil {
		fmt.Fprintf(os.Stderr, "JSON encode failed: %v\n", err)
		os.Exit(1)
	}
}
