// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"net/url"
	"strings"

	"github.com/kopexa-grc/kspec/core"
)

// Host is the root resource for network scanning that aggregates
// all sub-resources (certificate, dns, http, tls) for a target host.
type Host struct{}

// NewHost creates a new Host resource instance.
func NewHost() *Host {
	return &Host{}
}

// Name returns the resource type name.
func (r *Host) Name() string {
	return "host"
}

// ChildResources implements core.ChildResourceProvider.
// Returns the child resource types that are scanned for a host.
func (r *Host) ChildResources() []string {
	return []string{
		"certificate",
		"dns",
		"http",
		"tls",
	}
}

// Fetch retrieves the host resource (essentially returns metadata about the target).
func (r *Host) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	target := asset.Config["target"]
	if target == "" {
		target = asset.Name
	}

	// Normalize target to get domain
	domain := normalizeDomain(target)

	resource := core.Resource{
		"target": target,
		"domain": domain,
		"name":   domain,
	}

	return []core.Resource{resource}, nil
}

// normalizeDomain extracts and normalizes the domain from a target string.
func normalizeDomain(target string) string {
	// Handle URLs
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		if u, err := url.Parse(target); err == nil {
			return u.Hostname()
		}
	}

	// Handle host:port
	if idx := strings.LastIndex(target, ":"); idx != -1 {
		// Check if it's not an IPv6 address
		if !strings.Contains(target, "[") {
			return target[:idx]
		}
	}

	return target
}
