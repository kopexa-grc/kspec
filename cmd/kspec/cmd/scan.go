// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/kopexa-grc/kspec/cli"
	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/core"
	_ "github.com/kopexa-grc/kspec/provider/all" // Import all providers to register them
	"github.com/kopexa-grc/kspec/provider/registry"
	"github.com/kopexa-grc/kspec/provider/scanner"
	"github.com/kopexa-grc/kspec/report"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan resources using policy checks",
	Long: `Scan resources across different providers using policy-as-code checks.

Run 'kspec scan <provider> --help' to see provider-specific options.`,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Dynamically create subcommands for each registered provider
	for _, def := range registry.All() {
		providerCmd := createProviderCommand(def)
		scanCmd.AddCommand(providerCmd)
	}
}

// createProviderCommand creates a cobra command for a specific provider.
func createProviderCommand(def *registry.ProviderDefinition) *cobra.Command {
	cmd := &cobra.Command{
		Use:     def.Name,
		Short:   def.Description,
		Long:    def.Description,
		Aliases: def.Aliases,
	}

	// If provider has multiple asset types, create subcommands for each
	// If only one asset type, make it the default behavior
	if len(def.AssetTypes) == 1 {
		// Single asset type - run directly on provider command
		at := def.AssetTypes[0]
		cmd.Use = buildAssetUsage(def.Name, &at)
		cmd.Example = buildAssetExampleWithSubcmd(def.Name, def.Name, &at, def.Flags)
		cmd.Args = buildArgsValidator(&at)
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return runAssetScan(cmd, def, &at, args)
		}
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		// Register flags
		registerCommonFlags(cmd)
		registry.RegisterFlags(cmd, def)
	} else {
		// Multiple asset types - create subcommands
		for i := range def.AssetTypes {
			at := &def.AssetTypes[i]
			assetCmd := createAssetCommand(def, at)
			cmd.AddCommand(assetCmd)
		}
	}

	return cmd
}

