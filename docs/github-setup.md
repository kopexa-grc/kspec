# GitHub Provider Setup Guide

This guide explains how to configure and use the GitHub provider in kspec to scan your GitHub organizations and repositories for security compliance.

## Prerequisites

1. **GitHub Account**: Access to the GitHub organization or repositories you want to scan
2. **Personal Access Token (PAT)**: A GitHub token with appropriate permissions

## Authentication Methods

### Method 1: Personal Access Token (Recommended)

#### Creating a Personal Access Token

1. Go to **GitHub Settings** > **Developer settings** > **Personal access tokens** > **Tokens (classic)**
2. Click **Generate new token (classic)**
3. Set the following scopes based on your needs:

**For Organization Scanning:**
- `read:org` - Read organization membership and teams
- `repo` - Full control of private repositories (or `public_repo` for public only)
- `read:user` - Read user profile data

**For Repository Scanning:**
- `repo` - Full control of private repositories
- `read:user` - Read user profile data

4. Click **Generate token** and copy the token

#### Using the Token

**Option A: Environment Variable (Recommended)**

```bash
# Set the token as an environment variable
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"

# Scan an organization
kspec scan github org <organization-name> -f policies/github-security.yml

# Scan a repository
kspec scan github repo <owner>/<repo> -f policies/github-security.yml
```

**Option B: Command Line Flag**

```bash
# Using --token flag
kspec scan github org <organization-name> \
  --token "ghp_xxxxxxxxxxxxxxxxxxxx" \
  -f policies/github-security.yml
```

### Method 2: GitHub App (Enterprise)

For larger organizations, consider using a GitHub App for authentication:

1. Create a GitHub App in your organization settings
2. Grant necessary permissions (Repository: Read, Organization: Read)
3. Install the app on your organization
4. Generate a private key and use it for authentication

## Scanning Commands

### Scan an Organization

Scans all repositories, teams, and organization settings:

```bash
kspec scan github org <organization-name> -f policies/github-security.yml
```

**Example:**
```bash
kspec scan github org kopexa-grc -f policies/github-security.yml
```

### Scan a Single Repository

Scans a specific repository and its branches:

```bash
kspec scan github repo <owner>/<repo> -f policies/github-security.yml
```

**Example:**
```bash
kspec scan github repo kopexa-grc/kspec -f policies/github-security.yml
```

## Resources Discovered

The GitHub provider discovers and scans the following resources:

| Resource Type | Description |
|--------------|-------------|
| `github_organization` | Organization settings, security configuration |
| `github_repo` | Repository settings, branch protection, security features |
| `github_branch` | Default branch protection rules |
| `github_team` | Team membership and permissions |

## Creating Security Policies

### Basic Policy Structure

```yaml
policies:
  - uid: github-security
    name: GitHub Security Policy
    version: 1.0.0
    require:
      - provider: github
    groups:
      - title: GitHub Security
        checks:
          - uid: check-branch-protection
          - uid: check-2fa-required

queries:
  - uid: check-branch-protection
    title: Ensure branch protection is enabled on default branch
    resource: github_branch
    impact: 90
    query: |
      has(resource.protected) && resource.protected == true
    docs:
      desc: Branch protection prevents force pushes and requires reviews.
      remediation: Enable branch protection in repository settings.

  - uid: check-2fa-required
    title: Ensure two-factor authentication is required
    resource: github_organization
    impact: 100
    query: |
      has(resource.two_factor_requirement_enabled) &&
      resource.two_factor_requirement_enabled == true
    docs:
      desc: 2FA adds an extra layer of security for organization members.
      remediation: Enable 2FA requirement in organization security settings.
```

### Common Security Checks

