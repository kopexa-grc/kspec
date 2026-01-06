# Cloudflare Provider

The Cloudflare provider scans your Cloudflare accounts and zones for security compliance, covering DNS, WAF, Zero Trust, Workers, Pages, R2, and Tunnels.

## Overview

Use the Cloudflare provider to validate:
- Zone security settings (SSL, development mode, status)
- DNS record configuration (SPF, DKIM, DMARC, wildcards)
- WAF managed and custom rulesets
- Zero Trust Access applications and policies
- Workers and Pages projects
- R2 storage buckets
- Cloudflare Tunnels

## Quick Start

```bash
# Set your API token
export CLOUDFLARE_API_TOKEN="your-api-token"

# Scan all accounts
kspec scan cloudflare account -f policies/cloudflare-security.yml

# Scan a specific zone
kspec scan cloudflare zone <zone-id> -f policies/cloudflare-security.yml
```

## Prerequisites

- Cloudflare account
- API Token with appropriate permissions (recommended) or API Key + Email

## Authentication

### Method 1: API Token (Recommended)

API Tokens provide fine-grained access control.

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

**Using the token:**

```bash
# Via environment variable (recommended)
export CLOUDFLARE_API_TOKEN="your-api-token"
kspec scan cloudflare account -f policy.yml

# Via command line flag
kspec scan cloudflare account --api-token "your-api-token" -f policy.yml
```

### Method 2: API Key + Email (Legacy)

This method uses your Global API Key and grants full account access.

```bash
# Via environment variables
export CLOUDFLARE_API_KEY="your-api-key"
export CLOUDFLARE_EMAIL="your-email@example.com"
kspec scan cloudflare account -f policy.yml

# Via command line flags
kspec scan cloudflare account \
  --api-key "your-api-key" \
  --email "your-email@example.com" \
  -f policy.yml
```

---

## Resources

The Cloudflare provider discovers the following resources:

### Account & Zone

| Resource | Description |
|----------|-------------|
| `cloudflare_account` | Account information and settings |
| `cloudflare_zone` | Domain/zone configuration |
| `cloudflare_zone_settings` | Zone security settings |

### DNS

| Resource | Description |
|----------|-------------|
| `cloudflare_dns_record` | DNS records (A, AAAA, CNAME, MX, TXT, etc.) |

### Security

| Resource | Description |
|----------|-------------|
| `cloudflare_waf_rule` | WAF managed rulesets |
| `cloudflare_firewall_rule` | Custom firewall rules |

### Zero Trust

| Resource | Description |
|----------|-------------|
| `cloudflare_access_application` | Zero Trust Access applications |
| `cloudflare_access_policy` | Zero Trust Access policies |

### Platform

| Resource | Description |
|----------|-------------|
| `cloudflare_worker` | Workers scripts |
| `cloudflare_pages_project` | Pages static site projects |
| `cloudflare_r2_bucket` | R2 object storage buckets |
| `cloudflare_tunnel` | Cloudflare Tunnels (Argo) |

---

## Resource Fields

### cloudflare_zone

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Zone ID |
| `name` | `string` | Domain name |
| `status` | `string` | Zone status (`active`, `pending`, etc.) |
| `paused` | `bool` | Zone is paused |
| `type` | `string` | Zone type (`full`, `partial`) |
| `development_mode` | `int` | Development mode seconds remaining |
| `name_servers` | `[]string` | Cloudflare nameservers |
| `plan` | `object` | Current plan details |

### cloudflare_zone_settings

| Field | Type | Description |
|-------|------|-------------|
| `zone_id` | `string` | Zone ID |
| `status` | `string` | Zone status |
| `is_active` | `bool` | Zone is active (computed) |
| `is_pending` | `bool` | Zone is pending (computed) |
| `is_paused` | `bool` | Zone is paused (computed) |
| `plan_name` | `string` | Plan name |
| `is_free_plan` | `bool` | On free plan (computed) |
| `is_pro_plan` | `bool` | On Pro plan (computed) |
| `is_business_plan` | `bool` | On Business plan (computed) |
| `is_enterprise_plan` | `bool` | On Enterprise plan (computed) |
| `development_mode_enabled` | `bool` | Development mode active (computed) |

### cloudflare_dns_record

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Record ID |
| `name` | `string` | Record name |
| `type` | `string` | Record type (A, AAAA, CNAME, MX, TXT, etc.) |
| `content` | `string` | Record content/value |
| `ttl` | `int` | TTL in seconds |
| `proxied` | `bool` | Proxied through Cloudflare |
| `priority` | `int` | MX record priority |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_spf` | `bool` | Is SPF record |
| `is_dkim` | `bool` | Is DKIM record |
| `is_dmarc` | `bool` | Is DMARC record |
| `is_mx` | `bool` | Is MX record |
| `is_wildcard` | `bool` | Is wildcard record (*) |
| `is_proxied` | `bool` | Is proxied (orange cloud) |
| `ttl_seconds` | `int` | TTL as integer |
| `ttl_auto` | `bool` | TTL is automatic |

### cloudflare_waf_rule

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Ruleset ID |
| `name` | `string` | Ruleset name |
| `kind` | `string` | Ruleset kind (`managed`, `custom`, `zone`) |
| `phase` | `string` | Execution phase |
| `rules` | `[]object` | Rules in the ruleset |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_managed` | `bool` | Is managed ruleset |
| `is_custom` | `bool` | Is custom ruleset |
| `is_owasp` | `bool` | Is OWASP/Cloudflare managed ruleset |
| `is_http_request_firewall` | `bool` | Is HTTP request firewall phase |

