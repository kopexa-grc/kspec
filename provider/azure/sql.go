package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"

	"github.com/kopexa-grc/kspec/core"
)

// SqlServerResource represents an Azure SQL Server resource scanner.
type SqlServerResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *SqlServerResource) Name() string {
	return "azure_sql_server"
}

// Fetch retrieves SQL servers from Azure.
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
							auditData, marshalErr := json.Marshal(auditPolicy)
							if marshalErr == nil {
								var auditMap map[string]interface{}
								if unmarshalErr := json.Unmarshal(auditData, &auditMap); unmarshalErr == nil {
									resourceMap["auditingPolicy"] = auditMap
								}
							}
						}
					}
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

// SubResources returns the sub-resources associated with SQL servers.
func (r *SqlServerResource) SubResources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&SqlDatabaseResource{
			credential:     r.credential,
			subscriptionID: r.subscriptionID,
			// serverName and resourceGroup will be populated during Fetch
		},
	}
}
