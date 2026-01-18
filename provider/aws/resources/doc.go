// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package resources provides AWS resource fetchers for security scanning.
// Each resource type implements the core.ResourceSpec interface and fetches
// cloud resources with their security-relevant properties.
//
// The resources in this package are designed to be used with the AWS Client
// from the parent aws package, which provides rate limiting and consistent
// access to AWS service clients.
package resources
