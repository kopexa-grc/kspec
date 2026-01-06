package sbom

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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

	resources := conn.Resources()
	if len(resources) != 4 {
		t.Errorf("Resources() returned %d resources, want 4", len(resources))
	}

	expectedNames := map[string]bool{
		"sbom_document":      false,
		"sbom_component":     false,
		"sbom_vulnerability": false,
		"sbom_dependency":    false,
	}

	for _, r := range resources {
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
		want Format
	}{
		{
			name: "CycloneDX JSON",
			data: `{"bomFormat": "CycloneDX", "specVersion": "1.4"}`,
			want: FormatCycloneDX,
		},
		{
			name: "CycloneDX XML-style",
			data: `bomFormat=CycloneDX`,
			want: FormatCycloneDX,
		},
		{
			name: "SPDX JSON",
			data: `{"spdxVersion": "SPDX-2.3"}`,
			want: FormatSPDX,
		},
		{
			name: "SPDX tag-value",
			data: `SPDXVersion: SPDX-2.3`,
			want: FormatSPDX,
		},
		{
			name: "unknown format",
			data: `{"name": "some-package"}`,
			want: FormatUnknown,
		},
		{
			name: "empty content",
			data: ``,
			want: FormatUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectFormat([]byte(tt.data)); got != tt.want {
				t.Errorf("detectFormat() = %v, want %v", got, tt.want)
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
			result, err := parseSBOMFile(tt.file)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSBOMFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("parseSBOMFile() returned nil")
					return
				}

				format, ok := result["_format"].(string)
				if !ok || format != tt.wantFormat {
					t.Errorf("parseSBOMFile() format = %v, want %v", format, tt.wantFormat)
				}

				filePath, ok := result["_file_path"].(string)
				if !ok || filePath != tt.file {
					t.Errorf("parseSBOMFile() file_path = %v, want %v", filePath, tt.file)
				}
			}
		})
	}
}

func TestParseSBOMFiles_Directory(t *testing.T) {
	results, err := parseSBOMFiles("testdata", true)
	if err != nil {
		t.Fatalf("parseSBOMFiles() error = %v", err)
	}

	if len(results) < 2 {
		t.Errorf("parseSBOMFiles() found %d files, want at least 2", len(results))
	}

	hasFormat := map[string]bool{}
	for _, r := range results {
		if format, ok := r["_format"].(string); ok {
			hasFormat[format] = true
		}
	}

	if !hasFormat["cyclonedx"] {
		t.Error("parseSBOMFiles() did not find CycloneDX file")
	}
	if !hasFormat["spdx"] {
		t.Error("parseSBOMFiles() did not find SPDX file")
	}
}

func TestParseSBOMFile_NonExistent(t *testing.T) {
	_, err := parseSBOMFile("/nonexistent/file.json")
	if err == nil {
		t.Error("parseSBOMFile() expected error for nonexistent file")
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

	_, err := parseSBOMFile(tmpFile)
	if err == nil {
		t.Error("parseSBOMFile() expected error for invalid JSON")
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

	result, err := parseSBOMFile(tmpFile)
	if err != nil {
		t.Errorf("parseSBOMFile() unexpected error: %v", err)
	}
	if result != nil {
		t.Error("parseSBOMFile() expected nil result for unknown format")
	}
}
