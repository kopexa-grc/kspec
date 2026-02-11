// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Example: Advanced scanner with events, scoring, and report export.
//
// This example demonstrates:
//   - Listening to scan events for progress tracking
//   - Using the scoring API (NewGraphScorer + ScoreTree)
//   - Exporting reports to JSON and CSV
//
// Usage:
//
//	go run ./_examples/check-engine/advanced/ example.com
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

	// --- Load policy ---
	policyPath := policyFile()

	policies, err := policy.Load(policyPath, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load policy: %v\n", err)
		os.Exit(1)
	}

	policies = policy.FilterByProvider(policies, "network", "network")

	// --- Create scanner ---
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

	// --- Listen to events ---
	s.OnEvent(func(event scanner.ScanEvent) {
		switch event.Type {
		case scanner.EventDiscoveryStarted:
			fmt.Fprintln(os.Stderr, "[event] Discovery started")
		case scanner.EventDiscoveryComplete:
			fmt.Fprintln(os.Stderr, "[event] Discovery complete")
		case scanner.EventScanStarted:
			fmt.Fprintln(os.Stderr, "[event] Scan started")
		case scanner.EventResourceScanning:
			fmt.Fprintf(os.Stderr, "[event] Scanning: %s\n", event.ResourceType)
		case scanner.EventResourceComplete:
			fmt.Fprintf(os.Stderr, "[event] Complete: %s\n", event.ResourceType)
		case scanner.EventScanComplete:
			fmt.Fprintln(os.Stderr, "[event] Scan complete")
		case scanner.EventError:
			fmt.Fprintf(os.Stderr, "[event] Error: %v\n", event.Error)
		}
	})

	// --- Initialize and run ---
	if err := s.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Initialize failed: %v\n", err)
		os.Exit(1)
	}

	result := s.Run(ctx)
	for _, scanErr := range result.Errors {
		fmt.Fprintf(os.Stderr, "Error: %v\n", scanErr)
	}

	if result.Tree == nil || result.Tree.Root == nil {
		fmt.Fprintln(os.Stderr, "No results.")
		os.Exit(1)
	}

	// --- Scoring ---
	scorer := scoring.NewGraphScorer(scoring.SystemBanded)
	score := scorer.ScoreTree(result.Tree)
	findings := scorer.GetRootFindings(result.Tree)

	fmt.Println("\n=== Score Summary ===")
	fmt.Printf("Score:      %d/100\n", score.Value)
	fmt.Printf("Grade:      %s\n", score.Grade)
	fmt.Printf("Risk Level: %s\n", score.RiskLevel)
	fmt.Printf("Findings:   %d critical, %d high, %d medium, %d low\n",
		findings.Critical, findings.High, findings.Medium, findings.Low)

	// --- Export reports ---
	rep := report.FromResourceTreeWithScoring(result.Tree, "network", scoring.SystemBanded)

	// JSON report
	jsonPath := "report.json"
	if err := report.NewJSONExporter(true).Export(rep, jsonPath); err != nil {
		fmt.Fprintf(os.Stderr, "JSON export failed: %v\n", err)
	} else {
		fmt.Printf("\nJSON report: %s\n", jsonPath)
	}

	// CSV report
	csvPath := "report.csv"
	if err := report.NewCSVExporter().Export(rep, csvPath); err != nil {
		fmt.Fprintf(os.Stderr, "CSV export failed: %v\n", err)
	} else {
		fmt.Printf("CSV report:  %s\n", csvPath)
	}
}

// policyFile returns the path to the bundled policy.yaml next to this example.
func policyFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "policy.yaml")
}
