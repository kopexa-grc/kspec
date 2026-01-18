// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

//nolint:dupl // Azure database resource scanners have similar structures by design
package resources

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"

	"github.com/kopexa-grc/kspec/core"
)

// PostgreSQLServer represents an Azure PostgreSQL Server resource scanner.
type PostgreSQLServer struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewPostgreSQLServer creates a new PostgreSQLServer resource scanner.
func NewPostgreSQLServer(credential azcore.TokenCredential, subscriptionID string) *PostgreSQLServer {
	return &PostgreSQLServer{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *PostgreSQLServer) Name() string {
	return "azure_postgresql_server"
}

// Fetch retrieves PostgreSQL servers from Azure.
func (r *PostgreSQLServer) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
