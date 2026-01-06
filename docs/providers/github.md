# GitHub Provider

The GitHub provider scans your GitHub organizations and repositories for security compliance, covering organization settings, repository configurations, branch protection rules, and teams.

## Overview

Use the GitHub provider to validate:
- Organization security (2FA requirement, verified domain, default permissions)
- Repository branch protection rules
- Force push restrictions
- Signed commit requirements
- Status check enforcement
- Code review requirements
- Dependabot configuration

## Quick Start

```bash
# Set your GitHub token
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Scan an organization
kspec scan github org <organization-name> -f policies/github-security.yml

# Scan a single repository
kspec scan github repo <owner>/<repo> -f policies/github-security.yml
```

## Prerequisites

- GitHub account with access to the target organization/repository
- Personal Access Token (PAT) with appropriate permissions

## Authentication

### Creating a Personal Access Token

1. Go to **GitHub Settings** > **Developer settings** > **Personal access tokens** > **Tokens (classic)**
2. Click **Generate new token (classic)**
3. Set the following scopes:

**For Organization Scanning:**

| Scope | Description |
|-------|-------------|
| `read:org` | Read organization membership and teams |
| `repo` | Full control of private repositories |
| `read:user` | Read user profile data |

**For Repository Scanning:**

| Scope | Description |
|-------|-------------|
| `repo` | Full control of private repositories |
| `public_repo` | Access public repositories only (alternative) |

4. Click **Generate token** and copy immediately

### Configuration Methods

**Environment Variable (Recommended):**
```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
kspec scan github org <organization-name> -f policy.yml
```

**Command Line Flag:**
```bash
kspec scan github org <organization-name> \
  --token "ghp_xxxxxxxxxxxx" \
  -f policy.yml
```

### GitHub App Authentication (Enterprise)

For larger organizations, consider using a GitHub App:

1. Create a GitHub App in organization settings
2. Grant permissions (Repository: Read, Organization: Read)
3. Install the app on your organization
4. Generate a private key for authentication

---

## Resources

The GitHub provider discovers the following resources:

| Resource | Description |
|----------|-------------|
| `github_organization` | Organization settings and security configuration |
| `github_repo` | Repository settings and metadata |
| `github_branch` | Branch protection rules (default branch) |
| `github_team` | Team membership and permissions |

---

## Resource Fields

### github_organization

| Field | Type | Description |
|-------|------|-------------|
| `login` | `string` | Organization login name |
| `name` | `string` | Organization display name |
| `description` | `string` | Organization description |
| `two_factor_requirement_enabled` | `bool` | 2FA required for all members |
| `default_repository_permission` | `string` | Default repo permission (`none`, `read`, `write`, `admin`) |
| `is_verified` | `bool` | Domain is verified |
| `members_can_create_repositories` | `bool` | Members can create repos |
| `members_can_create_public_repositories` | `bool` | Members can create public repos |
| `members_can_create_private_repositories` | `bool` | Members can create private repos |
| `members_can_fork_private_repositories` | `bool` | Members can fork private repos |

### github_repo

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | Repository ID |
| `name` | `string` | Repository name |
| `full_name` | `string` | Full repository name (owner/repo) |
| `description` | `string` | Repository description |
| `private` | `bool` | Repository is private |
| `archived` | `bool` | Repository is archived |
| `disabled` | `bool` | Repository is disabled |
| `default_branch` | `string` | Default branch name |
| `visibility` | `string` | Visibility (`public`, `private`, `internal`) |
| `has_issues` | `bool` | Issues are enabled |
| `has_wiki` | `bool` | Wiki is enabled |
| `has_discussions` | `bool` | Discussions are enabled |
| `allow_forking` | `bool` | Forking is allowed |
| `allow_squash_merge` | `bool` | Squash merging allowed |
| `allow_merge_commit` | `bool` | Merge commits allowed |
| `allow_rebase_merge` | `bool` | Rebase merging allowed |
| `delete_branch_on_merge` | `bool` | Delete branch after merge |
| `collaborators` | `[]object` | Repository collaborators |
| `files` | `[]object` | Repository file tree |

**files Entry Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `path` | `string` | File path in repository |
| `type` | `string` | Entry type (`blob` for files) |
| `size` | `int` | File size in bytes |
| `sha` | `string` | File SHA hash |

