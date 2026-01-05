package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/juliankoehn/kspec/core"
	"github.com/juliankoehn/kspec/provider"
)

type Job struct {
	ProviderName string            `json:"provider"`
	ResourceType string            `json:"resource_type"`
	Config       map[string]string `json:"config"`
	Policy       string            `json:"policy"`
	PolicyName   string            `json:"policy_name"`
}

type AgentConfig struct {
	TenantID string `json:"tenant_id"`
	Jobs     []Job  `json:"jobs"`
	Interval string `json:"interval"`
}

type Runner struct {
	Config    AgentConfig
	Evaluator *core.Evaluator
	// Registry maps "resource_type" -> ResourceSpec
	Registry map[string]core.ResourceSpec
}

func NewRunner(config AgentConfig) (*Runner, error) {
	// In a real app, providers might be injected or configured dynamically.
	// Here we initialize the GitHub provider based on one of the jobs or a global config?
	// For simplicity, let's assume we scan jobs to find tokens OR just initialize providers once.
	// But providers might need specific config (Token).
	// The current Job-centric config makes it hard to share provider instances across jobs if config varies.
	// However, usually Provider config (Token) is global per Tenant/Agent.

	// Let's assume we find a token from the first github job or env var for now to Init the provider.
	// Ideally AgentConfig should have "Providers" section.

	// Workaround: Look for "token" in the first github job config we find.
	var ghToken string
	for _, j := range config.Jobs {
		if j.ProviderName == "github" {
			if t, ok := j.Config["token"]; ok {
				ghToken = t
			}
			break
		}
	}

	// We need to hydrate providers with config (e.g. GitHub Token)
	// Extract from Job config or AgentConfig?
	// For simplicity, checking if any Job has a token
	providerConfig := make(map[string]string)
	if ghToken != "" {
		providerConfig["token"] = ghToken
	}

	registry, err := provider.InitProviders(context.Background(), providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to init providers: %v", err)
	}

	eval, err := core.NewEvaluator(registry)
	if err != nil {
		return nil, err
	}

	return &Runner{
		Config:    config,
		Evaluator: eval,
		Registry:  registry,
	}, nil
}

func (r *Runner) Start(ctx context.Context) error {
	duration, err := time.ParseDuration(r.Config.Interval)
	if err != nil {
		duration = 5 * time.Minute
	}
	log.Printf("Starting runner for tenant %s with interval %v", r.Config.TenantID, duration)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	// Run immediately once
	r.RunOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			r.RunOnce(ctx)
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context) {
	for _, job := range r.Config.Jobs {
		// Lookup ResourceSpec
		spec, ok := r.Registry[job.ResourceType]
		if !ok {
			log.Printf("Unknown resource type: %s", job.ResourceType)
			continue
		}

		log.Printf("Fetching resources (%s)...", job.ResourceType)
		asset := core.Asset{
			Type:   job.ResourceType,
			Config: job.Config,
		}
		// Try to populate common fields if present in config
		if domain, ok := job.Config["domain"]; ok {
			asset.FQDN = domain
			if asset.Name == "" {
				asset.Name = domain
			}
		}

		resources, err := spec.Fetch(ctx, asset)
		if err != nil {
			log.Printf("Error fetching resource %s: %v", job.ResourceType, err)
			continue
		}

		for _, res := range resources {
			passed, err := r.Evaluator.Evaluate(job.Policy, res, job.Config, nil, asset)

			// Identify the resource (e.g., by name or ID)
			resourceID := "unknown"
			if name, ok := res["name"].(string); ok {
				resourceID = name
			}

			result := core.CheckResult{
				TenantID:   r.Config.TenantID,
				ResourceID: resourceID,
				Passed:     passed,
				Metadata: map[string]interface{}{
					"policy_name": job.PolicyName,
					"error":       nil,
				},
			}

			if err != nil {
				result.Passed = false
				result.Metadata["error"] = err.Error()
			}

			// Output result (Log for now)
			output, _ := json.Marshal(result)
			fmt.Println(string(output))
		}
	}
}
