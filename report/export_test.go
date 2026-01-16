// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"testing"

	"github.com/kopexa-grc/kspec/cli/components/common"
)

func TestFromResourceTree(t *testing.T) {
	// Create a test tree
	tree := common.NewResourceTree("Test Scan", "root")
	tree.Root.ResourceType = "test"

	// Add some checks
	tree.Root.Checks = []common.CheckResult{
		{ID: "check1", Name: "Test Check 1", Status: "pass", Severity: "high"},
		{ID: "check2", Name: "Test Check 2", Status: "fail", Severity: "medium"},
		{ID: "check3", Name: "Test Check 3", Status: "skip", Severity: "low"},
	}

	// Convert to report
	report := FromResourceTree(tree, "test-provider")

	// Verify metadata
	if report.Metadata.Provider != "test-provider" {
		t.Errorf("expected provider 'test-provider', got '%s'", report.Metadata.Provider)
	}
	if report.Metadata.ScanTarget != "Test Scan" {
		t.Errorf("expected scan target 'Test Scan', got '%s'", report.Metadata.ScanTarget)
	}
	if report.Metadata.TotalChecks != 3 {
		t.Errorf("expected 3 total checks, got %d", report.Metadata.TotalChecks)
	}
	if report.Metadata.PassedChecks != 1 {
		t.Errorf("expected 1 passed check, got %d", report.Metadata.PassedChecks)
	}
	if report.Metadata.FailedChecks != 1 {
		t.Errorf("expected 1 failed check, got %d", report.Metadata.FailedChecks)
	}
	if report.Metadata.SkippedCheck != 1 {
		t.Errorf("expected 1 skipped check, got %d", report.Metadata.SkippedCheck)
	}

	// Verify rows
	if len(report.Rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(report.Rows))
	}
}

func TestFromResourceTreeNil(t *testing.T) {
	rep := FromResourceTree(nil, "test-provider")
	if rep == nil {
		t.Error("expected non-nil report from nil tree")
		return
	}
	if len(rep.Rows) != 0 {
		t.Errorf("expected 0 rows from nil tree, got %d", len(rep.Rows))
	}
}

func TestReportSummary(t *testing.T) {
	report := &Report{
		Rows: []Row{
			{ResourceType: "type1", ResourceID: "id1", CheckStatus: StatusPass, CheckSeverity: "high"},
			{ResourceType: "type1", ResourceID: "id1", CheckStatus: StatusFail, CheckSeverity: "high"},
			{ResourceType: "type2", ResourceID: "id2", CheckStatus: StatusPass, CheckSeverity: "medium"},
			{ResourceType: "type2", ResourceID: "id3", CheckStatus: StatusSkip, CheckSeverity: "low"},
		},
	}

	summary := report.Summary()

	if summary.TotalResources != 3 {
		t.Errorf("expected 3 total resources, got %d", summary.TotalResources)
	}
	if summary.TotalChecks != 4 {
		t.Errorf("expected 4 total checks, got %d", summary.TotalChecks)
	}
	if summary.PassedChecks != 2 {
		t.Errorf("expected 2 passed checks, got %d", summary.PassedChecks)
	}
	if summary.FailedChecks != 1 {
		t.Errorf("expected 1 failed check, got %d", summary.FailedChecks)
	}
	if summary.SkippedChecks != 1 {
		t.Errorf("expected 1 skipped check, got %d", summary.SkippedChecks)
	}
	if summary.FailedBySeverity["high"] != 1 {
		t.Errorf("expected 1 high severity failure, got %d", summary.FailedBySeverity["high"])
	}
}