### cloudflare_access_application

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Application ID |
| `name` | `string` | Application name |
| `domain` | `string` | Application domain |
| `type` | `string` | Application type |
| `session_duration` | `string` | Session duration |
| `cors_headers` | `object` | CORS configuration |
| `auto_redirect_to_identity` | `bool` | Auto-redirect to IdP |
| `skip_interstitial` | `bool` | Skip interstitial page |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_self_hosted` | `bool` | Self-hosted application |
| `is_saas` | `bool` | SaaS application |
| `is_ssh` | `bool` | SSH application |
| `is_vnc` | `bool` | VNC application |
| `is_bookmark` | `bool` | Bookmark application |
| `has_session_limit` | `bool` | Has session duration limit |
| `has_cors_config` | `bool` | Has CORS configuration |
| `cors_allow_all_origins` | `bool` | CORS allows all origins |
| `skips_interstitial` | `bool` | Skips interstitial page |
| `auto_redirects_to_idp` | `bool` | Auto-redirects to IdP |

### cloudflare_access_policy

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Policy ID |
| `name` | `string` | Policy name |
| `decision` | `string` | Policy decision (`allow`, `deny`, `bypass`) |
| `precedence` | `int` | Policy precedence |
| `application_id` | `string` | Parent application ID |
| `include` | `[]object` | Include rules |
| `exclude` | `[]object` | Exclude rules |
| `require` | `[]object` | Require rules |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_allow` | `bool` | Decision is allow |
| `is_deny` | `bool` | Decision is deny |
| `is_bypass` | `bool` | Decision is bypass |
| `requires_mfa` | `bool` | Requires MFA |
| `has_geo_restriction` | `bool` | Has geographic restriction |
| `has_device_posture` | `bool` | Requires device posture |
| `include_rule_count` | `int` | Number of include rules |
| `exclude_rule_count` | `int` | Number of exclude rules |
| `require_rule_count` | `int` | Number of require rules |

---

## Example Policies

### Zone Security

```yaml
queries:
  - uid: zone-active
    title: Zone should be active
    resource: cloudflare_zone
    severity: high
    query: resource.status == "active"
    docs: |
      Zones should be active to serve traffic through Cloudflare.
    remediation: |
      Verify DNS is pointing to Cloudflare nameservers.

  - uid: zone-not-paused
    title: Zone should not be paused
    resource: cloudflare_zone_settings
    severity: high
    query: resource.is_paused == false
    docs: |
      Paused zones bypass Cloudflare protection.
    remediation: |
      Unpause the zone in the Cloudflare dashboard.

  - uid: development-mode-disabled
    title: Development mode should be disabled
    resource: cloudflare_zone_settings
    severity: medium
    query: resource.development_mode_enabled == false
    docs: |
      Development mode bypasses caching and reduces performance.
    remediation: |
      Disable development mode in zone settings.
```

### DNS Security

```yaml
queries:
  - uid: has-spf-record
    title: Domain has SPF record
    resource: cloudflare_dns_record
    severity: high
    query: |
      resource.type != "TXT" || !resource.name.endsWith(resource.zone_name) ||
      resource.is_spf == true
    docs: |
      SPF records prevent email spoofing.
    remediation: |
      Add a TXT record with your SPF policy:
      v=spf1 include:_spf.example.com -all

  - uid: has-dmarc-record
    title: Domain has DMARC record
    resource: cloudflare_dns_record
    severity: high
    query: |
      !resource.name.startsWith("_dmarc.") || resource.is_dmarc == true
    docs: |
      DMARC records define email authentication policy.
    remediation: |
      Add a TXT record at _dmarc.yourdomain.com.

  - uid: no-wildcard-dns
    title: Avoid wildcard DNS records
    resource: cloudflare_dns_record
    severity: medium
    query: resource.is_wildcard == false
    docs: |
      Wildcard records can expose unintended subdomains.
    remediation: |
      Remove wildcard records and create specific records.

  - uid: records-proxied
    title: DNS records should be proxied
    resource: cloudflare_dns_record
    severity: low
    query: |
      resource.type != "A" && resource.type != "AAAA" &&
      resource.type != "CNAME" || resource.is_proxied == true
    docs: |
      Proxied records benefit from Cloudflare protection.
    remediation: |
      Enable the proxy (orange cloud) for the record.
```

