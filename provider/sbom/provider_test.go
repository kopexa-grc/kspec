// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package sbom

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kopexa-grc/kspec/provider/sbom/resources"
)

func TestProvider_Name(t *testing.T) {
	p := NewProvider()
	if got := p.Name(); got != "sbom" {
		t.Errorf("Name() = %v, want %v", got, "sbom")
	}
}

func TestProvider_Connect(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]string
		envVar  string
		wantErr bool
	}{
		{
			name:    "no path provided",
			config:  map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid path",
			config:  map[string]string{"sbom_path": "/nonexistent/path"},
			wantErr: true,
		},
		{
			name:    "valid file path",
			config:  map[string]string{"sbom_path": "testdata/cyclonedx.json"},
			wantErr: false,
		},
		{
			name:    "valid directory path",
			config:  map[string]string{"sbom_path": "testdata"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				os.Setenv("SBOM_PATH", tt.envVar)
				defer os.Unsetenv("SBOM_PATH")
			}

			p := NewProvider()
			conn, err := p.Connect(context.Background(), tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && conn == nil {
				t.Error("Connect() returned nil connection")
			}
		})
	}
}

func TestProvider_ConnectWithEnv(t *testing.T) {
	os.Setenv("SBOM_PATH", "testdata/cyclonedx.json")
	defer os.Unsetenv("SBOM_PATH")

	p := NewProvider()
	conn, err := p.Connect(context.Background(), map[string]string{})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if conn == nil {
		t.Error("Connect() returned nil connection")
	}
}

func TestConnection_Resources(t *testing.T) {
	p := NewProvider()
	conn, err := p.Connect(context.Background(), map[string]string{
		"sbom_path": "testdata/cyclonedx.json",
	})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	resourceSpecs := conn.Resources()
	if len(resourceSpecs) != 4 {
		t.Errorf("Resources() returned %d resources, want 4", len(resourceSpecs))
	}

	expectedNames := map[string]bool{
		"sbom_document":      false,
		"sbom_component":     false,
		"sbom_vulnerability": false,
		"sbom_dependency":    false,
	}

	for _, r := range resourceSpecs {
		expectedNames[r.Name()] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Missing resource: %s", name)
		}
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name string
		data string
		want resources.Format
	}{
		{
			name: "CycloneDX JSON",
			data: `{"bomFormat": "CycloneDX", "specVersion": "1.4"}`,
			want: resources.FormatCycloneDX,
		},
		{
			name: "CycloneDX XML-style",
			data: `bomFormat=CycloneDX`,
			want: resources.FormatCycloneDX,
		},
		{
			name: "SPDX JSON",
			data: `{"spdxVersion": "SPDX-2.3"}`,
			want: resources.FormatSPDX,
		},
		{
			name: "SPDX tag-value",
			data: `SPDXVersion: SPDX-2.3`,
			want: resources.FormatSPDX,
		},
		{
			name: "unknown format",
			data: `{"name": "some-package"}`,
			want: resources.FormatUnknown,
		},
		{
			name: "empty content",
			data: ``,
			want: resources.FormatUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resources.DetectFormat([]byte(tt.data)); got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSBOMFile(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		wantFormat string
		wantErr    bool
	}{
		{
			name:       "CycloneDX file",
			file:       "testdata/cyclonedx.json",
			wantFormat: "cyclonedx",
			wantErr:    false,
		},
		{
			name:       "SPDX file",
			file:       "testdata/spdx.json",
			wantFormat: "spdx",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resources.ParseSBOMFile(tt.file)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSBOMFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("ParseSBOMFile() returned nil")
					return
				}

				format, ok := result["_format"].(string)
				if !ok || format != tt.wantFormat {
					t.Errorf("ParseSBOMFile() format = %v, want %v", format, tt.wantFormat)
				}

				filePath, ok := result["_file_path"].(string)
				if !ok || filePath != tt.file {
					t.Errorf("ParseSBOMFile() file_path = %v, want %v", filePath, tt.file)
				}
			}
		})
	}
}

func TestParseSBOMFiles_Directory(t *testing.T) {
	results, err := resources.ParseSBOMFiles("testdata", true)
	if err != nil {
		t.Fatalf("ParseSBOMFiles() error = %v", err)
	}

	if len(results) < 2 {
		t.Errorf("ParseSBOMFiles() found %d files, want at least 2", len(results))
	}

	hasFormat := map[string]bool{}
	for _, r := range results {
		if format, ok := r["_format"].(string); ok {
			hasFormat[format] = true
		}
	}

	if !hasFormat["cyclonedx"] {
		t.Error("ParseSBOMFiles() did not find CycloneDX file")
	}
	if !hasFormat["spdx"] {
		t.Error("ParseSBOMFiles() did not find SPDX file")
	}
}

func TestParseSBOMFile_NonExistent(t *testing.T) {
	_, err := resources.ParseSBOMFile("/nonexistent/file.json")
	if err == nil {
		t.Error("ParseSBOMFile() expected error for nonexistent file")
	}
}

func TestParseSBOMFile_InvalidJSON(t *testing.T) {
	// Create a temp file with invalid JSON but valid format marker
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")

	content := []byte(`{"bomFormat": "CycloneDX", invalid json`)
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err := resources.ParseSBOMFile(tmpFile)
	if err == nil {
		t.Error("ParseSBOMFile() expected error for invalid JSON")
	}
}

