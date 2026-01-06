package atlassian

import (
	"context"
	"os"
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

func TestAtlassianProvider_Name(t *testing.T) {
	p := NewAtlassianProvider()
	if got := p.Name(); got != "atlassian" {
		t.Errorf("Name() = %v, want %v", got, "atlassian")
	}
}

func TestAtlassianProvider_Connect_MissingCredentials(t *testing.T) {
	// Clear any environment variables
	os.Unsetenv("ATLASSIAN_EMAIL")
	os.Unsetenv("ATLASSIAN_API_TOKEN")
	os.Unsetenv("ATLASSIAN_SITE")

	tests := []struct {
		name    string
		config  map[string]string
		wantErr string
	}{
		{
			name:    "no credentials at all",
			config:  map[string]string{},
			wantErr: "no valid credentials provided",
		},
		{
			name: "email only",
			config: map[string]string{
				"email": "test@example.com",
			},
			wantErr: "no valid credentials provided",
		},
		{
			name: "token only",
			config: map[string]string{
				"api_token": "test-token",
			},
			wantErr: "no valid credentials provided",
		},
		{
			name: "email and token but no site",
			config: map[string]string{
				"email":     "test@example.com",
				"api_token": "test-token",
			},
			wantErr: "site not specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewAtlassianProvider()
			_, err := p.Connect(context.Background(), tt.config)

			if err == nil {
				t.Error("Connect() expected error, got nil")
				return
			}

			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("Connect() error = %v, want error containing %v", err, tt.wantErr)
			}
		})
	}
}

func TestAtlassianProvider_Connect_WithEnvVars(t *testing.T) {
	os.Setenv("ATLASSIAN_EMAIL", "env@example.com")
	os.Setenv("ATLASSIAN_API_TOKEN", "env-token")
	defer os.Unsetenv("ATLASSIAN_EMAIL")
	defer os.Unsetenv("ATLASSIAN_API_TOKEN")

	p := NewAtlassianProvider()
	_, err := p.Connect(context.Background(), map[string]string{
		"site": "test.atlassian.net",
	})

	// Should fail on actual connection (no real server), but credentials should be accepted
	// The error should NOT be about missing credentials
	if err != nil && containsString(err.Error(), "no valid credentials") {
		t.Error("Connect() should have accepted credentials from environment")
	}
}

func TestAtlassianProvider_Connect_ConfigOverridesEnv(t *testing.T) {
	os.Setenv("ATLASSIAN_EMAIL", "env@example.com")
	os.Setenv("ATLASSIAN_API_TOKEN", "env-token")
	os.Setenv("ATLASSIAN_SITE", "env.atlassian.net")
	defer os.Unsetenv("ATLASSIAN_EMAIL")
	defer os.Unsetenv("ATLASSIAN_API_TOKEN")
	defer os.Unsetenv("ATLASSIAN_SITE")

	p := NewAtlassianProvider()
	_, err := p.Connect(context.Background(), map[string]string{
		"email":     "config@example.com",
		"api_token": "config-token",
		"site":      "config.atlassian.net",
	})

	// Connection will fail but we're testing credential handling
	if err != nil && containsString(err.Error(), "no valid credentials") {
		t.Error("Connect() should have accepted credentials from config")
	}
}

func TestDiscoverCloudID_UnreachableHost(t *testing.T) {
	// Test that discoverCloudID returns an error for an unreachable host
	_, err := discoverCloudID("nonexistent-host-12345.atlassian.net", "test@example.com", "test-token")
	if err == nil {
		t.Error("discoverCloudID() expected error for unreachable host")
	}
}

// Note: Additional discoverCloudID tests would require mocking the HTTP client
// since the function uses HTTPS. The unreachable host test above verifies
// basic error handling.

func TestAtlassianConnection_Resources(t *testing.T) {
	conn := &AtlassianConnection{
		jiraClient:       nil,
		confluenceClient: nil,
		adminClient:      nil,
		site:             "test.atlassian.net",
		cloudID:          "test-cloud-id",
		orgID:            "test-org-id",
	}

	resources := conn.Resources()

	expectedNames := []string{
		"atlassian_jira_project",
		"atlassian_jira_permission_scheme",
		"atlassian_jira_security_scheme",
		"atlassian_jira_user",
		"atlassian_confluence_space",
		"atlassian_confluence_page",
		"atlassian_organization",
		"atlassian_admin_user",
		"atlassian_admin_group",
		"atlassian_auth_policy",
		"atlassian_scim_user",
		"atlassian_scim_group",
	}

	if len(resources) != len(expectedNames) {
		t.Errorf("Resources() returned %d resources, want %d", len(resources), len(expectedNames))
	}

	nameMap := make(map[string]bool)
	for _, r := range resources {
		nameMap[r.Name()] = true
	}

	for _, name := range expectedNames {
		if !nameMap[name] {
			t.Errorf("Resources() missing resource: %s", name)
		}
	}
}

