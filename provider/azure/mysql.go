package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/juliankoehn/kspec/core"
)

type MySQLServerResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (r *MySQLServerResource) Name() string {
	return "azure_mysql_server"
}

func (r *MySQLServerResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create MySQL servers client
	client, err := armmysql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL servers client: %w", err)
	}

	var resources []core.Resource

	// List all MySQL servers in the subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list MySQL servers: %w", err)
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
