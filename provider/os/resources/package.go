// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// Package provides access to installed package information via Homebrew.
type Package struct{}

// NewPackage creates a new Package resource.
func NewPackage() *Package {
	return &Package{}
}

// Name returns the resource identifier for packages.
func (r *Package) Name() string {
	return "package"
}

// Fetch retrieves installed packages from Homebrew, optionally filtered by name.
func (r *Package) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	targetName := config["name"]

	args := []string{"list", "--versions"}
	if targetName != "" {
		args = append(args, targetName)
	}

	cmd := exec.CommandContext(ctx, "brew", args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && targetName != "" {
			return []core.Resource{}, nil
		}
		return nil, fmt.Errorf("failed to run brew list: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var resources []core.Resource

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			version := parts[len(parts)-1]

			resources = append(resources, core.Resource{
				"name":      name,
				"version":   version,
				"manager":   "homebrew",
				"installed": true,
			})
		}
	}

	return resources, nil
}
