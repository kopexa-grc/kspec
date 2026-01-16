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
	"github.com/kopexa-grc/kspec/provider/scanner"
	"github.com/kopexa-grc/kspec/report"
)

// Provider name constants
const (
	ProviderHost       = "host"
	ProviderGitHub     = "github"
	ProviderAWS        = "aws"
	ProviderAzure      = "azure"
	ProviderMS365      = "ms365"
	ProviderCloudflare = "cloudflare"
	ProviderAtlassian  = "atlassian"
	ProviderSBOM       = "sbom"
	ProviderHetzner    = "hetzner"
	ProviderFactorial  = "factorial"
)

// Asset type constants
const (
	AssetTypeSBOMFile = "sbom-file"
	AssetTypeDefault  = "default"
)

// Credential type constants
const (
	CredentialTypeEnv = "env"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan <provider> [resource-type] [target]",
	Short: "Scan resources using policy checks",
	Long: `Scan resources across different providers using policy-as-code checks.

Examples:
  kspec scan local -f policy.yml
  kspec scan host example.com -f policy.yml
  kspec scan github org kopexa-grc -f policy.yml
  kspec scan github repo owner/repo -f policy.yml
  kspec scan aws account -f policy.yml
  kspec scan aws account --profile production -f policy.yml
  kspec scan aws account --region us-west-2 --regions us-east-1,eu-west-1 -f policy.yml
  kspec scan aws account --role-arn arn:aws:iam::123456789:role/SecurityAudit -f policy.yml
  kspec scan azure subscription <sub-id> -f policy.yml
  kspec scan ms365 tenant <tenant-id> --client-id <id> --client-secret <secret> -f policy.yml
  kspec scan cloudflare account -f policy.yml
  kspec scan cloudflare zone <zone-id> --api-token <token> -f policy.yml
  kspec scan atlassian site yoursite.atlassian.net -f policy.yml
  kspec scan atlassian org <org-id> --site yoursite.atlassian.net -f policy.yml
  kspec scan sbom file ./sbom.json -f policy.yml
  kspec scan sbom dir ./sboms/ -f policy.yml
  kspec scan hetzner project -f policy.yml
  kspec scan hcloud project my-project --api-token <token> -f policy.yml
  kspec scan factorial -f policy.yml
  kspec scan factorial --api-key <key> -f policy.yml`,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE:         runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Common flags
	scanCmd.Flags().StringP("policy", "f", "", "Policy file to use")
	scanCmd.Flags().StringP("policy-dir", "d", "", "Directory containing policy files")
	scanCmd.Flags().String("token", "", "Authentication token")
	scanCmd.Flags().String("credential-type", "", "Credential type (bearer, env, password)")
	scanCmd.Flags().String("env-var", "GITHUB_TOKEN", "Environment variable for credentials")

	// Azure-specific flags
	scanCmd.Flags().String("client-id", "", "Azure/MS365 Client ID (Service Principal)")
	scanCmd.Flags().String("tenant-id", "", "Azure/MS365 Tenant ID")
	scanCmd.Flags().String("resource-group", "", "Azure Resource Group (optional filter)")

	// MS365-specific flags
	scanCmd.Flags().String("client-secret", "", "MS365 Client Secret (Service Principal)")

	// Cloudflare-specific flags
	scanCmd.Flags().String("api-token", "", "Cloudflare API Token")
	scanCmd.Flags().String("api-key", "", "Cloudflare API Key (legacy)")
	scanCmd.Flags().String("email", "", "Cloudflare/Atlassian account email")
	scanCmd.Flags().String("account-id", "", "Cloudflare Account ID")

	// Atlassian-specific flags
	scanCmd.Flags().String("site", "", "Atlassian site (e.g., yoursite.atlassian.net)")
	scanCmd.Flags().String("org-id", "", "Atlassian Organization ID (for admin APIs)")

	// SBOM-specific flags
	scanCmd.Flags().String("sbom-path", "", "Path to SBOM file or directory")

	// Hetzner-specific flags
	scanCmd.Flags().String("hcloud-token", "", "Hetzner Cloud API token")
	scanCmd.Flags().String("project", "", "Hetzner Cloud project name (for identification)")

	// Factorial-specific flags
	scanCmd.Flags().String("factorial-api-key", "", "Factorial HR API key")
	scanCmd.Flags().String("access-token", "", "Factorial HR OAuth access token")

	// AWS-specific flags
	scanCmd.Flags().String("profile", "", "AWS profile name from ~/.aws/credentials")
	scanCmd.Flags().String("region", "", "AWS region (default: us-east-1)")
	scanCmd.Flags().StringSlice("regions", nil, "Multiple AWS regions for multi-region scanning")
	scanCmd.Flags().String("access-key-id", "", "AWS access key ID")
	scanCmd.Flags().String("secret-access-key", "", "AWS secret access key")
	scanCmd.Flags().String("session-token", "", "AWS session token")
	scanCmd.Flags().String("role-arn", "", "IAM role ARN to assume for cross-account access")
	scanCmd.Flags().String("external-id", "", "External ID for assume role")

	// Export flags
	scanCmd.Flags().StringP("export", "o", "", "Export results to file (supported formats: csv, xlsx, json)")
	scanCmd.Flags().String("export-format", "", "Export format (csv, xlsx, json). Auto-detected from filename if not specified")

	// Output mode flags
	scanCmd.Flags().Bool("no-ui", false, "Disable interactive UI and log output to stdout (useful for CI/CD)")
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse CLI arguments
	scanConfig, providerAlias, err := parseScanArgs(cmd, args)
	if err != nil {
		return err
	}

	// Load and filter policies
	policies, err := loadAndFilterPolicies(cmd, scanConfig.ProviderName, providerAlias)
	if err != nil {
		return err
	}
	scanConfig.Policies = policies

	// Create scanner
	s := scanner.NewScanner(scanConfig)

	// Initialize provider
	if err := s.Initialize(ctx); err != nil {
		return err
	}

	// Check for no-ui mode
	noUI, _ := cmd.Flags().GetBool("no-ui") //nolint:errcheck // Flag is defined
	if noUI {
		return runScanNoUI(ctx, cmd, s, scanConfig.ProviderName)
	}

	return runScanWithUI(ctx, cmd, s, scanConfig.ProviderName)
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
		// Log any errors that occurred during scanning
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
		// Get the final tree from the model
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
	// Configure zerolog for console output
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Timestamp().
		Logger()

	var finalTree *common.ResourceTree

	// Set up event handler to log events
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
			// Tree updated, no specific log needed
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

	// Run scan synchronously
	result := s.Run(ctx)

	// Log any errors that occurred during scanning
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			logger.Error().Err(err).Msg("Scan error")
		}
	}

	// Check if export is requested
	exportPath, _ := cmd.Flags().GetString("export") //nolint:errcheck // Flag is defined
	if exportPath != "" && finalTree != nil {
		if err := exportResults(cmd, finalTree, providerName, exportPath); err != nil {
			return fmt.Errorf("export failed: %w", err)
		}
		logger.Info().Str("path", exportPath).Msg("Results exported")
	}

	return nil
}

