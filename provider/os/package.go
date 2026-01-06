package os

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// PackageResource provides access to installed package information via Homebrew.
type PackageResource struct{}

// Name returns the resource identifier for packages.
func (r *PackageResource) Name() string {
	return "package"
}

// Fetch retrieves installed packages from Homebrew, optionally filtered by name.
func (r *PackageResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Use config from asset
	config := asset.Config
	// If 'name' is specified, we check for that specific package.
	// If not, we list all installed packages (discovery mode).
	// Note: 'name' might be passed as a config, OR we might be doing a query like 'resource.name == "foo"'.
	// If the user does cnquery --resource package, we should probably return ALL packages so they can filter in CEL.
	// But if they provide config 'name=foo', we optimize.

	targetName := config["name"]

	// Construct brew command
	// brew list --versions [name]
	args := []string{"list", "--versions"}
	if targetName != "" {
		args = append(args, targetName)
	}

	cmd := exec.CommandContext(ctx, "brew", args...)
	out, err := cmd.Output()
	if err != nil {
		// If specific package not found, brew returns exit status 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && targetName != "" {
			// Return empty list if specific package requested but not found?
			// Or return resource with installed=false?
			// Unlike File resource, package managers usually just Error or Empty.
			// Let's return empty list for "not found".
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
		// Output format: "package_name version1 version2 ..."
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			// There might be multiple versions installed, we take the last one (usually latest) or list all?
			// Let's take the last one as "version"
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
