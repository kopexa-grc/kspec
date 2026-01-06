# Atlassian Provider

The Atlassian provider scans your Atlassian Cloud environment (Jira, Confluence, Admin, and SCIM) for security compliance.

## Overview

Use the Atlassian provider to validate:
- Jira project security (leads, permission schemes)
- Jira permission schemes (anonymous access, group permissions)
- Confluence space security (anonymous access, unlicensed access)
- Organization security policies
- SCIM user provisioning
- Authentication policies

## Quick Start

```bash
# Set your credentials
export ATLASSIAN_EMAIL="your-email@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"
export ATLASSIAN_SITE="yoursite.atlassian.net"

# Scan a site
kspec scan atlassian site yoursite.atlassian.net -f policies/atlassian-security.yml

# Scan organization (requires org ID)
export ATLASSIAN_ORG_ID="your-org-id"
kspec scan atlassian org <org-id> -f policies/atlassian-security.yml
```

## Prerequisites

- Atlassian Cloud account
- API Token for authentication
- Appropriate permissions for resources to scan

## Authentication

### Create an API Token

1. Log in to [Atlassian Account](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **Create API token**
3. Enter a label (e.g., "kspec security scanning")
4. Click **Create**
5. Copy the token and store it securely

### Configuration Methods

**Environment Variables (Recommended):**
```bash
export ATLASSIAN_EMAIL="your-email@example.com"
export ATLASSIAN_API_TOKEN="your-api-token"
export ATLASSIAN_SITE="yoursite.atlassian.net"

# Optional: For Admin API access
export ATLASSIAN_ORG_ID="your-org-id"

kspec scan atlassian site $ATLASSIAN_SITE -f policy.yml
```

**Command Line Flags:**
```bash
kspec scan atlassian site yoursite.atlassian.net \
  --email "your-email@example.com" \
  --api-token "your-api-token" \
  -f policy.yml
```

### Finding Your Organization ID

For Admin API features, you need your Organization ID:

1. Go to [admin.atlassian.com](https://admin.atlassian.com)
2. Select your organization
3. The Organization ID is in the URL: `https://admin.atlassian.com/o/{org-id}/...`

---

## Resources

The Atlassian provider discovers the following resources:

### Jira Resources

| Resource | Description |
|----------|-------------|
| `atlassian_jira_project` | Jira projects with lead and insight info |
| `atlassian_jira_permission_scheme` | Permission schemes and grants |
| `atlassian_jira_security_scheme` | Issue security schemes |
| `atlassian_jira_user` | Jira users and account types |

### Confluence Resources

| Resource | Description |
|----------|-------------|
| `atlassian_confluence_space` | Spaces with permissions |
| `atlassian_confluence_page` | Pages with metadata |

### Admin Resources

| Resource | Description |
|----------|-------------|
| `atlassian_organization` | Organization details and domains |
| `atlassian_admin_user` | Organization users with product access |
| `atlassian_admin_group` | Organization groups |
| `atlassian_auth_policy` | Authentication policies |

### SCIM Resources

| Resource | Description |
|----------|-------------|
| `atlassian_scim_user` | SCIM-provisioned users |
| `atlassian_scim_group` | SCIM-provisioned groups |

---

## Resource Fields

### atlassian_jira_project

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Project ID |
| `key` | `string` | Project key |
| `name` | `string` | Project name |
| `description` | `string` | Project description |
| `project_type_key` | `string` | Project type |
| `style` | `string` | Project style |
| `is_private` | `bool` | Project is private |
| `lead_account_id` | `string` | Lead's account ID |
| `lead_display_name` | `string` | Lead's display name |
| `lead_email` | `string` | Lead's email address |
| `insight_total_issues` | `int` | Total issue count |
| `insight_last_issue_update` | `string` | Last issue update time |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `has_lead` | `bool` | Project has an assigned lead |

### atlassian_jira_permission_scheme

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int` | Scheme ID |
| `name` | `string` | Scheme name |
| `description` | `string` | Scheme description |
| `permissions` | `[]object` | Permission grants |
| `permission_count` | `int` | Number of permissions |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `has_anonymous_access` | `bool` | Has "anyone" holder type |
| `has_group_anyone_access` | `bool` | Has "anyone" group access |
| `has_application_role_access` | `bool` | Has application role access |
| `is_default` | `bool` | Is default permission scheme |

**permissions Entry Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int` | Permission ID |
| `permission` | `string` | Permission type |
| `holder_type` | `string` | Holder type (`user`, `group`, `anyone`, etc.) |
| `holder_parameter` | `string` | Holder parameter (user ID, group name, etc.) |

### atlassian_confluence_space

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int` | Space ID |
| `key` | `string` | Space key |
| `name` | `string` | Space name |
| `type` | `string` | Space type (`global`, `personal`) |
| `status` | `string` | Space status |
| `permission_count` | `int` | Number of permissions |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `has_anonymous_access` | `bool` | Allows anonymous access |
| `has_unlicensed_access` | `bool` | Allows unlicensed access |
| `is_global` | `bool` | Is global space |
| `is_personal` | `bool` | Is personal space |
| `is_active` | `bool` | Space is active |

### atlassian_jira_user

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | `string` | User account ID |
| `account_type` | `string` | Account type (`atlassian`, `customer`, `app`) |
| `display_name` | `string` | Display name |
| `email_address` | `string` | Email address |
| `active` | `bool` | User is active |
| `locale` | `string` | User locale |
| `timezone` | `string` | User timezone |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_atlassian_account` | `bool` | Is Atlassian account |
| `is_customer_account` | `bool` | Is customer account |
| `is_app_account` | `bool` | Is app account |

### atlassian_admin_user

| Field | Type | Description |
|-------|------|-------------|
| `account_id` | `string` | User account ID |
| `account_type` | `string` | Account type |
| `email` | `string` | Email address |
| `name` | `string` | Display name |
| `account_status` | `string` | Account status |
| `product_access` | `[]object` | Product access list |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_active` | `bool` | Account is active |
| `has_jira_access` | `bool` | Has Jira product access |
| `has_confluence_access` | `bool` | Has Confluence product access |
| `product_count` | `int` | Number of products accessed |

### atlassian_auth_policy

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Policy ID |
| `type` | `string` | Policy type |
| `name` | `string` | Policy name |
| `status` | `string` | Policy status |
| `attributes` | `object` | Policy attributes |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_enabled` | `bool` | Policy is enabled |
| `requires_sso` | `bool` | Requires SSO |
| `requires_2fa` | `bool` | Requires 2FA |

---

## Example Policies

### Jira Project Security

```yaml
queries:
  - uid: projects-have-leads
    title: Projects should have assigned leads
    resource: atlassian_jira_project
    severity: medium
    query: resource.has_lead == true
    docs: |
      Every project should have an assigned lead for accountability.
    remediation: |
      Assign a project lead in Project Settings > Details.

  - uid: no-private-projects
    title: Projects should be visible to organization
    resource: atlassian_jira_project
    severity: low
    query: resource.is_private == false
    docs: |
      Private projects may hide important work from the organization.
    remediation: |
      Review project visibility in Project Settings > Access.
```

### Permission Scheme Security

```yaml
queries:
  - uid: no-anonymous-jira-access
    title: Permission schemes should not allow anonymous access
    resource: atlassian_jira_permission_scheme
    severity: critical
    query: resource.has_anonymous_access == false
    docs: |
      Anonymous access allows unauthenticated users to view Jira data.
    remediation: |
      Remove "anyone" holder from permission scheme grants.

  - uid: no-group-anyone-access
    title: Permission schemes should not use "anyone" group
    resource: atlassian_jira_permission_scheme
    severity: high
    query: resource.has_group_anyone_access == false
    docs: |
      The "anyone" group grants access to all authenticated users.
    remediation: |
      Replace "anyone" group with specific groups.

  - uid: review-default-scheme
    title: Review default permission scheme
    resource: atlassian_jira_permission_scheme
    severity: medium
    query: |
      resource.is_default == false ||
      resource.has_anonymous_access == false
    docs: |
      The default scheme should have restrictive permissions.
    remediation: |
      Review and restrict permissions in the default scheme.
```

### Confluence Space Security

```yaml
queries:
  - uid: no-anonymous-confluence-access
    title: Spaces should not allow anonymous access
    resource: atlassian_confluence_space
    severity: critical
    query: resource.has_anonymous_access == false
    docs: |
      Anonymous access exposes content to unauthenticated users.
    remediation: |
      Remove anonymous access in Space Settings > Permissions.

  - uid: no-unlicensed-access
    title: Spaces should not allow unlicensed access
    resource: atlassian_confluence_space
    severity: high
    query: resource.has_unlicensed_access == false
    docs: |
      Unlicensed access allows users without licenses to view content.
    remediation: |
      Remove unlicensed access in Space Settings > Permissions.

  - uid: global-spaces-reviewed
    title: Global spaces should be reviewed
    resource: atlassian_confluence_space
    severity: low
    query: |
      resource.is_global == false ||
      (resource.has_anonymous_access == false &&
       resource.has_unlicensed_access == false)
    docs: |
      Global spaces are visible to all users and need careful review.
    remediation: |
      Review permissions for all global spaces.
```

### Organization Security

```yaml
queries:
  - uid: users-have-product-access
    title: Users should have appropriate product access
    resource: atlassian_admin_user
    severity: low
    query: resource.product_count > 0
    docs: |
      Users without product access may be unnecessary.
    remediation: |
      Remove users without any product access.

  - uid: sso-required
    title: SSO should be required
    resource: atlassian_auth_policy
    severity: high
    query: |
      resource.type != "SSO" ||
      resource.is_enabled == true
    docs: |
      SSO provides centralized authentication and security.
    remediation: |
      Enable SSO in organization security settings.

  - uid: 2fa-required
    title: Two-factor authentication should be required
    resource: atlassian_auth_policy
    severity: high
    query: |
      resource.type != "2FA" ||
      resource.requires_2fa == true
    docs: |
      2FA adds an extra layer of security for user accounts.
    remediation: |
      Enable 2FA requirement in organization settings.
```

---

## CLI Reference

```bash
# Scan a site (Jira/Confluence)
kspec scan atlassian site <site-name> -f <policy-file>

# Scan organization (Admin APIs)
kspec scan atlassian org <org-id> -f <policy-file>

# Using aliases
kspec scan jira site <site-name> -f <policy-file>
kspec scan confluence site <site-name> -f <policy-file>

# With explicit credentials
kspec scan atlassian site <site-name> \
  --email <email> \
  --api-token <token> \
  -f <policy-file>
```

### Options

| Flag | Description |
|------|-------------|
| `-f, --policy` | Policy file to use |
| `-d, --policy-dir` | Directory containing policy files |
| `--email` | Atlassian account email |
| `--api-token` | API token |
| `--site` | Site URL (e.g., yoursite.atlassian.net) |
| `--org-id` | Organization ID (for Admin APIs) |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ATLASSIAN_EMAIL` | Account email |
| `ATLASSIAN_API_TOKEN` | API token |
| `ATLASSIAN_SITE` | Site URL |
| `ATLASSIAN_ORG_ID` | Organization ID |

---

## Required Permissions

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

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Atlassian Security Scan

on:
  schedule:
    - cron: '0 6 * * *'
  workflow_dispatch:

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build kspec
        run: go build -o kspec ./cmd/kspec

      - name: Run Atlassian scan
        env:
          ATLASSIAN_EMAIL: ${{ secrets.ATLASSIAN_EMAIL }}
          ATLASSIAN_API_TOKEN: ${{ secrets.ATLASSIAN_API_TOKEN }}
          ATLASSIAN_SITE: ${{ vars.ATLASSIAN_SITE }}
        run: |
          ./kspec scan atlassian site $ATLASSIAN_SITE \
            -f policies/atlassian-security.yml
```

### GitLab CI

```yaml
atlassian-scan:
  image: golang:1.21
  script:
    - go build -o kspec ./cmd/kspec
    - ./kspec scan atlassian site $ATLASSIAN_SITE -f policies/atlassian-security.yml
  variables:
    ATLASSIAN_EMAIL: $ATLASSIAN_EMAIL
    ATLASSIAN_API_TOKEN: $ATLASSIAN_API_TOKEN
    ATLASSIAN_SITE: $ATLASSIAN_SITE
```

---

## Troubleshooting

### No Valid Credentials

```
Error: atlassian: no valid credentials provided
```

**Solutions:**
- Set `ATLASSIAN_EMAIL` and `ATLASSIAN_API_TOKEN`
- Or use `--email` and `--api-token` flags

### Site Not Specified

```
Error: atlassian: site not specified
```

**Solutions:**
- Set `ATLASSIAN_SITE` environment variable
- Or use `--site` flag
- Or provide site as argument

### 401 Unauthorized

**Causes:**
- Invalid or expired API token
- Wrong email address
- Token lacks permissions

**Solutions:**
- Verify API token is valid
- Ensure email matches the account
- Check token has required permissions

### 403 Forbidden

**Causes:**
- Insufficient permissions
- Admin API requires org admin access

**Solutions:**
- Request appropriate permissions
- For Admin API, ensure org admin access

### Rate Limiting

**Causes:**
- Too many API requests
- Large number of resources

**Solutions:**
- Add delays between scans
- Scan during off-peak hours
- Reduce scan scope

---

## Best Practices

1. **Use API Tokens**: Always use API tokens instead of passwords
2. **Least Privilege**: Create tokens with only required permissions
3. **Regular Scans**: Schedule scans to detect configuration drift
4. **Token Rotation**: Rotate tokens every 90 days
5. **Secure Storage**: Store tokens in secret management systems
6. **Audit Logging**: Enable Atlassian audit logs to track API access
