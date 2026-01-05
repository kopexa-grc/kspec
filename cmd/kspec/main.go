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
