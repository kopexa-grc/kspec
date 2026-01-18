package sbom

import (
	"context"
	"testing"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/sbom/resources"
)

func TestComponentResource_Name(t *testing.T) {
	r := resources.NewComponent("", false)
	if got := r.Name(); got != "sbom_component" {
		t.Errorf("Name() = %v, want %v", got, "sbom_component")
	}
}

func TestComponentResource_FetchCycloneDX(t *testing.T) {
	r := resources.NewComponent("testdata/cyclonedx.json", false)

	results, err := r.Fetch(context.Background(), core.Asset{})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(results) != 4 {
		t.Errorf("Fetch() returned %d components, want 4", len(results))
	}

	// Find lodash component
	var lodash core.Resource
	for _, res := range results {
		if res["name"] == "lodash" {
			lodash = res
			break
		}
	}

	if lodash == nil {
		t.Fatal("lodash component not found")
	}

	// Verify lodash fields
	tests := []struct {
		field string
		want  interface{}
	}{
		{"format", "cyclonedx"},
		{"name", "lodash"},
		{"version", "4.17.21"},
		{"type", "library"},
		{"purl", "pkg:npm/lodash@4.17.21"},
		{"bom_ref", "pkg:npm/lodash@4.17.21"},
		{"group", "npm"},
		{"publisher", "John-David Dalton"},
		{"has_purl", true},
		{"has_version", true},
		{"has_license", true},
		{"has_hashes", true},
		{"is_library", true},
		{"is_application", false},
		{"is_framework", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if got := lodash[tt.field]; got != tt.want {
				t.Errorf("lodash[%s] = %v (%T), want %v (%T)", tt.field, got, got, tt.want, tt.want)
			}
		})
	}

	// Check license parsing
	licenses, ok := lodash["licenses"].([]string)
	if !ok || len(licenses) == 0 {
		t.Error("lodash licenses not parsed correctly")
	} else if licenses[0] != "MIT" {
		t.Errorf("lodash license = %v, want MIT", licenses[0])
	}
}

func TestComponentResource_FetchSPDX(t *testing.T) {
	r := resources.NewComponent("testdata/spdx.json", false)

	results, err := r.Fetch(context.Background(), core.Asset{})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Fetch() returned %d packages, want 3", len(results))
	}

	// Find lodash package
	var lodash core.Resource
	for _, res := range results {
		if res["name"] == "lodash" {
			lodash = res
			break
		}
	}

	if lodash == nil {
		t.Fatal("lodash package not found")
	}

	// Verify lodash fields
	tests := []struct {
		field string
		want  interface{}
	}{
		{"format", "spdx"},
		{"name", "lodash"},
		{"version", "4.17.21"},
		{"spdx_id", "SPDXRef-Package-lodash"},
		{"license_concluded", "MIT"},
		{"license_declared", "MIT"},
		{"purl", "pkg:npm/lodash@4.17.21"},
		{"has_version", true},
		{"has_license_concluded", true},
		{"has_license_declared", true},
		{"has_purl", true},
		{"has_checksums", true},
		{"has_supplier", true},
		{"is_noassertion_license", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			if got := lodash[tt.field]; got != tt.want {
				t.Errorf("lodash[%s] = %v (%T), want %v (%T)", tt.field, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestComponentResource_SPDXNoAssertion(t *testing.T) {
	r := resources.NewComponent("testdata/spdx.json", false)

	results, err := r.Fetch(context.Background(), core.Asset{})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Find unknown-pkg package
	var unknown core.Resource
	for _, res := range results {
		if res["name"] == "unknown-pkg" {
			unknown = res
			break
		}
	}

	if unknown == nil {
		t.Fatal("unknown-pkg package not found")
	}

	// Should detect NOASSERTION licenses
	if unknown["is_noassertion_license"] != true {
		t.Error("is_noassertion_license should be true for unknown-pkg")
	}

	if unknown["has_license_concluded"] != false {
		t.Error("has_license_concluded should be false for NOASSERTION")
	}
}

func TestComponentResource_FetchDirectory(t *testing.T) {
	r := resources.NewComponent("testdata", true)

	results, err := r.Fetch(context.Background(), core.Asset{})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Should have components from both files
	if len(results) < 5 {
		t.Errorf("Fetch() returned %d components, want at least 5 (from both files)", len(results))
	}

	// Check we have both formats
	formats := map[string]bool{}
	for _, res := range results {
		if format, ok := res["format"].(string); ok {
			formats[format] = true
		}
	}

	if !formats["cyclonedx"] {
		t.Error("No CycloneDX components found")
	}
	if !formats["spdx"] {
		t.Error("No SPDX components found")
	}
}

func TestComponentResource_ComponentTypes(t *testing.T) {
	r := resources.NewComponent("testdata/cyclonedx.json", false)

	results, err := r.Fetch(context.Background(), core.Asset{})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	types := map[string]int{}
	for _, resource := range results {
		if resource["is_library"] == true {
			types["library"]++
		}
		if resource["is_framework"] == true {
			types["framework"]++
		}
		if resource["is_application"] == true {
			types["application"]++
		}
	}

	// We have 2 libraries, 1 framework, 0 applications in test data
	if types["library"] < 2 {
		t.Errorf("Expected at least 2 libraries, got %d", types["library"])
	}
	if types["framework"] != 1 {
		t.Errorf("Expected 1 framework, got %d", types["framework"])
	}
}

func TestComponentResource_ComponentWithoutVersion(t *testing.T) {
	r := resources.NewComponent("testdata/cyclonedx.json", false)

	results, err := r.Fetch(context.Background(), core.Asset{})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	// Find no-version-lib
	var noVersion core.Resource
	for _, res := range results {
		if res["name"] == "no-version-lib" {
			noVersion = res
			break
		}
	}

	if noVersion == nil {
		t.Fatal("no-version-lib component not found")
	}

	if noVersion["has_version"] != false {
		t.Error("has_version should be false for component without version")
	}

	if noVersion["has_purl"] != false {
		t.Error("has_purl should be false for component without purl")
	}
}
