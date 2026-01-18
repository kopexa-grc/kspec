// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package resources contains OS resource implementations.
package resources

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

// AppleCare provides access to Apple Care coverage information for macOS devices.
type AppleCare struct{}

// NewAppleCare creates a new AppleCare resource.
func NewAppleCare() *AppleCare {
	return &AppleCare{}
}

// Name returns the resource identifier for Apple Care.
func (r *AppleCare) Name() string {
	return "apple_care"
}

// Fetch retrieves Apple Care coverage details from the local device cache.
func (r *AppleCare) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	serial, err := getSerialNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get serial number: %w", err)
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	jsonPath := filepath.Join(usr.HomeDir, "Library", "Application Support", "com.apple.NewDeviceOutreach", "caches", "coverageDetails", serial+".json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("coverage file not found for serial %s: %w", serial, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse coverage json: %w", err)
	}

	res := core.Resource{
		"serialNumber":           raw["serialNumber"],
		"coverageLabel":          raw["coverageLabel"],
		"hasTheftAndLossBenefit": raw["hasTheftAndLossBenefit"],
		"expiration":             "0",
	}

	if settings, ok := raw["settingsCoverageSection"].(map[string]interface{}); ok {
		if offer, ok := settings["offer"].(map[string]interface{}); ok {
			if exp, ok := offer["expiration"].(string); ok {
				res["expiration"] = exp
			}
		}
	}

	if expStr, ok := res["expiration"].(string); ok {
		if expInt, err := strconv.ParseInt(expStr, 10, 64); err == nil {
			res["expirationInt"] = expInt
		} else {
			res["expirationInt"] = 0
		}
	}

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
