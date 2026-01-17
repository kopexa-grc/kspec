# CLI Reference

Complete reference for kspec command-line interface.

## Global Options

```bash
kspec [command] [flags]
```

| Flag | Description |
|------|-------------|
| `-h, --help` | Help for kspec |
| `-v, --version` | Print version |

## Commands

### scan

Scan resources using policy checks. Each provider is a subcommand with its own flags and asset types.

```bash
kspec scan <provider> [asset-type] [target] [flags]
```

#### Common Scan Flags

These flags are available for all scan commands:

| Flag | Description |
|------|-------------|
| `-f, --policy` | Path to policy YAML file |
| `-d, --policy-dir` | Path to directory containing policy files |
| `-o, --export` | Export results to file (csv, xlsx, json, html) |
| `--export-format` | Export format (auto-detected from filename) |
| `--no-ui` | Disable interactive UI (useful for CI/CD) |

#### Default Policy Directory

If no `--policy` or `--policy-dir` flag is specified, kspec automatically uses `./policies` as the default directory if it exists. If the directory doesn't exist, an error is shown with instructions.

```bash
# These are equivalent if ./policies/ exists:
kspec scan github org my-org
kspec scan github org my-org -d policies
```

---

## Providers

### Provider Overview

| Provider | Asset Types | Example |
|----------|-------------|---------|
| `aws` | `account` | `kspec scan aws` |
| `azure` | `subscription` | `kspec scan azure <subscription-id>` |
| `github` | `org`, `repo` | `kspec scan github org my-org` |
| `ms365` | `tenant` | `kspec scan ms365 <tenant-id>` |
| `cloudflare` | `account`, `zone` | `kspec scan cloudflare account` |
| `atlassian` | `site`, `org` | `kspec scan atlassian site mysite.atlassian.net` |
| `hetzner` | `project` | `kspec scan hetzner` |
| `factorial` | `company` | `kspec scan factorial` |
| `network` | `host` | `kspec scan network host example.com` |
| `os` | `local` | `kspec scan os` |
| `sbom` | `file`, `dir` | `kspec scan sbom file ./sbom.json` |

### Provider Aliases

| Alias | Provider |
|-------|----------|
| `amazon`, `ec2`, `s3` | `aws` |
| `gh` | `github` |
| `m365`, `microsoft365` | `ms365` |
| `cf` | `cloudflare` |
| `jira`, `confluence` | `atlassian` |
| `hcloud` | `hetzner` |
| `factorial-hr`, `hris` | `factorial` |
| `bom`, `cyclonedx`, `spdx` | `sbom` |

---

## Provider-Specific Options

### AWS

```bash
kspec scan aws [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--profile` | `AWS_PROFILE` | AWS profile from ~/.aws/credentials |
| `--region` | `AWS_REGION` | AWS region |
| `--regions` | - | Multiple regions (comma-separated) |
| `--access-key-id` | `AWS_ACCESS_KEY_ID` | AWS access key ID |
| `--secret-access-key` | `AWS_SECRET_ACCESS_KEY` | AWS secret access key |
| `--session-token` | `AWS_SESSION_TOKEN` | AWS session token |
| `--role-arn` | - | IAM role ARN to assume |
| `--external-id` | - | External ID for assume role |

**Examples:**

```bash
# Using default credentials
kspec scan aws -d policies

# Using specific profile
kspec scan aws --profile production -d policies

# Multi-region scan
kspec scan aws --regions us-east-1,eu-west-1 -d policies

# Cross-account with assume role
kspec scan aws --role-arn arn:aws:iam::123456789:role/SecurityAudit -d policies
```

---

### Azure

```bash
kspec scan azure <subscription-id> [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--tenant-id` | `AZURE_TENANT_ID` | Azure tenant ID |
| `--client-id` | `AZURE_CLIENT_ID` | Service principal client ID |
| `--client-secret` | `AZURE_CLIENT_SECRET` | Client secret |
| `--resource-group` | - | Filter to resource group |

