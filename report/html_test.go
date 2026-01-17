// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestHTMLExporter(t *testing.T) {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Provider:     "aws",
			ScanTarget:   "test-account",
			TotalChecks:  3,
			PassedChecks: 1,
			FailedChecks: 1,
			SkippedCheck: 1,
		},
		Rows: []Row{
			{
				ResourceType:  "aws_s3_bucket",
				ResourceName:  "my-bucket",
				ResourceID:    "bucket-123",
				ResourcePath:  "test-account > my-bucket",
				CheckID:       "s3-encryption",
				CheckGroup:    "S3",
				CheckName:     "S3 bucket encryption enabled",
				CheckStatus:   StatusPass,
				CheckSeverity: "high",
			},
			{
				ResourceType:  "aws_s3_bucket",
				ResourceName:  "insecure-bucket",
				ResourceID:    "bucket-456",
				ResourcePath:  "test-account > insecure-bucket",
				CheckID:       "s3-public-access",
				CheckGroup:    "S3",
				CheckName:     "S3 bucket public access blocked",
				CheckStatus:   StatusFailed,
				CheckSeverity: "critical",
				CheckDetails:  "Bucket allows public access",
			},
			{
				ResourceType:  "aws_ec2_instance",
				ResourceName:  "test-instance",
				ResourceID:    "i-123",
				ResourcePath:  "test-account > test-instance",
				CheckID:       "ec2-monitoring",
				CheckGroup:    "EC2",
				CheckName:     "EC2 detailed monitoring enabled",
				CheckStatus:   StatusSkipped,
				CheckSeverity: "low",
				CheckDetails:  "Not applicable for this instance type",
			},
		},
	}

	exporter := NewHTMLExporter()
	var buf bytes.Buffer
	err := exporter.ExportToWriter(report, &buf)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	output := buf.String()

	// Verify basic HTML structure
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("expected HTML to contain DOCTYPE declaration")
	}
	if !strings.Contains(output, "<html") {
		t.Error("expected HTML to contain <html> tag")
	}
	if !strings.Contains(output, "</html>") {
		t.Error("expected HTML to contain closing </html> tag")
	}

	// Verify metadata is present
	if !strings.Contains(output, "test-account") {
		t.Error("expected HTML to contain scan target 'test-account'")
	}
	if !strings.Contains(output, "aws") {
		t.Error("expected HTML to contain provider 'aws'")
	}

	// Verify summary counts
	if !strings.Contains(output, ">3<") { // Total checks
		t.Error("expected HTML to contain total checks count '3'")
	}

	// Verify failed check details are present
	if !strings.Contains(output, "S3 bucket public access blocked") {
		t.Error("expected HTML to contain failed check name")
	}
	if !strings.Contains(output, "Bucket allows public access") {
		t.Error("expected HTML to contain check details")
	}
	if !strings.Contains(output, "critical") {
		t.Error("expected HTML to contain severity 'critical'")
	}

	// Verify passed check is present
	if !strings.Contains(output, "S3 bucket encryption enabled") {
		t.Error("expected HTML to contain passed check name")
	}

	// Verify skipped check is present
	if !strings.Contains(output, "EC2 detailed monitoring enabled") {
		t.Error("expected HTML to contain skipped check name")
	}

	// Verify CSS is embedded
	if !strings.Contains(output, "<style>") {
		t.Error("expected HTML to contain embedded styles")
	}

	// Verify footer is present
	if !strings.Contains(output, "kspec") {
		t.Error("expected HTML to contain kspec branding in footer")
	}
}

func TestHTMLExporterEmptyReport(t *testing.T) {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt: time.Now(),
			Provider:    "github",
			ScanTarget:  "my-org",
		},
		Rows: []Row{},
	}

	exporter := NewHTMLExporter()
	var buf bytes.Buffer
	err := exporter.ExportToWriter(report, &buf)
	if err != nil {
		t.Fatalf("failed to export empty report: %v", err)
	}

	output := buf.String()

	// Should still produce valid HTML
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("expected empty report to produce valid HTML")
	}
	if !strings.Contains(output, "my-org") {
		t.Error("expected scan target in empty report")
	}
}

func TestHTMLExporterSeverityClass(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "severity-critical"},
		{"CRITICAL", "severity-critical"},
		{"high", "severity-high"},
		{"HIGH", "severity-high"},
		{"medium", "severity-medium"},
		{"low", "severity-low"},
		{"info", "severity-info"},
		{"", "severity-info"},
		{"unknown", "severity-info"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := severityClass(tt.severity)
			if result != tt.expected {
				t.Errorf("severityClass(%q) = %q, want %q", tt.severity, result, tt.expected)
			}
		})
	}
}

func TestHTMLExporterStatusClass(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"pass", "status-pass"},
		{"PASS", "status-pass"},
		{"passed", "status-pass"},
		{"fail", "status-fail"},
		{"FAIL", "status-fail"},
		{"failed", "status-fail"},
		{"skip", "status-skip"},
		{"skipped", "status-skip"},
		{"", "status-unknown"},
		{"unknown", "status-unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := statusClass(tt.status)
			if result != tt.expected {
				t.Errorf("statusClass(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

func TestHTMLExporterProgressPercentages(t *testing.T) {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt: time.Now(),
			Provider:    "test",
		},
		Rows: []Row{
			{CheckStatus: StatusPass},
			{CheckStatus: StatusPass},
			{CheckStatus: StatusFailed},
			{CheckStatus: StatusSkipped},
		},
	}

	exporter := NewHTMLExporter()
	var buf bytes.Buffer
	err := exporter.ExportToWriter(report, &buf)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	output := buf.String()

	// 2 passed out of 4 = 50%
	if !strings.Contains(output, "50.0%") {
		t.Error("expected HTML to contain 50.0% for passed checks")
	}
	// 1 failed out of 4 = 25%
	if !strings.Contains(output, "25.0%") {
		t.Error("expected HTML to contain 25.0% for failed/skipped checks")
	}
}

func TestHTMLExporterSeverityBadges(t *testing.T) {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt: time.Now(),
			Provider:    "test",
		},
		Rows: []Row{
			{CheckStatus: StatusFailed, CheckSeverity: "critical"},
			{CheckStatus: StatusFailed, CheckSeverity: "critical"},
			{CheckStatus: StatusFailed, CheckSeverity: "high"},
			{CheckStatus: StatusFailed, CheckSeverity: "medium"},
		},
	}

	exporter := NewHTMLExporter()
	var buf bytes.Buffer
	err := exporter.ExportToWriter(report, &buf)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	output := buf.String()

	// Should show severity badges with counts
	if !strings.Contains(output, "critical") {
		t.Error("expected HTML to contain 'critical' severity badge")
	}
	if !strings.Contains(output, "high") {
		t.Error("expected HTML to contain 'high' severity badge")
	}
}
