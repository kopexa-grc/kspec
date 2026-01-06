# Cloudflare Provider Setup

This guide explains how to configure kspec to scan your Cloudflare resources for security compliance.

## Prerequisites

- A Cloudflare account
- API Token with appropriate permissions (recommended) or API Key + Email
- `kspec` installed on your system

## Authentication

### Method 1: API Token (Recommended)

API Tokens provide fine-grained access control and are the recommended authentication method.

1. Log in to the [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Go to **My Profile** > **API Tokens**
3. Click **Create Token**
4. Use **Create Custom Token** with these permissions:

| Permission | Access Level |
|------------|--------------|
| Account > Account Settings | Read |
| Zone > Zone | Read |
| Zone > Zone Settings | Read |
| Zone > DNS | Read |
| Zone > Firewall Services | Read |
| Account > Workers Scripts | Read |
| Account > Cloudflare Pages | Read |
| Account > Cloudflare Tunnel | Read |
| Account > Access: Apps and Policies | Read |
| Account > R2 | Read |

5. Click **Continue to summary** and **Create Token**
6. Copy the token and store it securely

**Using the token:**

```bash
# Via environment variable (recommended)
export CLOUDFLARE_API_TOKEN="your-api-token"
kspec scan cloudflare account -f policies/cloudflare-security.yml

# Via command line flag
kspec scan cloudflare account --api-token "your-api-token" -f policies/cloudflare-security.yml
```

### Method 2: API Key + Email (Legacy)

This method uses your Global API Key and grants full account access. Use API Tokens instead when possible.

1. Log in to the [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Go to **My Profile** > **API Tokens**
3. View your **Global API Key**

```bash
# Via environment variables
export CLOUDFLARE_API_KEY="your-api-key"
export CLOUDFLARE_EMAIL="your-email@example.com"
kspec scan cloudflare account -f policies/cloudflare-security.yml

# Via command line flags
kspec scan cloudflare account --api-key "your-api-key" --email "your-email@example.com" -f policies/cloudflare-security.yml
```

## Scanning Commands

### Scan All Accounts

Scan all Cloudflare accounts accessible with your credentials:

```bash
kspec scan cloudflare account -f policies/cloudflare-security.yml
```

### Scan a Specific Account

```bash
kspec scan cloudflare account <account-id> -f policies/cloudflare-security.yml
```

### Scan a Specific Zone

Scan a specific domain/zone by ID:

```bash
kspec scan cloudflare zone <zone-id> -f policies/cloudflare-security.yml
```

To find your zone ID:
1. Go to the Cloudflare Dashboard
2. Select your domain
3. The Zone ID is displayed on the right side of the Overview page

## Resources Discovered

kspec discovers the following Cloudflare resources:

| Resource Type | Description |
|---------------|-------------|
| `cloudflare_account` | Account information and settings |
| `cloudflare_zone` | Domain/zone configuration |
| `cloudflare_zone_settings` | Zone security settings |
| `cloudflare_dns_record` | DNS records (A, AAAA, CNAME, MX, TXT, etc.) |
| `cloudflare_waf_rule` | WAF managed rulesets |
| `cloudflare_firewall_rule` | Custom firewall rules |
| `cloudflare_r2_bucket` | R2 object storage buckets |
| `cloudflare_worker` | Workers scripts |
| `cloudflare_pages_project` | Pages static site projects |
| `cloudflare_tunnel` | Cloudflare Tunnels (Argo) |
| `cloudflare_access_application` | Zero Trust Access applications |
| `cloudflare_access_policy` | Zero Trust Access policies |

## Example Policy

Create a policy file `cloudflare-security.yml`:

```yaml
policies:
  - uid: cloudflare-security
    name: Cloudflare Security Policy
    version: "1.0"

    require:
      - provider: cloudflare

    groups:
      - title: Zone Security
        filter: asset.type == "cloudflare-zone"
        checks:
          - uid: cf-zone-active
            title: Zone should be active
            resource: cloudflare_zone
            query: resource.status == "active"
            severity: high

      - title: DNS Security
        filter: asset.type == "cloudflare-zone"
        checks:
          - uid: cf-dns-no-wildcard
            title: Avoid wildcard DNS records
            resource: cloudflare_dns_record
            query: "!resource.is_wildcard"
            severity: medium
            docs:
              desc: Wildcard DNS records can expose unintended subdomains
              remediation: Remove wildcard records and create specific records instead
```

## Troubleshooting

### Error: No valid credentials provided

Ensure you have set either:
- `CLOUDFLARE_API_TOKEN` environment variable, or
- `CLOUDFLARE_API_KEY` and `CLOUDFLARE_EMAIL` environment variables

### Error: Permission denied

Your API token may not have the required permissions. Check that your token has the permissions listed in the Authentication section.

### Error: Zone not found

Verify the zone ID is correct. You can find it in the Cloudflare Dashboard under your domain's Overview page.

### Rate limiting

Cloudflare has API rate limits. If you encounter rate limiting:
- Use API tokens with only the permissions you need
- Scan specific zones instead of all zones
- Add delays between scans if running in CI/CD

## CI/CD Integration

### GitHub Actions

```yaml
name: Cloudflare Security Scan

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

      - name: Run Cloudflare scan
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
        run: |
          kspec scan cloudflare account -f policies/cloudflare-security.yml
```

### GitLab CI

```yaml
cloudflare-scan:
  image: golang:1.21
  script:
    - go install github.com/kopexa-grc/kspec/cmd/kspec@latest
    - kspec scan cloudflare account -f policies/cloudflare-security.yml
  variables:
    CLOUDFLARE_API_TOKEN: $CLOUDFLARE_API_TOKEN
  only:
    - schedules
```

## Quick Start Script

```bash
#!/bin/bash

# Cloudflare Security Scan Quick Start
# Usage: ./cloudflare-scan.sh

# Check for API token
if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
    echo "Error: CLOUDFLARE_API_TOKEN environment variable not set"
    echo "Get your token from: https://dash.cloudflare.com/profile/api-tokens"
    exit 1
fi

# Run scan
echo "Starting Cloudflare security scan..."
kspec scan cloudflare account -f policies/cloudflare-security.yml

echo "Scan complete!"
```

## Best Practices

1. **Use API Tokens**: Always prefer API Tokens over Global API Keys for better security
2. **Least Privilege**: Create tokens with only the permissions needed for scanning
3. **Regular Scans**: Schedule regular scans to catch configuration drift
4. **Zone-Specific Tokens**: For production environments, create zone-specific tokens
5. **Audit Logging**: Enable Cloudflare Audit Logs to track API access
6. **Token Rotation**: Rotate API tokens regularly (every 90 days recommended)

## Additional Resources

- [Cloudflare API Documentation](https://developers.cloudflare.com/api/)
- [Cloudflare API Token Permissions](https://developers.cloudflare.com/fundamentals/api/reference/permissions/)
- [Cloudflare Security Best Practices](https://developers.cloudflare.com/fundamentals/reference/policies-compliances/security-best-practices/)
