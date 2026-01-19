// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package common

import "github.com/kopexa-grc/kspec/policy"

// AssetState represents the current state of an asset
type AssetState string

// AssetState constants define the possible states of an asset during scanning.
const (
	AssetStatePending   AssetState = "pending"
	AssetStateDiscovery AssetState = "discovering"
	AssetStateScanning  AssetState = "scanning"
	AssetStateComplete  AssetState = "complete"
	AssetStateError     AssetState = "error"
)

// Asset represents a scan target with its current state and results
type Asset struct {
	Type          string
	Name          string
	ID            string
	State         AssetState
	ResourceCount int
	ChecksPassed  int
	ChecksFailed  int
	ChecksSkipped int
	Error         error
}

// ViewMode represents the current UI view
type ViewMode string

// ViewMode constants define the possible view modes in the CLI.
const (
	ViewModeDiscovery ViewMode = "discovery"
	ViewModeOverview  ViewMode = "overview"
	ViewModeDetail    ViewMode = "detail"
)

// CheckResult is an alias for policy.Result for backwards compatibility.
type CheckResult = policy.Result
