// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package sbom provides SBOM (Software Bill of Materials) scanning capabilities
// for security policy evaluation. It supports CycloneDX and SPDX formats.
package sbom

import (
	"context"
	"fmt"
	"os"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/sbom/resources"
)

// Provider implements the core.Provider interface for SBOM (Software Bill of Materials) scanning.
type Provider struct{}

// NewProvider creates a new SBOM provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "sbom"
}

// Connect establishes a connection to scan SBOM files.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
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

	return &Connection{
		path:  sbomPath,
		isDir: info.IsDir(),
	}, nil
}

// Connection represents an active connection for SBOM scanning.
type Connection struct {
	path  string
	isDir bool
}

// Resources returns all available SBOM resources.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		resources.NewDocument(c.path, c.isDir),
		resources.NewComponent(c.path, c.isDir),
		resources.NewVulnerability(c.path, c.isDir),
		resources.NewDependency(c.path, c.isDir),
	}
}

// EntryResourceType returns the entry point resource type for a given asset type.
func (c *Connection) EntryResourceType(assetType string) string {
	switch assetType {
	case "sbom-file", "sbom-directory":
		return "sbom_document"
	default:
		return ""
	}
}
