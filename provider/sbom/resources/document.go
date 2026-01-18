// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// Document fetches SBOM document metadata.
type Document struct {
	path  string
	isDir bool
}

// NewDocument creates a new Document resource.
func NewDocument(path string, isDir bool) *Document {
	return &Document{path: path, isDir: isDir}
}

// Name returns the resource type name.
func (r *Document) Name() string {
	return "sbom_document"
}

// Fetch retrieves SBOM document information.
func (r *Document) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	sboms, err := ParseSBOMFiles(r.path, r.isDir)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(sboms))

	for _, sbom := range sboms {
		resource := make(core.Resource)

		format, ok := sbom["_format"].(string)
		if !ok {
			continue
		}
		resource["format"] = format
		resource["file_path"] = sbom["_file_path"]

		if format == string(FormatCycloneDX) {
			r.parseCycloneDXDocument(sbom, resource)
		} else if format == string(FormatSPDX) {
			r.parseSPDXDocument(sbom, resource)
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (r *Document) parseCycloneDXDocument(sbom map[string]interface{}, resource core.Resource) {
	resource["bom_format"] = sbom["bomFormat"]
	resource["spec_version"] = sbom["specVersion"]
	resource["serial_number"] = sbom["serialNumber"]
	resource["version"] = sbom["version"]

	// Metadata
	if metadata, ok := sbom["metadata"].(map[string]interface{}); ok {
		resource["timestamp"] = metadata["timestamp"]

		if tools, ok := metadata["tools"].([]interface{}); ok {
			resource["tool_count"] = len(tools)
		}

		if component, ok := metadata["component"].(map[string]interface{}); ok {
			resource["subject_name"] = component["name"]
			resource["subject_version"] = component["version"]
			resource["subject_type"] = component["type"]
		}
	}

	// Component count
	if components, ok := sbom["components"].([]interface{}); ok {
		resource["component_count"] = len(components)
	}

	// Vulnerability count
	if vulnerabilities, ok := sbom["vulnerabilities"].([]interface{}); ok {
		resource["vulnerability_count"] = len(vulnerabilities)
		resource["has_vulnerabilities"] = len(vulnerabilities) > 0
	} else {
		resource["vulnerability_count"] = 0
		resource["has_vulnerabilities"] = false
	}

	// Security flags
	resource["is_cyclonedx"] = true
	resource["is_spdx"] = false
	resource["has_metadata"] = sbom["metadata"] != nil
	componentCount := 0
	if cc, ok := resource["component_count"].(int); ok {
		componentCount = cc
	}
	resource["has_components"] = componentCount > 0
}

func (r *Document) parseSPDXDocument(sbom map[string]interface{}, resource core.Resource) {
	resource["spdx_version"] = sbom["spdxVersion"]
	resource["spdx_id"] = sbom["SPDXID"]
	resource["document_name"] = sbom["name"]
	resource["document_namespace"] = sbom["documentNamespace"]

	// Creation info
	if creationInfo, ok := sbom["creationInfo"].(map[string]interface{}); ok {
		resource["created"] = creationInfo["created"]
		if creators, ok := creationInfo["creators"].([]interface{}); ok {
			resource["creator_count"] = len(creators)
		}
	}

	// Package count
	if packages, ok := sbom["packages"].([]interface{}); ok {
		resource["package_count"] = len(packages)
		resource["component_count"] = len(packages)
	}

	// Relationship count
	if relationships, ok := sbom["relationships"].([]interface{}); ok {
		resource["relationship_count"] = len(relationships)
	}

	// Check for hasExtractedLicensingInfos
	if licensingInfo, ok := sbom["hasExtractedLicensingInfos"].([]interface{}); ok {
		resource["extracted_licensing_count"] = len(licensingInfo)
	}

	// Security flags
	resource["is_cyclonedx"] = false
	resource["is_spdx"] = true
	packageCount := 0
	if pc, ok := resource["package_count"].(int); ok {
		packageCount = pc
	}
	resource["has_packages"] = packageCount > 0

	// Extract document name from path if not available
	if resource["document_name"] == nil || resource["document_name"] == "" {
		if filePath, ok := sbom["_file_path"].(string); ok {
			resource["document_name"] = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		}
	}
}
