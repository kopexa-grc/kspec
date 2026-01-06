package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ctreminiom/go-atlassian/admin"
	"github.com/ctreminiom/go-atlassian/confluence"
	v3 "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/kopexa-grc/kspec/core"
)

// AtlassianProvider implements the core.Provider interface for Atlassian Cloud.
type AtlassianProvider struct{}

// NewAtlassianProvider creates a new Atlassian provider.
func NewAtlassianProvider() *AtlassianProvider {
	return &AtlassianProvider{}
}

// Name returns the provider name.
func (p *AtlassianProvider) Name() string {
	return "atlassian"
}

// Connect establishes a connection to Atlassian Cloud APIs.
func (p *AtlassianProvider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Get authentication credentials
	email := config["email"]
	if email == "" {
		email = os.Getenv("ATLASSIAN_EMAIL")
	}

	apiToken := config["api_token"]
	if apiToken == "" {
		apiToken = os.Getenv("ATLASSIAN_API_TOKEN")
	}

	site := config["site"]
	if site == "" {
		site = os.Getenv("ATLASSIAN_SITE")
	}

	if email == "" || apiToken == "" {
		return nil, fmt.Errorf("atlassian: no valid credentials provided. Set email and api_token or ATLASSIAN_EMAIL and ATLASSIAN_API_TOKEN")
	}

	if site == "" {
		return nil, fmt.Errorf("atlassian: site not specified. Set site or ATLASSIAN_SITE (e.g., yoursite.atlassian.net)")
	}

	// Create Jira client
	jiraClient, err := v3.New(nil, site)
	if err != nil {
		return nil, fmt.Errorf("atlassian: failed to create Jira client: %w", err)
	}
	jiraClient.Auth.SetBasicAuth(email, apiToken)

	// Create Confluence client
	confluenceClient, err := confluence.New(nil, site)
	if err != nil {
		return nil, fmt.Errorf("atlassian: failed to create Confluence client: %w", err)
	}
	confluenceClient.Auth.SetBasicAuth(email, apiToken)

	// Create Admin client (uses different base URL)
	adminClient, err := admin.New(nil)
	if err != nil {
		return nil, fmt.Errorf("atlassian: failed to create Admin client: %w", err)
	}
	adminClient.Auth.SetBearerToken(apiToken)

	// Discover Cloud ID from site
	cloudID, err := discoverCloudID(site, email, apiToken)
	if err != nil {
		// Cloud ID is optional for some operations
		cloudID = ""
	}

	// Get org ID from config or environment
	orgID := config["org_id"]
	if orgID == "" {
		orgID = os.Getenv("ATLASSIAN_ORG_ID")
	}

	return &AtlassianConnection{
		jiraClient:       jiraClient,
		confluenceClient: confluenceClient,
		adminClient:      adminClient,
		site:             site,
		cloudID:          cloudID,
		orgID:            orgID,
		email:            email,
		apiToken:         apiToken,
	}, nil
}

// AtlassianConnection represents an active connection to Atlassian Cloud.
type AtlassianConnection struct {
	jiraClient       *v3.Client
	confluenceClient *confluence.Client
	adminClient      *admin.Client
	site             string
	cloudID          string
	orgID            string
	email            string
	apiToken         string
}

// Resources returns all available Atlassian resources.
func (c *AtlassianConnection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Jira resources
		&JiraProjectResource{client: c.jiraClient},
		&JiraPermissionSchemeResource{client: c.jiraClient},
		&JiraSecuritySchemeResource{client: c.jiraClient},
		&JiraUserResource{client: c.jiraClient},
		// Confluence resources
		&ConfluenceSpaceResource{client: c.confluenceClient},
		&ConfluencePageResource{client: c.confluenceClient},
		// Admin resources (require org ID)
		&AdminOrganizationResource{client: c.adminClient, orgID: c.orgID},
		&AdminUserResource{client: c.adminClient, orgID: c.orgID},
		&AdminGroupResource{client: c.adminClient, orgID: c.orgID},
		&AdminPolicyResource{client: c.adminClient, orgID: c.orgID},
		// SCIM resources
		&SCIMUserResource{client: c.adminClient, orgID: c.orgID},
		&SCIMGroupResource{client: c.adminClient, orgID: c.orgID},
	}
}

// discoverCloudID fetches the Cloud ID from the Atlassian site.
func discoverCloudID(site, email, apiToken string) (string, error) {
	url := fmt.Sprintf("https://%s/_edge/tenant_info", site)

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(email, apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get tenant info: %s", resp.Status)
	}

	var info struct {
		CloudID string `json:"cloudId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	return info.CloudID, nil
}
