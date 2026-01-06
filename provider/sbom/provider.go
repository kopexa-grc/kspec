package sbom

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// SBOMProvider implements the core.Provider interface for SBOM (Software Bill of Materials) scanning.
type SBOMProvider struct{}

// NewSBOMProvider creates a new SBOM provider.
func NewSBOMProvider() *SBOMProvider {
	return &SBOMProvider{}
}

// Name returns the provider name.
func (p *SBOMProvider) Name() string {
	return "sbom"
}

// Connect establishes a connection to scan SBOM files.
func (p *SBOMProvider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Get SBOM file path from config or environment
	sbomPath := config["sbom_path"]
	if sbomPath == "" {
		sbomPath = os.Getenv("SBOM_PATH")
	}

	if sbomPath == "" {
		return nil, fmt.Errorf("sbom: no SBOM file path provided. Set sbom_path or SBOM_PATH")
	}

	// Check if path exists
	info, err := os.Stat(sbomPath)
	if err != nil {
		return nil, fmt.Errorf("sbom: failed to access path %s: %w", sbomPath, err)
	}

	return &SBOMConnection{
		path:  sbomPath,
		isDir: info.IsDir(),
	}, nil
}

// SBOMConnection represents an active connection for SBOM scanning.
type SBOMConnection struct {
	path  string
	isDir bool
}

// Resources returns all available SBOM resources.
func (c *SBOMConnection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&SBOMDocumentResource{path: c.path, isDir: c.isDir},
		&SBOMComponentResource{path: c.path, isDir: c.isDir},
		&SBOMVulnerabilityResource{path: c.path, isDir: c.isDir},
		&SBOMDependencyResource{path: c.path, isDir: c.isDir},
	}
}

// SBOMFormat represents the detected SBOM format.
type SBOMFormat string

const (
	FormatCycloneDX SBOMFormat = "cyclonedx"
	FormatSPDX      SBOMFormat = "spdx"
	FormatUnknown   SBOMFormat = "unknown"
)

// detectFormat detects the SBOM format from file content.
func detectFormat(data []byte) SBOMFormat {
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

// parseSBOMFiles parses SBOM files from the given path.
func parseSBOMFiles(path string, isDir bool) ([]map[string]interface{}, error) {
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
				sbom, err := parseSBOMFile(filePath)
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
		sbom, err := parseSBOMFile(path)
		if err != nil {
			return nil, err
		}
		if sbom != nil {
			results = append(results, sbom)
		}
	}

	return results, nil
}

// parseSBOMFile parses a single SBOM file.
func parseSBOMFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	format := detectFormat(data)
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