// createAssetCommand creates a cobra command for a specific asset type.
func createAssetCommand(def *registry.ProviderDefinition, at *registry.AssetDefinition) *cobra.Command {
	cmd := &cobra.Command{
		Use:     buildAssetUsage(at.Name, at),
		Short:   at.Description,
		Long:    at.Description,
		Aliases: at.Aliases,
		Example: buildAssetExample(def.Name, at, def.Flags),
		Args:    buildArgsValidator(at),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAssetScan(cmd, def, at, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Register flags
	registerCommonFlags(cmd)
	registry.RegisterFlags(cmd, def)

	return cmd
}

// buildAssetUsage builds the usage string for an asset type.
func buildAssetUsage(name string, at *registry.AssetDefinition) string {
	usage := name
	for _, arg := range at.Args {
		if arg.Required {
			usage += fmt.Sprintf(" <%s>", arg.Name)
		} else {
			usage += fmt.Sprintf(" [%s]", arg.Name)
		}
	}
	return usage
}

// buildAssetExample builds example usage for an asset type.
func buildAssetExample(providerName string, at *registry.AssetDefinition, flags []registry.FlagDefinition) string {
	return buildAssetExampleWithSubcmd(providerName, at.Name, at, flags)
}

// buildAssetExampleWithSubcmd builds example usage, optionally including asset type as subcommand.
func buildAssetExampleWithSubcmd(providerName, assetCmd string, at *registry.AssetDefinition, flags []registry.FlagDefinition) string {
	var examples []string

	// Basic example
	example := fmt.Sprintf("  kspec scan %s", providerName)
	if assetCmd != providerName {
		example += " " + assetCmd
	}
	if len(at.Args) > 0 && at.Args[0].Required {
		example += fmt.Sprintf(" <%s>", at.Args[0].Name)
	}
	example += " -f policy.yml"
	examples = append(examples, example)

	// Example with a flag if provider has flags
	if len(flags) > 0 {
		flagExample := fmt.Sprintf("  kspec scan %s", providerName)
		if assetCmd != providerName {
			flagExample += " " + assetCmd
		}
		if len(at.Args) > 0 && at.Args[0].Required {
			flagExample += fmt.Sprintf(" <%s>", at.Args[0].Name)
		}
		flagExample += fmt.Sprintf(" --%s <value> -f policy.yml", flags[0].Name)
		examples = append(examples, flagExample)
	}

	return strings.Join(examples, "\n")
}

// buildArgsValidator builds an argument validator for an asset type.
func buildArgsValidator(at *registry.AssetDefinition) cobra.PositionalArgs {
	// Count required args
	requiredCount := 0
	for _, arg := range at.Args {
		if arg.Required {
			requiredCount++
		}
	}

	if requiredCount > 0 {
		return func(cmd *cobra.Command, args []string) error {
			if len(args) < requiredCount {
				var missing []string
				for i, arg := range at.Args {
					if arg.Required && i >= len(args) {
						missing = append(missing, arg.Name)
					}
				}
				return fmt.Errorf("missing required argument(s): %s", strings.Join(missing, ", "))
			}
			return nil
		}
	}

	return cobra.ArbitraryArgs
}

// registerCommonFlags registers flags common to all scan commands.
func registerCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("policy", "f", "", "Policy file to use")
	cmd.Flags().StringP("policy-dir", "d", "", "Directory containing policy files")
	cmd.Flags().StringP("export", "o", "", "Export results to file (csv, xlsx, json)")
	cmd.Flags().String("export-format", "", "Export format (auto-detected from filename)")
	cmd.Flags().Bool("no-ui", false, "Disable interactive UI (for CI/CD)")
}

// runAssetScan executes the scan for a specific provider and asset type.
func runAssetScan(cmd *cobra.Command, def *registry.ProviderDefinition, at *registry.AssetDefinition, args []string) error {
	ctx := context.Background()

	// Build asset config from positional args
	assetConfig := make(map[string]string)
	assetName := at.DefaultName
	if assetName == "" {
		assetName = "default"
	}

	for i, arg := range at.Args {
		if i < len(args) {
			assetConfig[arg.ConfigKey] = args[i]
			if i == 0 {
				assetName = args[i]
			}
		}
	}

	// Build asset type key
	fullAssetType := def.GetScannerKey(at.Name)

	// Build provider config from flags
	providerConfig := registry.BuildConfig(cmd, def)

	// Merge asset config into provider config.
	// Asset-level configuration takes precedence over provider-level
	// configuration when keys collide (more specific wins).
	for k, v := range assetConfig {
		providerConfig[k] = v
	}

	// Build asset with merged config so providers have access to all configuration
	asset := core.Asset{
		Type:   fullAssetType,
		Name:   assetName,
		Config: providerConfig,
	}

	// Load and filter policies
	policies, err := loadAndFilterPolicies(cmd, def.Name)
	if err != nil {
		return err
	}

	// Create scan config
	scanConfig := scanner.ScanConfig{
		ProviderName:   def.Name,
		ProviderConfig: providerConfig,
		Asset:          asset,
		Policies:       policies,
	}

	// Create scanner
	s := scanner.NewScanner(scanConfig)

	// Initialize provider
	if err := s.Initialize(ctx); err != nil {
		return err
	}

	// Check for no-ui mode
	noUI, _ := cmd.Flags().GetBool("no-ui") //nolint:errcheck // Flag is defined
	if noUI {
		return runScanNoUI(ctx, cmd, s, def.Name)
	}

	return runScanWithUI(ctx, cmd, s, def.Name)
}

// runScanWithUI runs the scan with the interactive TUI
func runScanWithUI(ctx context.Context, cmd *cobra.Command, s *scanner.Scanner, providerName string) error {
	// Initialize bubbletea UI
	uiModel := cli.InitialModel()
	p := tea.NewProgram(uiModel)

	// Set up event handler to forward events to UI
	s.OnEvent(func(event scanner.ScanEvent) {
		switch event.Type {
		case scanner.EventTreeCreated:
			p.Send(cli.SetTreeMsg{Tree: event.Tree})
		case scanner.EventDiscoveryStarted,
			scanner.EventTreeUpdated,
			scanner.EventDiscoveryComplete,
			scanner.EventScanStarted,
			scanner.EventScanComplete,
			scanner.EventResourceScanning,
			scanner.EventResourceComplete:
			p.Send(cli.UpdateTreeMsg{Tree: event.Tree})
		case scanner.EventError:
			p.Send(cli.UpdateTreeMsg{Tree: event.Tree})
		}
	})

	// Run scan in background
	go func() {
		result := s.Run(ctx)
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				p.Send(cli.ScanErrorMsg{Error: err})
			}
		}
	}()

	// Run the UI
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("UI error: %w", err)
	}

	// Check if export is requested
	exportPath, _ := cmd.Flags().GetString("export") //nolint:errcheck // Flag is defined
	if exportPath != "" {
		if model, ok := finalModel.(cli.Model); ok {
			tree := model.GetTree()
			if tree != nil {
				if err := exportResults(cmd, tree, providerName, exportPath); err != nil {
					return fmt.Errorf("export failed: %w", err)
				}
				fmt.Printf("Results exported to: %s\n", exportPath)
			}
		}
	}

	return nil
}

