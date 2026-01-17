// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package all imports all provider packages to trigger their init() registrations.
// Import this package to register all providers with the registry.
package all

import (
	// Import all providers to trigger their init() registration
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
