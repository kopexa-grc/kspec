// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// JSONExporter exports reports to JSON format.
type JSONExporter struct {
	// Pretty enables indented JSON output.
	Pretty bool
}

// NewJSONExporter creates a new JSON exporter.
func NewJSONExporter(pretty bool) *JSONExporter {
	return &JSONExporter{Pretty: pretty}
}

// Export writes the report to a JSON file.
func (e *JSONExporter) Export(report *Report, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Best effort close
	}()

	return e.ExportToWriter(report, file)
}

// ExportToWriter writes the report to an io.Writer in JSON format.
func (e *JSONExporter) ExportToWriter(report *Report, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if e.Pretty {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode report to JSON: %w", err)
	}

	return nil
}
