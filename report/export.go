// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"time"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/policy/scoring"
)

// FromResourceTree converts a ResourceTree to a Report.
func FromResourceTree(tree *common.ResourceTree, provider string) *Report {
	return FromResourceTreeWithScoring(tree, provider, scoring.SystemBanded)
}

// FromResourceTreeWithScoring converts a ResourceTree to a Report with scoring.
func FromResourceTreeWithScoring(tree *common.ResourceTree, provider string, scoringSystem scoring.System) *Report {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt:   time.Now(),
			Provider:      provider,
			ScoringSystem: scoringSystem.String(),
		},
		Rows: make([]Row, 0),
	}

	if tree == nil || tree.Root == nil {
		report.Metadata.Score = 100
		report.Metadata.Grade = "A"
		report.Metadata.RiskLevel = "None"
		return report
	}

	// Set scan target from root
	report.Metadata.ScanTarget = tree.Root.Name

	// Collect all rows recursively
	collectRows(tree.Root, "", report)

	// Calculate totals
	for _, row := range report.Rows {
		report.Metadata.TotalChecks++
		switch row.CheckStatus {
		case StatusPass, StatusPassed:
			report.Metadata.PassedChecks++
		case StatusFail, StatusFailed:
			report.Metadata.FailedChecks++
		case StatusSkip, StatusSkipped:
			report.Metadata.SkippedCheck++
		}
	}

	// Calculate score using the graph scorer
	scorer := scoring.NewGraphScorer(scoringSystem)
	score := scorer.ScoreTree(tree)
	findings := scorer.GetRootFindings(tree)

	report.Metadata.Score = score.Value
	report.Metadata.Grade = score.Grade
	report.Metadata.RiskLevel = score.RiskLevel
	report.Metadata.CriticalFindings = findings.Critical
	report.Metadata.HighFindings = findings.High
	report.Metadata.MediumFindings = findings.Medium
	report.Metadata.LowFindings = findings.Low
	report.Metadata.InfoFindings = findings.Info

	return report
}

// collectRows recursively collects rows from a ResourceNode and its children.
func collectRows(node *common.ResourceNode, path string, report *Report) {
	if node == nil {
		return
	}

	// Build current path
	currentPath := path
	if currentPath != "" {
		currentPath += " > " + node.Name
	} else {
		currentPath = node.Name
	}

	// Add rows for each check on this node
	for _, check := range node.Checks {
		row := Row{
			ResourceType:     node.ResourceType,
			ResourceName:     node.Name,
			ResourceID:       node.ID,
			ResourcePath:     currentPath,
			CheckID:          check.ID,
			CheckGroup:       check.Group,
			CheckName:        check.Name,
			CheckStatus:      check.Status.String(),
			CheckSeverity:    check.Severity,
			CheckDetails:     check.Details,
			CheckRemediation: check.Remediation,
			CheckDocs:        check.Docs,
			CheckAudit:       check.Audit,
		}
		report.Rows = append(report.Rows, row)
	}

	// Process children
	for _, child := range node.Children {
		collectRows(child, currentPath, report)
	}
}

// Summary generates a summary from a Report.
func (r *Report) Summary() *Summary {
	summary := &Summary{
		BySeverity:       make(map[string]int),
		ByResourceType:   make(map[string]int),
		FailedBySeverity: make(map[string]int),
	}

	resourcesSeen := make(map[string]bool)

	for _, row := range r.Rows {
		// Count unique resources
		resourceKey := row.ResourceType + ":" + row.ResourceID
		if !resourcesSeen[resourceKey] {
			resourcesSeen[resourceKey] = true
			summary.TotalResources++
		}

		summary.TotalChecks++

		// Count by severity
		if row.CheckSeverity != "" {
			summary.BySeverity[row.CheckSeverity]++
		}

		// Count by resource type
		if row.ResourceType != "" {
			summary.ByResourceType[row.ResourceType]++
		}

		// Count by status
		switch row.CheckStatus {
		case StatusPass, StatusPassed:
			summary.PassedChecks++
		case StatusFail, StatusFailed:
			summary.FailedChecks++
			if row.CheckSeverity != "" {
				summary.FailedBySeverity[row.CheckSeverity]++
			}
		case StatusSkip, StatusSkipped:
			summary.SkippedChecks++
		}
	}

	return summary
}
