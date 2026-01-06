package cloudflare

import (
	"context"

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
	accountID, err := resolveAccountID(r.accountID, asset, "Pages")
	if err != nil {
		return nil, err
	}

	result, err := r.client.Pages.Projects.List(ctx, pages.ProjectListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return nil, err
	}

	return itemsToResources(result.Result, accountID, r.addComputedFields), nil
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
