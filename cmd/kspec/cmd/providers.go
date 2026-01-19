// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package cmd

// Blank imports to trigger provider init() registration.
// This is done here in the CLI package to avoid circular imports in the provider package.
import (
	_ "github.com/kopexa-grc/kspec/provider/atlassian"  // provider registration
	_ "github.com/kopexa-grc/kspec/provider/aws"        // provider registration
	_ "github.com/kopexa-grc/kspec/provider/azure"      // provider registration
	_ "github.com/kopexa-grc/kspec/provider/cloudflare" // provider registration
	_ "github.com/kopexa-grc/kspec/provider/factorial"  // provider registration
	_ "github.com/kopexa-grc/kspec/provider/github"     // provider registration
	_ "github.com/kopexa-grc/kspec/provider/hetzner"    // provider registration
	_ "github.com/kopexa-grc/kspec/provider/ms365"      // provider registration
	_ "github.com/kopexa-grc/kspec/provider/network"    // provider registration
	_ "github.com/kopexa-grc/kspec/provider/os"         // provider registration
	_ "github.com/kopexa-grc/kspec/provider/sbom"       // provider registration
)
