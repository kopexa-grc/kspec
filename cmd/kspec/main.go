// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package main is the entry point for the kspec CLI tool.
package main

import (
	"github.com/kopexa-grc/kspec/cmd/kspec/cmd"
)

// Build information - injected at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