### github_branch

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Branch name |
| `is_default` | `bool` | Is the default branch |
| `protected` | `bool` | Branch protection enabled |
| `repo` | `string` | Repository name (org scans) |
| `repo_full_name` | `string` | Full repo name (org scans) |
| `protection_rules` | `object` | Full protection configuration |
| `required_status_checks` | `[]string` | Required status check contexts |
| `enforce_admins` | `object` | Admin enforcement settings |
| `required_pull_request_reviews` | `bool` | PR reviews required |
| `allow_force_pushes` | `object` | Force push settings |
| `required_conversation_resolution` | `object` | Conversation resolution settings |
| `required_signatures` | `bool` | Signed commits required |

**enforce_admins / allow_force_pushes / required_conversation_resolution Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | `bool` | Setting is enabled |

### github_team

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | Team ID |
| `name` | `string` | Team name |
| `slug` | `string` | Team slug |
| `description` | `string` | Team description |
| `privacy` | `string` | Team privacy (`secret`, `closed`) |
| `permission` | `string` | Default permission level |
| `members_count` | `int` | Number of members |
| `repos_count` | `int` | Number of repositories |

---

## Example Policies

### Organization Security

```yaml
queries:
  - uid: two-factor-required
    title: Two-factor authentication required
    resource: github_organization
    severity: critical
    query: resource.two_factor_requirement_enabled == true
    docs: |
      2FA provides an additional layer of security for all members.
    remediation: |
      Go to Settings > Security > Authentication security and enable
      "Require two-factor authentication for all members".

  - uid: verified-domain
    title: Organization has verified domain
    resource: github_organization
    severity: high
    query: resource.is_verified == true
    docs: |
      A verified domain adds trust and credibility.
    remediation: |
      Go to Settings > Verified & approved domains to verify your domain.

  - uid: restrictive-permissions
    title: Default permissions are restrictive
    resource: github_organization
    severity: medium
    query: |
      resource.default_repository_permission == "read" ||
      resource.default_repository_permission == "none"
    docs: |
      Following least privilege, default permissions should be minimal.
    remediation: |
      Go to Settings > Member privileges and set base permissions to "Read".

  - uid: no-public-repos
    title: Members cannot create public repositories
    resource: github_organization
    severity: medium
    query: resource.members_can_create_public_repositories == false
    docs: |
      Preventing public repo creation reduces accidental data exposure.
    remediation: |
      Go to Settings > Member privileges and disable public repo creation.
```

### Branch Protection

```yaml
queries:
  - uid: default-branch-protected
    title: Default branch is protected
    resource: github_branch
    severity: critical
    query: |
      resource.is_default == false || resource.protected == true
    docs: |
      Branch protection ensures only authorized changes are merged.
    remediation: |
      Go to Settings > Branches and add a protection rule for the default branch.

  - uid: no-force-push
    title: Force pushes are disabled
    resource: github_branch
    severity: high
    query: |
      !resource.is_default ||
      !resource.protected ||
      (has(resource.allow_force_pushes) &&
       resource.allow_force_pushes.enabled == false)
    docs: |
      Force pushes can overwrite history and cause data loss.
    remediation: |
      Edit branch protection and ensure "Allow force pushes" is disabled.

  - uid: signed-commits
    title: Signed commits are required
    resource: github_branch
    severity: high
    query: |
      !resource.is_default ||
      !resource.protected ||
      resource.required_signatures == true
    docs: |
      Signed commits verify author identity cryptographically.
    remediation: |
      Edit branch protection and enable "Require signed commits".

  - uid: status-checks-required
    title: Status checks are required
    resource: github_branch
    severity: high
    query: |
      !resource.is_default ||
      !resource.protected ||
      (has(resource.required_status_checks) &&
       size(resource.required_status_checks) > 0)
    docs: |
      Status checks ensure CI tests pass before merging.
    remediation: |
      Edit branch protection and add required status checks.

  - uid: admin-enforcement
    title: Admins cannot bypass protection
    resource: github_branch
    severity: high
    query: |
      !resource.is_default ||
      !resource.protected ||
      (has(resource.enforce_admins) &&
       resource.enforce_admins.enabled == true)
    docs: |
      Even admins should follow branch protection rules.
    remediation: |
      Edit branch protection and enable "Do not allow bypassing the above settings".

  - uid: conversation-resolution
    title: Conversation resolution required
    resource: github_branch
    severity: medium
    query: |
      !resource.is_default ||
      !resource.protected ||
      (has(resource.required_conversation_resolution) &&
       resource.required_conversation_resolution.enabled == true)
    docs: |
      All review comments should be resolved before merging.
    remediation: |
      Edit branch protection and enable "Require conversation resolution".
```

