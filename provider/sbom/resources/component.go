// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package resources provides SBOM resource implementations for security policy evaluation.
// It supports both CycloneDX and SPDX SBOM formats.
package resources

import (
	"context"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// SPDX constants
const (
	SPDXNoAssertion = "NOASSERTION"
)

// Component fetches SBOM components/packages.
type Component struct {
	path  string
	isDir bool
}

// NewComponent creates a new Component resource.
func NewComponent(path string, isDir bool) *Component {
	return &Component{path: path, isDir: isDir}
}

// Name returns the resource type name.
func (r *Component) Name() string {
	return "sbom_component"
}

// Fetch retrieves SBOM components.
func (r *Component) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	sboms, err := ParseSBOMFiles(r.path, r.isDir)
	if err != nil {
		return nil, err
	}

	var resources []core.Resource

	for _, sbom := range sboms {
		format, ok := sbom["_format"].(string)
		if !ok {
			continue
		}
		filePath, ok := sbom["_file_path"].(string)
		if !ok {
			continue
		}

		if format == string(FormatCycloneDX) {
			components := r.parseCycloneDXComponents(sbom, filePath)
			resources = append(resources, components...)
		} else if format == string(FormatSPDX) {
			components := r.parseSPDXPackages(sbom, filePath)
			resources = append(resources, components...)
		}
	}

	return resources, nil
}

func (r *Component) parseCycloneDXComponents(sbom map[string]interface{}, filePath string) []core.Resource {
	components, ok := sbom["components"].([]interface{})
	if !ok {
		return nil
	}

	resources := make([]core.Resource, 0, len(components))
	for _, comp := range components {
		component, ok := comp.(map[string]interface{})
		if !ok {
			continue
		}

		resource := make(core.Resource)
		resource["source_file"] = filePath
		resource["format"] = string(FormatCycloneDX)

		resource["name"] = component["name"]
		resource["version"] = component["version"]
		resource["type"] = component["type"]
		resource["purl"] = component["purl"]
		resource["bom_ref"] = component["bom-ref"]
		resource["description"] = component["description"]
		resource["group"] = component["group"]
		resource["publisher"] = component["publisher"]

		// License information
		if licenses, ok := component["licenses"].([]interface{}); ok {
			var licenseNames []string
			for _, lic := range licenses {
				if license, ok := lic.(map[string]interface{}); ok {
					if licData, ok := license["license"].(map[string]interface{}); ok {
						if id, ok := licData["id"].(string); ok {
							licenseNames = append(licenseNames, id)
						} else if name, ok := licData["name"].(string); ok {
							licenseNames = append(licenseNames, name)
						}
					}
				}
			}
			resource["licenses"] = licenseNames
			resource["license_count"] = len(licenseNames)
		}

		// Hash information
		if hashes, ok := component["hashes"].([]interface{}); ok {
			resource["hash_count"] = len(hashes)
			for _, h := range hashes {
				if hash, ok := h.(map[string]interface{}); ok {
					if alg, ok := hash["alg"].(string); ok {
						resource["hash_"+strings.ToLower(alg)] = hash["content"]
					}
				}
			}
		}

		// External references
		if extRefs, ok := component["externalReferences"].([]interface{}); ok {
			resource["external_reference_count"] = len(extRefs)
		}

		// Security flags
		resource["has_purl"] = component["purl"] != nil && component["purl"] != ""
		resource["has_version"] = component["version"] != nil && component["version"] != ""
		licenseCount := 0
		if lc, ok := resource["license_count"].(int); ok {
			licenseCount = lc
		}
		hashCount := 0
		if hc, ok := resource["hash_count"].(int); ok {
			hashCount = hc
		}
		resource["has_license"] = licenseCount > 0
		resource["has_hashes"] = hashCount > 0
		resource["is_library"] = component["type"] == "library"
		resource["is_application"] = component["type"] == "application"
		resource["is_framework"] = component["type"] == "framework"

		resources = append(resources, resource)
	}

	return resources
}

func (r *Component) parseSPDXPackages(sbom map[string]interface{}, filePath string) []core.Resource {
	packages, ok := sbom["packages"].([]interface{})
	if !ok {
		return nil
	}

	resources := make([]core.Resource, 0, len(packages))
	for _, pkg := range packages {
		pack, ok := pkg.(map[string]interface{})
		if !ok {
			continue
		}

		resource := make(core.Resource)
		resource["source_file"] = filePath
		resource["format"] = string(FormatSPDX)

		resource["spdx_id"] = pack["SPDXID"]
		resource["name"] = pack["name"]
		resource["version"] = pack["versionInfo"]
		resource["download_location"] = pack["downloadLocation"]
		resource["files_analyzed"] = pack["filesAnalyzed"]
		resource["supplier"] = pack["supplier"]
		resource["originator"] = pack["originator"]
		resource["homepage"] = pack["homepage"]

		// License information
		resource["license_concluded"] = pack["licenseConcluded"]
		resource["license_declared"] = pack["licenseDeclared"]
		resource["copyright_text"] = pack["copyrightText"]

		// External references
		if extRefs, ok := pack["externalRefs"].([]interface{}); ok {
			resource["external_reference_count"] = len(extRefs)

			// Look for PURL
			for _, ref := range extRefs {
				if extRef, ok := ref.(map[string]interface{}); ok {
					if extRef["referenceType"] == "purl" {
						resource["purl"] = extRef["referenceLocator"]
					}
				}
			}
		}

		// Checksums
		if checksums, ok := pack["checksums"].([]interface{}); ok {
			resource["checksum_count"] = len(checksums)
			for _, cs := range checksums {
				if checksum, ok := cs.(map[string]interface{}); ok {
					if alg, ok := checksum["algorithm"].(string); ok {
						resource["checksum_"+strings.ToLower(alg)] = checksum["checksumValue"]
					}
				}
			}
		}

		// Security flags
		resource["has_version"] = pack["versionInfo"] != nil && pack["versionInfo"] != ""
		resource["has_license_concluded"] = pack["licenseConcluded"] != nil && pack["licenseConcluded"] != "" && pack["licenseConcluded"] != SPDXNoAssertion
		resource["has_license_declared"] = pack["licenseDeclared"] != nil && pack["licenseDeclared"] != "" && pack["licenseDeclared"] != SPDXNoAssertion
		resource["has_purl"] = resource["purl"] != nil && resource["purl"] != ""
		checksumCount := 0
		if cc, ok := resource["checksum_count"].(int); ok {
			checksumCount = cc
		}
		resource["has_checksums"] = checksumCount > 0
		resource["has_supplier"] = pack["supplier"] != nil && pack["supplier"] != "" && pack["supplier"] != SPDXNoAssertion
		resource["is_noassertion_license"] = pack["licenseConcluded"] == SPDXNoAssertion || pack["licenseDeclared"] == SPDXNoAssertion

		resources = append(resources, resource)
	}

	return resources
}