// Resource name tests
func TestJiraProjectResource_Name(t *testing.T) {
	r := &JiraProjectResource{}
	if got := r.Name(); got != "atlassian_jira_project" {
		t.Errorf("Name() = %v, want atlassian_jira_project", got)
	}
}

func TestJiraPermissionSchemeResource_Name(t *testing.T) {
	r := &JiraPermissionSchemeResource{}
	if got := r.Name(); got != "atlassian_jira_permission_scheme" {
		t.Errorf("Name() = %v, want atlassian_jira_permission_scheme", got)
	}
}

func TestJiraSecuritySchemeResource_Name(t *testing.T) {
	r := &JiraSecuritySchemeResource{}
	if got := r.Name(); got != "atlassian_jira_security_scheme" {
		t.Errorf("Name() = %v, want atlassian_jira_security_scheme", got)
	}
}

func TestJiraUserResource_Name(t *testing.T) {
	r := &JiraUserResource{}
	if got := r.Name(); got != "atlassian_jira_user" {
		t.Errorf("Name() = %v, want atlassian_jira_user", got)
	}
}

func TestConfluenceSpaceResource_Name(t *testing.T) {
	r := &ConfluenceSpaceResource{}
	if got := r.Name(); got != "atlassian_confluence_space" {
		t.Errorf("Name() = %v, want atlassian_confluence_space", got)
	}
}

func TestConfluencePageResource_Name(t *testing.T) {
	r := &ConfluencePageResource{}
	if got := r.Name(); got != "atlassian_confluence_page" {
		t.Errorf("Name() = %v, want atlassian_confluence_page", got)
	}
}

func TestAdminOrganizationResource_Name(t *testing.T) {
	r := &AdminOrganizationResource{}
	if got := r.Name(); got != "atlassian_organization" {
		t.Errorf("Name() = %v, want atlassian_organization", got)
	}
}

func TestAdminUserResource_Name(t *testing.T) {
	r := &AdminUserResource{}
	if got := r.Name(); got != "atlassian_admin_user" {
		t.Errorf("Name() = %v, want atlassian_admin_user", got)
	}
}

func TestAdminGroupResource_Name(t *testing.T) {
	r := &AdminGroupResource{}
	if got := r.Name(); got != "atlassian_admin_group" {
		t.Errorf("Name() = %v, want atlassian_admin_group", got)
	}
}

func TestAdminPolicyResource_Name(t *testing.T) {
	r := &AdminPolicyResource{}
	if got := r.Name(); got != "atlassian_auth_policy" {
		t.Errorf("Name() = %v, want atlassian_auth_policy", got)
	}
}

func TestSCIMUserResource_Name(t *testing.T) {
	r := &SCIMUserResource{}
	if got := r.Name(); got != "atlassian_scim_user" {
		t.Errorf("Name() = %v, want atlassian_scim_user", got)
	}
}

func TestSCIMGroupResource_Name(t *testing.T) {
	r := &SCIMGroupResource{}
	if got := r.Name(); got != "atlassian_scim_group" {
		t.Errorf("Name() = %v, want atlassian_scim_group", got)
	}
}

// JiraSecuritySchemeResource returns nil (placeholder)
func TestJiraSecuritySchemeResource_Fetch(t *testing.T) {
	r := &JiraSecuritySchemeResource{}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Errorf("Fetch() unexpected error = %v", err)
	}
	if resources != nil {
		t.Errorf("Fetch() expected nil (placeholder), got %v", resources)
	}
}

// AdminGroupResource returns nil (placeholder)
func TestAdminGroupResource_Fetch(t *testing.T) {
	r := &AdminGroupResource{}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Errorf("Fetch() unexpected error = %v", err)
	}
	if resources != nil {
		t.Errorf("Fetch() expected nil (placeholder), got %v", resources)
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