### WAF Security

```yaml
queries:
  - uid: has-managed-waf
    title: Zone has managed WAF enabled
    resource: cloudflare_waf_rule
    severity: high
    query: resource.is_managed == true
    docs: |
      Managed WAF rulesets protect against common attacks.
    remediation: |
      Enable Cloudflare Managed Ruleset in Security > WAF.

  - uid: has-owasp-ruleset
    title: OWASP ruleset is enabled
    resource: cloudflare_waf_rule
    severity: high
    query: resource.is_owasp == true
    docs: |
      OWASP ruleset provides protection against OWASP Top 10.
    remediation: |
      Enable Cloudflare OWASP Core Ruleset in Security > WAF.
```

### Zero Trust Security

```yaml
queries:
  - uid: access-app-has-session-limit
    title: Access applications have session limits
    resource: cloudflare_access_application
    severity: medium
    query: resource.has_session_limit == true
    docs: |
      Session limits force re-authentication periodically.
    remediation: |
      Set a session duration in the Access application settings.

  - uid: access-no-cors-allow-all
    title: Access apps should not allow all CORS origins
    resource: cloudflare_access_application
    severity: high
    query: |
      !has(resource.cors_allow_all_origins) ||
      resource.cors_allow_all_origins == false
    docs: |
      Allowing all CORS origins is a security risk.
    remediation: |
      Restrict CORS to specific trusted origins.

  - uid: access-policy-requires-mfa
    title: Access policies should require MFA
    resource: cloudflare_access_policy
    severity: high
    query: |
      resource.is_bypass == true ||
      resource.requires_mfa == true
    docs: |
      MFA provides additional security for sensitive applications.
    remediation: |
      Add MFA requirement to the policy include rules.

  - uid: access-policy-no-bypass
    title: Access policies should not bypass authentication
    resource: cloudflare_access_policy
    severity: critical
    query: resource.is_bypass == false
    docs: |
      Bypass policies skip authentication entirely.
    remediation: |
      Review and remove unnecessary bypass policies.
```

---

## CLI Reference

```bash
# Scan all accounts
kspec scan cloudflare account -f <policy-file>

# Scan a specific account
kspec scan cloudflare account <account-id> -f <policy-file>

# Scan a specific zone
kspec scan cloudflare zone <zone-id> -f <policy-file>

# Using explicit credentials
kspec scan cloudflare account \
  --api-token <api-token> \
  -f <policy-file>
```

### Options

| Flag | Description |
|------|-------------|
| `-f, --policy` | Policy file to use |
| `-d, --policy-dir` | Directory containing policy files |
| `--api-token` | Cloudflare API Token |
| `--api-key` | Cloudflare API Key (legacy) |
| `--email` | Cloudflare account email (with API Key) |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CLOUDFLARE_API_TOKEN` | API Token (recommended) |
| `CLOUDFLARE_API_KEY` | API Key (legacy) |
| `CLOUDFLARE_EMAIL` | Account email (with API Key) |
| `CLOUDFLARE_ACCOUNT_ID` | Account ID (optional) |

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Cloudflare Security Scan

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

      - name: Run Cloudflare scan
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
        run: |
          ./kspec scan cloudflare account \
            -f policies/cloudflare-security.yml
```

### GitLab CI

```yaml
cloudflare-scan:
  image: golang:1.21
  script:
    - go build -o kspec ./cmd/kspec
    - ./kspec scan cloudflare account -f policies/cloudflare-security.yml
  variables:
    CLOUDFLARE_API_TOKEN: $CLOUDFLARE_API_TOKEN
```

---

## Troubleshooting

### No Valid Credentials

```
Error: cloudflare: no valid credentials provided
```

**Solutions:**
- Set `CLOUDFLARE_API_TOKEN` environment variable, or
- Set `CLOUDFLARE_API_KEY` and `CLOUDFLARE_EMAIL`

### Permission Denied

```
Error: permission denied
```

**Solutions:**
- Verify API token has required permissions
- Check token hasn't expired
- Ensure token has access to the zone/account

### Zone Not Found

```
Error: zone not found
```

**Solutions:**
- Verify zone ID is correct (found in zone Overview page)
- Check token has access to the zone

### Rate Limiting

```
Error: rate limit exceeded
```

**Solutions:**
- Use API tokens with only required permissions
- Scan specific zones instead of all zones
- Add delays between scans in CI/CD

---

## Best Practices

1. **Use API Tokens**: Always prefer API Tokens over Global API Keys
2. **Least Privilege**: Create tokens with only required permissions
3. **Zone-Specific Tokens**: For production, use zone-specific tokens
4. **Regular Scans**: Schedule scans to detect configuration drift
5. **Token Rotation**: Rotate tokens every 90 days
6. **Audit Logging**: Enable Cloudflare Audit Logs to track API access
