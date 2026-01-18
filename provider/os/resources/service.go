// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// Service provides access to launchctl service information on macOS.
type Service struct{}

// NewService creates a new Service resource.
func NewService() *Service {
	return &Service{}
}

// Name returns the resource identifier for services.
func (r *Service) Name() string {
	return "service"
}

// Fetch retrieves launchctl services, optionally filtered by label.
func (r *Service) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	labelFilter := config["label"]

	cmd := exec.CommandContext(ctx, "launchctl", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run launchctl list: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var resources []core.Resource

	startIndex := 0
	if len(lines) > 0 && strings.Contains(lines[0], "Label") {
		startIndex = 1
	}

	for i := startIndex; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			pid := parts[0]
			status := parts[1]
			label := parts[2]

			if labelFilter != "" && label != labelFilter {
				continue
			}

			running := pid != "-"

			res := core.Resource{
				"label":     label,
				"name":      label,
				"pid":       pid,
				"exit_code": status,
				"running":   running,
				"type":      "launchctl",
			}
			resources = append(resources, res)
		}
	}

	return resources, nil
}
