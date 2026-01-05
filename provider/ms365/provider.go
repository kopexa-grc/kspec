package ms365

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

type MS365Provider struct{}

func NewMS365Provider() core.Provider {
	return &MS365Provider{}
}

func (p *MS365Provider) Name() string {
	return "ms365"
}

type MS365Connection struct {
	client   *msgraphsdk.GraphServiceClient
	tenantID string
}

func (p *MS365Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	var credential azcore.TokenCredential
	var err error

	// Get tenant ID from config
	tenantID := config["tenant_id"]
	clientID := config["client_id"]
	clientSecret := config["client_secret"]

	if clientID != "" && clientSecret != "" && tenantID != "" {
		// Service Principal authentication with client secret
		credential, err = azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create client secret credential: %w", err)
		}
	} else {
		// Use default Azure credential (tries az cli, managed identity, env vars, etc.)
		credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create default credential: %w", err)
		}
	}

	// Create Graph client
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(credential, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return &MS365Connection{
		client:   client,
		tenantID: tenantID,
	}, nil
}

func (c *MS365Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Tenant and organization
		&TenantResource{client: c.client},

		// Users and groups
		&UserResource{client: c.client},
		&GroupResource{client: c.client},

		// Applications and service principals
		&ApplicationResource{client: c.client},
		&ServicePrincipalResource{client: c.client},

		// Devices
		&DeviceResource{client: c.client},
		&ManagedDeviceResource{client: c.client},
		&DeviceConfigurationResource{client: c.client},
		&DeviceCompliancePolicyResource{client: c.client},

		// Domains
		&DomainResource{client: c.client},

		// Identity and access
		&ConditionalAccessPolicyResource{client: c.client},
		&NamedLocationResource{client: c.client},

		// Directory roles
		&DirectoryRoleResource{client: c.client},

		// Policies
		&AuthorizationPolicyResource{client: c.client},
		&AuthenticationMethodPolicyResource{client: c.client},

		// Security
		&RiskyUserResource{client: c.client},
		&SecureScoreResource{client: c.client},

		// Teams
		&TeamResource{client: c.client},

		// Settings
		&DirectorySettingResource{client: c.client},
	}
}
