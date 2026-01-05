package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/juliankoehn/kspec/core"
)

// SqlDatabaseResource represents individual SQL databases (sub-resource of SQL Server)
type SqlDatabaseResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
	serverName     string
	resourceGroup  string
}

func (r *SqlDatabaseResource) Name() string {
	return "azure_sql_database"
}

func (r *SqlDatabaseResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}
	if r.resourceGroup == "" || r.serverName == "" {
		return []core.Resource{}, nil // No server context
	}

	// Create SQL databases client
	client, err := armsql.NewDatabasesClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL databases client: %w", err)
	}

	var resources []core.Resource

	// List all databases in this server
	pager := client.NewListByServerPager(r.resourceGroup, r.serverName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list SQL databases: %w", err)
		}

		for _, db := range page.Value {
			// Convert to map[string]interface{}
			data, err := json.Marshal(db)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			// Add server context
			resourceMap["serverName"] = r.serverName
			resourceMap["resourceGroup"] = r.resourceGroup

			// Fetch TDE (Transparent Data Encryption) status
			if db.Name != nil {
				tdeClient, err := armsql.NewTransparentDataEncryptionsClient(r.subscriptionID, r.credential, nil)
				if err == nil {
					tde, err := tdeClient.Get(ctx, r.resourceGroup, r.serverName, *db.Name, armsql.TransparentDataEncryptionNameCurrent, nil)
					if err == nil {
						tdeData, _ := json.Marshal(tde)
						var tdeMap map[string]interface{}
						json.Unmarshal(tdeData, &tdeMap)
						resourceMap["transparentDataEncryption"] = tdeMap
					}
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