// runScanNoUI runs the scan with logging output instead of the TUI
func runScanNoUI(ctx context.Context, cmd *cobra.Command, s *scanner.Scanner, providerName string) error {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Timestamp().
		Logger()

	var finalTree *common.ResourceTree

	s.OnEvent(func(event scanner.ScanEvent) {
		finalTree = event.Tree

		switch event.Type {
		case scanner.EventTreeCreated:
			logger.Info().Str("target", event.Tree.Root.Name).Msg("Scan initialized")
		case scanner.EventDiscoveryStarted:
			logger.Info().Msg("Discovery started")
		case scanner.EventDiscoveryComplete:
			logger.Info().Msg("Discovery complete")
		case scanner.EventScanStarted:
			logger.Info().Msg("Scan started")
		case scanner.EventTreeUpdated:
			// no log needed
		case scanner.EventResourceScanning:
			if event.Tree.Root != nil {
				logger.Info().Msg("Scanning resources")
			}
		case scanner.EventResourceComplete:
			if event.Tree.Root != nil {
				logResourceResults(&logger, event.Tree.Root)
			}
		case scanner.EventScanComplete:
			logger.Info().Msg("Scan complete")
			if event.Tree.Root != nil {
				logSummary(&logger, event.Tree)
			}
		case scanner.EventError:
			logger.Error().Msg("Scan error occurred")
		}
	})

	result := s.Run(ctx)

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			logger.Error().Err(err).Msg("Scan error")
		}
	}

	exportPath, _ := cmd.Flags().GetString("export") //nolint:errcheck // Flag is defined
	if exportPath != "" && finalTree != nil {
		if err := exportResults(cmd, finalTree, providerName, exportPath); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		logger.Info().Str("path", exportPath).Msg("Results exported")
	}

	return nil
}

func logResourceResults(logger *zerolog.Logger, node *common.ResourceNode) {
	for _, check := range node.Checks {
		switch check.Status {
		case "pass", "passed":
			logger.Info().Str("id", check.ID).Str("status", "PASS").Msg(check.Name)
		case "fail", "failed":
			logger.Warn().Str("id", check.ID).Str("status", "FAIL").Str("severity", check.Severity).Str("details", check.Details).Msg(check.Name)
		case "skip", "skipped":
			logger.Info().Str("id", check.ID).Str("status", "SKIP").Msg(check.Name)
		default:
			logger.Info().Str("id", check.ID).Str("status", strings.ToUpper(check.Status)).Msg(check.Name)
		}
	}
	for _, child := range node.Children {
		logResourceResults(logger, child)
	}
}

func logSummary(logger *zerolog.Logger, tree *common.ResourceTree) {
	if tree.Root == nil {
		return
	}
	total := tree.Root.ChecksPassed + tree.Root.ChecksFailed + tree.Root.ChecksSkipped
	logger.Info().Int("total", total).Int("passed", tree.Root.ChecksPassed).Int("failed", tree.Root.ChecksFailed).Int("skipped", tree.Root.ChecksSkipped).Msg("Scan summary")
}

func loadAndFilterPolicies(cmd *cobra.Command, providerName string) ([]core.Policy, error) {
	policyFile, _ := cmd.Flags().GetString("policy")    //nolint:errcheck // Flag is defined
	policyDir, _ := cmd.Flags().GetString("policy-dir") //nolint:errcheck // Flag is defined

	// Default to ./policies if no policy specified
	if policyFile == "" && policyDir == "" {
		if info, err := os.Stat("policies"); err == nil && info.IsDir() {
			policyDir = "policies"
		} else {
			return nil, fmt.Errorf("no policy specified and no ./policies directory found. Use --policy (-f) or --policy-dir (-d), or create a ./policies directory")
		}
	}

	policies, err := scanner.LoadPolicies(policyFile, policyDir)
	if err != nil {
		return nil, err
	}

	policies = scanner.FilterPoliciesByProvider(policies, providerName, providerName)

	if len(policies) == 0 {
		return nil, fmt.Errorf("no policies found for provider %s", providerName)
	}

	return policies, nil
}

func exportResults(cmd *cobra.Command, tree *common.ResourceTree, provider, exportPath string) error {
	rep := report.FromResourceTree(tree, provider)

	format, _ := cmd.Flags().GetString("export-format") //nolint:errcheck // Flag is defined
	if format == "" {
		ext := strings.ToLower(filepath.Ext(exportPath))
		switch ext {
		case ".csv":
			format = "csv"
		case ".xlsx":
			format = "xlsx"
		case ".json":
			format = "json"
		default:
			return fmt.Errorf("unknown export format, use --export-format or specify .csv, .xlsx, or .json extension")
		}
	}

	switch report.Format(format) {
	case report.FormatCSV:
		return report.NewCSVExporter().Export(rep, exportPath)
	case report.FormatXLSX:
		return report.NewXLSXExporter().Export(rep, exportPath)
	case report.FormatJSON:
		return report.NewJSONExporter(true).Export(rep, exportPath)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}
