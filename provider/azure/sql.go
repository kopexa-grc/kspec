package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/kopexa-grc/kspec/core"
)

type SqlServerResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (r *SqlServerResource) Name() string {
	return "azure_sql_server"
}

func (r *SqlServerResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create SQL servers client
	client, err := armsql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL servers client: %w", err)
	}

	var resources []core.Resource

	// List all SQL servers in the subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list SQL servers: %w", err)
		}

		for _, server := range page.Value {
			// Convert to map[string]interface{}
			data, err := json.Marshal(server)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			// Fetch auditing policy
			if server.Name != nil {
				// Extract resource group from server ID
				// Simplified - in production you'd parse the full ARM resource ID
				resourceGroup := asset.Config["resource_group"]
				if resourceGroup != "" {
					auditingClient, err := armsql.NewServerBlobAuditingPoliciesClient(r.subscriptionID, r.credential, nil)
					if err == nil {
						auditPolicy, err := auditingClient.Get(ctx, resourceGroup, *server.Name, nil)
						if err == nil {
							auditData, _ := json.Marshal(auditPolicy)
							var auditMap map[string]interface{}
							json.Unmarshal(auditData, &auditMap)
							resourceMap["auditingPolicy"] = auditMap
						}
					}
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

// SubResources implements core.SubResourceProvider
// This allows SQL Server to dynamically register SQL Database as a sub-resource
func (r *SqlServerResource) SubResources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&SqlDatabaseResource{
			credential:     r.credential,
			subscriptionID: r.subscriptionID,
			// serverName and resourceGroup will be populated during Fetch
		},
	}
}
