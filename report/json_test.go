// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestJSONExporter(t *testing.T) {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt:  time.Now(),
			Provider:     "test",
			ScanTarget:   "target",
			TotalChecks:  1,
			PassedChecks: 1,
		},
		Rows: []Row{
			{
				ResourceType:  "vm",
				ResourceName:  "test-vm",
				ResourceID:    "vm-123",
				CheckID:       "check-1",
				CheckStatus:   StatusPass,
				CheckSeverity: "high",
			},
		},
	}

	exporter := NewJSONExporter(false)
	var buf bytes.Buffer
	err := exporter.ExportToWriter(report, &buf)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var parsed Report
	err = json.Unmarshal(buf.Bytes(), &parsed)
	if err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.Metadata.Provider != "test" {
		t.Errorf("expected provider 'test', got '%s'", parsed.Metadata.Provider)
	}
	if len(parsed.Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(parsed.Rows))
	}
	if parsed.Rows[0].CheckID != "check-1" {
		t.Errorf("expected check ID 'check-1', got '%s'", parsed.Rows[0].CheckID)
	}
}

func TestJSONExporterPretty(t *testing.T) {
	report := &Report{
		Metadata: Metadata{
			GeneratedAt: time.Now(),
			Provider:    "test",
		},
		Rows: []Row{},
	}

	exporter := NewJSONExporter(true)
	var buf bytes.Buffer
	err := exporter.ExportToWriter(report, &buf)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	output := buf.String()

	// Pretty output should have newlines and indentation
	if !bytes.Contains(buf.Bytes(), []byte("\n")) {
		t.Error("expected pretty JSON to have newlines")
	}
	if !bytes.Contains(buf.Bytes(), []byte("  ")) {
		t.Error("expected pretty JSON to have indentation")
	}

	// Still valid JSON
	var parsed Report
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Fatalf("pretty output is not valid JSON: %v", err)
	}
}
