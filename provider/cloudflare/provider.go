package cloudflare

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/kopexa-grc/kspec/core"
)

// CloudflareProvider implements the core.Provider interface for Cloudflare.
type CloudflareProvider struct{}

// NewCloudflareProvider creates a new Cloudflare provider.
func NewCloudflareProvider() *CloudflareProvider {
	return &CloudflareProvider{}
}

// Name returns the provider name.
func (p *CloudflareProvider) Name() string {
	return "cloudflare"
}

// Connect establishes a connection to Cloudflare API.
func (p *CloudflareProvider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	var opts []option.RequestOption

	// Try API Token first (preferred method)
	apiToken := config["api_token"]
	if apiToken == "" {
		apiToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	}

	if apiToken != "" {
		opts = append(opts, option.WithAPIToken(apiToken))
	} else {
		// Fall back to API Key + Email (legacy)
		apiKey := config["api_key"]
		if apiKey == "" {
			apiKey = os.Getenv("CLOUDFLARE_API_KEY")
		}

		email := config["email"]
		if email == "" {
			email = os.Getenv("CLOUDFLARE_EMAIL")
		}

		if apiKey != "" && email != "" {
			opts = append(opts, option.WithAPIKey(apiKey), option.WithAPIEmail(email))
		} else {
			return nil, fmt.Errorf("cloudflare: no valid credentials provided. Set api_token or (api_key + email)")
		}
	}

	client := cloudflare.NewClient(opts...)

	// Get account ID from config or discover it
	accountID := config["account_id"]
	if accountID == "" {
		accountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	}

	return &CloudflareConnection{
		client:    client,
		accountID: accountID,
	}, nil
}

// CloudflareConnection represents an active connection to Cloudflare.
type CloudflareConnection struct {
	client    *cloudflare.Client
	accountID string
}

// Resources returns all available Cloudflare resources.
func (c *CloudflareConnection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Account resources
		&AccountResource{client: c.client},
		// Zone resources
		&ZoneResource{client: c.client, accountID: c.accountID},
		&ZoneSettingsResource{client: c.client},
		&DNSRecordResource{client: c.client},
		// Security resources
		&WAFRuleResource{client: c.client},
		&FirewallRuleResource{client: c.client},
		// Platform resources
		&R2BucketResource{client: c.client, accountID: c.accountID},
		&WorkerResource{client: c.client, accountID: c.accountID},
		&PagesProjectResource{client: c.client, accountID: c.accountID},
		&TunnelResource{client: c.client, accountID: c.accountID},
		// Zero Trust resources
		&AccessApplicationResource{client: c.client, accountID: c.accountID},
		&AccessPolicyResource{client: c.client, accountID: c.accountID},
	}
}
