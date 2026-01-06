package cmd

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kopexa-grc/kspec/cli"
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/scanner"
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
  kspec scan azure subscription <sub-id> -f policy.yml
  kspec scan ms365 tenant <tenant-id> --client-id <id> --client-secret <secret> -f policy.yml
  kspec scan cloudflare account -f policy.yml
  kspec scan cloudflare zone <zone-id> --api-token <token> -f policy.yml
  kspec scan atlassian site yoursite.atlassian.net -f policy.yml
  kspec scan atlassian org <org-id> --site yoursite.atlassian.net -f policy.yml
  kspec scan sbom file ./sbom.json -f policy.yml
  kspec scan sbom dir ./sboms/ -f policy.yml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runScan,
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

	// Initialize bubbletea UI
	uiModel := cli.InitialModel()
	p := tea.NewProgram(uiModel)

	// Set up event handler to forward events to UI
	s.OnEvent(func(event scanner.ScanEvent) {
		switch event.Type {
		case scanner.EventTreeCreated:
			p.Send(cli.SetTreeMsg{Tree: event.Tree})
		case scanner.EventTreeUpdated,
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
		_, _ = s.Run(ctx)
	}()

	// Run the UI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("UI error: %w", err)
	}

	return nil
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

	case "host":
		if len(args) < 2 {
			return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan host <target> -f policy.yml")
		}
		providerName = "network"
		assetType = "host"
		assetName = args[1]
		assetConfig["target"] = assetName

	case "github":
		if len(args) < 3 {
			return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan github <org|repo> <target> -f policy.yml")
		}

		providerName = "github"
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

	case "azure":
		if len(args) < 3 {
			return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan azure subscription <subscription-id> -f policy.yml")
		}

		providerName = "azure"
		resourceType := args[1]
		target := args[2]

		switch resourceType {
		case "subscription", "sub":
			assetType = "azure"
			assetName = target
			assetConfig["subscription_id"] = target

			// Optional: client_id, tenant_id from flags
			if clientID, _ := cmd.Flags().GetString("client-id"); clientID != "" {
				assetConfig["client_id"] = clientID
			}
			if tenantID, _ := cmd.Flags().GetString("tenant-id"); tenantID != "" {
				assetConfig["tenant_id"] = tenantID
			}
			if resourceGroup, _ := cmd.Flags().GetString("resource-group"); resourceGroup != "" {
				assetConfig["resource_group"] = resourceGroup
			}

		default:
			return scanner.ScanConfig{}, "", fmt.Errorf("unknown azure resource type: %s (use 'subscription')", resourceType)
		}

	case "ms365", "m365", "microsoft365":
		providerName = "ms365"
		assetType = "ms365"

		// Check if tenant ID is provided
		if len(args) >= 3 && args[1] == "tenant" {
			assetName = args[2]
			assetConfig["tenant_id"] = args[2]
		} else if len(args) >= 2 {
			// Just the tenant ID directly
			assetName = args[1]
			assetConfig["tenant_id"] = args[1]
		} else {
			assetName = "default"
		}

		// Optional: client_id, client_secret from flags
		if clientID, _ := cmd.Flags().GetString("client-id"); clientID != "" {
			assetConfig["client_id"] = clientID
		}
		if tenantID, _ := cmd.Flags().GetString("tenant-id"); tenantID != "" {
			assetConfig["tenant_id"] = tenantID
			assetName = tenantID
		}
		if clientSecret, _ := cmd.Flags().GetString("client-secret"); clientSecret != "" {
			assetConfig["client_secret"] = clientSecret
		}

	case "cloudflare", "cf":
		providerName = "cloudflare"

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
		if apiToken, _ := cmd.Flags().GetString("api-token"); apiToken != "" {
			assetConfig["api_token"] = apiToken
		}
		if apiKey, _ := cmd.Flags().GetString("api-key"); apiKey != "" {
			assetConfig["api_key"] = apiKey
		}
		if email, _ := cmd.Flags().GetString("email"); email != "" {
			assetConfig["email"] = email
		}
		if accountID, _ := cmd.Flags().GetString("account-id"); accountID != "" {
			assetConfig["account_id"] = accountID
		}

	case "atlassian", "jira", "confluence":
		providerName = "atlassian"

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
					site, _ := cmd.Flags().GetString("site")
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
			site, _ := cmd.Flags().GetString("site")
			if site != "" {
				assetName = site
				assetConfig["site"] = site
			} else {
				return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan atlassian site <site-name> -f policy.yml")
			}
		}

		// Get credentials from flags
		if apiToken, _ := cmd.Flags().GetString("api-token"); apiToken != "" {
			assetConfig["api_token"] = apiToken
		}
		if email, _ := cmd.Flags().GetString("email"); email != "" {
			assetConfig["email"] = email
		}
		if site, _ := cmd.Flags().GetString("site"); site != "" && assetConfig["site"] == "" {
			assetConfig["site"] = site
		}
		if orgID, _ := cmd.Flags().GetString("org-id"); orgID != "" {
			assetConfig["org_id"] = orgID
		}

	case "sbom", "bom", "cyclonedx", "spdx":
		providerName = "sbom"

		// Check resource type
		if len(args) >= 2 {
			resourceType := args[1]
			switch resourceType {
			case "file":
				if len(args) < 3 {
					return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan sbom file <path> -f policy.yml")
				}
				assetType = "sbom-file"
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
				assetType = "sbom-file"
				assetName = resourceType
				assetConfig["sbom_path"] = resourceType
			}
		} else {
			// Get path from flag
			sbomPath, _ := cmd.Flags().GetString("sbom-path")
			if sbomPath != "" {
				assetType = "sbom-file"
				assetName = sbomPath
				assetConfig["sbom_path"] = sbomPath
			} else {
				return scanner.ScanConfig{}, "", fmt.Errorf("usage: kspec scan sbom file <path> -f policy.yml")
			}
		}

		// Get path from flag if not set
		if sbomPath, _ := cmd.Flags().GetString("sbom-path"); sbomPath != "" && assetConfig["sbom_path"] == "" {
			assetConfig["sbom_path"] = sbomPath
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
	if assetType == "host" {
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
	token, _ := cmd.Flags().GetString("token")
	credentialType, _ := cmd.Flags().GetString("credential-type")
	envVar, _ := cmd.Flags().GetString("env-var")

	providerConfig := make(map[string]string)

	switch providerName {
	case "github":
		if token != "" {
			providerConfig["credential_type"] = "bearer"
			providerConfig["secret"] = token
		} else if credentialType != "" {
			providerConfig["credential_type"] = credentialType
			if credentialType == "env" {
				providerConfig["env_var"] = envVar
			}
		} else {
			providerConfig["credential_type"] = "env"
			providerConfig["env_var"] = "GITHUB_TOKEN"
		}

	case "azure":
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		if token != "" {
			providerConfig["credential_type"] = "bearer"
			providerConfig["secret"] = token
		} else if credentialType != "" {
			providerConfig["credential_type"] = credentialType
			if credentialType == "env" {
				providerConfig["env_var"] = "AZURE"
			}
		} else {
			providerConfig["credential_type"] = "env"
			providerConfig["env_var"] = "AZURE"
		}

	case "ms365":
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get client secret from flag
		clientSecret, _ := cmd.Flags().GetString("client-secret")
		if clientSecret != "" {
			providerConfig["client_secret"] = clientSecret
		}

		// Default credential type for MS365
		if credentialType != "" {
			providerConfig["credential_type"] = credentialType
		} else {
			providerConfig["credential_type"] = "client_credentials"
		}

	case "cloudflare":
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get API token from flag
		apiToken, _ := cmd.Flags().GetString("api-token")
		if apiToken != "" {
			providerConfig["api_token"] = apiToken
		}

		// Get API key and email from flags (legacy auth)
		apiKey, _ := cmd.Flags().GetString("api-key")
		if apiKey != "" {
			providerConfig["api_key"] = apiKey
		}
		email, _ := cmd.Flags().GetString("email")
		if email != "" {
			providerConfig["email"] = email
		}

		// Get account ID from flag
		accountID, _ := cmd.Flags().GetString("account-id")
		if accountID != "" {
			providerConfig["account_id"] = accountID
		}

	case "atlassian":
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get API token from flag
		apiToken, _ := cmd.Flags().GetString("api-token")
		if apiToken != "" {
			providerConfig["api_token"] = apiToken
		}

		// Get email from flag
		email, _ := cmd.Flags().GetString("email")
		if email != "" {
			providerConfig["email"] = email
		}

		// Get site from flag
		site, _ := cmd.Flags().GetString("site")
		if site != "" {
			providerConfig["site"] = site
		}

		// Get org ID from flag
		orgID, _ := cmd.Flags().GetString("org-id")
		if orgID != "" {
			providerConfig["org_id"] = orgID
		}

	case "sbom":
		// Copy asset config to provider config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Get SBOM path from flag
		sbomPath, _ := cmd.Flags().GetString("sbom-path")
		if sbomPath != "" {
			providerConfig["sbom_path"] = sbomPath
		}
	}

	return providerConfig
}

// loadAndFilterPolicies loads policies and filters them for the given provider
func loadAndFilterPolicies(cmd *cobra.Command, providerName, providerAlias string) ([]core.Policy, error) {
	policyFile, _ := cmd.Flags().GetString("policy")
	policyDir, _ := cmd.Flags().GetString("policy-dir")

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
