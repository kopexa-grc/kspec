package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/pages"

	"github.com/kopexa-grc/kspec/core"
)

// PagesProjectResource fetches Cloudflare Pages projects.
type PagesProjectResource struct {
	client    *cloudflare.Client
	accountID string
}

// Name returns the resource type name.
func (r *PagesProjectResource) Name() string {
	return "cloudflare_pages_project"
}

// Fetch retrieves all Pages projects for the account.
func (r *PagesProjectResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	accountID := r.accountID
	if accountID == "" {
		accountID = asset.Config["account_id"]
	}
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required for Pages scanning")
	}

	var resources []core.Resource

	// List all Pages projects
	result, err := r.client.Pages.Projects.List(ctx, pages.ProjectListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return resources, err
	}

	for _, project := range result.Result {
		data, err := json.Marshal(project)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["account_id"] = accountID

		// Add computed fields
		r.addComputedFields(resourceMap)

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

func (r *PagesProjectResource) addComputedFields(project map[string]interface{}) {
	// Check source configuration
	if source, ok := project["source"].(map[string]interface{}); ok {
		if config, ok := source["config"].(map[string]interface{}); ok {
			project["has_preview_deployment"] = config["preview_deployment_setting"] != "none"
			project["production_branch"] = config["production_branch"]
		}

		if sourceType, ok := source["type"].(string); ok {
			project["source_type"] = sourceType
			project["is_git_connected"] = sourceType == "github" || sourceType == "gitlab"
		}
	}

	// Check deployment configs
	if deploymentConfigs, ok := project["deployment_configs"].(map[string]interface{}); ok {
		// Check production config
		if prod, ok := deploymentConfigs["production"].(map[string]interface{}); ok {
			if envVars, ok := prod["env_vars"].(map[string]interface{}); ok {
				project["production_env_var_count"] = len(envVars)
			}
		}
		// Check preview config
		if preview, ok := deploymentConfigs["preview"].(map[string]interface{}); ok {
			if envVars, ok := preview["env_vars"].(map[string]interface{}); ok {
				project["preview_env_var_count"] = len(envVars)
			}
		}
	}

	// Check custom domains
	if domains, ok := project["domains"].([]interface{}); ok {
		project["has_custom_domain"] = len(domains) > 0
		project["domain_count"] = len(domains)
	}
}
