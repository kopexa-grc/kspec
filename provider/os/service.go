package os

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/juliankoehn/kspec/core"
)

type ServiceResource struct{}

func (r *ServiceResource) Name() string {
	return "service"
}

func (r *ServiceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// launchctl list returns: PID Status Label
	// If name/label is specified, we could grep it, but listing all is fast enough usually.
	config := asset.Config

	labelFilter := config["label"] // or "name"

	cmd := exec.CommandContext(ctx, "launchctl", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run launchctl list: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var resources []core.Resource

	// First line is header usually? "PID\tStatus\tLabel"
	// Let's skip header if detected
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
			// PID Status Label
			pid := parts[0]
			status := parts[1]
			label := parts[2]

			if labelFilter != "" && label != labelFilter {
				continue
			}

			// Convert PID to integer? It might be "-" if not running but loaded.
			// Status 0 means OK/Running usually? No, Status is exit code of last run.
			// If PID is present (not "-"), it's running.

			running := pid != "-"

			res := core.Resource{
				"label":     label,
				"name":      label, // alias
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
