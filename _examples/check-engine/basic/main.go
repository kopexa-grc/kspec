// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Example: Basic scanner API usage (end-to-end scan).
//
// This example loads a policy, creates a scanner, and runs a full
// scan against a network host. Results are printed as a summary.
//
// Usage:
//
//	go run ./_examples/check-engine/basic/ example.com
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/policy"
	"github.com/kopexa-grc/kspec/policy/scoring"
	"github.com/kopexa-grc/kspec/provider/scanner"
	"github.com/kopexa-grc/kspec/report"

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

	// --- Step 1: Load policy ---
	policyPath := policyFile()

	policies, err := policy.Load(policyPath, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load policy: %v\n", err)
		os.Exit(1)
	}

	// Filter policies for the network provider.
	policies = policy.FilterByProvider(policies, "network", "network")
	if len(policies) == 0 {
		fmt.Fprintf(os.Stderr, "No policies found for network provider\n")
		os.Exit(1)
	}

	// --- Step 2: Create scanner ---
	s := scanner.NewScanner(scanner.ScanConfig{
		ProviderName:   "network",
		ProviderConfig: map[string]string{"target": target},
		Asset: core.Asset{
			Type:   "host",
			Name:   target,
			Config: map[string]string{"target": target},
		},
		Policies: policies,
	})

	// --- Step 3: Initialize provider ---
	if err := s.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	// --- Step 4: Run scan ---
	result := s.Run(ctx)

	// Print errors if any.
	for _, scanErr := range result.Errors {
		fmt.Fprintf(os.Stderr, "Error: %v\n", scanErr)
	}

	if result.Tree == nil || result.Tree.Root == nil {
		fmt.Fprintln(os.Stderr, "No results.")
		os.Exit(1)
	}

	// --- Step 5: Score and report ---
	rep := report.FromResourceTreeWithScoring(result.Tree, "network", scoring.SystemBanded)

	fmt.Printf("Target:  %s\n", rep.Metadata.ScanTarget)
	fmt.Printf("Score:   %d/100 (Grade: %s)\n", rep.Metadata.Score, rep.Metadata.Grade)
	fmt.Printf("Risk:    %s\n", rep.Metadata.RiskLevel)
	fmt.Printf("Checks:  %d total, %d passed, %d failed, %d skipped\n",
		rep.Metadata.TotalChecks,
		rep.Metadata.PassedChecks,
		rep.Metadata.FailedChecks,
		rep.Metadata.SkippedCheck,
	)

	// Exit with non-zero code if there are failed checks.
	if rep.Metadata.FailedChecks > 0 {
		os.Exit(2)
	}
}

// policyFile returns the path to the bundled policy.yaml next to this example.
func policyFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "policy.yaml")
}