**Examples:**

```bash
# Using environment variables (DefaultAzureCredential)
export AZURE_TENANT_ID="tenant-id"
export AZURE_CLIENT_ID="client-id"
export AZURE_CLIENT_SECRET="secret"
kspec scan azure 00000000-0000-0000-0000-000000000000 -d policies

# Using flags
kspec scan azure <sub-id> \
  --tenant-id <tenant-id> \
  --client-id <client-id> \
  --client-secret <secret> \
  -d policies

# Filter by resource group
kspec scan azure <sub-id> --resource-group my-rg -d policies
```

---

### GitHub

```bash
kspec scan github org <organization> [flags]
kspec scan github repo <owner/repo> [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--token` | `GITHUB_TOKEN` | GitHub Personal Access Token |

**Examples:**

```bash
# Scan organization (uses GITHUB_TOKEN env var)
export GITHUB_TOKEN="ghp_xxx"
kspec scan github org kopexa-grc -d policies

# Scan repository
kspec scan github repo kopexa-grc/kspec -d policies

# With explicit token
kspec scan github org my-org --token ghp_xxx -d policies
```

---

### Microsoft 365

```bash
kspec scan ms365 <tenant-id> [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--tenant-id` | `AZURE_TENANT_ID` | Azure AD tenant ID |
| `--client-id` | `AZURE_CLIENT_ID` | Application client ID |
| `--client-secret` | `AZURE_CLIENT_SECRET` | Client secret |

**Examples:**

```bash
# Using environment variables
export AZURE_TENANT_ID="tenant-id"
export AZURE_CLIENT_ID="client-id"
export AZURE_CLIENT_SECRET="secret"
kspec scan ms365 $AZURE_TENANT_ID -d policies

# Using flags
kspec scan ms365 <tenant-id> \
  --client-id <client-id> \
  --client-secret <secret> \
  -d policies
```

---

### Cloudflare

```bash
kspec scan cloudflare account [account-id] [flags]
kspec scan cloudflare zone <zone-id> [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--api-token` | `CLOUDFLARE_API_TOKEN` | API token (recommended) |
| `--api-key` | `CLOUDFLARE_API_KEY` | API key (legacy) |
| `--email` | `CLOUDFLARE_EMAIL` | Account email (with API key) |
| `--account-id` | - | Account ID |

**Examples:**

```bash
# Using API token
export CLOUDFLARE_API_TOKEN="token"
kspec scan cloudflare account -d policies

# Scan specific zone
kspec scan cloudflare zone abc123 --api-token token -d policies

# Using legacy API key
kspec scan cloudflare account \
  --api-key key \
  --email user@example.com \
  -d policies
```

---

### Atlassian

```bash
kspec scan atlassian site <site-name> [flags]
kspec scan atlassian org <org-id> [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--email` | `ATLASSIAN_EMAIL` | Account email |
| `--api-token` | `ATLASSIAN_API_TOKEN` | API token |
| `--site` | - | Site URL |
| `--org-id` | - | Organization ID |

**Examples:**

```bash
# Using environment variables
export ATLASSIAN_EMAIL="user@example.com"
export ATLASSIAN_API_TOKEN="token"
kspec scan atlassian site mysite.atlassian.net -d policies

# Scan organization (Admin APIs)
kspec scan atlassian org abc123 --site mysite.atlassian.net -d policies
```

---

### Hetzner Cloud

```bash
kspec scan hetzner [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--api-token` | `HCLOUD_TOKEN` | API token |
| `--project` | - | Project name (for identification) |

**Examples:**

```bash
# Using environment variable
export HCLOUD_TOKEN="token"
kspec scan hetzner -d policies

# With project name
kspec scan hetzner --project my-project --api-token token -d policies

# Using alias
kspec scan hcloud -d policies
```

