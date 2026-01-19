// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/discovery"
	"github.com/kopexa-grc/kspec/pkg/concurrency"
	_ "github.com/kopexa-grc/kspec/provider/all" // Import all providers to register them
	"github.com/kopexa-grc/kspec/provider"
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover resources without policy evaluation",
	Long: `Discover resources across different providers to inventory your infrastructure.

This command discovers and lists all resources without evaluating policies.
Useful for:
  - Inventory audits
  - Pre-scan resource counts
  - Infrastructure documentation
  - Integration with other tools (JSON output)

Run 'kspec discover <provider> --help' to see provider-specific options.`,
}

func init() {
	rootCmd.AddCommand(discoverCmd)

	// Dynamically create subcommands for each registered provider
	for _, def := range provider.All() {
		providerCmd := createDiscoverProviderCommand(def)
		discoverCmd.AddCommand(providerCmd)
	}
}

// createDiscoverProviderCommand creates a cobra command for a specific provider.
func createDiscoverProviderCommand(def *provider.ProviderDefinition) *cobra.Command {
	cmd := &cobra.Command{
		Use:     def.Name,
		Short:   fmt.Sprintf("Discover %s resources", def.Name),
		Long:    buildDiscoverMultiAssetDescription(def),
		Aliases: def.Aliases,
	}

	// Always create subcommands for asset types
	for i := range def.AssetTypes {
		at := &def.AssetTypes[i]
		assetCmd := createDiscoverAssetCommand(def, at)
		cmd.AddCommand(assetCmd)
	}

	// For single-asset providers, also allow direct execution without subcommand
	// e.g., "kspec discover host julian.pro" instead of "kspec discover network host julian.pro"
	if len(def.AssetTypes) == 1 {
		at := def.AssetTypes[0]
		cmd.Use = buildAssetUsage(def.Name, &at)
		cmd.Args = buildArgsValidator(&at)
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return runAssetDiscover(cmd, def, &at, args)
		}
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		// Register flags for direct execution
		registerDiscoverFlags(cmd)
		provider.RegisterFlags(cmd, def)
	}

	return cmd
}

// createDiscoverAssetCommand creates a cobra command for a specific asset type.
func createDiscoverAssetCommand(def *provider.ProviderDefinition, at *provider.AssetDefinition) *cobra.Command {
	cmd := &cobra.Command{
		Use:     buildAssetUsage(at.Name, at),
		Short:   at.Description,
		Long:    at.Description,
		Aliases: at.Aliases,
		Args:    buildArgsValidator(at),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAssetDiscover(cmd, def, at, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Register flags
	registerDiscoverFlags(cmd)
	provider.RegisterFlags(cmd, def)

	return cmd
}

// buildDiscoverMultiAssetDescription builds a detailed description for providers with multiple asset types.
func buildDiscoverMultiAssetDescription(def *provider.ProviderDefinition) string {
	var sb strings.Builder
	sb.WriteString(def.Description)
	sb.WriteString("\n\nAvailable asset types:\n")
	for _, at := range def.AssetTypes {
		sb.WriteString(fmt.Sprintf("  %s", at.Name))
		if len(at.Aliases) > 0 {
			sb.WriteString(fmt.Sprintf(" (aliases: %s)", strings.Join(at.Aliases, ", ")))
		}
		if at.Description != "" {
			sb.WriteString(fmt.Sprintf(" - %s", at.Description))
		}
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("\nRun 'kspec discover %s <asset-type> --help' for asset-specific options.", def.Name))
	return sb.String()
}

// registerDiscoverFlags registers flags specific to the discover command.
func registerDiscoverFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "Output format (json, table, tree)")
	cmd.Flags().StringP("export", "e", "", "Export results to file")
	cmd.Flags().Bool("include-instances", false, "Include individual resource instances (required for graph)")
	cmd.Flags().Bool("graph", false, "Include graph data in JSON output (implies --include-instances)")

	// Concurrency flags
	cmd.Flags().Int("max-workers", concurrency.DefaultMaxWorkers, "Maximum number of concurrent workers")
	cmd.Flags().Bool("sequential", false, "Disable concurrency (run sequentially)")
}