### Repository Configuration

```yaml
queries:
  - uid: has-dependabot
    title: Dependabot is configured
    resource: github_repo
    severity: medium
    query: |
      resource.files.exists(f,
        f.path == ".github/dependabot.yml" ||
        f.path == ".github/dependabot.yaml"
      )
    docs: |
      Dependabot keeps dependencies up to date and secure.
    remediation: |
      Create a .github/dependabot.yml file to configure Dependabot.

  - uid: private-repos
    title: Repository is private
    resource: github_repo
    severity: medium
    query: resource.private == true
    docs: |
      Repositories with sensitive code should be private.
    remediation: |
      Go to Settings > General > Danger Zone and change visibility.

  - uid: delete-branch-on-merge
    title: Branches deleted after merge
    resource: github_repo
    severity: low
    query: resource.delete_branch_on_merge == true
    docs: |
      Automatically deleting branches keeps the repository clean.
    remediation: |
      Go to Settings > General and enable "Automatically delete head branches".
```

---

## CLI Reference

```bash
# Scan an organization
kspec scan github org <organization-name> -f <policy-file>

# Scan a single repository
kspec scan github repo <owner>/<repo> -f <policy-file>

# Scan with policy directory
kspec scan github org <organization-name> -d <policy-directory>

# Using explicit token
kspec scan github org <organization-name> \
  --token <github-token> \
  -f <policy-file>
```

### Options

| Flag | Description |
|------|-------------|
| `-f, --policy` | Policy file to use |
| `-d, --policy-dir` | Directory containing policy files |
| `--token` | GitHub Personal Access Token |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | Personal Access Token |
| `GITHUB_API_URL` | GitHub API URL (for Enterprise) |

---

## CI/CD Integration

### GitHub Actions

```yaml
name: GitHub Security Scan

on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
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

      - name: Run Security Scan
        env:
          GITHUB_TOKEN: ${{ secrets.KSPEC_GITHUB_TOKEN }}
        run: |
          ./kspec scan github org ${{ github.repository_owner }} \
            -f policies/github-security.yml
```

### GitLab CI

```yaml
security-scan:
  image: golang:1.21
  script:
    - go build -o kspec ./cmd/kspec
    - ./kspec scan github org ${GITHUB_ORG} -f policies/github-security.yml
  variables:
    GITHUB_TOKEN: ${GITHUB_TOKEN}
```

---

## Troubleshooting

### Authentication Errors

```
Error: 401 Unauthorized
```

**Causes:**
- Invalid or expired token
- Token lacks required scopes
- Token doesn't have access to organization

**Solutions:**
- Verify token is correct and not expired
- Check token has required scopes (`read:org`, `repo`)
- Ensure you're a member of the organization

### Rate Limiting

```
Error: API rate limit exceeded
```

**Causes:**
- Too many API requests
- Token rate limit exhausted

**Solutions:**
- Wait for rate limit to reset
- Use a GitHub App for higher limits (5000 requests/hour)
- Reduce scan scope

### Permission Errors

```
Error: Resource not accessible by integration
```

**Causes:**
- Token lacks permissions
- Organization restricts third-party access
- Repository is inaccessible

**Solutions:**
- Add required scopes to token
- Check organization settings for OAuth app restrictions
- Verify access to the target repository

### No Resources Found

**Causes:**
- Incorrect organization/repository name
- Token lacks access
- Empty organization

**Solutions:**
- Verify names are correct
- Check token permissions
- Ensure organization has repositories to scan

### Branch Protection Not Found

```
protection_rules: null
```

**Causes:**
- Branch protection not configured
- Token lacks admin access

**Solutions:**
- Configure branch protection rules
- Use a token with admin access to fetch protection details

---

## Best Practices

1. **Use Fine-Grained Tokens**: Create tokens with minimal required permissions
2. **Rotate Tokens Regularly**: Set up token rotation (90 days recommended)
3. **Store Secrets Securely**: Never commit tokens to repositories
4. **Scan Regularly**: Schedule daily/weekly scans to detect drift
5. **Review Results**: Act on findings to maintain security posture
