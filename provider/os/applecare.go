package os

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

type AppleCareResource struct{}

func (r *AppleCareResource) Name() string {
	return "apple_care"
}

func (r *AppleCareResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// 1. Get Serial Number
	serial, err := getSerialNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get serial number: %w", err)
	}

	// 2. Locate JSON file
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// Path: ~/Library/Application Support/com.apple.NewDeviceOutreach/caches/coverageDetails/<SERIAL>.json
	jsonPath := filepath.Join(usr.HomeDir, "Library/Application Support/com.apple.NewDeviceOutreach/caches/coverageDetails", serial+".json")

	// 3. Read and Parse
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		// If file doesn't exist, we might return empty resource or error.
		// Returning error implies we can't verify coverage.
		return nil, fmt.Errorf("coverage file not found for serial %s: %w", serial, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse coverage json: %w", err)
	}

	// 4. Flatten interesting fields
	// logic: "expiration" is often "0" if expired. If active, might be a timestamp?
	// We'll expose raw fields + parsed expiration.

	res := core.Resource{
		"serialNumber":           raw["serialNumber"],
		"coverageLabel":          raw["coverageLabel"],
		"hasTheftAndLossBenefit": raw["hasTheftAndLossBenefit"],
		"expiration":             "0", // Default
	}

	// Nested fields
	if settings, ok := raw["settingsCoverageSection"].(map[string]interface{}); ok {
		if offer, ok := settings["offer"].(map[string]interface{}); ok {
			if exp, ok := offer["expiration"].(string); ok {
				res["expiration"] = exp
			}
		}
	}

	// Try parsing expiration to int
	if expStr, ok := res["expiration"].(string); ok {
		if expInt, err := strconv.ParseInt(expStr, 10, 64); err == nil {
			res["expirationInt"] = expInt
		} else {
			res["expirationInt"] = 0
		}
	}

	// Add params for detailed inspection
	res["params"] = raw

	return []core.Resource{res}, nil
}

func getSerialNumber(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "system_profiler", "SPHardwareDataType").Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Serial Number (system):") {
			parts := strings.Split(trimmed, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}
	return "", fmt.Errorf("serial number not found in system_profiler output")
}
