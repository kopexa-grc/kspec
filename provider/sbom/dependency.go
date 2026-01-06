package sbom

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

// DependencyResource fetches SBOM dependency relationships.
type DependencyResource struct {
	path  string
	isDir bool
}

// Name returns the resource type name.
func (r *DependencyResource) Name() string {
	return "sbom_dependency"
}

// Fetch retrieves SBOM dependencies.
func (r *DependencyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	sboms, err := parseSBOMFiles(r.path, r.isDir)
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
			deps := r.parseCycloneDXDependencies(sbom, filePath)
			resources = append(resources, deps...)
		} else if format == string(FormatSPDX) {
			deps := r.parseSPDXRelationships(sbom, filePath)
			resources = append(resources, deps...)
		}
	}

	return resources, nil
}

func (r *DependencyResource) parseCycloneDXDependencies(sbom map[string]interface{}, filePath string) []core.Resource {
	var resources []core.Resource

	dependencies, ok := sbom["dependencies"].([]interface{})
	if !ok {
		return resources
	}

	for _, dep := range dependencies {
		dependency, ok := dep.(map[string]interface{})
		if !ok {
			continue
		}

		resource := make(core.Resource)
		resource["source_file"] = filePath
		resource["format"] = "cyclonedx"

		resource["ref"] = dependency["ref"]

		// Dependencies on this component
		if dependsOn, ok := dependency["dependsOn"].([]interface{}); ok {
			var deps []string
			for _, d := range dependsOn {
				if depRef, ok := d.(string); ok {
					deps = append(deps, depRef)
				}
			}
			resource["depends_on"] = deps
			resource["dependency_count"] = len(deps)
		} else {
			resource["dependency_count"] = 0
		}

		// Security flags
		depCount, _ := resource["dependency_count"].(int)
		resource["has_dependencies"] = depCount > 0
		resource["is_leaf"] = depCount == 0

		resources = append(resources, resource)
	}

	return resources
}

func (r *DependencyResource) parseSPDXRelationships(sbom map[string]interface{}, filePath string) []core.Resource {
	var resources []core.Resource

	relationships, ok := sbom["relationships"].([]interface{})
	if !ok {
		return resources
	}

	for _, rel := range relationships {
		relationship, ok := rel.(map[string]interface{})
		if !ok {
			continue
		}

		resource := make(core.Resource)
		resource["source_file"] = filePath
		resource["format"] = "spdx"

		resource["element_id"] = relationship["spdxElementId"]
		resource["related_element_id"] = relationship["relatedSpdxElement"]
		resource["relationship_type"] = relationship["relationshipType"]
		resource["comment"] = relationship["comment"]

		// Security flags
		relType, _ := relationship["relationshipType"].(string)
		resource["is_depends_on"] = relType == "DEPENDS_ON"
		resource["is_dependency_of"] = relType == "DEPENDENCY_OF"
		resource["is_contains"] = relType == "CONTAINS"
		resource["is_contained_by"] = relType == "CONTAINED_BY"
		resource["is_generates"] = relType == "GENERATES"
		resource["is_generated_from"] = relType == "GENERATED_FROM"
		resource["is_build_tool_of"] = relType == "BUILD_TOOL_OF"
		resource["is_dev_tool_of"] = relType == "DEV_TOOL_OF"
		resource["is_test_tool_of"] = relType == "TEST_TOOL_OF"
		resource["is_documentation_of"] = relType == "DOCUMENTATION_OF"
		resource["is_optional_dependency"] = relType == "OPTIONAL_DEPENDENCY_OF"
		resource["is_runtime_dependency"] = relType == "RUNTIME_DEPENDENCY_OF"

		resources = append(resources, resource)
	}

	return resources
}