// logResourceResults logs results for a resource node using zerolog
func logResourceResults(logger *zerolog.Logger, node *common.ResourceNode) {
	for _, check := range node.Checks {
		switch check.Status {
		case "pass", "passed":
			logger.Info().
				Str("id", check.ID).
				Str("status", "PASS").
				Msg(check.Name)
		case "fail", "failed":
			logger.Warn().
				Str("id", check.ID).
				Str("status", "FAIL").
				Str("severity", check.Severity).
				Str("details", check.Details).
				Msg(check.Name)
		case "skip", "skipped":
			logger.Info().
				Str("id", check.ID).
				Str("status", "SKIP").
				Msg(check.Name)
		default:
			logger.Info().
				Str("id", check.ID).
				Str("status", strings.ToUpper(check.Status)).
				Msg(check.Name)
		}
	}

	for _, child := range node.Children {
		logResourceResults(logger, child)
	}
}

// logSummary logs a summary of the scan results using zerolog
func logSummary(logger *zerolog.Logger, tree *common.ResourceTree) {
	if tree.Root == nil {
		return
	}

	total := tree.Root.ChecksPassed + tree.Root.ChecksFailed + tree.Root.ChecksSkipped
	logger.Info().
		Int("total", total).
		Int("passed", tree.Root.ChecksPassed).
		Int("failed", tree.Root.ChecksFailed).
		Int("skipped", tree.Root.ChecksSkipped).
		Msg("Scan summary")
}

