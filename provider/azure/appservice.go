package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"

	"github.com/kopexa-grc/kspec/core"
)

type AppServiceResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (r *AppServiceResource) Name() string {
	return "azure_app_service"
}

func (r *AppServiceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create App Service web apps client
	client, err := armappservice.NewWebAppsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create App Service client: %w", err)
	}

	var resources []core.Resource

	// List all web apps in the subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list web apps: %w", err)
		}

		for _, app := range page.Value {
			// Convert to map[string]interface{}
			data, err := json.Marshal(app)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			// Fetch additional app configuration if needed
			if app.Name != nil {
				resourceGroup := asset.Config["resource_group"]
				if resourceGroup != "" {
					// Get app configuration
					configClient, err := armappservice.NewWebAppsClient(r.subscriptionID, r.credential, nil)
					if err == nil {
						config, err := configClient.GetConfiguration(ctx, resourceGroup, *app.Name, nil)
						if err == nil {
							configData, _ := json.Marshal(config)
							var configMap map[string]interface{}
							json.Unmarshal(configData, &configMap)
							resourceMap["configuration"] = configMap
						}
					}
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
