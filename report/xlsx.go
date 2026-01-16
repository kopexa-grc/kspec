// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// XLSXExporter exports reports to XLSX (Excel) format.
type XLSXExporter struct{}

// NewXLSXExporter creates a new XLSX exporter.
func NewXLSXExporter() *XLSXExporter {
	return &XLSXExporter{}
}

// Export writes the report to an XLSX file.
func (e *XLSXExporter) Export(report *Report, path string) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close() //nolint:errcheck // Best effort close
	}()

	// Create Results sheet
	resultsSheet := "Results"
	index, err := f.NewSheet(resultsSheet)
	if err != nil {
		return fmt.Errorf("failed to create Results sheet: %w", err)
	}

	// Delete default Sheet1
	if err := f.DeleteSheet("Sheet1"); err != nil {
		// Ignore error if sheet doesn't exist
		_ = err
	}

	// Set Results as active sheet
	f.SetActiveSheet(index)

	// Define header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create header style: %w", err)
	}

	// Write header
	headers := []string{
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
	}

	for col, header := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return fmt.Errorf("failed to get cell name: %w", err)
		}
		if err := f.SetCellValue(resultsSheet, cell, header); err != nil {
			return fmt.Errorf("failed to set header cell: %w", err)
		}
		if err := f.SetCellStyle(resultsSheet, cell, cell, headerStyle); err != nil {
			return fmt.Errorf("failed to set header style: %w", err)
		}
	}

	// Define status styles
	passStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#C6EFCE"},
			Pattern: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create pass style: %w", err)
	}
	failStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFC7CE"},
			Pattern: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create fail style: %w", err)
	}
	skipStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFEB9C"},
			Pattern: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create skip style: %w", err)
	}

	// Write data rows
	for rowIdx, row := range report.Rows {
		rowNum := rowIdx + 2 // Start from row 2 (after header)

		values := []interface{}{
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
		}

		for col, value := range values {
			cell, err := excelize.CoordinatesToCellName(col+1, rowNum)
			if err != nil {
				return fmt.Errorf("failed to get cell name: %w", err)
			}
			if err := f.SetCellValue(resultsSheet, cell, value); err != nil {
				return fmt.Errorf("failed to set cell value: %w", err)
			}

			// Apply status color to the Status column (column 8)
			if col == 7 {
				var style int
				switch row.CheckStatus {
				case StatusPass, StatusPassed:
					style = passStyle
				case StatusFail, StatusFailed:
					style = failStyle
				case StatusSkip, StatusSkipped:
					style = skipStyle
				}
				if style != 0 {
					if err := f.SetCellStyle(resultsSheet, cell, cell, style); err != nil {
						return fmt.Errorf("failed to set status style: %w", err)
					}
				}
			}
		}
	}

	// Auto-fit column widths (approximate)
	colWidths := map[string]float64{
		"A": 15, // Resource Type
		"B": 25, // Resource Name
		"C": 30, // Resource ID
		"D": 40, // Resource Path
		"E": 20, // Check ID
		"F": 20, // Check Group
		"G": 30, // Check Name
		"H": 10, // Status
		"I": 10, // Severity
		"J": 50, // Details
	}
	for col, width := range colWidths {
		if err := f.SetColWidth(resultsSheet, col, col, width); err != nil {
			return fmt.Errorf("failed to set column width: %w", err)
		}
	}

	// Create Summary sheet
	summarySheet := "Summary"
	if _, err := f.NewSheet(summarySheet); err != nil {
		return fmt.Errorf("failed to create Summary sheet: %w", err)
	}

	summary := report.Summary()
	summaryData := [][]interface{}{
		{"Summary", ""},
		{"", ""},
		{"Total Resources", summary.TotalResources},
		{"Total Checks", summary.TotalChecks},
		{"Passed Checks", summary.PassedChecks},
		{"Failed Checks", summary.FailedChecks},
		{"Skipped Checks", summary.SkippedChecks},
		{"", ""},
		{"Report Generated", report.Metadata.GeneratedAt.Format("2006-01-02 15:04:05")},
		{"Provider", report.Metadata.Provider},
		{"Scan Target", report.Metadata.ScanTarget},
	}

	for rowIdx, rowData := range summaryData {
		for col, value := range rowData {
			cell, err := excelize.CoordinatesToCellName(col+1, rowIdx+1)
			if err != nil {
				return fmt.Errorf("failed to get summary cell name: %w", err)
			}
			if err := f.SetCellValue(summarySheet, cell, value); err != nil {
				return fmt.Errorf("failed to set summary cell: %w", err)
			}
		}
	}

	// Style summary header
	if err := f.SetCellStyle(summarySheet, "A1", "A1", headerStyle); err != nil {
		return fmt.Errorf("failed to set summary header style: %w", err)
	}

	// Set summary column widths
	if err := f.SetColWidth(summarySheet, "A", "A", 20); err != nil {
		return fmt.Errorf("failed to set summary column A width: %w", err)
	}
	if err := f.SetColWidth(summarySheet, "B", "B", 30); err != nil {
		return fmt.Errorf("failed to set summary column B width: %w", err)
	}

	// Save file
	if err := f.SaveAs(path); err != nil {
		return fmt.Errorf("failed to save XLSX file: %w", err)
	}

	return nil
}