// runAssetDiscover executes the discovery for a specific provider and asset type.
func runAssetDiscover(cmd *cobra.Command, def *provider.ProviderDefinition, at *provider.AssetDefinition, args []string) error {
	ctx := context.Background()

	// Build asset config from positional args
	assetConfig := make(map[string]string)
	assetName := at.DefaultName
	if assetName == "" {
		assetName = "default"
	}

	for i, arg := range at.Args {
		if i < len(args) {
			if arg.ConfigKey != "" {
				assetConfig[arg.ConfigKey] = args[i]
			}
			if i == 0 {
				assetName = args[i]
			}
		}
	}

	// Build asset type key
	fullAssetType := def.GetScannerKey(at.Name)

	// Build provider config from flags
	providerConfig := provider.BuildConfig(cmd, def)

	// Merge asset config into provider config
	for k, v := range assetConfig {
		providerConfig[k] = v
	}

	// Build asset
	asset := core.Asset{
		Type:   fullAssetType,
		Name:   assetName,
		Config: providerConfig,
	}

	// Build concurrency config
	maxWorkers, _ := cmd.Flags().GetInt("max-workers") //nolint:errcheck // Flag is always defined
	sequential, _ := cmd.Flags().GetBool("sequential") //nolint:errcheck // Flag is always defined
	if sequential {
		maxWorkers = 1
	}

	// Check flags
	includeInstances, _ := cmd.Flags().GetBool("include-instances") //nolint:errcheck // Flag is always defined
	includeGraph, _ := cmd.Flags().GetBool("graph")                 //nolint:errcheck // Flag is always defined
	if includeGraph {
		includeInstances = true // Graph requires instances
	}

	// Create discovery config
	discConfig := discovery.Config{
		ProviderName:     def.Name,
		ProviderConfig:   providerConfig,
		Asset:            asset,
		Concurrency:      maxWorkers,
		IncludeInstances: includeInstances,
	}

	// Create discoverer
	disc := discovery.NewDiscoverer(discConfig)

	// Set up logging
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()

	disc.OnEvent(func(event discovery.Event) {
		switch event.Type {
		case discovery.EventDiscoveryStarted:
			logger.Info().Msg("Discovery started")
		case discovery.EventTreeCreated, discovery.EventTreeUpdated:
			// Tree events are handled by UI, not logging
		case discovery.EventResourceDiscovering, discovery.EventNodeScanning:
			// Only log for resource type nodes, not instances
			if event.Node != nil && (event.Node.Type == discovery.NodeTypeRoot || event.Node.Type == discovery.NodeTypeResourceType) {
				logger.Debug().Str("resource", event.ResourceType).Msg("Discovering")
			}
		case discovery.EventResourceDiscovered, discovery.EventNodeComplete:
			// Only log for resource type nodes, not instances
			if event.Node != nil && (event.Node.Type == discovery.NodeTypeRoot || event.Node.Type == discovery.NodeTypeResourceType) {
				logger.Info().Str("resource", event.ResourceType).Int("count", event.Node.ResourceCount).Msg("Discovered")
			}
		case discovery.EventDiscoveryComplete:
			logger.Info().Msg("Discovery complete")
		case discovery.EventDiscoveryError, discovery.EventNodeError:
			logger.Error().Str("resource", event.ResourceType).Err(event.Error).Msg("Error")
		case discovery.EventEvaluateStarted:
			logger.Info().Msg("Evaluation started")
		case discovery.EventEvaluateComplete:
			logger.Info().Msg("Evaluation complete")
		}
	})

	// Initialize provider
	if err := disc.Initialize(ctx); err != nil {
		return err
	}

	// Run discovery
	result, err := disc.Discover(ctx, asset)
	if err != nil {
		return err
	}

	// Get output format
	outputFormat, _ := cmd.Flags().GetString("output") //nolint:errcheck // Flag is always defined
	format := discovery.OutputFormat(outputFormat)

	// Output results
	if err := discovery.Export(result, format, os.Stdout); err != nil {
		return err
	}

	// Export to file if requested
	exportPath, _ := cmd.Flags().GetString("export") //nolint:errcheck // Flag is always defined
	if exportPath != "" {
		f, err := os.Create(exportPath)
		if err != nil {
			return fmt.Errorf("failed to create export file: %w", err)
		}
		defer f.Close() //nolint:errcheck // Best effort close

		// Always export as JSON to file
		if err := discovery.Export(result, discovery.FormatJSON, f); err != nil {
			return fmt.Errorf("failed to export: %w", err)
		}
		logger.Info().Str("path", exportPath).Msg("Results exported")
	}

	return nil
}
