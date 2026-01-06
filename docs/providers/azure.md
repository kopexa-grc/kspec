# Azure Provider

The Azure provider scans your Azure subscriptions for security compliance, covering storage accounts, SQL databases, Key Vaults, virtual machines, networking, and more.

## Overview

Use the Azure provider to validate:
- Storage account security (HTTPS enforcement, public access, network rules)
- SQL Server and database configurations (auditing, TDE, firewall rules)
- Key Vault protection (purge protection, network access)
- Network security groups (SSH/RDP restrictions)
- Virtual machine security (disk encryption)
- App Service configurations (managed identity, TLS)

## Quick Start

```bash
# Set Azure credentials
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"

# Scan a subscription with all Azure policies
kspec scan azure subscription <subscription-id> -d policies

# Scan with specific policy
kspec scan azure subscription <subscription-id> -f policies/azure-security.yml
```

## Prerequisites

- Azure subscription
- Service Principal with **Reader** permissions (minimum)

## Authentication

### Creating a Service Principal

```bash
# Create a service principal with Reader role
az ad sp create-for-rbac \
  --name "kspec-reader" \
  --role Reader \
  --scopes /subscriptions/<SUBSCRIPTION_ID>
```

This outputs credentials:
```json
{
  "appId": "<CLIENT_ID>",
  "displayName": "kspec-reader",
  "password": "<CLIENT_SECRET>",
  "tenant": "<TENANT_ID>"
}
```

### Configuration Methods

**Environment Variables (Recommended):**
```bash
export AZURE_SUBSCRIPTION_ID="your-subscription-id"
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"

kspec scan azure subscription $AZURE_SUBSCRIPTION_ID -f policy.yml
```

**Command Line Flags:**
```bash
kspec scan azure subscription <subscription-id> \
  --tenant-id <tenant-id> \
  --client-id <client-id> \
  --token <client-secret> \
  -f policy.yml
```

**Azure CLI Authentication:**
```bash
# Login with Azure CLI
az login

# kspec will use DefaultAzureCredential
kspec scan azure subscription <subscription-id> -f policy.yml
```

---

## Resources

The Azure provider discovers the following resources:

### Storage

| Resource | Description |
|----------|-------------|
| `azure_storage_account` | Storage accounts with blob, file, queue, and table services |

### Databases

| Resource | Description |
|----------|-------------|
| `azure_sql_server` | Azure SQL Server instances |
| `azure_sql_database` | SQL databases (sub-resource of SQL Server) |
| `azure_mysql_server` | Azure Database for MySQL servers |
| `azure_postgresql_server` | Azure Database for PostgreSQL servers |

### Security

| Resource | Description |
|----------|-------------|
| `azure_keyvault_vault` | Key Vault instances |
| `azure_network_security_group` | Network Security Groups (NSGs) |

### Compute

| Resource | Description |
|----------|-------------|
| `azure_virtual_machine` | Virtual machine instances |
| `azure_app_service` | App Service web applications |

### Management

| Resource | Description |
|----------|-------------|
| `azure_subscription` | Subscription-level configuration |

---

## Resource Fields

### azure_storage_account

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | Storage account name |
| `location` | `string` | Azure region |
| `tags` | `map` | Resource tags |
| `properties.supportsHttpsTrafficOnly` | `bool` | HTTPS traffic only enforcement |
| `properties.allowBlobPublicAccess` | `bool` | Allow public blob access |
| `properties.minimumTlsVersion` | `string` | Minimum TLS version |
| `properties.networkAcls.defaultAction` | `string` | Default network action (`Allow`/`Deny`) |
| `properties.networkAcls.bypass` | `string` | Services bypassing network rules |
| `properties.networkAcls.ipRules` | `[]object` | IP-based access rules |
| `properties.networkAcls.virtualNetworkRules` | `[]object` | VNet rules |

### azure_sql_server

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | SQL Server name |
| `location` | `string` | Azure region |
| `properties.administratorLogin` | `string` | Admin username |
| `properties.fullyQualifiedDomainName` | `string` | Server FQDN |
| `properties.publicNetworkAccess` | `string` | Public network access status |
| `auditingPolicy` | `object` | Server auditing configuration |
| `auditingPolicy.state` | `string` | Auditing state (`Enabled`/`Disabled`) |
| `firewallRules` | `[]object` | Firewall rules |
| `virtualNetworkRules` | `[]object` | VNet rules |

