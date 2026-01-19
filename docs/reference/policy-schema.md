# Policy Schema Reference

Complete reference for kspec policy file format.

## File Format

Policies are defined in YAML format with the `.yml` or `.yaml` extension.

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | Yes | API version (e.g., `kopexa.io/v1alpha1`) |
| `kind` | string | Yes | Resource kind (always `Policy`) |
| `metadata` | Metadata | Yes | Policy metadata |
| `require` | []Requirement | No | Provider requirements |
| `authors` | []Author | No | Policy authors |
| `groups` | []Group | Yes | Check groups |
| `queries` | []Check | Yes | Query/check definitions |
| `scoring_system` | string | No | Scoring method |

## Schema Definitions

### Metadata

```yaml
metadata:
  name: my-security-policy
  title: My Security Policy
  version: 1.0.0
  license: ELv2
  tags:
    category: security
    platform: azure
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique policy identifier |
| `title` | string | No | Human-readable title |
| `version` | string | No | Semantic version (e.g., `1.0.0`) |
| `license` | string | No | License identifier |
| `tags` | map[string]string | No | Key-value metadata tags |

### Requirement

```yaml
require:
  - provider: azure
  - provider: github
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `provider` | string | Yes | Provider name |

### Author

```yaml
authors:
  - name: Security Team
    email: security@example.com
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Author name |
| `email` | string | No | Author email |

### Group

```yaml
groups:
  - title: Storage Security
    filter: asset.type == "azure-subscription"
    checks:
      - uid: check-1
      - uid: check-2
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | Yes | Group title |
| `filter` | string | No | CEL expression for asset filtering |
| `checks` | []CheckRef | Yes | List of check references |

### CheckRef

References a check defined in the `queries` section by UID.

```yaml
checks:
  - uid: my-check-uid
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `uid` | string | Yes | Reference to query UID |

### Check (Query)

Checks are defined in the `queries` section and referenced by groups.

```yaml
queries:
  - uid: storage-https-required
    title: Storage accounts require HTTPS
    resource: azure_storage_account
    severity: critical
    query: |
      resource.properties.supportsHttpsTrafficOnly == true
    docs: Description of the check.
    audit: Manual audit steps.
    remediation: How to fix the issue.
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `uid` | string | Yes | Unique identifier |
| `id` | string | No | Alternative identifier |
| `title` | string | Yes | Check title |
| `resource` | string | Yes | Resource type to check |
| `query` | string | Yes | CEL expression |
| `severity` | string | No | Severity level |
| `docs` | string | No | Check documentation |
| `audit` | string | No | Manual audit instructions |
| `remediation` | string | No | Remediation steps |
| `config` | map[string]string | No | Configuration key-value pairs |
| `props` | []Prop | No | Property extractions |

### Prop

Property extractions for check results.

```yaml
props:
  - uid: my-prop
    title: Property Title
    mql: resource.field
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `uid` | string | No | Property identifier |
| `title` | string | No | Property title |
| `mql` | any | No | MQL expression for extraction |

### Severity Values

| Value | Description |
|-------|-------------|
| `critical` | Immediate security risk |
| `high` | Significant security risk |
| `medium` | Moderate security risk |
| `low` | Minor security risk |

### Scoring System Values

| Value | Description |
|-------|-------------|
| `highest_impact` | Use highest severity from failed checks |
| `average` | Average of all severity scores |

## Complete Example

```yaml
apiVersion: kopexa.io/v1alpha1
kind: Policy
metadata:
  name: azure-security-policy
  title: Azure Security Policy
  version: 2.0.0
  license: Apache-2.0
  tags:
    category: security
    platform: azure
    compliance: cis

require:
  - provider: azure

authors:
  - name: Security Team
    email: security@example.com

groups:
  - title: Storage Security
    filter: asset.type == "azure-subscription"
    checks:
      - uid: storage-https-required
      - uid: storage-public-access-disabled
      - uid: storage-network-deny

  - title: SQL Security
    checks:
      - uid: sql-auditing-enabled
      - uid: sql-tde-enabled

  - title: Key Vault Security
    checks:
      - uid: keyvault-purge-protection

  - title: Network Security
    checks:
      - uid: nsg-no-ssh-from-internet
      - uid: nsg-no-rdp-from-internet

scoring_system: highest impact

