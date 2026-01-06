# Hetzner Cloud Provider Setup

This guide explains how to set up and use the Hetzner Cloud provider for kspec.

## Prerequisites

- A Hetzner Cloud account
- An API Token with read access to your project

## Authentication

### Creating an API Token

1. Log in to your [Hetzner Cloud Console](https://console.hetzner.cloud/)
2. Select your project
3. Navigate to **Security** > **API Tokens**
4. Click **Generate API Token**
5. Give it a name (e.g., "kspec-scanner")
6. Select **Read** permission (write is not required for scanning)
7. Click **Generate API Token**
8. Copy the token - you won't be able to see it again

### Configuration Options

You can provide credentials in several ways:

**Environment Variables (Recommended):**
```bash
export HCLOUD_TOKEN="your-api-token"
# Alternative:
export HETZNER_API_TOKEN="your-api-token"
```

**Command Line Flags:**
```bash
kspec scan hetzner project --api-token "your-api-token" -f policy.yml
# or
kspec scan hetzner project --hcloud-token "your-api-token" -f policy.yml
```

**Config Map:**
```yaml
# In your scan configuration
provider_config:
  api_token: "your-api-token"
  project: "my-project"
```

## Scanning Commands

### Scan all resources in a project
```bash
kspec scan hetzner project -f policy.yml
```

### Scan with a named project
```bash
kspec scan hetzner project my-project -f policy.yml
```

### Using hcloud alias
```bash
kspec scan hcloud project -f policy.yml
```

## Resources Discovered

The Hetzner provider discovers the following resource types:

### Infrastructure

| Resource Type | Description |
|--------------|-------------|
| `hcloud_location` | Data center locations (regions) |
| `hcloud_datacenter` | Physical datacenters |
| `hcloud_server_type` | Available server configurations |
| `hcloud_iso` | ISO images for server installation |

### Compute

| Resource Type | Description |
|--------------|-------------|
| `hcloud_server` | Cloud servers/instances |
| `hcloud_image` | Server images (system, snapshot, backup) |

### Storage

| Resource Type | Description |
|--------------|-------------|
| `hcloud_volume` | Block storage volumes |

### Networking

| Resource Type | Description |
|--------------|-------------|
| `hcloud_network` | Private networks (VPCs) |
| `hcloud_floating_ip` | Floating IP addresses |
| `hcloud_primary_ip` | Primary IP addresses |

### Security

| Resource Type | Description |
|--------------|-------------|
| `hcloud_firewall` | Cloud firewalls with rules |
| `hcloud_ssh_key` | SSH public keys |

## Resource Fields

### hcloud_server

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Server ID |
| `name` | string | Server name |
| `status` | string | Server status (running, off, etc.) |
| `server_type` | string | Server type name |
| `location` | string | Location name |
| `public_ipv4` | string | Public IPv4 address |
| `public_ipv6` | string | Public IPv6 address |
| `is_running` | bool | True if server is running |
| `is_locked` | bool | True if server is locked |
| `has_backup_window` | bool | True if backups enabled |
| `has_public_ipv4` | bool | True if has public IPv4 |
| `has_private_network` | bool | True if in private network |
| `has_delete_protection` | bool | True if delete protected |

### hcloud_firewall

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Firewall ID |
| `name` | string | Firewall name |
| `rules` | array | Firewall rules |
| `rule_count` | int | Total number of rules |
| `inbound_rule_count` | int | Number of inbound rules |
| `outbound_rule_count` | int | Number of outbound rules |
| `has_rules` | bool | True if has any rules |
| `is_applied` | bool | True if applied to resources |
| `allows_all_inbound` | bool | True if allows 0.0.0.0/0 inbound |
| `allows_all_outbound` | bool | True if allows 0.0.0.0/0 outbound |

### hcloud_ssh_key

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | SSH key ID |
| `name` | string | Key name |
| `fingerprint` | string | Key fingerprint |
| `key_type` | string | Key type (rsa, ed25519, ecdsa, dsa) |
| `is_rsa` | bool | True if RSA key |
| `is_ed25519` | bool | True if Ed25519 key |
| `is_weak_key_type` | bool | True if DSA or RSA (prefer ed25519) |

## Example Policies

### Security Policy Example

```yaml
name: hetzner-security
version: "1.0"
provider: hetzner

checks:
  - name: servers-have-delete-protection
    description: "Servers should have delete protection enabled"
    resource: hcloud_server
    expr: has_delete_protection == true
    severity: medium

  - name: no-overly-permissive-firewalls
    description: "Firewalls should not allow all inbound traffic"
    resource: hcloud_firewall
    expr: allows_all_inbound == false
    severity: high

  - name: prefer-modern-ssh-keys
    description: "SSH keys should use Ed25519 instead of RSA/DSA"
    resource: hcloud_ssh_key
    expr: is_weak_key_type == false
    severity: low

  - name: servers-have-private-networks
    description: "Servers should be connected to private networks"
    resource: hcloud_server
    expr: has_private_network == true
    severity: medium

  - name: volumes-have-labels
    description: "Volumes should have labels for organization"
    resource: hcloud_volume
    expr: has_labels == true
    severity: info

  - name: firewalls-are-applied
    description: "Firewalls should be applied to resources"
    resource: hcloud_firewall
    expr: is_applied == true
    severity: medium
```

## Troubleshooting

### "no API token provided"

Make sure you have set either:
- `HCLOUD_TOKEN` environment variable
- `HETZNER_API_TOKEN` environment variable
- `--api-token` or `--hcloud-token` flag

### "failed to connect"

- Verify your API token is valid
- Check that the token has read permissions
- Ensure you have network connectivity to Hetzner's API

### No resources found

- Verify you're scanning the correct project
- Check that resources exist in your Hetzner Cloud project
- Ensure the API token has access to the project

## CI/CD Integration

### GitHub Actions

```yaml
- name: Scan Hetzner Cloud
  env:
    HCLOUD_TOKEN: ${{ secrets.HCLOUD_TOKEN }}
  run: |
    kspec scan hetzner project -f policies/hetzner-security.yml
```

### GitLab CI

```yaml
hetzner-scan:
  script:
    - kspec scan hetzner project -f policies/hetzner-security.yml
  variables:
    HCLOUD_TOKEN: $HCLOUD_TOKEN
```
