# Writing Policies

Policies in kspec define security expectations for your infrastructure. This guide covers how to write effective security policies.

## Policy Structure

A policy file is a YAML document with the following sections:

```yaml
# Metadata
name: My Security Policy
version: 1.0.0
license: Apache-2.0

# Provider requirements
require:
  - provider: azure

# Author information
authors:
  - name: Security Team
    email: security@example.com

# Documentation
docs:
  desc: |
    Description of what this policy checks.

# Check groups
groups:
  - title: Storage Security
    checks:
      - uid: storage-https-required

# Scoring configuration
scoring_system: highest_impact

# Query definitions
queries:
  - uid: storage-https-required
    title: Storage accounts require HTTPS
    resource: azure_storage_account
    severity: critical
    query: |
      resource.properties.supportsHttpsTrafficOnly == true
```

## Metadata

### Required Fields

| Field | Description |
|-------|-------------|
| `name` | Policy name |
| `version` | Semantic version |

### Optional Fields

| Field | Description |
|-------|-------------|
| `license` | License identifier (e.g., `Apache-2.0`, `BUSL-1.1`) |
| `tags` | Key-value metadata tags |
| `authors` | List of author objects |
| `docs.desc` | Policy description |

## Provider Requirements

Specify which providers are needed:

```yaml
require:
  - provider: azure
  - provider: github
```

## Policy Imports

Policies can import checks from other policy files, allowing you to build modular, reusable policy libraries. Imported queries are available for reference in your groups.

### Import Syntax

```yaml
imports:
  - ./base-checks.yaml           # Relative local file
  - /path/to/checks.yaml         # Absolute local file
  - ./checks/*.yaml              # Glob pattern
  - https://example.com/policy.yaml  # Remote URL
```

### Import Types

| Type | Pattern | Example |
|------|---------|---------|
| Local file | Relative or absolute path | `./checks.yaml`, `/policies/base.yaml` |
| Glob | Wildcard patterns | `./checks/*.yaml`, `./lib/**/*.yaml` |
| URL | HTTP/HTTPS URLs | `https://raw.githubusercontent.com/org/repo/main/policy.yaml` |

### Example: Extending a Base Policy

**base-checks.yaml** - Reusable check definitions:
```yaml
apiVersion: kspec/v1
kind: Policy
metadata:
  name: base-security-checks
queries:
  - uid: https-required
    title: HTTPS must be enabled
    resource: azure_storage_account
    severity: critical
    query: resource.properties.supportsHttpsTrafficOnly == true

  - uid: encryption-enabled
    title: Encryption must be enabled
    resource: azure_storage_account
    severity: critical
    query: resource.properties.encryption.services.blob.enabled == true
```

**my-policy.yaml** - Policy that imports and uses base checks:
```yaml
apiVersion: kspec/v1
kind: Policy
metadata:
  name: my-storage-policy
  version: 1.0.0

imports:
  - ./base-checks.yaml

require:
  - provider: azure

groups:
  - title: Storage Security
    checks:
      - uid: https-required      # From import
      - uid: encryption-enabled  # From import
      - uid: custom-check        # Local check

queries:
  - uid: custom-check
    title: Custom storage check
    resource: azure_storage_account
    severity: medium
    query: has(resource.tags.environment)
```

### Import Resolution

1. **Relative paths** are resolved relative to the importing policy file
2. **Glob patterns** match files in the specified directory
3. **URLs** are fetched via HTTP/HTTPS with a 30-second timeout
4. **Circular imports** are automatically detected and prevented
5. **Import order**: Imported queries are prepended to local queries (local queries take precedence for duplicate UIDs)

### Nested Imports

Imported policies can themselves have imports, which are resolved recursively:

```
main-policy.yaml
  └── imports: security-checks.yaml
        └── imports: common-checks.yaml
```

All queries from the entire import chain become available in the main policy.

### Best Practices for Imports

1. **Organize by domain**: Group related checks into separate files
2. **Use relative paths**: Keep policies portable within a repository
3. **Version remote imports**: Pin to specific commits or tags for stability
4. **Document dependencies**: Note required imports in policy documentation

## Groups

Groups organize checks into logical sections:

```yaml
groups:
  - title: Storage Security
    filter: asset.type == "azure-subscription"  # Optional filter
    checks:
      - uid: check-1
      - uid: check-2

  - title: Network Security
    checks:
      - uid: check-3
```

### Group Filters

Filter which assets a group applies to:

```yaml
groups:
  - title: Organization Checks
    filter: asset.type == "github-org"
    checks:
      - uid: org-2fa-required

  - title: Repository Checks
    filter: asset.type == "github-repo"
    checks:
      - uid: branch-protection
```

## Queries

Queries define individual security checks:

```yaml
queries:
  - uid: storage-https-required
    title: Storage accounts require HTTPS
    resource: azure_storage_account
    severity: critical
    impact: 90
    query: |
      resource.properties.supportsHttpsTrafficOnly == true
    docs: |
      HTTPS ensures data is encrypted in transit.
    audit: |
      1. Go to Azure Portal
      2. Navigate to Storage Accounts
      3. Check "Secure transfer required" setting
    remediation: |
      Enable "Secure transfer required" in storage account settings.
```

### Query Fields

