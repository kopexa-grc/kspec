// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package cmd

// Import all providers to trigger their init() registration.
// This is done here in the CLI package to avoid circular imports
// in the provider package.
import (
	_ "github.com/kopexa-grc/kspec/provider/atlassian"
	_ "github.com/kopexa-grc/kspec/provider/aws"
	_ "github.com/kopexa-grc/kspec/provider/azure"
	_ "github.com/kopexa-grc/kspec/provider/cloudflare"
	_ "github.com/kopexa-grc/kspec/provider/factorial"
	_ "github.com/kopexa-grc/kspec/provider/github"
	_ "github.com/kopexa-grc/kspec/provider/hetzner"
	_ "github.com/kopexa-grc/kspec/provider/ms365"
	_ "github.com/kopexa-grc/kspec/provider/network"
	_ "github.com/kopexa-grc/kspec/provider/os"
	_ "github.com/kopexa-grc/kspec/provider/sbom"
)
