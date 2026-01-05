package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/juliankoehn/kspec/cli"
	"github.com/juliankoehn/kspec/cli/components/common"
	"github.com/juliankoehn/kspec/core"
	"github.com/juliankoehn/kspec/provider"
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
  kspec scan azure subscription <sub-id> -f policy.yml`,
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
	scanCmd.Flags().String("client-id", "", "Azure Client ID (Service Principal)")
	scanCmd.Flags().String("tenant-id", "", "Azure Tenant ID")
	scanCmd.Flags().String("resource-group", "", "Azure Resource Group (optional filter)")
}

func runScan(cmd *cobra.Command, args []string) error {
	// Get common flags
	policyFile, _ := cmd.Flags().GetString("policy")
	policyDir, _ := cmd.Flags().GetString("policy-dir")
	token, _ := cmd.Flags().GetString("token")
	credentialType, _ := cmd.Flags().GetString("credential-type")
	envVar, _ := cmd.Flags().GetString("env-var")

	// Parse provider and determine asset type
	providerArg := args[0]

	var assetType string
	var assetName string
	var assetConfig = make(map[string]string)
	var providerName string

	// Handle different provider patterns
	switch providerArg {
	case "local":
		providerName = "os"
		assetType = "local"
		assetName = "localhost"

	case "host":
		if len(args) < 2 {
			return fmt.Errorf("usage: kspec scan host <target> -f policy.yml")
		}
		providerName = "network"
		assetType = "host"
		assetName = args[1]
		assetConfig["target"] = assetName

	case "github":
		if len(args) < 3 {
			return fmt.Errorf("usage: kspec scan github <org|repo> <target> -f policy.yml")
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
				return fmt.Errorf("repository must be in format: owner/repo")
			}
			assetType = "github-repo"
			assetName = target
			assetConfig["owner"] = parts[0]
			assetConfig["repo"] = parts[1]

		default:
			return fmt.Errorf("unknown github resource type: %s (use 'org' or 'repo')", resourceType)
		}

	case "azure":
		if len(args) < 3 {
			return fmt.Errorf("usage: kspec scan azure subscription <subscription-id> -f policy.yml")
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
			return fmt.Errorf("unknown azure resource type: %s (use 'subscription')", resourceType)
		}

	default:
		return fmt.Errorf("unknown provider: %s", providerArg)
	}

	// Validate policy flags
	if policyFile == "" && policyDir == "" {
		return fmt.Errorf("policy file (-f) or directory (-d) is required")
	}

	// Load policies
	var policies []core.Policy
	loadPolicy := func(path string) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var p core.Policy
		if err := yaml.Unmarshal(data, &p); err != nil {
			return err
		}
		policies = append(policies, p)
		return nil
	}

	if policyDir != "" {
		files, err := os.ReadDir(policyDir)
		if err != nil {
			return fmt.Errorf("failed to read policy directory: %w", err)
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
				path := filepath.Join(policyDir, file.Name())
				if err := loadPolicy(path); err != nil {
					log.Printf("Warning: failed to load policy %s: %v", file.Name(), err)
				}
			}
		}
	} else if policyFile != "" {
		if err := loadPolicy(policyFile); err != nil {
			return fmt.Errorf("failed to read policy file: %w", err)
		}
	}

	// Filter policies based on their 'require' field
	// Only keep policies that match the current provider or have no requirements
	var filteredPolicies []core.Policy
	for _, policy := range policies {
		if len(policy.Require) == 0 {
			// No requirements - policy applies to all providers
			filteredPolicies = append(filteredPolicies, policy)
			continue
		}

		// Check if any requirement matches the current provider
		for _, req := range policy.Require {
			if req.Provider == providerName || req.Provider == providerArg {
				filteredPolicies = append(filteredPolicies, policy)
				break
			}
		}
	}
	policies = filteredPolicies

	if len(policies) == 0 {
		return fmt.Errorf("no policies found for provider %s", providerName)
	}

	// Setup provider config with credentials
	providerConfig := make(map[string]string)

	// Handle GitHub-specific credentials
	if providerName == "github" {
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
	}

	// Handle Azure-specific credentials
	if providerName == "azure" {
		// Copy subscription_id and other config
		for k, v := range assetConfig {
			providerConfig[k] = v
		}

		// Add credentials
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
	}

	// Initialize only the needed provider (lazy-loading)
	registry, err := provider.InitProvider(context.Background(), providerName, providerConfig)
	if err != nil {
		return fmt.Errorf("failed to init provider %s: %w", providerName, err)
	}

	// Create asset
	asset := core.Asset{
		Type:   assetType,
		Name:   assetName,
		Config: assetConfig,
	}
	if assetType == "host" {
		asset.FQDN = assetName
	}

	ctx := context.Background()

	// Initialize bubbletea UI
	uiModel := cli.InitialModel()
	p := tea.NewProgram(uiModel)

	// Start discovery and scan in background
	go func() {
		// Create resource tree with root node
		tree := common.NewResourceTree(assetName, assetType)
		tree.Root.State = common.AssetStateDiscovery

		// Send initial tree to UI
		p.Send(cli.SetTreeMsg{Tree: tree})

		// Track discovered resources for discovery-capable providers
		discoveredResources := make(map[string]int)
		hasDiscovery := false

		// Phase 1: Discovery (if supported)
		for _, resSpec := range registry {
			if discoverer, ok := resSpec.(core.DiscoveryResource); ok {
				hasDiscovery = true
				discovered, err := discoverer.Discover(ctx, asset)
				if err != nil {
					continue
				}

				// Store discovered resources
				for resourceType, count := range discovered {
					discoveredResources[resourceType] = count
				}
			}
		}

		// Create nodes in a sensible order based on provider
		if hasDiscovery {
			// Define preferred resource order per provider
			// Note: github_branch is a sub-resource of github_repo, not a top-level type
			var orderedTypes []string
			switch providerName {
			case "github":
				orderedTypes = []string{"github_organization", "github_team", "github_repo"}
			case "azure":
				orderedTypes = []string{"azure_subscription", "azure_storage_account", "azure_sql_server", "azure_sql_database",
					"azure_mysql_server", "azure_postgresql_server", "azure_keyvault_vault",
					"azure_network_security_group", "azure_virtual_machine", "azure_app_service"}
			default:
				// For unknown providers, use discovered order
				for resourceType := range discoveredResources {
					orderedTypes = append(orderedTypes, resourceType)
				}
			}

			// Create nodes in order
			for _, resourceType := range orderedTypes {
				count, exists := discoveredResources[resourceType]
				if !exists || count == 0 {
					continue
				}

				resourceNode := &common.ResourceNode{
					ID:                fmt.Sprintf("%s-%s", tree.Root.ID, resourceType),
					Name:              resourceType,
					Type:              "resource_type",
					ResourceType:      resourceType,
					ResourceCount:     count,
					State:             common.AssetStatePending,
					Children:          []*common.ResourceNode{},
					SubResources:      make(map[string][]*common.ResourceNode),
					SubResourceCounts: make(map[string]int),
					Metadata:          make(map[string]interface{}),
				}

				// Add to tree
				if err := tree.AddNode(tree.Root.ID, resourceNode); err != nil {
					log.Printf("Error adding resource node: %v", err)
					continue
				}

				// Update root's total resource count
				tree.Root.ResourceCount += count

				// Check if this resource has sub-resources
				if resSpec, ok := registry[resourceType]; ok {
					if subProvider, ok := resSpec.(core.SubResourceProvider); ok {
						subSpecs := subProvider.SubResources()
						for _, subSpec := range subSpecs {
							if subDiscoverer, ok := subSpec.(core.DiscoveryResource); ok {
								subDiscovered, err := subDiscoverer.Discover(ctx, asset)
								if err != nil {
									continue
								}

								for subResType, subCount := range subDiscovered {
									if subCount > 0 {
										resourceNode.SubResourceCounts[subResType] = subCount
									}
								}
							}
						}
					}
				}

				// Send updated tree to UI after each resource type
				p.Send(cli.UpdateTreeMsg{Tree: tree})
			}
		}

		// For non-discovery providers, create nodes based on registry resources
		// Only create nodes for resource types that exist in the current provider's registry
		if !hasDiscovery {
			for _, resSpec := range registry {
				resourceType := resSpec.Name()
				nodeID := fmt.Sprintf("%s-%s", tree.Root.ID, resourceType)

				// Check if we already have a node for this resource type
				if _, exists := tree.GetNode(nodeID); !exists {
					resourceNode := &common.ResourceNode{
						ID:                nodeID,
						Name:              resourceType,
						Type:              "resource_type",
						ResourceType:      resourceType,
						ResourceCount:     1, // Will be updated during fetch
						State:             common.AssetStatePending,
						Children:          []*common.ResourceNode{},
						SubResources:      make(map[string][]*common.ResourceNode),
						SubResourceCounts: make(map[string]int),
						Metadata:          make(map[string]interface{}),
					}

					if err := tree.AddNode(tree.Root.ID, resourceNode); err != nil {
						log.Printf("Error adding resource node: %v", err)
					}
				}
			}
			p.Send(cli.UpdateTreeMsg{Tree: tree})
		}

		// Mark root as scanning
		tree.Root.State = common.AssetStateScanning
		p.Send(cli.UpdateTreeMsg{Tree: tree})

		// Phase 2: Fetch and scan resources
		// Process resources in a deterministic order based on discovered resources
		var resourceOrder []string
		for _, child := range tree.Root.Children {
			resourceOrder = append(resourceOrder, child.ResourceType)
		}

		for _, resourceType := range resourceOrder {
			resSpec, ok := registry[resourceType]
			if !ok {
				continue
			}
			nodeID := fmt.Sprintf("%s-%s", tree.Root.ID, resourceType)

			resourceNode, exists := tree.GetNode(nodeID)
			if !exists {
				// For non-discovery mode, only process resources that have nodes
				continue
			}

			// Discovery-based skip: if discovered with 0 count, skip
			if count, discovered := discoveredResources[resourceType]; discovered && count == 0 {
				resourceNode.State = common.AssetStateComplete
				p.Send(cli.UpdateTreeMsg{Tree: tree})
				continue
			}

			// Mark as scanning
			resourceNode.State = common.AssetStateScanning
			p.Send(cli.UpdateTreeMsg{Tree: tree})

			// Fetch resources
			resources, err := resSpec.Fetch(ctx, asset)
			if err != nil {
				resourceNode.State = common.AssetStateError
				resourceNode.Error = err
				p.Send(cli.UpdateTreeMsg{Tree: tree})
				continue
			}

			// Update actual resource count
			resourceNode.ResourceCount = len(resources)

			// Create instance nodes for each resource
			for i, resource := range resources {
				// Get resource name/ID for display
				resourceName := fmt.Sprintf("%s-%d", resourceType, i)
				if name, ok := resource["name"]; ok {
					resourceName = fmt.Sprintf("%v", name)
				} else if id, ok := resource["id"]; ok {
					resourceName = fmt.Sprintf("%v", id)
				} else if login, ok := resource["login"]; ok {
					resourceName = fmt.Sprintf("%v", login)
				} else if fullName, ok := resource["full_name"]; ok {
					resourceName = fmt.Sprintf("%v", fullName)
				}

				instanceID := fmt.Sprintf("%s-%d", nodeID, i)
				instanceNode := &common.ResourceNode{
					ID:           instanceID,
					Name:         resourceName,
					Type:         "resource_instance",
					ResourceType: resourceType,
					State:        common.AssetStateScanning,
					Checks:       []common.CheckResult{},
					Metadata:     resource,
				}

				// Add to tree
				if err := tree.AddNode(nodeID, instanceNode); err != nil {
					log.Printf("Error adding instance node: %v", err)
					continue
				}

				// Run policy checks for this resource
				checkResults := runPolicyChecksForResource(ctx, resource, resourceType, policies, registry, asset)

				// Add check results to instance node
				for _, checkResult := range checkResults {
					instanceNode.Checks = append(instanceNode.Checks, checkResult)

					// Update counts
					switch checkResult.Status {
					case "passed":
						instanceNode.ChecksPassed++
					case "failed":
						instanceNode.ChecksFailed++
					case "skipped":
						instanceNode.ChecksSkipped++
					}
				}

				instanceNode.State = common.AssetStateComplete

				// Update parent counts
				resourceNode.ChecksPassed += instanceNode.ChecksPassed
				resourceNode.ChecksFailed += instanceNode.ChecksFailed
				resourceNode.ChecksSkipped += instanceNode.ChecksSkipped

				// Update root counts
				tree.Root.ChecksPassed += instanceNode.ChecksPassed
				tree.Root.ChecksFailed += instanceNode.ChecksFailed
				tree.Root.ChecksSkipped += instanceNode.ChecksSkipped

				// Fetch sub-resources for this specific instance (e.g., branches for a repo)
				if subProvider, ok := resSpec.(core.SubResourceProvider); ok {
					subSpecs := subProvider.SubResources()
					for _, subSpec := range subSpecs {
						subResType := subSpec.Name()

						// Create a modified asset with instance-specific config
						instanceAsset := core.Asset{
							Type:   asset.Type,
							Name:   asset.Name,
							Config: make(map[string]string),
						}
						for k, v := range asset.Config {
							instanceAsset.Config[k] = v
						}

						// Add instance-specific config for sub-resource fetching
						// For repos, we need to specify which repo to get branches for
						if resourceType == "github_repo" {
							if repoName, ok := resource["name"]; ok {
								instanceAsset.Config["repo"] = fmt.Sprintf("%v", repoName)
							}
						}

						// Fetch sub-resources for this specific instance
						subResources, err := subSpec.Fetch(ctx, instanceAsset)
						if err != nil {
							continue
						}

						// Create sub-resource instance nodes as children of this instance
						for j, subResource := range subResources {
							subResName := fmt.Sprintf("%s-%d", subResType, j)
							if name, ok := subResource["name"]; ok {
								subResName = fmt.Sprintf("%v", name)
							}

							subInstanceID := fmt.Sprintf("%s-sub-%s-%d", instanceID, subResType, j)

							// Run checks for sub-resource
							subCheckResults := runPolicyChecksForResource(ctx, subResource, subResType, policies, registry, asset)

							subInstanceNode := &common.ResourceNode{
								ID:           subInstanceID,
								Name:         subResName,
								Type:         "sub_resource_instance",
								ResourceType: subResType,
								State:        common.AssetStateComplete,
								Checks:       subCheckResults,
								Metadata:     subResource,
							}

							// Count check results
							for _, cr := range subCheckResults {
								switch cr.Status {
								case "passed":
									subInstanceNode.ChecksPassed++
								case "failed":
									subInstanceNode.ChecksFailed++
								case "skipped":
									subInstanceNode.ChecksSkipped++
								}
							}

							// Add to parent instance's sub-resources
							if err := tree.AddSubResource(instanceID, subResType, subInstanceNode); err != nil {
								log.Printf("Error adding sub-resource: %v", err)
								continue
							}

							// Update counts up the tree
							instanceNode.ChecksPassed += subInstanceNode.ChecksPassed
							instanceNode.ChecksFailed += subInstanceNode.ChecksFailed
							instanceNode.ChecksSkipped += subInstanceNode.ChecksSkipped
							resourceNode.ChecksPassed += subInstanceNode.ChecksPassed
							resourceNode.ChecksFailed += subInstanceNode.ChecksFailed
							resourceNode.ChecksSkipped += subInstanceNode.ChecksSkipped
							tree.Root.ChecksPassed += subInstanceNode.ChecksPassed
							tree.Root.ChecksFailed += subInstanceNode.ChecksFailed
							tree.Root.ChecksSkipped += subInstanceNode.ChecksSkipped
						}
					}
				}
			}

			// Mark resource type as complete
			resourceNode.State = common.AssetStateComplete
			p.Send(cli.UpdateTreeMsg{Tree: tree})
		}

		// Mark root as complete
		tree.Root.State = common.AssetStateComplete
		p.Send(cli.UpdateTreeMsg{Tree: tree})
	}()

	// Run the UI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("UI error: %w", err)
	}

	return nil
}

// evaluateGroupFilter performs a simple evaluation of group filters
// This handles common patterns like: asset.type == "github-org"
func evaluateGroupFilter(filter string, asset core.Asset) bool {
	// Handle common asset.type filters
	filter = strings.TrimSpace(filter)

	// Pattern: asset.type == "value" or asset.type == 'value'
	if strings.Contains(filter, "asset.type") {
		// Extract the expected value
		for _, sep := range []string{`== "`, `== '`, `=="`, `=='`} {
			if idx := strings.Index(filter, sep); idx != -1 {
				start := idx + len(sep)
				end := strings.IndexAny(filter[start:], `"'`)
				if end != -1 {
					expectedType := filter[start : start+end]
					return asset.Type == expectedType
				}
			}
		}
	}

	// If we can't parse the filter, default to true (include the group)
	return true
}

