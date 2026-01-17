// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package report provides functionality to export scan results to various formats.
package report

import "time"

// Status constants for check results.
const (
	StatusPass    = "pass"
	StatusPassed  = "passed"
	StatusFail    = "fail"
	StatusFailed  = "failed"
	StatusSkip    = "skip"
	StatusSkipped = "skipped"
)

// Row represents a single row in the exported report.
type Row struct {
	// Resource identification
	ResourceType string `json:"resource_type" csv:"Resource Type"`
	ResourceName string `json:"resource_name" csv:"Resource Name"`
	ResourceID   string `json:"resource_id" csv:"Resource ID"`
	ResourcePath string `json:"resource_path" csv:"Resource Path"`

	// Check information
	CheckID       string `json:"check_id" csv:"Check ID"`
	CheckGroup    string `json:"check_group" csv:"Check Group"`
	CheckName     string `json:"check_name" csv:"Check Name"`
	CheckStatus   string `json:"check_status" csv:"Status"`
	CheckSeverity string `json:"check_severity" csv:"Severity"`
	CheckDetails  string `json:"check_details" csv:"Details"`
}

// Metadata contains metadata about the report.
type Metadata struct {
	GeneratedAt  time.Time `json:"generated_at"`
	ScanTarget   string    `json:"scan_target"`
	Provider     string    `json:"provider"`
	TotalChecks  int       `json:"total_checks"`
	PassedChecks int       `json:"passed_checks"`
	FailedChecks int       `json:"failed_checks"`
	SkippedCheck int       `json:"skipped_checks"`
}

// Report contains the full report data.
type Report struct {
	Metadata Metadata `json:"metadata"`
	Rows     []Row    `json:"rows"`
}

// Summary returns a summary of the report results.
type Summary struct {
	TotalResources   int            `json:"total_resources"`
	TotalChecks      int            `json:"total_checks"`
	PassedChecks     int            `json:"passed_checks"`
	FailedChecks     int            `json:"failed_checks"`
	SkippedChecks    int            `json:"skipped_checks"`
	BySeverity       map[string]int `json:"by_severity"`
	ByResourceType   map[string]int `json:"by_resource_type"`
	FailedBySeverity map[string]int `json:"failed_by_severity"`
}

// Format represents the export format.
type Format string

// Export format constants.
const (
	FormatCSV  Format = "csv"
	FormatXLSX Format = "xlsx"
	FormatJSON Format = "json"
	FormatHTML Format = "html"
)