// parseScanArgs parses command line arguments and returns a ScanConfig
func parseScanArgs(cmd *cobra.Command, args []string) (scanner.ScanConfig, string, error) {
	providerArg := args[0]

	var assetType string
	var assetName string
	var assetConfig = make(map[string]string)
	var providerName string

	switch providerArg {
	case "local":
		providerName = "os"
		assetType = "local"
		assetName = "localhost"

	case ProviderHost:
		if len(args) < 2 {
			return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan host <target> -f policy.yml")
		}
		providerName = "network"
		assetType = ProviderHost
		assetName = args[1]
		assetConfig["target"] = assetName

	case ProviderGitHub:
		if len(args) < 3 {
			return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan github <org|repo> <target> -f policy.yml")
		}

		providerName = ProviderGitHub
		resourceType := args[1]
		target := args[2]

		switch resourceType {
		case "org":
			assetType = "github-org"
			assetName = target
			assetConfig["owner"] = target

		case "repo":
			parts := strings.Split(target, "/")
			if len(parts) != 2 {
				return scanner.ScanConfig{}, "", fmt.Errorf("repository must be in format: owner/repo")
			}
			assetType = "github-repo"
			assetName = target
			assetConfig["owner"] = parts[0]
			assetConfig["repo"] = parts[1]

		default:
			return scanner.ScanConfig{}, "", fmt.Errorf("unknown github resource type: %s (use 'org' or 'repo')", resourceType)
		}

	case ProviderAWS, "amazon", "ec2", "s3":
		providerName = ProviderAWS
		assetType = "aws-account"

		// Check resource type
		if len(args) >= 2 {
			resourceType := args[1]
			switch resourceType {
			case "account":
				assetName = AssetTypeDefault
			default:
				return scanner.ScanConfig{}, "", fmt.Errorf("unknown aws resource type: %s (use 'account')", resourceType)
			}
		} else {
			assetName = AssetTypeDefault
		}

		// Get profile from flag
		if profile, _ := cmd.Flags().GetString("profile"); profile != "" { //nolint:errcheck // Flag is defined
			assetConfig["profile"] = profile
		}

		// Get region from flag
		if region, _ := cmd.Flags().GetString("region"); region != "" { //nolint:errcheck // Flag is defined
			assetConfig["region"] = region
		}

		// Get regions from flag
		if regions, _ := cmd.Flags().GetStringSlice("regions"); len(regions) > 0 { //nolint:errcheck // Flag is defined
			assetConfig["regions"] = strings.Join(regions, ",")
		}

		// Get credentials from flags
		if accessKeyID, _ := cmd.Flags().GetString("access-key-id"); accessKeyID != "" { //nolint:errcheck // Flag is defined
			assetConfig["access_key"] = accessKeyID
		}
		if secretAccessKey, _ := cmd.Flags().GetString("secret-access-key"); secretAccessKey != "" { //nolint:errcheck // Flag is defined
			assetConfig["secret_key"] = secretAccessKey
		}
		if sessionToken, _ := cmd.Flags().GetString("session-token"); sessionToken != "" { //nolint:errcheck // Flag is defined
			assetConfig["session_token"] = sessionToken
		}

		// Get assume role config
		if roleArn, _ := cmd.Flags().GetString("role-arn"); roleArn != "" { //nolint:errcheck // Flag is defined
			assetConfig["role_arn"] = roleArn
		}
		if externalID, _ := cmd.Flags().GetString("external-id"); externalID != "" { //nolint:errcheck // Flag is defined
			assetConfig["external_id"] = externalID
		}

	case ProviderAzure:
		if len(args) < 3 {
			return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan azure subscription <subscription-id> -f policy.yml")
		}

		providerName = ProviderAzure
		resourceType := args[1]
		target := args[2]

		switch resourceType {
		case "subscription", "sub":
			assetType = "azure"
			assetName = target
			assetConfig["subscription_id"] = target

			// Optional: client_id, tenant_id from flags
			clientID, _ := cmd.Flags().GetString("client-id") //nolint:errcheck // Flag is defined
			if clientID != "" {
				assetConfig["client_id"] = clientID
			}
			tenantID, _ := cmd.Flags().GetString("tenant-id") //nolint:errcheck // Flag is defined
			if tenantID != "" {
				assetConfig["tenant_id"] = tenantID
			}
			resourceGroup, _ := cmd.Flags().GetString("resource-group") //nolint:errcheck // Flag is defined
			if resourceGroup != "" {
				assetConfig["resource_group"] = resourceGroup
			}

		default:
			return scanner.ScanConfig{}, "", fmt.Errorf("unknown azure resource type: %s (use 'subscription')", resourceType)
		}

	case ProviderMS365, "m365", "microsoft365":
		providerName = ProviderMS365
		assetType = ProviderMS365

		// Check if tenant ID is provided
		switch {
		case len(args) >= 3 && args[1] == "tenant":
			assetName = args[2]
			assetConfig["tenant_id"] = args[2]
		case len(args) >= 2:
			// Just the tenant ID directly
			assetName = args[1]
			assetConfig["tenant_id"] = args[1]
		default:
			assetName = AssetTypeDefault
		}

		// Optional: client_id, client_secret from flags
		clientID, _ := cmd.Flags().GetString("client-id") //nolint:errcheck // Flag is defined
		if clientID != "" {
			assetConfig["client_id"] = clientID
		}
		tenantID, _ := cmd.Flags().GetString("tenant-id") //nolint:errcheck // Flag is defined
		if tenantID != "" {
			assetConfig["tenant_id"] = tenantID
			assetName = tenantID
		}
		clientSecret, _ := cmd.Flags().GetString("client-secret") //nolint:errcheck // Flag is defined
		if clientSecret != "" {
			assetConfig["client_secret"] = clientSecret
		}

	case ProviderCloudflare, "cf":
		providerName = ProviderCloudflare

		// Check resource type
		if len(args) >= 2 {
			resourceType := args[1]
			switch resourceType {
			case "account":
				assetType = "cloudflare-account"
				if len(args) >= 3 {
					assetName = args[2]
					assetConfig["account_id"] = args[2]
				} else {
					assetName = "all"
				}
			case "zone":
				if len(args) < 3 {
					return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan cloudflare zone <zone-id> -f policy.yml")
				}
				assetType = "cloudflare-zone"
				assetName = args[2]
				assetConfig["zone_id"] = args[2]
			default:
				return scanner.ScanConfig{}, "", fmt.Errorf("unknown cloudflare resource type: %s (use 'account' or 'zone')", resourceType)
			}
		} else {
			// Default to account scan
			assetType = "cloudflare-account"
			assetName = "all"
		}

		// Get credentials from flags
		apiToken, _ := cmd.Flags().GetString("api-token") //nolint:errcheck // Flag is defined
		if apiToken != "" {
			assetConfig["api_token"] = apiToken
		}
		apiKey, _ := cmd.Flags().GetString("api-key") //nolint:errcheck // Flag is defined
		if apiKey != "" {
			assetConfig["api_key"] = apiKey
		}
		email, _ := cmd.Flags().GetString("email") //nolint:errcheck // Flag is defined
		if email != "" {
			assetConfig["email"] = email
		}
		accountID, _ := cmd.Flags().GetString("account-id") //nolint:errcheck // Flag is defined
		if accountID != "" {
			assetConfig["account_id"] = accountID
		}

	case ProviderAtlassian, "jira", "confluence":
		providerName = ProviderAtlassian

		// Check resource type
		if len(args) >= 2 {
			resourceType := args[1]
			switch resourceType {
			case "site":
				assetType = "atlassian-site"
				if len(args) >= 3 {
					assetName = args[2]
					assetConfig["site"] = args[2]
				} else {
					// Get site from flag
					site, _ := cmd.Flags().GetString("site") //nolint:errcheck // Flag is defined
					if site != "" {
						assetName = site
						assetConfig["site"] = site
					} else {
						return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan atlassian site <site-name> -f policy.yml")
					}
				}
			case "org":
				if len(args) < 3 {
					return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan atlassian org <org-id> -f policy.yml")
				}
				assetType = "atlassian-org"
				assetName = args[2]
				assetConfig["org_id"] = args[2]
			default:
				return scanner.ScanConfig{}, "", fmt.Errorf("unknown atlassian resource type: %s (use 'site' or 'org')", resourceType)
			}
		} else {
			// Default to site scan
			assetType = "atlassian-site"
			site, _ := cmd.Flags().GetString("site") //nolint:errcheck // Flag is defined
			if site != "" {
				assetName = site
				assetConfig["site"] = site
			} else {
				return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan atlassian site <site-name> -f policy.yml")
			}
		}

		// Get credentials from flags
		if apiToken, _ := cmd.Flags().GetString("api-token"); apiToken != "" { //nolint:errcheck // Flag is defined
			assetConfig["api_token"] = apiToken
		}
		if email, _ := cmd.Flags().GetString("email"); email != "" { //nolint:errcheck // Flag is defined
			assetConfig["email"] = email
		}
		if site, _ := cmd.Flags().GetString("site"); site != "" && assetConfig["site"] == "" { //nolint:errcheck // Flag is defined
			assetConfig["site"] = site
		}
		if orgID, _ := cmd.Flags().GetString("org-id"); orgID != "" { //nolint:errcheck // Flag is defined
			assetConfig["org_id"] = orgID
		}

	case ProviderSBOM, "bom", "cyclonedx", "spdx":
		providerName = ProviderSBOM

		// Check resource type
		if len(args) >= 2 {
			resourceType := args[1]
			switch resourceType {
			case "file":
				if len(args) < 3 {
					return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan sbom file <path> -f policy.yml")
				}
				assetType = AssetTypeSBOMFile
				assetName = args[2]
				assetConfig["sbom_path"] = args[2]
			case "dir", "directory":
				if len(args) < 3 {
					return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan sbom dir <path> -f policy.yml")
				}
				assetType = "sbom-directory"
				assetName = args[2]
				assetConfig["sbom_path"] = args[2]
			default:
				// Assume it's a file path directly
				assetType = AssetTypeSBOMFile
				assetName = resourceType
				assetConfig["sbom_path"] = resourceType
			}
		} else {
			// Get path from flag
			sbomPath, _ := cmd.Flags().GetString("sbom-path") //nolint:errcheck // Flag is defined
			if sbomPath != "" {
				assetType = AssetTypeSBOMFile
				assetName = sbomPath
				assetConfig["sbom_path"] = sbomPath
			} else {
				return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan sbom file <path> -f policy.yml")
			}
		}

		// Get path from flag if not set
		if sbomPath, _ := cmd.Flags().GetString("sbom-path"); sbomPath != "" && assetConfig["sbom_path"] == "" { //nolint:errcheck // Flag is defined
			assetConfig["sbom_path"] = sbomPath
		}

	case ProviderHetzner, "hcloud":
		providerName = ProviderHetzner
		assetType = "hetzner-project"

		// Check if project name is provided
		if len(args) >= 2 {
			resourceType := args[1]
			switch resourceType {
			case "project":
				if len(args) >= 3 {
					assetName = args[2]
					assetConfig["project"] = args[2]
				} else {
					// Get project from flag or default
					project, _ := cmd.Flags().GetString("project") //nolint:errcheck // Flag is defined
					if project != "" {
						assetName = project
						assetConfig["project"] = project
					} else {
						assetName = AssetTypeDefault
						assetConfig["project"] = AssetTypeDefault
					}
				}
			default:
				return scanner.ScanConfig{}, "", fmt.Errorf("unknown hetzner resource type: %s (use 'project')", resourceType)
			}
		} else {
			// Default to project scan
			project, _ := cmd.Flags().GetString("project") //nolint:errcheck // Flag is defined
			if project != "" {
				assetName = project
				assetConfig["project"] = project
			} else {
				assetName = AssetTypeDefault
				assetConfig["project"] = AssetTypeDefault
			}
		}

		// Get API token from flag
		if apiToken, _ := cmd.Flags().GetString("hcloud-token"); apiToken != "" { //nolint:errcheck // Flag is defined
			assetConfig["api_token"] = apiToken
		}
		if apiToken, _ := cmd.Flags().GetString("api-token"); apiToken != "" { //nolint:errcheck // Flag is defined
			assetConfig["api_token"] = apiToken
		}

	case ProviderFactorial, "factorial-hr", "hris":
		providerName = ProviderFactorial
		assetType = "factorial-company"
		assetName = "default"

		// Get API key from flag
		if apiKey, _ := cmd.Flags().GetString("factorial-api-key"); apiKey != "" { //nolint:errcheck // Flag is defined
			assetConfig["api_key"] = apiKey
		}
		if apiKey, _ := cmd.Flags().GetString("api-key"); apiKey != "" { //nolint:errcheck // Flag is defined
			assetConfig["api_key"] = apiKey
		}

		// Get access token from flag
		if accessToken, _ := cmd.Flags().GetString("access-token"); accessToken != "" { //nolint:errcheck // Flag is defined
			assetConfig["access_token"] = accessToken
		}

	default:
		return scanner.ScanConfig{}, "", fmt.Errorf("unknown provider: %s", providerArg)
	}

	// Build provider config with credentials
	providerConfig := buildProviderConfig(cmd, providerName, assetConfig)

	// Build asset
	asset := core.Asset{
		Type:   assetType,
		Name:   assetName,
		Config: assetConfig,
	}
	if assetType == ProviderHost {
		asset.FQDN = assetName
	}

	return scanner.ScanConfig{
		ProviderName:   providerName,
		ProviderConfig: providerConfig,
		Asset:          asset,
	}, providerArg, nil
}