// FuzzDetectFormat tests format detection with random input.
func FuzzDetectFormat(f *testing.F) {
	// Seed corpus
	seeds := []string{
		`{"bomFormat": "CycloneDX", "specVersion": "1.4"}`,
		`{"spdxVersion": "SPDX-2.3"}`,
		`bomFormat=CycloneDX`,
		`SPDXVersion: SPDX-2.3`,
		`{"name": "some-package"}`,
		``,
		`{}`,
		`[]`,
		`null`,
		`"bomFormat"`,
		`"spdxVersion"`,
		string(make([]byte, 1000)),
		`{"bomFormat": null}`,
		`{"spdxVersion": 123}`,
		`<<<>>>`,
		`\x00\x01\x02`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data string) {
		// Should never panic
		result := resources.DetectFormat([]byte(data))

		// Result should always be one of the known formats
		if result != resources.FormatCycloneDX && result != resources.FormatSPDX && result != resources.FormatUnknown {
			t.Errorf("DetectFormat() returned unexpected format: %v", result)
		}
	})
}

// FuzzParseSBOMContent tests SBOM JSON parsing with random input.
func FuzzParseSBOMContent(f *testing.F) {
	// Seed corpus with various JSON structures
	seeds := []string{
		`{"bomFormat": "CycloneDX", "specVersion": "1.4", "components": []}`,
		`{"spdxVersion": "SPDX-2.3", "SPDXID": "SPDXRef-DOCUMENT", "packages": []}`,
		`{}`,
		`[]`,
		`null`,
		`{"key": "value"}`,
		`{"bomFormat": "CycloneDX", "components": [{"name": "test", "version": "1.0"}]}`,
		`{"deeply": {"nested": {"structure": {"here": true}}}}`,
		`{"array": [1, 2, 3, "four", null, true, {"obj": "val"}]}`,
		`{"special": "chars: \n\t\r"}`,
		`{"unicode": "日本語テスト"}`,
		`{"numbers": [0, -1, 1.5, 1e10, -1.5e-10]}`,
		`{"empty": "", "null": null, "bool": false}`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, content string) {
		// Create temp file with fuzzed content
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "fuzz-sbom.json")

		// Add a format marker so it's recognized
		var markedContent string
		if len(content) > 1 && content[0] == '{' {
			// Insert format marker into existing JSON object
			markedContent = `{"bomFormat": "CycloneDX", ` + content[1:]
		} else {
			// Wrap content in a valid JSON structure with format marker
			markedContent = `{"bomFormat": "CycloneDX", "data": "` + escapeJSONString(content) + `"}`
		}

		if err := os.WriteFile(tmpFile, []byte(markedContent), 0o644); err != nil {
			t.Skip("Failed to write temp file")
		}

		// Should never panic
		_, _ = resources.ParseSBOMFile(tmpFile)
	})
}

// escapeJSONString escapes a string for use in JSON.
func escapeJSONString(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			result = append(result, '\\', '\\')
		case '"':
			result = append(result, '\\', '"')
		case '\n':
			result = append(result, '\\', 'n')
		case '\r':
			result = append(result, '\\', 'r')
		case '\t':
			result = append(result, '\\', 't')
		default:
			if s[i] < 32 {
				// Control character - skip or escape
				result = append(result, ' ')
			} else {
				result = append(result, s[i])
			}
		}
	}
	return string(result)
}

// TestDetectFormat_EdgeCases tests edge cases in format detection.
func TestDetectFormat_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		data string
		want resources.Format
	}{
		{
			name: "bomFormat in middle of string",
			data: `some text "bomFormat" more text`,
			want: resources.FormatCycloneDX,
		},
		{
			name: "spdxVersion case sensitive",
			data: `{"SPDXVERSION": "2.3"}`,
			want: resources.FormatUnknown, // Case sensitive match
		},
		{
			name: "bomFormat case sensitive",
			data: `{"BOMFORMAT": "CycloneDX"}`,
			want: resources.FormatUnknown, // Case sensitive match
		},
		{
			name: "very long content with format at end",
			data: string(make([]byte, 10000)) + `"bomFormat"`,
			want: resources.FormatCycloneDX,
		},
		{
			name: "binary-like content",
			data: "\x00\x01\x02\x03",
			want: resources.FormatUnknown,
		},
		{
			name: "unicode content",
			data: `{"名前": "bomFormat"}`,
			want: resources.FormatCycloneDX,
		},
		{
			name: "escaped quotes around bomFormat",
			data: `{"key": "\"bomFormat\""}`,
			want: resources.FormatUnknown, // Escaped quotes means it's just a string value, not a key
		},
		{
			name: "bomFormat as actual key",
			data: `{"key": "value", "bomFormat": "CycloneDX"}`,
			want: resources.FormatCycloneDX,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resources.DetectFormat([]byte(tt.data))
			if got != tt.want {
				t.Errorf("DetectFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSBOMFile_UnknownFormat(t *testing.T) {
	// Create a temp file with unknown format
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "unknown.json")

	content := []byte(`{"name": "not-an-sbom"}`)
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	result, err := resources.ParseSBOMFile(tmpFile)
	if err != nil {
		t.Errorf("ParseSBOMFile() unexpected error: %v", err)
	}
	if result != nil {
		t.Error("ParseSBOMFile() expected nil result for unknown format")
	}
}
