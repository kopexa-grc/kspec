# Microsoft 365 Provider Setup Guide

This guide explains how to configure and use the Microsoft 365 (MS365) provider in kspec to scan your Microsoft 365 tenant for security compliance.

## Prerequisites

1. **Microsoft 365 Tenant**: An active Microsoft 365 tenant
2. **Azure AD App Registration**: A registered application with Microsoft Graph API permissions
3. **Admin Consent**: Global Administrator consent for required permissions

## Setting Up Azure AD App Registration

### Step 1: Create App Registration

1. Navigate to the **Azure Portal** (https://portal.azure.com)
2. Go to **Azure Active Directory** > **App registrations**
3. Click **New registration**
4. Configure the application:
   - **Name**: `kspec-scanner` (or your preferred name)
   - **Supported account types**: Accounts in this organizational directory only
   - **Redirect URI**: Leave blank (not needed for client credentials flow)
5. Click **Register**

### Step 2: Note Application Details

After registration, note down:
- **Application (client) ID**: Found on the Overview page
- **Directory (tenant) ID**: Found on the Overview page

### Step 3: Create Client Secret

1. Go to **Certificates & secrets**
2. Click **New client secret**
3. Add a description and select expiration period
4. Click **Add**
5. **Important**: Copy the secret value immediately (it won't be shown again)

### Step 4: Configure API Permissions

1. Go to **API permissions**
2. Click **Add a permission**
3. Select **Microsoft Graph**
4. Select **Application permissions**
5. Add the following permissions:

#### Required Permissions

| Permission | Description | Required For |
|------------|-------------|--------------|
| `Organization.Read.All` | Read organization information | Tenant info |
| `User.Read.All` | Read all users' full profiles | Users scan |
| `Group.Read.All` | Read all groups | Groups scan |
| `Application.Read.All` | Read all applications | Apps scan |
| `Directory.Read.All` | Read directory data | Roles, policies |
| `Policy.Read.All` | Read all policies | Conditional access, auth policies |
| `SecurityEvents.Read.All` | Read security events | Security alerts, risky users |
| `IdentityRiskEvent.Read.All` | Read identity risk events | Risky users |
| `Domain.Read.All` | Read domain data | Domains scan |

#### Optional Permissions (for full scanning)

| Permission | Description | Required For |
|------------|-------------|--------------|
| `Team.ReadBasic.All` | Read Teams info | Teams scan |
| `TeamSettings.Read.All` | Read Teams settings | Teams settings |
| `DeviceManagementConfiguration.Read.All` | Read Intune configs | Device configurations |
| `DeviceManagementManagedDevices.Read.All` | Read managed devices | Managed devices |
| `RoleManagement.Read.Directory` | Read directory role info | Role assignments |

### Step 5: Grant Admin Consent

1. Click **Grant admin consent for [Your Organization]**
2. Confirm by clicking **Yes**
3. Verify all permissions show "Granted for [Your Organization]"

## Authentication Methods

### Method 1: Client Credentials (Recommended for Automation)

```bash
kspec scan ms365 tenant <tenant-id> \
  --client-id <application-id> \
  --client-secret <client-secret> \
  -f policies/ms365-security.yml
```

**Example:**
```bash
kspec scan ms365 tenant 12345678-1234-1234-1234-123456789abc \
  --client-id 87654321-4321-4321-4321-cba987654321 \
  --client-secret "your-client-secret-value" \
  -f policies/ms365-security.yml
```

### Method 2: Environment Variables

```bash
# Set credentials as environment variables
export AZURE_TENANT_ID="12345678-1234-1234-1234-123456789abc"
export AZURE_CLIENT_ID="87654321-4321-4321-4321-cba987654321"
export AZURE_CLIENT_SECRET="your-client-secret-value"

# Run scan (uses DefaultAzureCredential)
kspec scan ms365 tenant $AZURE_TENANT_ID \
  -f policies/ms365-security.yml
```

### Method 3: Azure CLI Authentication (Development)

For development and testing, you can use Azure CLI authentication:

```bash
# Login to Azure
az login

# Run scan
kspec scan ms365 tenant <tenant-id> \
  -f policies/ms365-security.yml
```

## Scanning Commands

### Basic Scan

```bash
kspec scan ms365 tenant <tenant-id> \
  --client-id <client-id> \
  --client-secret <client-secret> \
  -f policies/ms365-security.yml
```

### Scan with Policy Directory

```bash
kspec scan ms365 tenant <tenant-id> \
  --client-id <client-id> \
  --client-secret <client-secret> \
  -d policies/
```

## Resources Discovered

The MS365 provider discovers and scans the following resources:

| Resource Type | Description | Required Permission |
|--------------|-------------|---------------------|
| `ms365_tenant` | Tenant/organization information | Organization.Read.All |
| `ms365_user` | User accounts and settings | User.Read.All |
| `ms365_group` | Microsoft 365 groups | Group.Read.All |
| `ms365_application` | Registered applications | Application.Read.All |
| `ms365_service_principal` | Service principals | Application.Read.All |
| `ms365_device` | Azure AD devices | Device.Read.All |
| `ms365_managed_device` | Intune managed devices | DeviceManagementManagedDevices.Read.All |
| `ms365_device_configuration` | Intune device configs | DeviceManagementConfiguration.Read.All |
| `ms365_domain` | Verified domains | Domain.Read.All |
| `ms365_conditional_access_policy` | CA policies | Policy.Read.All |
| `ms365_directory_role` | Directory roles | Directory.Read.All |
| `ms365_authorization_policy` | Authorization policy | Policy.Read.All |
| `ms365_security_defaults_policy` | Security defaults | Policy.Read.All |
| `ms365_risky_user` | Risky users | IdentityRiskEvent.Read.All |
| `ms365_secure_score` | Secure score | SecurityEvents.Read.All |
| `ms365_team` | Microsoft Teams | Team.ReadBasic.All |

## Security Checks

The MS365 security policy includes checks for:

### Identity Protection
- Sign-in risk policies enabled
- User risk policies enabled
- MFA enabled for administrators
- MFA enabled for all users

### Access Management
- Legacy authentication blocked
- Security defaults configuration
- Global admin count (2-4 recommended)
- Conditional access policies active
- Guest access restrictions

### Data Protection
- Password expiration policy
- SPF records for email domains
- Third-party app restrictions

### Device Management
- Android device encryption
- Minimum password length requirements

## Troubleshooting

### Authentication Errors

```
Error: failed to create client secret credential
```

**Solutions:**
- Verify tenant ID, client ID, and client secret are correct
- Check that the client secret hasn't expired
- Ensure the app registration exists in the correct tenant

### Permission Errors

```
Error: Insufficient privileges to complete the operation
```

**Solutions:**
- Verify all required API permissions are granted
- Ensure admin consent has been granted
- Check that the app has the correct permission type (Application, not Delegated)

### Specific Resource Errors

```
Error: failed to get teams: Access denied
```

**Solutions:**
- Add the specific permission for that resource type
- Grant admin consent for the new permission
- Some resources require additional licensing (e.g., Identity Protection requires Azure AD P2)

### License Requirements

Some features require specific Microsoft 365 licenses:

| Feature | Required License |
|---------|-----------------|
| Identity Protection (risky users) | Azure AD Premium P2 |
| Conditional Access | Azure AD Premium P1 |
| Intune Device Management | Microsoft Intune |
| Advanced Security Features | Microsoft 365 E5 Security |

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/ms365-security-scan.yml
name: MS365 Security Scan

on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6 AM
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

      - name: Run MS365 Security Scan
        env:
          AZURE_TENANT_ID: ${{ secrets.MS365_TENANT_ID }}
          AZURE_CLIENT_ID: ${{ secrets.MS365_CLIENT_ID }}
          AZURE_CLIENT_SECRET: ${{ secrets.MS365_CLIENT_SECRET }}
        run: |
          ./kspec scan ms365 tenant $AZURE_TENANT_ID \
            -f policies/ms365-security.yml
```

### Azure DevOps Pipeline

```yaml
# azure-pipelines.yml
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
      ./kspec scan ms365 tenant $(MS365_TENANT_ID) \
        --client-id $(MS365_CLIENT_ID) \
        --client-secret $(MS365_CLIENT_SECRET) \
        -f policies/ms365-security.yml
    displayName: 'Run MS365 Security Scan'
```

## Quick Start Script

```bash
#!/bin/bash
# ms365-scan.sh

# Configuration - Set these values or use environment variables
TENANT_ID="${MS365_TENANT_ID:-}"
CLIENT_ID="${MS365_CLIENT_ID:-}"
CLIENT_SECRET="${MS365_CLIENT_SECRET:-}"
POLICY_FILE="${1:-policies/ms365-security.yml}"

# Validate required parameters
if [ -z "$TENANT_ID" ] || [ -z "$CLIENT_ID" ] || [ -z "$CLIENT_SECRET" ]; then
    echo "Error: Missing required credentials"
    echo ""
    echo "Set the following environment variables:"
    echo "  export MS365_TENANT_ID=<your-tenant-id>"
    echo "  export MS365_CLIENT_ID=<your-client-id>"
    echo "  export MS365_CLIENT_SECRET=<your-client-secret>"
    echo ""
    echo "Usage: ./ms365-scan.sh [policy-file]"
    exit 1
fi

# Build if needed
if [ ! -f "./kspec" ]; then
    echo "Building kspec..."
    go build -o kspec ./cmd/kspec
fi

# Run scan
echo "Scanning Microsoft 365 tenant: $TENANT_ID"
./kspec scan ms365 tenant "$TENANT_ID" \
    --client-id "$CLIENT_ID" \
    --client-secret "$CLIENT_SECRET" \
    -f "$POLICY_FILE"
```

Make executable and run:
```bash
chmod +x ms365-scan.sh
export MS365_TENANT_ID="your-tenant-id"
export MS365_CLIENT_ID="your-client-id"
export MS365_CLIENT_SECRET="your-secret"
./ms365-scan.sh policies/ms365-security.yml
```

## Best Practices

1. **Principle of Least Privilege**: Only request the permissions you need
2. **Rotate Secrets**: Set up secret rotation schedules (90 days recommended)
3. **Use Managed Identities**: When running in Azure, use managed identities instead of client secrets
4. **Monitor App Activity**: Review sign-in logs for your app registration
5. **Audit Permissions**: Regularly review and remove unused permissions
6. **Secure Storage**: Never commit secrets to source control; use secret managers
7. **Regular Scans**: Schedule daily or weekly scans to detect configuration drift

## Security Considerations

### App Registration Security

- Enable **Conditional Access** for the app if supported
- Configure **Certificate credentials** instead of client secrets for production
- Set **owner** for the app registration to track responsibility
- Enable **audit logging** for app activities

### Network Security

- Consider using **Private endpoints** if scanning from Azure
- Use **VPN or ExpressRoute** for on-premises scanning
- Implement **firewall rules** to restrict where the app can authenticate from

### Compliance

The MS365 provider helps you comply with:
- **CIS Microsoft 365 Benchmark**
- **ISO 27001** (Information Security)
- **NIST 800-53** (Security Controls)
- **SOC 2** (Trust Services Criteria)
- **GDPR** (Data Protection)
