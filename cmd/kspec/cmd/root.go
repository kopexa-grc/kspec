// Package cmd provides the command-line interface for kspec security scanning.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Build information - injected at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	// Persistent flags for all commands
	policyFile string
	policyDir  string

	// Credential flags
	token          string
	credentialType string
	envVar         string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "kspec",
	Short: "kspec - Kopexa Policy Specification Scanner",
	Long: `kspec is a powerful policy-as-code scanner that validates infrastructure,
security, and compliance requirements across various platforms including
GitHub organizations, repositories, and network resources.`,
}

// SetVersionInfo sets the version information for the CLI
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVarP(&policyFile, "policy", "f", "", "Path to policy YAML file")
	rootCmd.PersistentFlags().StringVarP(&policyDir, "policy-dir", "d", "", "Path to policy directory")

	// Credential flags
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Authentication token (e.g., GitHub token)")
	rootCmd.PersistentFlags().StringVar(&credentialType, "credential-type", "env", "Credential type (bearer, env, password)")
	rootCmd.PersistentFlags().StringVar(&envVar, "env-var", "GITHUB_TOKEN", "Environment variable name for token")
}