// buildProviderConfig builds provider-specific configuration from CLI flags
func buildProviderConfig(cmd *cobra.Command, providerName string, assetConfig map[string]string) map[string]string {
	token, _ := cmd.Flags().GetString("token")                    //nolint:errcheck // Flag is defined
	credentialType, _ := cmd.Flags().GetString("credential-type") //nolint:errcheck // Flag is defined
	envVar, _ := cmd.Flags().GetString("env-var")                 //nolint:errcheck // Flag is defined

	providerConfig := make(map[string]string)

	switch providerName {
	case ProviderGitHub:
		switch {
		case token != "":
			providerConfig["credential_type"] = "bearer"
			providerConfig["secret"] = token
		case credentialType != "":
			providerConfig["credential_type"] = credentialType
			if credentialType == CredentialTypeEnv {
				providerConfig["env_var"] = envVar
			}
		default:
			providerConfig["credential_type"] = CredentialTypeEnv
			providerConfig["env_var"] = "GITHUB_TOKEN"
		}

	case ProviderAWS:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get profile from flag
		if profile, _ := cmd.Flags().GetString("profile"); profile != "" { //nolint:errcheck // Flag is defined
			providerConfig["profile"] = profile
		}

		// Get region from flag
		if region, _ := cmd.Flags().GetString("region"); region != "" { //nolint:errcheck // Flag is defined
			providerConfig["region"] = region
		}

		// Get regions from flag
		if regions, _ := cmd.Flags().GetStringSlice("regions"); len(regions) > 0 { //nolint:errcheck // Flag is defined
			providerConfig["regions"] = strings.Join(regions, ",")
		}

		// Get credentials from flags
		if accessKeyID, _ := cmd.Flags().GetString("access-key-id"); accessKeyID != "" { //nolint:errcheck // Flag is defined
			providerConfig["access_key"] = accessKeyID
		}
		if secretAccessKey, _ := cmd.Flags().GetString("secret-access-key"); secretAccessKey != "" { //nolint:errcheck // Flag is defined
			providerConfig["secret_key"] = secretAccessKey
		}
		if sessionToken, _ := cmd.Flags().GetString("session-token"); sessionToken != "" { //nolint:errcheck // Flag is defined
			providerConfig["session_token"] = sessionToken
		}

		// Get assume role config
		if roleArn, _ := cmd.Flags().GetString("role-arn"); roleArn != "" { //nolint:errcheck // Flag is defined
			providerConfig["role_arn"] = roleArn
		}
		if externalID, _ := cmd.Flags().GetString("external-id"); externalID != "" { //nolint:errcheck // Flag is defined
			providerConfig["external_id"] = externalID
		}

	case ProviderAzure:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		switch {
		case token != "":
			providerConfig["credential_type"] = "bearer"
			providerConfig["secret"] = token
		case credentialType != "":
			providerConfig["credential_type"] = credentialType
			if credentialType == CredentialTypeEnv {
				providerConfig["env_var"] = "AZURE"
			}
		default:
			providerConfig["credential_type"] = CredentialTypeEnv
			providerConfig["env_var"] = "AZURE"
		}

	case ProviderMS365:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get client secret from flag
		clientSecret, _ := cmd.Flags().GetString("client-secret") //nolint:errcheck // Flag is defined
		if clientSecret != "" {
			providerConfig["client_secret"] = clientSecret
		}

		// Default credential type for MS365
		if credentialType != "" {
			providerConfig["credential_type"] = credentialType
		} else {
			providerConfig["credential_type"] = "client_credentials"
		}

	case ProviderCloudflare:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get API token from flag
		apiToken, _ := cmd.Flags().GetString("api-token") //nolint:errcheck // Flag is defined
		if apiToken != "" {
			providerConfig["api_token"] = apiToken
		}

		// Get API key and email from flags (legacy auth)
		apiKey, _ := cmd.Flags().GetString("api-key") //nolint:errcheck // Flag is defined
		if apiKey != "" {
			providerConfig["api_key"] = apiKey
		}
		email, _ := cmd.Flags().GetString("email") //nolint:errcheck // Flag is defined
		if email != "" {
			providerConfig["email"] = email
		}

		// Get account ID from flag
		accountID, _ := cmd.Flags().GetString("account-id") //nolint:errcheck // Flag is defined
		if accountID != "" {
			providerConfig["account_id"] = accountID
		}

	case ProviderAtlassian:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get API token from flag
		apiToken, _ := cmd.Flags().GetString("api-token") //nolint:errcheck // Flag is defined
		if apiToken != "" {
			providerConfig["api_token"] = apiToken
		}

		// Get email from flag
		email, _ := cmd.Flags().GetString("email") //nolint:errcheck // Flag is defined
		if email != "" {
			providerConfig["email"] = email
		}

		// Get site from flag
		site, _ := cmd.Flags().GetString("site") //nolint:errcheck // Flag is defined
		if site != "" {
			providerConfig["site"] = site
		}

		// Get org ID from flag
		orgID, _ := cmd.Flags().GetString("org-id") //nolint:errcheck // Flag is defined
		if orgID != "" {
			providerConfig["org_id"] = orgID
		}

	case ProviderSBOM:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get SBOM path from flag
		sbomPath, _ := cmd.Flags().GetString("sbom-path") //nolint:errcheck // Flag is defined
		if sbomPath != "" {
			providerConfig["sbom_path"] = sbomPath
		}

	case ProviderHetzner:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get API token from flag
		hcloudToken, _ := cmd.Flags().GetString("hcloud-token") //nolint:errcheck // Flag is defined
		if hcloudToken != "" {
			providerConfig["api_token"] = hcloudToken
		}
		apiToken, _ := cmd.Flags().GetString("api-token") //nolint:errcheck // Flag is defined
		if apiToken != "" {
			providerConfig["api_token"] = apiToken
		}

		// Get project from flag
		project, _ := cmd.Flags().GetString("project") //nolint:errcheck // Flag is defined
		if project != "" {
			providerConfig["project"] = project
		}

	case ProviderFactorial:
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get API key from flag
		factorialAPIKey, _ := cmd.Flags().GetString("factorial-api-key") //nolint:errcheck // Flag is defined
		if factorialAPIKey != "" {
			providerConfig["api_key"] = factorialAPIKey
		}
		apiKey, _ := cmd.Flags().GetString("api-key") //nolint:errcheck // Flag is defined
		if apiKey != "" {
			providerConfig["api_key"] = apiKey
		}

		// Get access token from flag
		accessToken, _ := cmd.Flags().GetString("access-token") //nolint:errcheck // Flag is defined
		if accessToken != "" {
			providerConfig["access_token"] = accessToken
		}
	}

	return providerConfig
}

