// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

// Common AWS string constants used across resource fetchers.
const (
	// Status values
	statusActive    = "ACTIVE"
	statusAvailable = "available"
	statusEnabled   = "ENABLED"
	valueEnabled    = "enabled"
	valueTrue       = "true"
	valueNone       = "NONE"

	// IAM policy effect
	policyEffectAllow = "Allow"

	// CloudTrail read/write type
	readWriteTypeAll = "All"

	// CIDR blocks for public access
	cidrAllIPv4 = "0.0.0.0/0"
	cidrAllIPv6 = "::/0"
)
