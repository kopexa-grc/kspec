// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"

	"github.com/kopexa-grc/kspec/core"
)

// MySQLServer represents an Azure MySQL Server resource scanner.
type MySQLServer struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewMySQLServer creates a new MySQLServer resource scanner.
func NewMySQLServer(credential azcore.TokenCredential, subscriptionID string) *MySQLServer {
	return &MySQLServer{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *MySQLServer) Name() string {
	return "azure_mysql_server"
}

// Fetch retrieves MySQL servers from Azure.
func (r *MySQLServer) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
