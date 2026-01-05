package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/juliankoehn/kspec/core"
	"github.com/juliankoehn/kspec/provider"
	"github.com/juliankoehn/kspec/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan <provider> <resource-type> <target>",
	Short: "Scan resources using policy checks",
	Long: `Scan resources across different providers using policy-as-code checks.
	
The scan command dynamically discovers available providers and their resource types.

Examples:
  kspec scan github org kopexa-grc -f policy.yml
  kspec scan github repo owner/repo -f policy.yml
  kspec scan host example.com -f policy.yml
  kspec scan local -f policy.yml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	// Parse provider and determine asset type
	providerName := args[0]
	
	var assetType string
	var assetName string
	var assetConfig map[string]string = make(map[string]string)
	
	// Handle different provider patterns
	switch providerName {
	case "local":
		assetType = "local"
		assetName = "localhost"
		
	case "host":
		if len(args) < 2 {
			return fmt.Errorf("usage: kspec scan host <target> -f policy.yml")
		}
		assetType = "host"
		assetName = args[1]
		assetConfig["target"] = assetName
		
	case "github":
		if len(args) < 3 {
			return fmt.Errorf("usage: kspec scan github <org|repo> <target> -f policy.yml")
		}
		
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
		
	default:
		return fmt.Errorf("unknown provider: %s", providerName)
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
	
	// Setup provider config with credentials
	providerConfig := make(map[string]string)
	if assetType == "github-org" || assetType == "github-repo" {
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
	
	// Initialize providers
	registry, err := provider.InitProviders(context.Background(), providerConfig)
	if err != nil {
		return fmt.Errorf("failed to init providers: %w", err)
	}
	
	// Initialize evaluator
	eval, err := core.NewEvaluator(registry)
	if err != nil {
		return fmt.Errorf("failed to init evaluator: %w", err)
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
	
	// Execute scan (reusing existing logic)
	return executeScan(ctx, policies, registry, eval, asset)
}

func executeScan(ctx context.Context, policies []core.Policy, registry map[string]core.ResourceSpec, eval *core.Evaluator, asset core.Asset) error {
	// Index Definitions from ALL policies
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
	
	type ResolvedCheck struct {
		GroupTitle string
		CheckDef   core.Check
	}
	var executionPlan []ResolvedCheck
	var uiItems []ui.CheckItem
	
	// Collect checks from ALL policies
	for _, policy := range policies {
		for _, group := range policy.Groups {
			// Group filter evaluation (simplified for now)
			if group.Filter != "" {
				pass, err := eval.Evaluate(group.Filter, make(map[string]interface{}), nil, nil, asset)
				if err != nil || !pass {
					continue
				}
			}
			
			for _, checkRef := range group.Checks {
				check := checkRef
				if check.ID == "" && check.UID != "" {
					check.ID = check.UID
				}
				if check.Query == "" {
					if def, ok := definitions[check.UID]; ok {
						check.Query = def.Query
						check.Resource = def.Resource
						check.Remediation = def.Remediation
						check.Severity = def.Severity
						if check.Title == "" {
							check.Title = def.Title
						}
						config := make(map[string]string)
						for k, v := range def.Config {
							config[k] = v
						}
						for k, v := range checkRef.Config {
							config[k] = v
						}
						check.Config = config
					} else if def, ok := definitions[check.ID]; ok {
						check.Query = def.Query
						check.Resource = def.Resource
						check.Remediation = def.Remediation
						check.Severity = def.Severity
						if check.Title == "" {
							check.Title = def.Title
						}
						config := make(map[string]string)
						for k, v := range def.Config {
							config[k] = v
						}
						for k, v := range checkRef.Config {
							config[k] = v
						}
						check.Config = config
					}
				}
				
				executionPlan = append(executionPlan, ResolvedCheck{
					GroupTitle: group.Title,
					CheckDef:   check,
				})
				
				uiItems = append(uiItems, ui.CheckItem{
					ID:     check.ID,
					Group:  group.Title,
					Name:   check.Title,
					Status: ui.StatusPending,
				})
			}
		}
	}
	
	// Init UI
	model := ui.NewModel(uiItems)
	p := tea.NewProgram(model)
	
	// Run execution in background
	go func() {
		for _, item := range executionPlan {
			check := item.CheckDef
			
			p.Send(ui.CheckResultMsg{
				ID:     check.ID,
				Status: ui.StatusRunning,
			})
			
			checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			
			resSpec, ok := registry[check.Resource]
			if !ok {
				p.Send(ui.CheckResultMsg{
					ID:       check.ID,
					Status:   ui.StatusFailed,
					ErrorMsg: "Unknown resource type: " + check.Resource,
				})
				cancel()
				continue
			}
			
			fetchAsset := asset
			fetchAsset.Config = make(map[string]string)
			for k, v := range asset.Config {
				fetchAsset.Config[k] = v
			}
			for k, v := range check.Config {
				fetchAsset.Config[k] = v
			}
			
			resources, err := resSpec.Fetch(checkCtx, fetchAsset)
			if err != nil {
				p.Send(ui.CheckResultMsg{
					ID:       check.ID,
					Status:   ui.StatusFailed,
					ErrorMsg: fmt.Sprintf("Fetch failed: %v", err),
				})
				cancel()
				continue
			}
			
			if len(resources) == 0 {
				p.Send(ui.CheckResultMsg{
					ID:       check.ID,
					Status:   ui.StatusSkipped,
					ErrorMsg: "No resources found",
				})
				cancel()
				continue
			}
			
			props := make(map[string]interface{})
			for _, prop := range check.Props {
				props[prop.UID] = prop.MQL
			}
			
			finalStatus := ui.StatusPassed
			var errorMsgs []string
			
			for _, res := range resources {
				pass, err := eval.Evaluate(check.Query, res, check.Config, props, asset)
				if err != nil {
					finalStatus = ui.StatusFailed
					errorMsgs = append(errorMsgs, err.Error())
				} else if !pass {
					finalStatus = ui.StatusFailed
				}
			}
			
			errStr := ""
			if len(errorMsgs) > 0 {
				errStr = strings.Join(errorMsgs, "; ")
			} else if finalStatus == ui.StatusFailed {
				errStr = "Check failed"
			}
			
			p.Send(ui.CheckResultMsg{
				ID:       check.ID,
				Status:   finalStatus,
				ErrorMsg: errStr,
			})
			
			cancel()
		}
	}()
	
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("UI error: %w", err)
	}
	
	return nil
}
