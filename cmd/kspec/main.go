package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/juliankoehn/kspec/core"
	"github.com/juliankoehn/kspec/provider"
	"github.com/juliankoehn/kspec/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// kspec scan local -f policy.yaml
	if len(os.Args) < 2 {
		log.Fatal("Usage: kspec <command> [args]")
	}

	command := os.Args[1]

	switch command {
	case "scan":
		runScan(os.Args[2:])
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func runScan(args []string) {
	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	policyPath := scanCmd.String("f", "", "Path to policy YAML file")
	policyDir := scanCmd.String("d", "", "Path to policy directory")

	// Credential flags
	token := scanCmd.String("token", "", "GitHub token (or use GITHUB_TOKEN env var)")
	credentialType := scanCmd.String("credential-type", "", "Credential type (bearer, env, password)")
	envVar := scanCmd.String("env-var", "GITHUB_TOKEN", "Environment variable name for token")

	var target string
	var assetType string
	var parseArgs []string
	var owner, repo string

	// Parse scan target type
	if len(args) > 0 {
		switch args[0] {
		case "local":
			assetType = "local"
			target = "localhost"
			parseArgs = args[1:]
		case "host":
			if len(args) < 2 {
				log.Fatal("Usage: kspec scan host <target> -f ...")
			}
			assetType = "host"
			target = args[1]
			parseArgs = args[2:]
		case "github":
			if len(args) < 2 {
				log.Fatal("Usage: kspec scan github <org|repo> <target> -f ...")
			}
			subCmd := args[1]
			if subCmd == "org" {
				if len(args) < 3 {
					log.Fatal("Usage: kspec scan github org <org-name> -f ...")
				}
				assetType = "github-org"
				owner = args[2]
				target = owner
				parseArgs = args[3:]
			} else if subCmd == "repo" {
				if len(args) < 3 {
					log.Fatal("Usage: kspec scan github repo <owner/repo> -f ...")
				}
				repoPath := args[2]
				parts := strings.Split(repoPath, "/")
				if len(parts) != 2 {
					log.Fatal("Repository must be in format: owner/repo")
				}
				assetType = "github-repo"
				owner = parts[0]
				repo = parts[1]
				target = repoPath
				parseArgs = args[3:]
			} else {
				log.Fatalf("Unknown github subcommand: %s (use 'org' or 'repo')", subCmd)
			}
		default:
			// Try parsing as flags
			assetType = "local"
			target = "localhost"
			parseArgs = args
		}
	} else {
		assetType = "local"
		target = "localhost"
		parseArgs = args
	}

	scanCmd.Parse(parseArgs)

	if *policyPath == "" && *policyDir == "" {
		log.Fatal("Policy file (-f) or directory (-d) is required")
	}

	// Build Asset
	asset := core.Asset{
		Type:   assetType,
		Name:   target,
		Config: make(map[string]string),
	}

	// Configure asset based on type
	switch assetType {
	case "host":
		asset.FQDN = target
		asset.Config["target"] = target
	case "github-org":
		asset.Config["owner"] = owner
	case "github-repo":
		asset.Config["owner"] = owner
		asset.Config["repo"] = repo
	}

	// 1. Load Policies
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

	if *policyDir != "" {
		files, err := os.ReadDir(*policyDir)
		if err != nil {
			log.Fatalf("Failed to read policy directory: %v", err)
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
				path := filepath.Join(*policyDir, file.Name())
				if err := loadPolicy(path); err != nil {
					log.Printf("Warning: failed to load policy %s: %v", file.Name(), err)
				}
			}
		}
	} else if *policyPath != "" {
		if err := loadPolicy(*policyPath); err != nil {
			log.Fatalf("Failed to read policy file: %v", err)
		}
	}

	// 2. Setup Provider Config with Credentials
	providerConfig := make(map[string]string)

	// Configure credentials for GitHub provider
	if assetType == "github-org" || assetType == "github-repo" {
		if *token != "" {
			// Explicit token provided
			providerConfig["credential_type"] = "bearer"
			providerConfig["secret"] = *token
		} else if *credentialType != "" {
			// Credential type specified
			providerConfig["credential_type"] = *credentialType
			if *credentialType == "env" {
				providerConfig["env_var"] = *envVar
			}
		} else {
			// Default: try environment variable
			providerConfig["credential_type"] = "env"
			providerConfig["env_var"] = "GITHUB_TOKEN"
		}
	}

	registry, err := provider.InitProviders(context.Background(), providerConfig)
	if err != nil {
		log.Fatalf("Failed to init providers: %v", err)
	}
	// 3. Init Evaluator
	eval, err := core.NewEvaluator(registry)
	if err != nil {
		log.Fatalf("Failed to init evaluator: %v", err)
	}

	ctx := context.Background()

	// 3.5 Pre-fetch DNS Context if targeting host
	// This allows filters like 'resource.mx != empty' to work
	dnsCtx := make(map[string]interface{})
	if target != "" {
		if dnsSpec, ok := registry["dns"]; ok {
			pfCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			dnsAsset := core.Asset{
				FQDN:   target,
				Config: map[string]string{"domain": target},
			}
			resources, err := dnsSpec.Fetch(pfCtx, dnsAsset)
			cancel() // cleanup context

			if err == nil && len(resources) > 0 {
				dnsCtx = resources[0]
			}
		}
	}

	// 4. PREPARE / PLAN
	// We resolve all checks first to populate the UI
	// We need to keep track of the resolved checks to run them later
	type ResolvedCheck struct {
		GroupTitle string
		CheckDef   core.Check
	}
	var executionPlan []ResolvedCheck
	var uiItems []ui.CheckItem

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

	// Collect checks from ALL policies
	for _, policy := range policies {
		for _, group := range policy.Groups {
			// Evaluator Filter
			if group.Filter != "" {
				// Pass dnsCtx as 'resource' to Evaluate
				pass, err := eval.Evaluate(group.Filter, dnsCtx, nil, nil, asset)
				if err != nil {
					continue
				}
				if !pass {
					continue
				}
			}

			for _, checkRef := range group.Checks {
				// Resolve Check
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
						// Merge config: local overrides definition
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

				// Add to plan
				executionPlan = append(executionPlan, ResolvedCheck{
					GroupTitle: group.Title,
					CheckDef:   check,
				})

				// Add to UI
				uiItems = append(uiItems, ui.CheckItem{
					ID:     check.ID, // Note: ID might not be unique if reused? Using ID for now.
					Group:  group.Title,
					Name:   check.Title,
					Status: ui.StatusPending,
				})
			}
		}
	}

	// 5. Init UI
	model := ui.NewModel(uiItems)
	p := tea.NewProgram(model)

	// 6. Run Execution in Background
	go func() {
		for _, item := range executionPlan {
			check := item.CheckDef

			// Notify RUNNING
			p.Send(ui.CheckResultMsg{
				ID:     check.ID,
				Status: ui.StatusRunning,
			})

			// Create timeout context for this check
			// Increased timeout for network checks
			checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

			// Resolve Resource Spec
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

			// Variable Substitution
			config := make(map[string]string)
			hasDomain := false
			for k, v := range check.Config {
				if k == "domain" {
					hasDomain = true
				}
				if target != "" {
					config[k] = strings.ReplaceAll(v, "${target}", target)
				} else {
					config[k] = v
				}
			}

			if target != "" && !hasDomain {
				config["domain"] = target
			}

			// Fetch with Timeout
			// Note: Provider Fetch needs to respect context.
			fetchAsset := asset
			// Merge/Override config for the check
			fetchAsset.Config = make(map[string]string)
			for k, v := range asset.Config {
				fetchAsset.Config[k] = v
			}
			for k, v := range config {
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

			// Prepare Props
			props := make(map[string]interface{})
			for _, prop := range check.Props {
				props[prop.UID] = prop.MQL
			}

			finalStatus := ui.StatusPassed
			var errorMsgs []string

			// Evaluate
			for _, res := range resources {
				// Evaluate might not take context yet, but it's fast usually.
				// If Evaluate eventually does network, pass checkCtx.

				// Standard evaluate call (5 args)
				pass, err := eval.Evaluate(check.Query, res, config, props, asset)
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

			// Small delay to visually see the running state
			// time.Sleep(50 * time.Millisecond)
		}

		// Optional: Quit automatically when done?
		// p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
