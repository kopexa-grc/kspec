# Hetzner Cloud Provider

The Hetzner Cloud provider scans your Hetzner Cloud projects for security compliance, including servers, firewalls, networks, and SSH keys.

## Overview

Use the Hetzner Cloud provider to validate:
- Server security configurations (protection, backups, networks)
- Firewall rules and their application
- SSH key security (algorithm strength)
- Network isolation and segmentation
- Volume and IP address management

## Quick Start

```bash
# Set your API token
export HCLOUD_TOKEN="your-api-token"

# Scan with all Hetzner policies
kspec scan hetzner project -d policies

# Scan with specific policy
kspec scan hetzner project -f policies/hetzner-security.yml
```

## Prerequisites

- Hetzner Cloud account
- API Token with **Read** permissions

## Authentication

### Creating an API Token

1. Log in to the [Hetzner Cloud Console](https://console.hetzner.cloud/)
2. Select your project
3. Navigate to **Security** → **API Tokens**
4. Click **Generate API Token**
5. Name it (e.g., `kspec-scanner`)
6. Select **Read** permission
7. Click **Generate API Token**
8. Copy the token immediately (shown only once)

### Configuration Methods

**Environment Variable (Recommended):**
```bash
export HCLOUD_TOKEN="your-api-token"
kspec scan hetzner project -f policy.yml
```

**Alternative Environment Variable:**
```bash
export HETZNER_API_TOKEN="your-api-token"
kspec scan hetzner project -f policy.yml
```

**Command Line Flag:**
```bash
kspec scan hetzner project --api-token "your-api-token" -f policy.yml
```

---

## Resources

The Hetzner provider discovers the following resources:

### Compute

| Resource | Description |
|----------|-------------|
| `hcloud_server` | Cloud servers/instances |
| `hcloud_image` | Server images (system, snapshot, backup) |
| `hcloud_server_type` | Available server configurations |

### Storage

| Resource | Description |
|----------|-------------|
| `hcloud_volume` | Block storage volumes |

### Networking

| Resource | Description |
|----------|-------------|
| `hcloud_network` | Private networks (VPCs) |
| `hcloud_floating_ip` | Floating IP addresses |
| `hcloud_primary_ip` | Primary IP addresses |

### Security

| Resource | Description |
|----------|-------------|
| `hcloud_firewall` | Cloud firewalls with rules |
| `hcloud_ssh_key` | SSH public keys |

### Infrastructure

| Resource | Description |
|----------|-------------|
| `hcloud_location` | Data center locations |
| `hcloud_datacenter` | Physical datacenters |

---

## Resource Fields

### hcloud_server

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | Server ID |
| `name` | `string` | Server name |
| `status` | `string` | Status (`running`, `off`, `starting`, etc.) |
| `server_type` | `string` | Server type name |
| `location` | `string` | Location name |
| `datacenter` | `string` | Datacenter name |
| `public_ipv4` | `string` | Public IPv4 address |
| `public_ipv6` | `string` | Public IPv6 network |
| `labels` | `map` | Resource labels |
| `created` | `timestamp` | Creation time |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_running` | `bool` | Server is running |
| `is_locked` | `bool` | Server is locked |
| `has_backup_window` | `bool` | Backups are enabled |
| `has_public_ipv4` | `bool` | Has public IPv4 |
| `has_public_ipv6` | `bool` | Has public IPv6 |
| `has_private_network` | `bool` | Connected to private network |
| `has_delete_protection` | `bool` | Delete protection enabled |
| `has_rebuild_protection` | `bool` | Rebuild protection enabled |
| `has_labels` | `bool` | Has any labels |

### hcloud_firewall

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | Firewall ID |
| `name` | `string` | Firewall name |
| `rules` | `[]object` | Firewall rules |
| `labels` | `map` | Resource labels |
| `created` | `timestamp` | Creation time |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `rule_count` | `int` | Total number of rules |
| `inbound_rule_count` | `int` | Inbound rule count |
| `outbound_rule_count` | `int` | Outbound rule count |
| `has_rules` | `bool` | Has any rules defined |
| `is_applied` | `bool` | Applied to any resource |
| `allows_all_inbound` | `bool` | Allows 0.0.0.0/0 inbound |
| `allows_all_outbound` | `bool` | Allows 0.0.0.0/0 outbound |
| `has_ssh_rule` | `bool` | Has SSH (port 22) rule |
| `has_rdp_rule` | `bool` | Has RDP (port 3389) rule |

### hcloud_ssh_key

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | SSH key ID |
| `name` | `string` | Key name |
| `fingerprint` | `string` | Key fingerprint |
| `public_key` | `string` | Public key data |
| `labels` | `map` | Resource labels |
| `created` | `timestamp` | Creation time |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `key_type` | `string` | Key type (`rsa`, `ed25519`, `ecdsa`, `dsa`) |
| `is_rsa` | `bool` | Is RSA key |
| `is_ed25519` | `bool` | Is Ed25519 key |
| `is_ecdsa` | `bool` | Is ECDSA key |
| `is_dsa` | `bool` | Is DSA key |
| `is_weak_key_type` | `bool` | DSA or older RSA (prefer ed25519) |

### hcloud_volume

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | Volume ID |
| `name` | `string` | Volume name |
| `size` | `int` | Size in GB |
| `location` | `string` | Location name |
| `linux_device` | `string` | Linux device path |
| `labels` | `map` | Resource labels |
| `created` | `timestamp` | Creation time |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_attached` | `bool` | Attached to a server |
| `has_delete_protection` | `bool` | Delete protection enabled |
| `has_labels` | `bool` | Has any labels |

### hcloud_network

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | Network ID |
| `name` | `string` | Network name |
| `ip_range` | `string` | IP range (CIDR) |
| `subnets` | `[]object` | Subnet definitions |
| `labels` | `map` | Resource labels |
| `created` | `timestamp` | Creation time |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `subnet_count` | `int` | Number of subnets |
| `has_subnets` | `bool` | Has any subnets |
| `has_delete_protection` | `bool` | Delete protection enabled |

### hcloud_floating_ip

| Field | Type | Description |
|-------|------|-------------|
| `id` | `int64` | Floating IP ID |
| `ip` | `string` | IP address |
| `type` | `string` | IP type (`ipv4`, `ipv6`) |
| `location` | `string` | Location name |
| `dns_ptr` | `string` | Reverse DNS pointer |
| `labels` | `map` | Resource labels |
| `created` | `timestamp` | Creation time |

**Computed Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `is_assigned` | `bool` | Assigned to a server |
| `has_dns_ptr` | `bool` | Has reverse DNS configured |
| `has_delete_protection` | `bool` | Delete protection enabled |
| `is_blocked` | `bool` | IP is blocked |

---

## Example Policies

### Server Security

```yaml
queries:
  - uid: server-delete-protection
    title: Servers should have delete protection
    resource: hcloud_server
    severity: medium
    query: resource.has_delete_protection == true
    docs: |
      Delete protection prevents accidental deletion of servers.
    remediation: |
      Enable delete protection in the Hetzner Cloud Console or via API.

  - uid: server-private-network
    title: Servers should use private networks
    resource: hcloud_server
    severity: medium
    query: resource.has_private_network == true
    docs: |
      Servers should be connected to private networks for
      internal communication instead of using public IPs.
    remediation: |
      Create a private network and attach the server to it.

  - uid: server-backups-enabled
    title: Servers should have backups enabled
    resource: hcloud_server
    severity: high
    query: resource.has_backup_window == true
    docs: |
      Automated backups protect against data loss.
    remediation: |
      Enable backups in the server settings.
```

### Firewall Security

```yaml
queries:
  - uid: firewall-no-allow-all
    title: Firewalls should not allow all inbound traffic
    resource: hcloud_firewall
    severity: critical
    query: resource.allows_all_inbound == false
    docs: |
      Allowing all inbound traffic (0.0.0.0/0) is a security risk.
    remediation: |
      Restrict inbound rules to specific IP ranges and ports.

  - uid: firewall-is-applied
    title: Firewalls should be applied to resources
    resource: hcloud_firewall
    severity: medium
    query: resource.is_applied == true
    docs: |
      Unused firewalls provide no protection.
    remediation: |
      Apply the firewall to servers that need protection.

  - uid: firewall-restrict-ssh
    title: SSH access should be restricted
    resource: hcloud_firewall
    severity: high
    query: |
      !resource.has_ssh_rule ||
      !resource.rules.exists(r,
        r.direction == 'in' &&
        r.port == '22' &&
        r.source_ips.exists(ip, ip == '0.0.0.0/0')
      )
    docs: |
      SSH should not be open to the entire internet.
    remediation: |
      Restrict SSH access to specific IP addresses or ranges.
```

### SSH Key Security

```yaml
queries:
  - uid: ssh-modern-algorithm
    title: SSH keys should use modern algorithms
    resource: hcloud_ssh_key
    severity: medium
    query: resource.is_weak_key_type == false
    docs: |
      Ed25519 keys are recommended over RSA and DSA.
      DSA is deprecated and RSA requires larger key sizes.
    remediation: |
      Generate a new Ed25519 key:
      ssh-keygen -t ed25519 -C "your@email.com"

  - uid: ssh-no-dsa
    title: SSH keys should not use DSA
    resource: hcloud_ssh_key
    severity: high
    query: resource.is_dsa == false
    docs: |
      DSA keys are considered weak and deprecated.
    remediation: |
      Replace DSA keys with Ed25519 keys.
```

---

## CLI Reference

```bash
# Scan all resources in current project
kspec scan hetzner project -f <policy-file>

# Scan with named project (for identification)
kspec scan hetzner project my-project -f <policy-file>

# Using hcloud alias
kspec scan hcloud project -f <policy-file>

# With explicit token
kspec scan hetzner project --api-token <token> -f <policy-file>
```

### Options

| Flag | Description |
|------|-------------|
| `-f, --policy` | Policy file to use |
| `-d, --policy-dir` | Directory containing policy files |
| `--api-token` | Hetzner Cloud API token |
| `--hcloud-token` | Alternative for API token |
| `--project` | Project name (for identification) |

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Hetzner Security Scan

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

      - name: Scan Hetzner Cloud
        env:
          HCLOUD_TOKEN: ${{ secrets.HCLOUD_TOKEN }}
        run: ./kspec scan hetzner project -f policies/hetzner-security.yml
```

### GitLab CI

```yaml
hetzner-scan:
  image: golang:1.21
  script:
    - go build -o kspec ./cmd/kspec
    - ./kspec scan hetzner project -f policies/hetzner-security.yml
  variables:
    HCLOUD_TOKEN: $HCLOUD_TOKEN
```

---

## Troubleshooting

### No API Token Provided

```
Error: no API token provided
```

**Solution:** Set the API token via environment variable or flag:
```bash
export HCLOUD_TOKEN="your-token"
```

### Authentication Failed

```
Error: authentication failed
```

**Causes:**
- Invalid API token
- Token expired or revoked
- Token doesn't have read permissions

**Solutions:**
- Verify the token in Hetzner Cloud Console
- Generate a new token with read permissions

### No Resources Found

```
No resources found
```

**Causes:**
- Empty project
- Token doesn't have access to the project
- Wrong project selected

**Solutions:**
- Verify resources exist in the project
- Check token permissions
- Ensure you're scanning the correct project
