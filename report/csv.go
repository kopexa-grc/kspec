// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// CSVExporter exports reports to CSV format.
type CSVExporter struct{}

// NewCSVExporter creates a new CSV exporter.
func NewCSVExporter() *CSVExporter {
	return &CSVExporter{}
}

// Export writes the report to a CSV file.
func (e *CSVExporter) Export(report *Report, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Best effort close
	}()

	return e.ExportToWriter(report, file)
}

// ExportToWriter writes the report to an io.Writer in CSV format.
func (e *CSVExporter) ExportToWriter(report *Report, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"Resource Type",
		"Resource Name",
		"Resource ID",
		"Resource Path",
		"Check ID",
		"Check Group",
		"Check Name",
		"Status",
		"Severity",
		"Details",
		"Remediation",
		"Documentation",
		"Audit",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, row := range report.Rows {
		record := []string{
			row.ResourceType,
			row.ResourceName,
			row.ResourceID,
			row.ResourcePath,
			row.CheckID,
			row.CheckGroup,
			row.CheckName,
			row.CheckStatus,
			row.CheckSeverity,
			row.CheckDetails,
			row.CheckRemediation,
			row.CheckDocs,
			row.CheckAudit,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
