# Atlassian Provider Setup

This guide explains how to configure kspec to scan your Atlassian Cloud resources (Jira, Confluence, and Admin APIs) for security compliance.

## Prerequisites

- An Atlassian Cloud account
- API Token for authentication
- `kspec` installed on your system

## Authentication

### Create an API Token

1. Log in to [Atlassian Account](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **Create API token**
3. Enter a label (e.g., "kspec security scanning")
4. Click **Create**
5. Copy the token and store it securely

### Configure Credentials

**Using environment variables (recommended):**

```bash
export ATLASSIAN_EMAIL="your-email@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"
export ATLASSIAN_SITE="yoursite.atlassian.net"

# Optional: For Admin API access
export ATLASSIAN_ORG_ID="your-org-id"
```

**Using command line flags:**

```bash
kspec scan atlassian site yoursite.atlassian.net \
  --email "your-email@example.com" \
  --api-token "your-api-token" \
  -f policies/atlassian-security.yml
```

## Finding Your Organization ID

To use Admin API features (organization users, policies), you need your Organization ID:

1. Go to [admin.atlassian.com](https://admin.atlassian.com)
2. Select your organization
3. The Organization ID is in the URL: `https://admin.atlassian.com/o/{org-id}/...`

## Scanning Commands

### Scan a Site

Scan all Jira and Confluence resources for a site:

```bash
kspec scan atlassian site yoursite.atlassian.net -f policies/atlassian-security.yml
```

### Scan Organization (Admin APIs)

Scan organization-level resources including users and policies:

```bash
kspec scan atlassian org <org-id> \
  --site yoursite.atlassian.net \
  -f policies/atlassian-security.yml
```

### Using Aliases

You can also use `jira` or `confluence` as aliases:

```bash
kspec scan jira site yoursite.atlassian.net -f policies/atlassian-security.yml
kspec scan confluence site yoursite.atlassian.net -f policies/atlassian-security.yml
```

## Resources Discovered

kspec discovers the following Atlassian resources:

### Jira Resources

| Resource Type | Description |
|---------------|-------------|
| `atlassian_jira_project` | Jira projects with lead and insight info |
| `atlassian_jira_permission_scheme` | Permission schemes and grants |
| `atlassian_jira_security_scheme` | Issue security schemes (placeholder) |
| `atlassian_jira_user` | Jira users and account types |

### Confluence Resources

| Resource Type | Description |
|---------------|-------------|
| `atlassian_confluence_space` | Spaces with permissions |
| `atlassian_confluence_page` | Pages with metadata |

### Admin Resources

| Resource Type | Description |
|---------------|-------------|
| `atlassian_organization` | Organization details and domains |
| `atlassian_admin_user` | Organization users with product access |
| `atlassian_admin_group` | Organization groups (placeholder) |
| `atlassian_auth_policy` | Authentication policies |

### SCIM Resources

| Resource Type | Description |
|---------------|-------------|
| `atlassian_scim_user` | SCIM-provisioned users |
| `atlassian_scim_group` | SCIM-provisioned groups |

## Required Permissions

Ensure your API token has appropriate access:

### For Jira Scanning
- Browse projects
- View project permissions
- View users

### For Confluence Scanning
- View spaces
- View pages
- View space permissions

### For Admin API Scanning
- Organization Admin access
- User management permissions (for SCIM)

## Example Policy

Create a policy file `atlassian-security.yml`:

```yaml
policies:
  - uid: atlassian-security
    name: Atlassian Security Policy
    version: "1.0"

    require:
      - provider: atlassian

    groups:
      - title: Jira Project Security
        filter: asset.type == "atlassian-site"
        checks:
          - uid: jira-projects-have-leads
            title: Projects should have assigned leads
            resource: atlassian_jira_project
            query: resource.has_lead == true
            severity: medium

      - title: Permission Scheme Security
        filter: asset.type == "atlassian-site"
        checks:
          - uid: no-anonymous-jira-access
            title: Permission schemes should not allow anonymous access
            resource: atlassian_jira_permission_scheme
            query: resource.has_anonymous_access != true
            severity: high

      - title: Confluence Space Security
        filter: asset.type == "atlassian-site"
        checks:
          - uid: no-anonymous-confluence-access
            title: Spaces should not allow anonymous access
            resource: atlassian_confluence_space
            query: resource.has_anonymous_access != true
            severity: high
```

## Troubleshooting

### Error: No valid credentials provided

Ensure you have set:
- `ATLASSIAN_EMAIL` and `ATLASSIAN_API_TOKEN` environment variables, or
- `--email` and `--api-token` command line flags

### Error: Site not specified

Set the site URL:
- `ATLASSIAN_SITE` environment variable, or
- `--site` command line flag, or
- Provide site as an argument: `kspec scan atlassian site yoursite.atlassian.net`

### Error: 401 Unauthorized

- Verify your API token is valid and not expired
- Ensure the email matches the account that created the token
- Check that the token has the required permissions

### Error: 403 Forbidden

Your API token may not have sufficient permissions. Verify:
- You have admin access for Admin API resources
- You have appropriate project/space access for Jira/Confluence resources

### Rate Limiting

Atlassian APIs have rate limits. If you encounter rate limiting:
- Reduce the number of resources being scanned
- Add delays between scans in CI/CD pipelines
- Consider scanning during off-peak hours

## CI/CD Integration

### GitHub Actions

```yaml
name: Atlassian Security Scan

on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6 AM
  workflow_dispatch:

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install kspec
        run: |
          go install github.com/kopexa-grc/kspec/cmd/kspec@latest

      - name: Run Atlassian scan
        env:
          ATLASSIAN_EMAIL: ${{ secrets.ATLASSIAN_EMAIL }}
          ATLASSIAN_API_TOKEN: ${{ secrets.ATLASSIAN_API_TOKEN }}
          ATLASSIAN_SITE: ${{ vars.ATLASSIAN_SITE }}
        run: |
          kspec scan atlassian site $ATLASSIAN_SITE -f policies/atlassian-security.yml
```

### GitLab CI

```yaml
atlassian-scan:
  image: golang:1.21
  script:
    - go install github.com/kopexa-grc/kspec/cmd/kspec@latest
    - kspec scan atlassian site $ATLASSIAN_SITE -f policies/atlassian-security.yml
  variables:
    ATLASSIAN_EMAIL: $ATLASSIAN_EMAIL
    ATLASSIAN_API_TOKEN: $ATLASSIAN_API_TOKEN
    ATLASSIAN_SITE: $ATLASSIAN_SITE
  only:
    - schedules
```

## Best Practices

1. **Use API Tokens**: Always use API tokens instead of passwords
2. **Least Privilege**: Create tokens with only the permissions needed for scanning
3. **Regular Scans**: Schedule regular scans to catch configuration drift
4. **Rotate Tokens**: Rotate API tokens regularly (every 90 days recommended)
5. **Secure Storage**: Store tokens in secure secret management systems
6. **Audit Logging**: Enable Atlassian audit logs to track API access

## Additional Resources

- [Atlassian API Documentation](https://developer.atlassian.com/cloud/)
- [Create API Tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
- [Admin API Documentation](https://developer.atlassian.com/cloud/admin/)
- [SCIM API Documentation](https://developer.atlassian.com/cloud/admin/user-provisioning/about-user-provisioning/)