// loadAndFilterPolicies loads policies and filters them for the given provider
func loadAndFilterPolicies(cmd *cobra.Command, providerName, providerAlias string) ([]core.Policy, error) {
	policyFile, _ := cmd.Flags().GetString("policy")    //nolint:errcheck // Flag is defined
	policyDir, _ := cmd.Flags().GetString("policy-dir") //nolint:errcheck // Flag is defined

	if policyFile == "" && policyDir == "" {
		return nil, fmt.Errorf("policy file (-f) or directory (-d) is required")
	}

	// Load policies
	policies, err := scanner.LoadPolicies(policyFile, policyDir)
	if err != nil {
		return nil, err
	}

	// Filter policies for this provider
	policies = scanner.FilterPoliciesByProvider(policies, providerName, providerAlias)

	if len(policies) == 0 {
		return nil, fmt.Errorf("no policies found for provider %s", providerName)
	}

	return policies, nil
}

// exportResults exports scan results to the specified file
func exportResults(cmd *cobra.Command, tree *common.ResourceTree, provider, exportPath string) error {
	// Convert tree to report
	rep := report.FromResourceTree(tree, provider)

	// Determine format from flag or filename extension
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

	// Export based on format
	switch report.Format(format) {
	case report.FormatCSV:
		exporter := report.NewCSVExporter()
		return exporter.Export(rep, exportPath)
	case report.FormatXLSX:
		exporter := report.NewXLSXExporter()
		return exporter.Export(rep, exportPath)
	case report.FormatJSON:
		exporter := report.NewJSONExporter(true)
		return exporter.Export(rep, exportPath)
	default:
		return fmt.Errorf("unsupported export format: %s (use csv, xlsx, or json)", format)
	}
}
