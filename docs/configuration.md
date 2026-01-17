# Configuration

This guide covers kspec configuration options for scanning and performance tuning.

## CLI Flags

### Common Flags

All scan commands support these flags:

| Flag | Short | Description |
|------|-------|-------------|
| `--policy` | `-f` | Policy file to use |
| `--policy-dir` | `-d` | Directory containing policy files |
| `--export` | `-o` | Export results to file (csv, xlsx, json, html) |
| `--export-format` | | Export format (auto-detected from filename) |
| `--no-ui` | | Disable interactive UI (for CI/CD) |
| `--max-workers` | | Maximum number of concurrent workers (default: 20) |
| `--sequential` | | Disable concurrency (run sequentially) |

### Examples

```bash
# Basic scan with policy file
kspec scan aws -f policies/aws-security.yml

# Scan with directory of policies
kspec scan azure <sub-id> -d policies/

# Export to Excel
kspec scan github org <org-name> -f policies/github.yml -o report.xlsx

# CI/CD mode with limited concurrency
kspec scan aws -f policies/aws-security.yml --no-ui --max-workers 5

# Debug mode (sequential execution)
kspec scan aws -f policies/aws-security.yml --sequential
```

## Concurrency

kspec uses adaptive concurrency to maximize scan performance while respecting API rate limits.

### How It Works

1. **Adaptive Scaling**: Worker count scales based on `runtime.GOMAXPROCS` (typically your CPU cores)
2. **Rate Limiting**: Each provider has built-in rate limits to prevent API throttling
3. **Workload Awareness**: Small scans use fewer workers to avoid overhead

### Parallelization Levels

kspec parallelizes at three levels:

| Level | Description | Example |
|-------|-------------|---------|
| **Discovery** | Resource types discovered concurrently | IAM, S3, EC2 discovered in parallel |
| **Fetching** | Resource instances fetched in parallel | 100 S3 buckets fetched concurrently |
| **Evaluation** | Policy checks run concurrently | CPU-bound, no rate limiting |

### Default Rate Limits

| Provider | Requests/Second | Burst |
|----------|-----------------|-------|
| AWS | 10 | 20 |
| GitHub | 5 | 10 |
| Azure | 10 | 20 |
| Cloudflare | 4 | 8 |
| Microsoft 365 | 10 | 20 |
| Hetzner | 10 | 20 |
| Factorial | 5 | 10 |

These defaults are conservative to prevent throttling. Most provider SDKs also handle 429 responses with automatic retry.

### Tuning Concurrency

**For large scans** (thousands of resources):
```bash
# Increase worker limit for better parallelism
kspec scan aws -f policies/aws-security.yml --max-workers 30
```

**For rate-limited APIs**:
```bash
# Reduce workers to stay under limits
kspec scan github org <org> -f policies/github.yml --max-workers 3
```

**For debugging**:
```bash
# Run sequentially to isolate issues
kspec scan aws -f policies/aws-security.yml --sequential
```

## Environment Variables

### AWS Provider

| Variable | Description |
|----------|-------------|
| `AWS_ACCESS_KEY_ID` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key |
| `AWS_SESSION_TOKEN` | Session token (for temporary credentials) |
| `AWS_REGION` | Default AWS region |
| `AWS_PROFILE` | Named profile from credentials file |

### Azure Provider

| Variable | Description |
|----------|-------------|
| `AZURE_TENANT_ID` | Azure AD tenant ID |
| `AZURE_CLIENT_ID` | Service principal client ID |
| `AZURE_CLIENT_SECRET` | Service principal secret |
| `AZURE_SUBSCRIPTION_ID` | Default subscription ID |

### GitHub Provider

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | Personal access token or GitHub App token |
| `GITHUB_ENTERPRISE_URL` | GitHub Enterprise base URL (optional) |

### Microsoft 365 Provider

| Variable | Description |
|----------|-------------|
| `MS365_TENANT_ID` | Microsoft 365 tenant ID |
| `MS365_CLIENT_ID` | Azure AD application client ID |
| `MS365_CLIENT_SECRET` | Azure AD application secret |

### Hetzner Provider

| Variable | Description |
|----------|-------------|
| `HCLOUD_TOKEN` | Hetzner Cloud API token |

### Cloudflare Provider

| Variable | Description |
|----------|-------------|
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token |
| `CLOUDFLARE_API_KEY` | Cloudflare API key (legacy) |
| `CLOUDFLARE_EMAIL` | Cloudflare account email (with API key) |

## Policy Configuration

Policies are YAML files with the following structure:

```yaml
policies:
  - uid: my-policy
    name: My Security Policy
    version: 1.0.0
    require:
      - provider: aws    # Filter: only run for AWS scans
    groups:
      - title: IAM Security
        filter: asset.type == "aws_account"  # Optional group filter
        checks:
          - uid: iam-mfa-enabled

queries:
  - uid: iam-mfa-enabled
    title: IAM users should have MFA enabled
    resource: aws_iam_user
    impact: 90
    query: |
      has(resource.mfa_active) && resource.mfa_active == true
    docs:
      desc: MFA provides an additional layer of security.
      remediation: Enable MFA for the user in the IAM console.
      audit: Check the MFA column in IAM users list.
```

### Policy Fields

| Field | Description |
|-------|-------------|
| `uid` | Unique identifier for the policy |
| `name` | Human-readable name |
| `version` | Semantic version |
| `require.provider` | Filter policies by provider |
| `groups` | Logical grouping of checks |
| `groups.filter` | CEL expression to filter when group applies |
| `groups.checks` | List of check UIDs to run |

### Query Fields

| Field | Description |
|-------|-------------|
| `uid` | Unique identifier (referenced by checks) |
| `title` | Check title shown in results |
| `resource` | Resource type this check applies to |
| `impact` | Severity score (0-100) |
| `query` | CEL expression that returns true for passing |
| `docs.desc` | Description of the check |
| `docs.remediation` | How to fix failures |
| `docs.audit` | How to manually verify |

See [Writing Policies](policies.md) for detailed policy authoring guidance.
