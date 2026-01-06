package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"github.com/kopexa-grc/kspec/core"
)

// SubscriptionResource represents an Azure Subscription resource scanner.
type SubscriptionResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *SubscriptionResource) Name() string {
	return "azure_subscription"
}

// Fetch retrieves subscription-level resources from Azure.
func (r *SubscriptionResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create a resource map for subscription-level data
	resourceMap := make(map[string]interface{})
	resourceMap["id"] = fmt.Sprintf("/subscriptions/%s", r.subscriptionID)
	resourceMap["subscriptionId"] = r.subscriptionID

	// Fetch diagnostic settings for the subscription
	diagnosticClient, err := armmonitor.NewDiagnosticSettingsClient(r.credential, nil)
	if err == nil {
		resourceURI := fmt.Sprintf("/subscriptions/%s", r.subscriptionID)
		pager := diagnosticClient.NewListPager(resourceURI, nil)

		var diagnosticSettings []map[string]interface{}
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}

			for _, setting := range page.Value {
				settingData, err := json.Marshal(setting)
				if err != nil {
					continue
				}
				var settingMap map[string]interface{}
				if err := json.Unmarshal(settingData, &settingMap); err != nil {
					continue
				}
				diagnosticSettings = append(diagnosticSettings, settingMap)
			}
		}

		// Create monitor object with diagnostic settings
		monitorData := map[string]interface{}{
			"diagnosticSettings": diagnosticSettings,
		}
		resourceMap["monitor"] = monitorData
	}

	// Fetch activity log alerts
	alertClient, err := armmonitor.NewActivityLogAlertsClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		pager := alertClient.NewListBySubscriptionIDPager(nil)

		var activityAlerts []map[string]interface{}
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}

			for _, alert := range page.Value {
				alertData, err := json.Marshal(alert)
				if err != nil {
					continue
				}
				var alertMap map[string]interface{}
				if err := json.Unmarshal(alertData, &alertMap); err != nil {
					continue
				}
				activityAlerts = append(activityAlerts, alertMap)
			}
		}

		// Add to monitor object
		if monitor, ok := resourceMap["monitor"].(map[string]interface{}); ok {
			activityLog := map[string]interface{}{
				"alerts": activityAlerts,
			}
			monitor["activityLog"] = activityLog
		}
	}

	// Placeholder for Cloud Defender security contacts
	// This would require Microsoft Defender for Cloud SDK
	resourceMap["cloudDefender"] = map[string]interface{}{
		"securityContacts": []interface{}{},
	}

	// Placeholder for web apps (App Services)
	// This would require App Service SDK
	resourceMap["web"] = map[string]interface{}{
		"apps": []interface{}{},
	}

	return []core.Resource{resourceMap}, nil
}

// Discover scans the subscription and discovers what resource types are present.
func (r *SubscriptionResource) Discover(ctx context.Context, asset core.Asset) (map[string]int, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	discovered := make(map[string]int)

	// The subscription itself is always 1
	discovered["azure_subscription"] = 1

	// Discover Storage Accounts
	storageClient, err := armstorage.NewAccountsClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := storageClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_storage_account"] = count
		}
	}

	// Discover SQL Servers
	sqlClient, err := armsql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := sqlClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_sql_server"] = count
		}
	}

	// Discover MySQL Servers
	mysqlClient, err := armmysql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := mysqlClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_mysql_server"] = count
		}
	}

	// Discover PostgreSQL Servers
	pgClient, err := armpostgresql.NewServersClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := pgClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_postgresql_server"] = count
		}
	}

	// Discover Key Vaults
	kvClient, err := armkeyvault.NewVaultsClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := kvClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_keyvault_vault"] = count
		}
	}

	// Discover NSGs
	nsgClient, err := armnetwork.NewSecurityGroupsClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := nsgClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_network_security_group"] = count
		}
	}

	// Discover VMs
	vmClient, err := armcompute.NewVirtualMachinesClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := vmClient.NewListAllPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_virtual_machine"] = count
		}
	}

	// Discover App Services
	appClient, err := armappservice.NewWebAppsClient(r.subscriptionID, r.credential, nil)
	if err == nil {
		count := 0
		pager := appClient.NewListPager(nil)
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			count += len(page.Value)
		}
		if count > 0 {
			discovered["azure_app_service"] = count
		}
	}

	return discovered, nil
}
