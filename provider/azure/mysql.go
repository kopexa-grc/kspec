package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"

	"github.com/kopexa-grc/kspec/core"
)

// MySQLServerResource represents an Azure MySQL Server resource scanner.
type MySQLServerResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *MySQLServerResource) Name() string {
	return "azure_mysql_server"
}

// Fetch retrieves MySQL servers from Azure.
func (r *MySQLServerResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armmysql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL servers client: %w", err)
	}

	pager := client.NewListPager(nil)
	return fetchWithPager(ctx, pager, func(page armmysql.ServersClientListResponse) []*armmysql.Server {
		return page.Value
	}, r.Name())
}