| Field | Required | Description |
|-------|----------|-------------|
| `uid` | Yes | Unique identifier |
| `title` | Yes | Human-readable title |
| `resource` | Yes | Resource type to check |
| `query` | Yes | CEL expression |
| `severity` | No | `critical`, `high`, `medium`, `low` |
| `impact` | No | Numeric score (0-100) |
| `docs` | No | Documentation text |
| `audit` | No | Manual audit steps |
| `remediation` | No | Fix instructions |

## Severity Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `critical` | Immediate risk | Public exposure, missing encryption |
| `high` | Significant risk | Missing access controls |
| `medium` | Moderate risk | Best practice violations |
| `low` | Minor risk | Informational findings |

## CEL Expressions

Queries use [CEL (Common Expression Language)](concepts/cel-expressions.md) for evaluation.

### Basic Comparisons

```yaml
# Equality
query: resource.status == "active"

# Inequality
query: resource.public_access != true

# Numeric comparison
query: resource.min_tls_version >= "1.2"
```

### Boolean Logic

```yaml
# AND
query: |
  resource.encrypted == true &&
  resource.public_access == false

# OR
query: |
  resource.type == "private" ||
  resource.restricted == true

# NOT
query: "!resource.public_access"
```

### Field Existence

```yaml
# Check field exists
query: has(resource.encryption)

# Check field exists and has value
query: |
  has(resource.tags) &&
  has(resource.tags.environment)
```

### Collections

```yaml
# Check all items match
query: |
  resource.rules.all(r, r.action == "deny")

# Check any item matches
query: |
  resource.rules.exists(r, r.port == 22)

# Check size
query: size(resource.members) >= 2
```

### String Operations

```yaml
# Contains
query: resource.name.contains("prod")

# Starts with
query: resource.name.startsWith("app-")

# Regex match
query: resource.name.matches("^prod-.*-[0-9]+$")
```

## Common Patterns

### Check for Non-Default Values

```yaml
query: |
  resource.name != "Default Permission Scheme"
```

### Check for Protected Resources

```yaml
# Only check default branch
query: |
  resource.is_default == false || resource.protected == true
```

### Check for Missing Configuration

```yaml
# Pass if field missing OR field has correct value
query: |
  !has(resource.public_access) ||
  resource.public_access == false
```

### Check Lists for Security Issues

```yaml
# No rule allows all inbound traffic
query: |
  !resource.rules.exists(r,
    r.direction == "inbound" &&
    r.source == "0.0.0.0/0"
  )
```

### Check for Specific Files

```yaml
# Repository has Dependabot config
query: |
  resource.files.exists(f,
    f.path == ".github/dependabot.yml" ||
    f.path == ".github/dependabot.yaml"
  )
```

## Scoring System

Configure how overall scores are calculated:

```yaml
# Use highest_impact score from failed checks
scoring_system: highest_impact

# Alternative: average of all impacts
scoring_system: average
```

## Complete Example

```yaml
name: GitHub Security Policy
version: 1.0.0
license: Apache-2.0

require:
  - provider: github

authors:
  - name: Security Team
    email: security@example.com

docs:
  desc: |
    Security policy for GitHub organizations and repositories.

groups:
  - title: Organization Security
    filter: asset.type == "github-org"
    checks:
      - uid: org-2fa-required
      - uid: org-verified-domain

  - title: Repository Security
    checks:
      - uid: repo-branch-protection
      - uid: repo-dependabot

scoring_system: highest_impact

queries:
  - uid: org-2fa-required
    title: Two-factor authentication required
    resource: github_organization
    severity: critical
    impact: 100
    query: resource.two_factor_requirement_enabled == true
    docs: |
      Organizations should require 2FA for all members.
    remediation: |
      Enable 2FA requirement in organization security settings.

  - uid: org-verified-domain
    title: Organization has verified domain
    resource: github_organization
    severity: high
    impact: 70
    query: resource.is_verified == true
    docs: |
      Verified domains provide identity confirmation.
    remediation: |
      Verify your domain in organization settings.

  - uid: repo-branch-protection
    title: Default branch is protected
    resource: github_branch
    severity: critical
    impact: 90
    query: |
      resource.is_default == false || resource.protected == true
    docs: |
      Branch protection prevents unauthorized changes.
    remediation: |
      Enable branch protection for the default branch.

  - uid: repo-dependabot
    title: Dependabot is configured
    resource: github_repo
    severity: medium
    impact: 50
    query: |
      resource.files.exists(f,
        f.path == ".github/dependabot.yml" ||
        f.path == ".github/dependabot.yaml"
      )
    docs: |
      Dependabot keeps dependencies up to date.
    remediation: |
      Create a .github/dependabot.yml configuration file.
```

## Best Practices

1. **Use Descriptive UIDs**: Make UIDs readable and unique
2. **Set Appropriate Severity**: Match severity to actual risk
3. **Include Documentation**: Help users understand findings
4. **Provide Remediation**: Give clear fix instructions
5. **Test Policies**: Verify queries work as expected
6. **Version Policies**: Track changes over time

## Next Steps

- [CEL Expressions](concepts/cel-expressions.md) - Deep dive into query language
- [Policy Schema](reference/policy-schema.md) - Full schema reference
- [Provider Guides](README.md#provider-guides) - Available resources per provider
