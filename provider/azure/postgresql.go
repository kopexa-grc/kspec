package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"

	"github.com/kopexa-grc/kspec/core"
)

type PostgreSQLServerResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (r *PostgreSQLServerResource) Name() string {
	return "azure_postgresql_server"
}

func (r *PostgreSQLServerResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create PostgreSQL servers client
	client, err := armpostgresql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL servers client: %w", err)
	}

	var resources []core.Resource

	// List all PostgreSQL servers in the subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list PostgreSQL servers: %w", err)
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

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