// runPolicyChecksForResource executes policy checks for a given resource
func runPolicyChecksForResource(
	ctx context.Context,
	resource core.Resource,
	resourceType string,
	policies []core.Policy,
	registry map[string]core.ResourceSpec,
	asset core.Asset,
) []common.CheckResult {
	var results []common.CheckResult

	// Create evaluator
	evaluator, err := core.NewEvaluator(registry)
	if err != nil {
		return results
	}

	// Build definitions index from all policies
	definitions := make(map[string]core.Check)
	for _, policy := range policies {
		for _, q := range policy.Queries {
			if q.UID != "" {
				definitions[q.UID] = q
			}
			if q.ID != "" {
				definitions[q.ID] = q
			}
		}
	}

	// Process each policy
	for _, policy := range policies {
		// Process groups to get checks that apply to this asset type
		for _, group := range policy.Groups {
			// Evaluate group filter
			if group.Filter != "" {
				if !evaluateGroupFilter(group.Filter, asset) {
					continue
				}
			}

			// Process checks in this group
			for _, checkRef := range group.Checks {
				check := checkRef

				// Resolve check definition if needed
				if check.Query == "" {
					var def core.Check
					var found bool

					if check.UID != "" {
						def, found = definitions[check.UID]
					}
					if !found && check.ID != "" {
						def, found = definitions[check.ID]
					}

					if !found {
						continue
					}

					// Merge definition into check
					check.Query = def.Query
					check.Resource = def.Resource
					check.Remediation = def.Remediation
					check.Severity = def.Severity
					check.Docs = def.Docs
					check.Audit = def.Audit
					if check.Title == "" {
						check.Title = def.Title
					}

					// Merge config
					config := make(map[string]string)
					for k, v := range def.Config {
						config[k] = v
					}
					for k, v := range checkRef.Config {
						config[k] = v
					}
					check.Config = config

					// Merge props
					if len(def.Props) > 0 && len(check.Props) == 0 {
						check.Props = def.Props
					}
				}

				// Check if this check applies to this resource type
				if check.Resource != resourceType {
					continue
				}

				// Build props map
				propsMap := make(map[string]interface{})
				for _, prop := range check.Props {
					propsMap[prop.UID] = prop.MQL
				}

				// Merge check config with asset config
				evalConfig := make(map[string]string)
				for k, v := range asset.Config {
					evalConfig[k] = v
				}
				for k, v := range check.Config {
					evalConfig[k] = v
				}

				// Evaluate the query
				passed, err := evaluator.Evaluate(
					check.Query,
					resource,
					evalConfig,
					propsMap,
					asset,
				)

				status := "skipped"
				var details string

				if err != nil {
					// Check if it's a compilation error (MQL syntax not supported by CEL)
					errStr := err.Error()
					if strings.Contains(errStr, "compile error") ||
						strings.Contains(errStr, "Syntax error") ||
						strings.Contains(errStr, "undeclared reference") {
						status = "skipped"
						details = "Query uses syntax not supported by CEL evaluator"
					} else {
						status = "failed"
						details = fmt.Sprintf("Evaluation error: %s", errStr)
					}
				} else if passed {
					status = "passed"
					details = ""
				} else {
					status = "failed"
					details = check.Remediation
				}

				// Use severity from check (default to medium if not specified)
				severity := check.Severity
				if severity == "" {
					severity = "medium"
				}

				// Get check ID
				checkID := check.UID
				if checkID == "" {
					checkID = check.ID
				}

				results = append(results, common.CheckResult{
					ID:       checkID,
					Group:    group.Title,
					Name:     check.Title,
					Status:   status,
					Severity: severity,
					Details:  details,
					Docs:     check.Docs,
					Audit:    check.Audit,
				})
			}
		}
	}

	return results
}