queries:
  # Storage checks
  - uid: storage-https-required
    title: Storage accounts require HTTPS
    resource: azure_storage_account
    severity: critical
    query: |
      has(resource.properties) &&
      resource.properties.supportsHttpsTrafficOnly == true
    docs: |
      HTTPS ensures data is encrypted in transit, protecting
      against man-in-the-middle attacks.
    audit: |
      1. Go to Azure Portal > Storage Accounts
      2. Select storage account
      3. Go to Configuration
      4. Check "Secure transfer required" is Enabled
    remediation: |
      Enable "Secure transfer required" in storage account
      configuration settings.

  - uid: storage-public-access-disabled
    title: Public blob access is disabled
    resource: azure_storage_account
    severity: critical
    query: |
      has(resource.properties) &&
      resource.properties.allowBlobPublicAccess == false
    docs: |
      Disabling public access prevents unauthorized data exposure.
    remediation: |
      Set "Allow Blob Anonymous Access" to Disabled.

  - uid: storage-network-deny
    title: Default network access is deny
    resource: azure_storage_account
    severity: high
    query: |
      has(resource.properties.networkAcls) &&
      resource.properties.networkAcls.defaultAction == "Deny"
    docs: |
      Denying by default ensures only allowed networks can access.
    remediation: |
      Set default action to "Deny" in networking settings.

  # SQL checks
  - uid: sql-auditing-enabled
    title: SQL Server auditing is enabled
    resource: azure_sql_server
    severity: high
    query: |
      has(resource.auditingPolicy) &&
      resource.auditingPolicy.state == "Enabled"
    docs: |
      Auditing tracks database events for security analysis.
    remediation: |
      Enable auditing in SQL Server settings.

  - uid: sql-tde-enabled
    title: SQL databases have TDE enabled
    resource: azure_sql_database
    severity: high
    query: |
      resource.name == "master" ||
      (has(resource.transparentDataEncryption) &&
       resource.transparentDataEncryption.state == "Enabled")
    docs: |
      Transparent Data Encryption protects data at rest.
    remediation: |
      Enable TDE for the SQL database.

  # Key Vault checks
  - uid: keyvault-purge-protection
    title: Key Vault has purge protection
    resource: azure_keyvault_vault
    severity: critical
    query: |
      has(resource.properties) &&
      resource.properties.enablePurgeProtection == true
    docs: |
      Purge protection prevents permanent deletion of secrets.
    remediation: |
      Enable purge protection for the Key Vault.

  # Network checks
  - uid: nsg-no-ssh-from-internet
    title: SSH not open to internet
    resource: azure_network_security_group
    severity: critical
    query: |
      !has(resource.securityRules) ||
      !resource.securityRules.exists(rule,
        rule.properties.access == "Allow" &&
        rule.properties.direction == "Inbound" &&
        rule.properties.destinationPortRange == "22" &&
        (rule.properties.sourceAddressPrefix == "*" ||
         rule.properties.sourceAddressPrefix == "0.0.0.0/0")
      )
    docs: |
      SSH should not be open to the entire internet.
    remediation: |
      Restrict SSH rules to specific IP addresses.

  - uid: nsg-no-rdp-from-internet
    title: RDP not open to internet
    resource: azure_network_security_group
    severity: critical
    query: |
      !has(resource.securityRules) ||
      !resource.securityRules.exists(rule,
        rule.properties.access == "Allow" &&
        rule.properties.direction == "Inbound" &&
        rule.properties.destinationPortRange == "3389" &&
        (rule.properties.sourceAddressPrefix == "*" ||
         rule.properties.sourceAddressPrefix == "0.0.0.0/0")
      )
    docs: |
      RDP should not be open to the entire internet.
    remediation: |
      Restrict RDP rules to specific IP addresses.
```

## Validation

Policy files are validated on load. Common validation errors:

| Error | Cause | Solution |
|-------|-------|----------|
| `missing uid` | Query without UID | Add unique `uid` field |
| `duplicate uid` | Same UID used twice | Use unique UIDs |
| `unknown resource` | Invalid resource type | Check provider documentation |
| `invalid query` | CEL syntax error | Validate CEL expression |
| `missing checks` | Group with no checks | Add check references |

## Best Practices

1. **Unique UIDs**: Use namespaced UIDs like `company-provider-category-check`
2. **Descriptive Titles**: Make titles actionable and clear
3. **Complete Documentation**: Include docs, audit, and remediation
4. **Appropriate Severity**: Match severity to actual risk
5. **Version Policies**: Increment version on changes
6. **Test Queries**: Validate CEL expressions before deploying
7. **Use Tags**: Add meaningful tags for categorization and filtering
