# kspec Documentation

Welcome to the kspec documentation. kspec is an enterprise-grade policy-as-code engine for security compliance scanning across cloud platforms, SaaS applications, and infrastructure.

## Getting Started

- **[Quickstart Guide](QUICKSTART.md)** - Get up and running in 5 minutes
- **[Installation](installation.md)** - Detailed installation instructions
- **[Writing Policies](policies.md)** - Learn to write security policies

## Provider Guides

kspec supports multiple providers for scanning different platforms:

| Provider | Description | Guide |
|----------|-------------|-------|
| **Network** | TLS, certificates, DNS, HTTP security | [Network Guide](providers/network.md) |
| **Azure** | Azure cloud resources and configurations | [Azure Guide](providers/azure.md) |
| **Microsoft 365** | M365 identity, security, and compliance | [MS365 Guide](providers/ms365.md) |
| **GitHub** | Organizations, repositories, and security settings | [GitHub Guide](providers/github.md) |
| **Hetzner Cloud** | Servers, firewalls, networks, and storage | [Hetzner Guide](providers/hetzner.md) |
| **Cloudflare** | DNS, WAF, zones, and security settings | [Cloudflare Guide](providers/cloudflare.md) |
| **Atlassian** | Jira, Confluence, and admin settings | [Atlassian Guide](providers/atlassian.md) |

## Concepts

- **[Discovery & Scanning](concepts/discovery-scan.md)** - How resource discovery works
- **[Sub-Resources](concepts/sub-resources.md)** - Understanding resource hierarchies
- **[CEL Expressions](concepts/cel-expressions.md)** - Writing policy queries

## Reference

- **[CLI Reference](reference/cli.md)** - Command line options
- **[Policy Schema](reference/policy-schema.md)** - Policy file format
- **[Resource Types](reference/resources.md)** - All available resources

## Integration

- **[CI/CD Integration](integration/cicd.md)** - GitHub Actions, GitLab CI, etc.
- **[API Usage](integration/api.md)** - Using kspec as a library

## Security

- **[Security Policy](../SECURITY.md)** - Reporting vulnerabilities
- **[Supply Chain](security/supply-chain.md)** - SBOM, signatures, provenance
