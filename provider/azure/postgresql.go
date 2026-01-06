package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"

	"github.com/kopexa-grc/kspec/core"
)

// PostgreSQLServerResource represents an Azure PostgreSQL Server resource scanner.
type PostgreSQLServerResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *PostgreSQLServerResource) Name() string {
	return "azure_postgresql_server"
}

// Fetch retrieves PostgreSQL servers from Azure.
func (r *PostgreSQLServerResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armpostgresql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL servers client: %w", err)
	}

	pager := client.NewListPager(nil)
	return fetchWithPager(ctx, pager, func(page armpostgresql.ServersClientListResponse) []*armpostgresql.Server {
		return page.Value
	}, r.Name())
}
