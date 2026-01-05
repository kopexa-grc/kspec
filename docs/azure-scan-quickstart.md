# Azure Provider Testing Guide

## Prerequisites

Before testing the Azure provider, you need:

1. **Azure Subscription**: An active Azure subscription ID
2. **Azure Credentials**: Service Principal with Reader permissions

### Creating a Service Principal

```bash
# Create a service principal with Reader role
az ad sp create-for-rbac --name "kspec-reader" --role Reader --scopes /subscriptions/<SUBSCRIPTION_ID>
```

This will output:
```json
{
  "appId": "<CLIENT_ID>",
  "displayName": "kspec-reader",
  "password": "<CLIENT_SECRET>",
  "tenant": "<TENANT_ID>"
}
```

## Testing Methods

### Method 1: Using Environment Variables (Recommended)

```bash
# Set Azure credentials
export AZURE_SUBSCRIPTION_ID="<your-subscription-id>"
export AZURE_TENANT_ID="<your-tenant-id>"
export AZURE_CLIENT_ID="<your-client-id>"
export AZURE_CLIENT_SECRET="<your-client-secret>"

# Run kspec scan
./kspec scan azure subscription $AZURE_SUBSCRIPTION_ID \
  --credential-type env \
  -f policies/azure-security.yml
```

### Method 2: Using CLI Flags

```bash
./kspec scan azure subscription <SUBSCRIPTION_ID> \
  --tenant-id <TENANT_ID> \
  --client-id <CLIENT_ID> \
  --token <CLIENT_SECRET> \
  -f policies/azure-security.yml
```

### Method 3: Using Azure CLI Authentication

If you're already logged in with `az login`:

```bash
# Azure SDK will automatically use az cli credentials
az login

# Run scan (kspec will use DefaultAzureCredential)
./kspec scan azure subscription <SUBSCRIPTION_ID> \
  --credential-type env \
  -f policies/azure-security.yml
```

## Resource Group Filtering (Optional)

To scan only specific resource groups:

```bash
./kspec scan azure subscription <SUBSCRIPTION_ID> \
  --credential-type env \
  --resource-group <RESOURCE_GROUP_NAME> \
  -f policies/azure-security.yml
```

## Testing Individual Resource Types

You can test individual Azure resources by creating minimal policy files:

### Test Storage Accounts

```yaml
# test-storage.yml
name: Test Azure Storage
version: 1.0.0
groups:
  - title: Storage Test
    filter: asset.type == "azure-storage"
    checks:
      - uid: test-storage-https
queries:
  - uid: test-storage-https
    title: Test HTTPS enforcement
    resource: azure_storage_account
    query: |
      resource.properties.EnableHTTPSTrafficOnly == true
    severity: high
```

```bash
./kspec scan azure subscription <SUB_ID> \
  --credential-type env \
  -f test-storage.yml
```

### Test SQL Servers

```yaml
# test-sql.yml
name: Test Azure SQL
version: 1.0.0
groups:
  - title: SQL Test
    filter: asset.type == "azure-sql"
    checks:
      - uid: test-sql-audit
queries:
  - uid: test-sql-audit
    title: Test SQL Auditing
    resource: azure_sql_server
    query: |
      resource.auditingPolicy.state == "Enabled"
    severity: high
```

```bash
./kspec scan azure subscription <SUB_ID> \
  --credential-type env \
  -f test-sql.yml
```

## Expected Output

A successful scan should show:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Azure Storage                                                                │
│ ✓ Ensure HTTPS enforcement                                     PASSED       │
│ ✗ Ensure public blob access disabled                           FAILED       │
│                                                                              │
│ Azure SQL                                                                    │
│ ✓ Ensure SQL auditing enabled                                  PASSED       │
│ ✗ Ensure TDE enabled                                           FAILED       │
│                                                                              │
│ Azure Key Vault                                                              │
│ ✓ Ensure purge protection enabled                              PASSED       │
└─────────────────────────────────────────────────────────────────────────────┘

Total: 28  Passed: 12  Failed: 16  Pending: 0  Skipped: 0
```

## Troubleshooting

### Authentication Errors

```
Error: failed to create default Azure credential
```

**Solution**: Ensure your credentials are set correctly:
- Check `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`
- Or run `az login` for interactive authentication

### Permission Errors

```
Error: failed to list storage accounts: PERMISSION_DENIED
```

**Solution**: Your Service Principal needs at least **Reader** role on the subscription:

```bash
az role assignment create \
  --assignee <CLIENT_ID> \
  --role Reader \
  --scope /subscriptions/<SUBSCRIPTION_ID>
```

### No Resources Found

If you see "0 resources found", ensure:
1. Your subscription has the resource types you're scanning
2. Your service principal has permissions to list those resources
3. The resource group filter (if used) is correct

## Quick Test Script

Create `test-azure.sh`:

```bash
#!/bin/bash

# Azure credentials (set these!)
export AZURE_SUBSCRIPTION_ID="your-sub-id"
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"

# Build kspec
echo "Building kspec..."
go build -o kspec ./cmd/kspec

# Run Azure security scan
echo "Running Azure security scan..."
./kspec scan azure subscription $AZURE_SUBSCRIPTION_ID \
  --credential-type env \
  -d policies/ \
  | tee azure-scan-results.txt

echo "Scan complete! Results saved to azure-scan-results.txt"
```

Make it executable and run:
```bash
chmod +x test-azure.sh
./test-azure.sh
```

## Resource Types Supported

The Azure provider currently supports:

- ✅ `azure_storage_account` - Storage Accounts
- ✅ `azure_sql_server` - SQL Servers  
- ✅ `azure_keyvault_vault` - Key Vaults
- ✅ `azure_network_security_group` - Network Security Groups
- ✅ `azure_virtual_machine` - Virtual Machines
- ✅ `azure_subscription` - Subscription-level resources

## CI/CD Integration

For automated testing in CI/CD pipelines:

```yaml
# .github/workflows/azure-security-scan.yml
name: Azure Security Scan

on: [push, pull_request]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      
      - name: Build kspec
        run: go build -o kspec ./cmd/kspec
      
      - name: Run Azure Security Scan
        env:
          AZURE_SUBSCRIPTION_ID: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
          AZURE_TENANT_ID: ${{ secrets.AZURE_TENANT_ID }}
          AZURE_CLIENT_ID: ${{ secrets.AZURE_CLIENT_ID }}
          AZURE_CLIENT_SECRET: ${{ secrets.AZURE_CLIENT_SECRET }}
        run: |
          ./kspec scan azure subscription $AZURE_SUBSCRIPTION_ID \
            --credential-type env \
            -f policies/azure-security.yml
```
