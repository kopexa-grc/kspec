<div align="center">
  <img src="docs/banner.png" alt="kspec banner" width="100%" />

  # kspec

  **The Enterprise-Grade Policy-as-Code Engine.**

  [![License](https://img.shields.io/badge/License-Elastic%202.0-blue.svg)](LICENSE)
  [![Go Report Card](https://goreportcard.com/badge/github.com/kopexa-grc/kspec)](https://goreportcard.com/report/github.com/kopexa-grc/kspec)

  <p align="center">
    <b>Validate. Secure. Comply.</b><br />
    A modern, extensible framework for defining and enforcing security policies across your digital infrastructure.
  </p>
</div>

---

## Overview

**kspec** is a powerful policy engine designed to bridge the gap between complex security requirements and automated validation. Built for cloud-native environments, it allows organizations to define security posture as code, ensuring consistent enforcement across cloud platforms, SaaS applications, networks, and infrastructure.

Whether you are auditing cloud configurations, verifying GitHub repository security, enforcing Microsoft 365 compliance, or validating TLS settings, **kspec** provides the primitives to build, test, and run policies at scale.

## Key Features

- **Multi-Cloud Support**: Scan Azure subscriptions, Microsoft 365 tenants, and GitHub organizations from a single tool
- **Policy-as-Code**: Define your security expectations in clear, version-controlled YAML with CEL expressions
- **Extensible Provider Architecture**: Modular design with providers for Azure, MS365, GitHub, Network, and more
- **Interactive TUI**: Beautiful terminal UI showing real-time scan progress and results
- **High Performance**: Built in Go for speed, portability, and minimal overhead
- **CI/CD Ready**: Easy integration with GitHub Actions, Azure DevOps, and other CI/CD platforms

## Supported Providers

| Provider | Description | Documentation |
|----------|-------------|---------------|
| **Azure** | Scan Azure subscriptions for security compliance | [Setup Guide](docs/azure-setup.md) |
| **Microsoft 365** | Scan M365 tenants for identity and security settings | [Setup Guide](docs/ms365-setup.md) |
| **GitHub** | Scan organizations and repositories for security best practices | [Setup Guide](docs/github-setup.md) |
| **Network** | Validate TLS, DNS, and HTTP security configurations | - |
| **Local** | Scan local system configurations | - |

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/kopexa-grc/kspec.git
cd kspec

# Build
go build -o kspec ./cmd/kspec

# Verify installation
./kspec --help
```

## Quick Start

### Scan GitHub Organization

```bash
# Set your GitHub token
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Scan an organization
kspec scan github org <organization-name> -f policies/github-security.yml
```

### Scan Azure Subscription

```bash
# Set Azure credentials
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"

# Scan a subscription
kspec scan azure subscription <subscription-id> -f policies/azure-security.yml
```

### Scan Microsoft 365 Tenant

```bash
# Scan M365 tenant
kspec scan ms365 tenant <tenant-id> \
  --client-id <client-id> \
  --client-secret <client-secret> \
  -f policies/ms365-security.yml
```

### Scan Network Host

```bash
# Scan TLS and HTTP security
kspec scan host example.com -f policies/tls-security.yml
```

## Policy Library

The repository includes pre-built security policies:

| Policy | Provider | Description |
|--------|----------|-------------|
| [Azure Security](policies/azure-security.yml) | Azure | Storage encryption, SQL auditing, Key Vault protection, NSG rules |
| [MS365 Security](policies/ms365-security.yml) | MS365 | MFA enforcement, Conditional Access, identity protection, Teams security |
| [GitHub Security](policies/github-security.yml) | GitHub | Branch protection, 2FA, repository security settings |
| [TLS Security](policies/tls-security.yml) | Network | Protocol versions, cipher suites, certificate validation |
| [Email Security](policies/email-security.yml) | Network | SPF records, DMARC enforcement, DNS hygiene |

## Writing Policies

Policies are defined in YAML with CEL (Common Expression Language) queries:

```yaml
policies:
  - uid: my-security-policy
    name: My Security Policy
    version: 1.0.0
    require:
      - provider: azure
    groups:
      - title: Storage Security
        checks:
          - uid: storage-https-required

queries:
  - uid: storage-https-required
    title: Ensure HTTPS is required for storage accounts
    resource: azure_storage_account
    impact: 90
    query: |
      has(resource.properties) &&
      resource.properties.supportsHttpsTrafficOnly == true
    docs:
      desc: |
        Storage accounts should require HTTPS to encrypt data in transit.
      remediation: |
        Enable "Secure transfer required" in the storage account settings.
```

## Architecture

kspec operates on a **Provider-Resource-Policy** model:

1. **Providers** (Azure, MS365, GitHub, Network) connect to target assets
2. **Resources** expose structured data from the target (storage accounts, users, repos)
3. **Policies** define expected security state using CEL expressions
4. **Scanner** orchestrates discovery, fetching, and policy evaluation
5. **TUI** displays real-time progress and results

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Provider  │────▶│   Resources  │────▶│   Scanner   │
│  (Azure)    │     │  (Storage,   │     │  (Evaluate  │
│             │     │   SQL, etc)  │     │   Policies) │
└─────────────┘     └──────────────┘     └──────┬──────┘
                                                │
                                                ▼
                                         ┌─────────────┐
                                         │     TUI     │
                                         │  (Results)  │
                                         └─────────────┘
```

## Documentation

- [Azure Setup Guide](docs/azure-setup.md) - Configure Azure provider and credentials
- [Microsoft 365 Setup Guide](docs/ms365-setup.md) - Configure MS365 provider and app registration
- [GitHub Setup Guide](docs/github-setup.md) - Configure GitHub provider and tokens
- [Discovery & Scanning](docs/discovery-scan.md) - How resource discovery works
- [Sub-Resources](docs/sub-resources.md) - Understanding resource hierarchies

## CI/CD Integration

### GitHub Actions

```yaml
name: Security Scan

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

## Contributing

Contributions are welcome! Please read the license terms before contributing.

## License

This project is licensed under the **Elastic License 2.0 (ELv2)**.

- **Commercial use**: You are free to use this software commercially, including for auditing, consulting, and security assessments
- **Managed Service**: You may **not** provide this software as a hosted or managed service to third parties
- **Modifications**: You may modify and distribute the software, subject to the license terms

See [LICENSE](LICENSE) for the full license text.

### What's Allowed

- Using kspec internally at your company
- Using kspec to audit or assess client infrastructure (consultants, auditors)
- Modifying kspec for your own use
- Distributing kspec with your modifications (with license notices)

### What's Not Allowed

- Offering kspec as a hosted/managed service (SaaS)
- Removing or circumventing license functionality

---

<div align="center">
  <p>Built with care by <a href="https://github.com/kopexa-grc">Kopexa GRC</a></p>
</div>