### azure_sql_database

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | Database name |
| `properties.status` | `string` | Database status |
| `transparentDataEncryption` | `object` | TDE configuration |
| `transparentDataEncryption.state` | `string` | TDE state (`Enabled`/`Disabled`) |

### azure_keyvault_vault

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | Key Vault name |
| `location` | `string` | Azure region |
| `properties.enablePurgeProtection` | `bool` | Purge protection enabled |
| `properties.enableSoftDelete` | `bool` | Soft delete enabled |
| `properties.publicNetworkAccess` | `string` | Public network access status |
| `properties.networkAcls` | `object` | Network access rules |

### azure_network_security_group

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | NSG name |
| `location` | `string` | Azure region |
| `securityRules` | `[]object` | Security rules |

**Security Rule Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `properties.access` | `string` | `Allow` or `Deny` |
| `properties.direction` | `string` | `Inbound` or `Outbound` |
| `properties.protocol` | `string` | Protocol (`TCP`, `UDP`, `*`) |
| `properties.sourceAddressPrefix` | `string` | Source CIDR or keyword |
| `properties.destinationPortRange` | `string` | Destination port or range |
| `properties.priority` | `int` | Rule priority |

### azure_virtual_machine

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | VM name |
| `location` | `string` | Azure region |
| `properties.vmId` | `string` | Unique VM ID |
| `properties.hardwareProfile.vmSize` | `string` | VM size |
| `osDisk` | `object` | OS disk configuration |
| `osDisk.properties.encryption.type` | `string` | Encryption type |

### azure_app_service

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | App name |
| `location` | `string` | Azure region |
| `identity.type` | `string` | Identity type (`SystemAssigned`, etc.) |
| `identity.principalId` | `string` | Managed identity principal ID |
| `configuration.properties.minTlsVersion` | `string` | Minimum TLS version |

### azure_mysql_server / azure_postgresql_server

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Resource ID |
| `name` | `string` | Server name |
| `location` | `string` | Azure region |
| `properties.sslEnforcement` | `string` | SSL enforcement (`Enabled`/`Disabled`) |
| `properties.minimalTlsVersion` | `string` | Minimum TLS version |
| `properties.publicNetworkAccess` | `string` | Public network access status |

---

## Example Policies

### Storage Security

```yaml
queries:
  - uid: storage-https-required
    title: Storage accounts require HTTPS
    resource: azure_storage_account
    severity: critical
    query: |
      has(resource.properties) &&
      resource.properties.supportsHttpsTrafficOnly == true
    docs: |
      HTTPS ensures data is encrypted in transit.
    remediation: |
      Enable "Secure transfer required" in storage account settings.

  - uid: storage-public-access-disabled
    title: Blob public access is disabled
    resource: azure_storage_account
    severity: critical
    query: |
      has(resource.properties) &&
      resource.properties.allowBlobPublicAccess == false
    docs: |
      Disabling public access prevents unauthorized data exposure.
    remediation: |
      Set "Allow Blob Anonymous Access" to Disabled.

  - uid: storage-network-deny-default
    title: Default network access set to Deny
    resource: azure_storage_account
    severity: high
    query: |
      has(resource.properties.networkAcls) &&
      resource.properties.networkAcls.defaultAction == "Deny"
    docs: |
      Denying by default ensures only allowed networks can access storage.
    remediation: |
      Set default action to "Deny" in networking settings.
```

### SQL Security

```yaml
queries:
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
      Enable auditing in the SQL Server settings.

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

  - uid: sql-no-public-access
    title: SQL Server public access is restricted
    resource: azure_sql_server
    severity: critical
    query: |
      has(resource.properties) &&
      resource.properties.publicNetworkAccess == "Disabled"
    docs: |
      Disabling public access reduces attack surface.
    remediation: |
      Set public network access to "Disabled".
```

### Key Vault Security

