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
| `-f, --policy` | Path to policy YAML file |
| `-d, --policy-dir` | Path to policy directory |
| `--credential-type` | Credential type (`bearer`, `env`, `password`) |
| `--env-var` | Environment variable for token (default: `GITHUB_TOKEN`) |
| `--token` | Authentication token |

## Commands

### scan

Scan resources using policy checks.

```bash
kspec scan <provider> [resource-type] [target] [flags]
```

#### Providers

| Provider | Resource Types | Example |
|----------|---------------|---------|
| `local` | - | `kspec scan local -f policy.yml` |
| `host` | - | `kspec scan host example.com -f policy.yml` |
| `github` | `org`, `repo` | `kspec scan github org my-org -f policy.yml` |
| `azure` | `subscription` | `kspec scan azure subscription <id> -f policy.yml` |
| `ms365` | `tenant` | `kspec scan ms365 tenant <id> -f policy.yml` |
| `cloudflare` | `account`, `zone` | `kspec scan cloudflare account -f policy.yml` |
| `atlassian` | `site`, `org` | `kspec scan atlassian site site.atlassian.net -f policy.yml` |
| `hetzner` | `project` | `kspec scan hetzner project -f policy.yml` |
| `sbom` | `file`, `dir` | `kspec scan sbom file ./sbom.json -f policy.yml` |

#### Provider Aliases

| Alias | Provider |
|-------|----------|
| `hcloud` | `hetzner` |
| `jira` | `atlassian` |
| `confluence` | `atlassian` |

---

## Provider-Specific Options

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
# Scan organization
export GITHUB_TOKEN="ghp_xxx"
kspec scan github org kopexa-grc -f policies/github-security.yml

# Scan repository
kspec scan github repo kopexa-grc/kspec --token ghp_xxx -f policy.yml
```

---

### Azure

```bash
kspec scan azure subscription <subscription-id> [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--tenant-id` | `AZURE_TENANT_ID` | Azure tenant ID |
| `--client-id` | `AZURE_CLIENT_ID` | Service principal client ID |
| `--token` | `AZURE_CLIENT_SECRET` | Client secret |
| `--resource-group` | - | Filter to resource group |
| `--credential-type` | - | Credential type (`env`) |

**Examples:**

```bash
# Using environment variables
export AZURE_TENANT_ID="tenant-id"
export AZURE_CLIENT_ID="client-id"
export AZURE_CLIENT_SECRET="secret"
kspec scan azure subscription 00000000-0000-0000-0000-000000000000 -f policy.yml

# Using flags
kspec scan azure subscription <sub-id> \
  --tenant-id <tenant-id> \
  --client-id <client-id> \
  --token <secret> \
  -f policy.yml

# Filter by resource group
kspec scan azure subscription <sub-id> \
  --resource-group my-rg \
  -f policy.yml
```

---

### Microsoft 365

```bash
kspec scan ms365 tenant <tenant-id> [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--client-id` | `AZURE_CLIENT_ID` | Application client ID |
| `--client-secret` | `AZURE_CLIENT_SECRET` | Client secret |

**Examples:**

```bash
# Using environment variables
export AZURE_TENANT_ID="tenant-id"
export AZURE_CLIENT_ID="client-id"
export AZURE_CLIENT_SECRET="secret"
kspec scan ms365 tenant $AZURE_TENANT_ID -f policy.yml

# Using flags
kspec scan ms365 tenant <tenant-id> \
  --client-id <client-id> \
  --client-secret <secret> \
  -f policy.yml
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
| `--account-id` | `CLOUDFLARE_ACCOUNT_ID` | Account ID |

**Examples:**

```bash
# Using API token
export CLOUDFLARE_API_TOKEN="token"
kspec scan cloudflare account -f policy.yml

# Scan specific zone
kspec scan cloudflare zone abc123 --api-token token -f policy.yml

# Using legacy API key
kspec scan cloudflare account \
  --api-key key \
  --email user@example.com \
  -f policy.yml
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
| `--site` | `ATLASSIAN_SITE` | Site URL |
| `--org-id` | `ATLASSIAN_ORG_ID` | Organization ID |

**Examples:**

```bash
# Using environment variables
export ATLASSIAN_EMAIL="user@example.com"
export ATLASSIAN_API_TOKEN="token"
kspec scan atlassian site mysite.atlassian.net -f policy.yml

# Using flags
kspec scan atlassian site mysite.atlassian.net \
  --email user@example.com \
  --api-token token \
  -f policy.yml

# Scan organization (Admin APIs)
kspec scan atlassian org abc123 \
  --site mysite.atlassian.net \
  -f policy.yml
```

---

### Hetzner Cloud

```bash
kspec scan hetzner project [project-name] [flags]
kspec scan hcloud project [project-name] [flags]
```

| Flag | Env Var | Description |
|------|---------|-------------|
| `--api-token` | `HCLOUD_TOKEN` | API token |
| `--hcloud-token` | `HETZNER_API_TOKEN` | Alternative token var |
| `--project` | - | Project name (identification) |

**Examples:**

```bash
# Using environment variable
export HCLOUD_TOKEN="token"
kspec scan hetzner project -f policy.yml

# With project name
kspec scan hetzner project my-project --api-token token -f policy.yml

# Using alias
kspec scan hcloud project -f policy.yml
```

---

### Network/Host

```bash
kspec scan host <hostname> [flags]
```

No authentication required for public endpoints.

**Examples:**

```bash
# Scan TLS and HTTP security
kspec scan host example.com -f policies/tls_security.yaml

# Scan with multiple policies
kspec scan host example.com -d policies/
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
kspec scan sbom file ./sbom.json -f policy.yml

# Scan directory of SBOMs
kspec scan sbom dir ./sboms/ -f policy.yml
```

---

## Policy Options

| Flag | Description |
|------|-------------|
| `-f, --policy` | Single policy file |
| `-d, --policy-dir` | Directory of policy files |

**Examples:**

```bash
# Single policy file
kspec scan host example.com -f policies/tls_security.yaml

# All policies in directory
kspec scan host example.com -d policies/

# Multiple policy files (run separately)
kspec scan host example.com -f policies/tls.yaml
kspec scan host example.com -f policies/http.yaml
```

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
| `GITHUB_TOKEN` | GitHub | Personal Access Token |
| `AZURE_TENANT_ID` | Azure, MS365 | Tenant ID |
| `AZURE_CLIENT_ID` | Azure, MS365 | Client ID |
| `AZURE_CLIENT_SECRET` | Azure, MS365 | Client secret |
| `CLOUDFLARE_API_TOKEN` | Cloudflare | API token |
| `CLOUDFLARE_API_KEY` | Cloudflare | API key (legacy) |
| `CLOUDFLARE_EMAIL` | Cloudflare | Account email |
| `ATLASSIAN_EMAIL` | Atlassian | Account email |
| `ATLASSIAN_API_TOKEN` | Atlassian | API token |
| `ATLASSIAN_SITE` | Atlassian | Site URL |
| `ATLASSIAN_ORG_ID` | Atlassian | Organization ID |
| `HCLOUD_TOKEN` | Hetzner | API token |
| `HETZNER_API_TOKEN` | Hetzner | API token (alternative) |