**Repository Security:**
```yaml
# Require code review before merging
- uid: require-code-review
  title: Ensure pull request reviews are required
  resource: github_branch
  query: |
    !has(resource.protection) ||
    (has(resource.protection.required_pull_request_reviews) &&
     resource.protection.required_pull_request_reviews.required_approving_review_count >= 1)

# Prevent force pushes
- uid: prevent-force-push
  title: Ensure force pushes are disabled
  resource: github_branch
  query: |
    !has(resource.protection) ||
    resource.protection.allow_force_pushes.enabled == false

# Require signed commits
- uid: require-signed-commits
  title: Ensure signed commits are required
  resource: github_branch
  query: |
    !has(resource.protection) ||
    resource.protection.required_signatures.enabled == true
```

**Organization Security:**
```yaml
# Private repositories by default
- uid: default-repo-permission
  title: Ensure default repository permission is restrictive
  resource: github_organization
  query: |
    resource.default_repository_permission == "read" ||
    resource.default_repository_permission == "none"

# Verify member privileges
- uid: member-can-create-repos
  title: Ensure members cannot create public repositories
  resource: github_organization
  query: |
    resource.members_can_create_public_repositories == false
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITHUB_TOKEN` | Personal Access Token | - |
| `GITHUB_API_URL` | GitHub API URL (for Enterprise) | `https://api.github.com` |

## Troubleshooting

### Authentication Errors

```
Error: 401 Unauthorized
```

**Solutions:**
- Verify your token is correct and not expired
- Check that the token has the required scopes
- Ensure the token has access to the organization/repository

### Rate Limiting

```
Error: API rate limit exceeded
```

**Solutions:**
- Wait for the rate limit to reset (check `X-RateLimit-Reset` header)
- Use a GitHub App for higher rate limits
- Reduce the scope of your scan

### Permission Errors

```
Error: Resource not accessible by integration
```

**Solutions:**
- Verify the token has the required scopes
- Check organization settings for third-party access
- Ensure you're a member of the organization

### No Resources Found

If the scan returns no resources:

1. Verify the organization/repository name is correct
2. Check that your token has access to the target
3. Ensure the organization has repositories/teams to scan

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/security-scan.yml
name: Security Scan

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight

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

      - name: Run GitHub Security Scan
        env:
          GITHUB_TOKEN: ${{ secrets.KSPEC_GITHUB_TOKEN }}
        run: |
          ./kspec scan github org ${{ github.repository_owner }} \
            -f policies/github-security.yml
```

### GitLab CI

```yaml
# .gitlab-ci.yml
security-scan:
  image: golang:1.21
  script:
    - go build -o kspec ./cmd/kspec
    - ./kspec scan github org ${GITHUB_ORG} -f policies/github-security.yml
  variables:
    GITHUB_TOKEN: ${GITHUB_TOKEN}
```

## Best Practices

1. **Use Fine-Grained Tokens**: When possible, use fine-grained PATs with minimal required permissions
2. **Rotate Tokens Regularly**: Set up token rotation schedules for security
3. **Store Secrets Securely**: Use secret management tools, never commit tokens to repositories
4. **Scan Regularly**: Set up scheduled scans to detect configuration drift
5. **Review Results**: Regularly review scan results and remediate findings

## Quick Start Script

```bash
#!/bin/bash
# github-scan.sh

# Configuration
GITHUB_ORG="${1:-your-org-name}"
POLICY_FILE="${2:-policies/github-security.yml}"

# Check for token
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN environment variable is not set"
    echo "Usage: export GITHUB_TOKEN=ghp_xxx && ./github-scan.sh [org-name] [policy-file]"
    exit 1
fi

# Build if needed
if [ ! -f "./kspec" ]; then
    echo "Building kspec..."
    go build -o kspec ./cmd/kspec
fi

# Run scan
echo "Scanning GitHub organization: $GITHUB_ORG"
./kspec scan github org "$GITHUB_ORG" -f "$POLICY_FILE"
```

Make executable and run:
```bash
chmod +x github-scan.sh
export GITHUB_TOKEN="ghp_xxxx"
./github-scan.sh my-org policies/github-security.yml
```