---

### Factorial HR

```bash
kspec scan factorial [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--api-key` | `FACTORIAL_API_KEY` | Factorial API key |
| `--access-token` | - | OAuth access token |

**Examples:**

```bash
# Using API key
export FACTORIAL_API_KEY="key"
kspec scan factorial -d policies

# Using access token
kspec scan factorial --access-token token -d policies
```

---

### Network

```bash
kspec scan network host <hostname> [flags]
```

No authentication required for public endpoints.

**Examples:**

```bash
# Scan TLS and HTTP security
kspec scan network host example.com -d policies

# Scan with single policy file
kspec scan network host example.com -f policies/tls_security.yaml
```

---

### Operating System

```bash
kspec scan os [flags]
```

Scans the local operating system.

**Examples:**

```bash
# Scan local system
kspec scan os -d policies
```

---

### SBOM

```bash
kspec scan sbom file <path> [flags]
kspec scan sbom dir <path> [flags]
```

| Flag | Description |
|------|-------------|
| `--sbom-path` | Path to SBOM file or directory |

**Examples:**

```bash
# Scan single SBOM
kspec scan sbom file ./sbom.json -d policies

# Scan directory of SBOMs
kspec scan sbom dir ./sboms/ -d policies
```

---

## Export Options

Export scan results to various formats:

```bash
# Export to CSV
kspec scan github org my-org -d policies -o results.csv

# Export to Excel
kspec scan github org my-org -d policies -o results.xlsx

# Export to JSON
kspec scan github org my-org -d policies -o results.json

# Export to HTML (visual report)
kspec scan github org my-org -d policies -o results.html

# Explicit format
kspec scan github org my-org -d policies -o results --export-format html
```

### HTML Reports

HTML reports provide a visual, shareable format with:
- Summary statistics and progress bar
- Severity breakdown for failed checks
- Collapsible sections (passed/skipped collapsed by default)
- Dark theme optimized for readability
- No external dependencies (standalone HTML file)

---

## CI/CD Mode

Disable the interactive TUI for CI/CD pipelines:

```bash
kspec scan github org my-org -d policies --no-ui
```

This outputs structured logs instead of the interactive interface.

---

## Output

kspec displays results in an interactive TUI (Terminal User Interface):

- **Discovery phase**: Shows resources found
- **Scanning phase**: Shows check progress
- **Results**: Interactive navigation of findings

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Switch focus between panels |
| `↑/↓` | Navigate items |
| `Enter` | View details |
| `q` | Quit |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All checks passed |
| `1` | One or more checks failed |
| `2` | Error during execution |

---

## Environment Variables Summary

| Variable | Provider | Description |
|----------|----------|-------------|
| `AWS_PROFILE` | AWS | AWS profile name |
| `AWS_REGION` | AWS | AWS region |
| `AWS_ACCESS_KEY_ID` | AWS | Access key ID |
| `AWS_SECRET_ACCESS_KEY` | AWS | Secret access key |
| `AWS_SESSION_TOKEN` | AWS | Session token |
| `GITHUB_TOKEN` | GitHub | Personal Access Token |
| `AZURE_TENANT_ID` | Azure, MS365 | Tenant ID |
| `AZURE_CLIENT_ID` | Azure, MS365 | Client ID |
| `AZURE_CLIENT_SECRET` | Azure, MS365 | Client secret |
| `CLOUDFLARE_API_TOKEN` | Cloudflare | API token |
| `CLOUDFLARE_API_KEY` | Cloudflare | API key (legacy) |
| `CLOUDFLARE_EMAIL` | Cloudflare | Account email |
| `ATLASSIAN_EMAIL` | Atlassian | Account email |
| `ATLASSIAN_API_TOKEN` | Atlassian | API token |
| `HCLOUD_TOKEN` | Hetzner | API token |
| `FACTORIAL_API_KEY` | Factorial | API key |