```yaml
queries:
  - uid: keyvault-purge-protection
    title: Key Vault has purge protection enabled
    resource: azure_keyvault_vault
    severity: critical
    query: |
      has(resource.properties) &&
      resource.properties.enablePurgeProtection == true
    docs: |
      Purge protection prevents permanent deletion.
    remediation: |
      Enable purge protection for the Key Vault.

  - uid: keyvault-private-access
    title: Key Vault public access is restricted
    resource: azure_keyvault_vault
    severity: high
    query: |
      has(resource.properties) &&
      (resource.properties.publicNetworkAccess == "Disabled" ||
       has(resource.properties.networkAcls))
    docs: |
      Restricting network access protects secrets.
    remediation: |
      Configure network rules or disable public access.
```

### Network Security

```yaml
queries:
  - uid: nsg-no-ssh-from-internet
    title: SSH access is restricted from internet
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
    title: RDP access is restricted from internet
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

---

## CLI Reference

```bash
# Scan all resources in subscription
kspec scan azure subscription <subscription-id> -f <policy-file>

# Scan with policy directory
kspec scan azure subscription <subscription-id> -d <policy-directory>

# Filter by resource group
kspec scan azure subscription <subscription-id> \
  --resource-group <resource-group-name> \
  -f <policy-file>

# Using explicit credentials
kspec scan azure subscription <subscription-id> \
  --tenant-id <tenant-id> \
  --client-id <client-id> \
  --token <client-secret> \
  -f <policy-file>
```

### Options

| Flag | Description |
|------|-------------|
| `-f, --policy` | Policy file to use |
| `-d, --policy-dir` | Directory containing policy files |
| `--tenant-id` | Azure tenant ID |
| `--client-id` | Service principal client ID |
| `--token` | Client secret |
| `--resource-group` | Filter to specific resource group |
| `--credential-type` | Credential type (`env` for environment-based) |

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Azure Security Scan

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

      - name: Run Security Scan
        env:
          AZURE_TENANT_ID: ${{ secrets.AZURE_TENANT_ID }}
          AZURE_CLIENT_ID: ${{ secrets.AZURE_CLIENT_ID }}
          AZURE_CLIENT_SECRET: ${{ secrets.AZURE_CLIENT_SECRET }}
        run: |
          ./kspec scan azure subscription ${{ secrets.AZURE_SUBSCRIPTION_ID }} \
            -f policies/azure-security.yml
```

### Azure DevOps

```yaml
trigger:
  - main

pool:
  vmImage: 'ubuntu-latest'

steps:
  - task: GoTool@0
    inputs:
      version: '1.21'

  - script: go build -o kspec ./cmd/kspec
    displayName: 'Build kspec'

  - script: |
      ./kspec scan azure subscription $(AZURE_SUBSCRIPTION_ID) \
        -f policies/azure-security.yml
    displayName: 'Run Security Scan'
    env:
      AZURE_TENANT_ID: $(AZURE_TENANT_ID)
      AZURE_CLIENT_ID: $(AZURE_CLIENT_ID)
      AZURE_CLIENT_SECRET: $(AZURE_CLIENT_SECRET)
```

---

## Troubleshooting

### Authentication Failed

```
Error: failed to create default Azure credential
```

**Causes:**
- Invalid credentials
- Missing environment variables
- Expired client secret

**Solutions:**
- Verify `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`
- Check credentials in Azure Portal
- Regenerate client secret if expired

### Permission Errors

```
Error: failed to list storage accounts: PERMISSION_DENIED
```

**Causes:**
- Service Principal lacks permissions
- Wrong subscription scope

**Solutions:**
- Assign Reader role to the subscription:
```bash
az role assignment create \
  --assignee <CLIENT_ID> \
  --role Reader \
  --scope /subscriptions/<SUBSCRIPTION_ID>
```

### No Resources Found

**Causes:**
- Empty subscription
- Wrong subscription ID
- Resource group filter too restrictive

**Solutions:**
- Verify resources exist in Azure Portal
- Check subscription ID is correct
- Remove or adjust resource group filter

### Auditing Policy Not Found

```
auditingPolicy: null
```

**Causes:**
- Auditing not configured
- Missing resource group in request

**Solutions:**
- Configure auditing for the SQL Server
- Ensure resource group is passed correctly
