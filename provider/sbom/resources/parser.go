// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Format represents the detected SBOM format.
type Format string

// SBOM format constants.
const (
	FormatCycloneDX Format = "cyclonedx"
	FormatSPDX      Format = "spdx"
	FormatUnknown   Format = "unknown"
)

// DetectFormat detects the SBOM format from file content.
func DetectFormat(data []byte) Format {
	// Try to detect CycloneDX
	if strings.Contains(string(data), `"bomFormat"`) || strings.Contains(string(data), `bomFormat=`) {
		return FormatCycloneDX
	}

	// Try to detect SPDX
	if strings.Contains(string(data), `"spdxVersion"`) || strings.Contains(string(data), `SPDXVersion:`) {
		return FormatSPDX
	}

	return FormatUnknown
}

// ParseSBOMFiles parses SBOM files from the given path.
func ParseSBOMFiles(path string, isDir bool) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	if isDir {
		// Walk directory for SBOM files
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Check for common SBOM file extensions
			ext := strings.ToLower(filepath.Ext(filePath))
			name := strings.ToLower(filepath.Base(filePath))
			if ext == ".json" || ext == ".xml" || strings.Contains(name, "sbom") || strings.Contains(name, "bom") {
				sbom, err := ParseSBOMFile(filePath)
				if err == nil && sbom != nil {
					results = append(results, sbom)
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Parse single file
		sbom, err := ParseSBOMFile(path)
		if err != nil {
			return nil, err
		}
		if sbom != nil {
			results = append(results, sbom)
		}
	}

	return results, nil
}

// ParseSBOMFile parses a single SBOM file.
func ParseSBOMFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	format := DetectFormat(data)
	if format == FormatUnknown {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	result["_file_path"] = filePath
	result["_format"] = string(format)

	return result, nil
}
