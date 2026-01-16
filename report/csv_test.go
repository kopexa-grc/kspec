// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestCSVExporter(t *testing.T) {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt:  time.Now(),
			Provider:     "test",
			ScanTarget:   "target",
			TotalChecks:  2,
			PassedChecks: 1,
			FailedChecks: 1,
		},
		Rows: []Row{
			{
				ResourceType:  "vm",
				ResourceName:  "test-vm",
				ResourceID:    "vm-123",
				ResourcePath:  "root > test-vm",
				CheckID:       "check-1",
				CheckGroup:    "security",
				CheckName:     "Test Check",
				CheckStatus:   StatusPass,
				CheckSeverity: "high",
				CheckDetails:  "All good",
			},
			{
				ResourceType:  "vm",
				ResourceName:  "test-vm-2",
				ResourceID:    "vm-456",
				ResourcePath:  "root > test-vm-2",
				CheckID:       "check-2",
				CheckGroup:    "compliance",
				CheckName:     "Another Check",
				CheckStatus:   StatusFail,
				CheckSeverity: "medium",
				CheckDetails:  "Issue found",
			},
		},
	}

	exporter := NewCSVExporter()
	var buf bytes.Buffer
	err := exporter.ExportToWriter(report, &buf)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "Resource Type,Resource Name,Resource ID,Resource Path,Check ID,Check Group,Check Name,Status,Severity,Details") {
		t.Error("expected CSV header not found")
	}

	// Check data rows
	if !strings.Contains(output, "vm,test-vm,vm-123") {
		t.Error("expected first row data not found")
	}
	if !strings.Contains(output, "vm,test-vm-2,vm-456") {
		t.Error("expected second row data not found")
	}
}
