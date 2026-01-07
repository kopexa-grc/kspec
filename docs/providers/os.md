# OS Provider

The OS (Operating System) provider scans local system resources for security compliance, including services, packages, files, and macOS-specific resources.

## Overview

Use the OS provider to validate:
- System services and their status
- Installed packages and versions
- File permissions and ownership
- macOS AppleCare and warranty status

## Quick Start

```bash
# Scan local system
kspec scan os local -f policies/os-security.yml

# Scan specific aspects
kspec scan os local -f policies/service-compliance.yml
```

## Prerequisites

- Local system access
- Appropriate permissions to read system information

## Authentication

The OS provider operates on the local system and typically requires no authentication. However, some resources may require elevated privileges:

```bash
# Run with elevated privileges for full access
sudo kspec scan os local -f policy.yml
```

---

## Resources

The OS provider discovers the following resources:

### System

| Resource | Description |
|----------|-------------|
| `os_service` | System services (systemd, launchd, etc.) |
| `os_package` | Installed packages |
| `os_file` | Files and directories |

### macOS Specific

| Resource | Description |
|----------|-------------|
| `os_applecare` | AppleCare warranty and coverage status |

---

## Resource Fields

### os_service

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Service identifier |
| `name` | string | Service name |
| `status` | string | Service status (running, stopped, etc.) |
| `enabled` | bool | Whether service starts on boot |
| `type` | string | Service type (systemd, launchd, etc.) |

### os_package

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Package identifier |
| `name` | string | Package name |
| `version` | string | Installed version |
| `manager` | string | Package manager (apt, brew, etc.) |

### os_file

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | File path |
| `path` | string | Full file path |
| `permissions` | string | File permissions (octal) |
| `owner` | string | File owner |
| `group` | string | File group |
| `size` | number | File size in bytes |
| `is_directory` | bool | Whether path is a directory |

### os_applecare

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Device serial number |
| `serial_number` | string | Device serial number |
| `coverage_status` | string | Coverage status |
| `coverage_end_date` | string | Coverage end date |

---

## Example Policies

### Service Security

```yaml
policies:
  - uid: os-service-security
    name: OS Service Security
    version: 1.0.0
    require:
      - provider: os
    groups:
      - title: Service Security
        checks:
          - uid: ssh-service-enabled
          - uid: firewall-enabled

queries:
  - uid: ssh-service-enabled
    title: Ensure SSH service is properly configured
    resource: os_service
    impact: 80
    query: |
      resource.name != "sshd" || resource.status == "running"
    docs:
      desc: SSH service should be running for remote management.
      remediation: Enable and start the SSH service.

  - uid: firewall-enabled
    title: Ensure firewall is enabled
    resource: os_service
    impact: 90
    query: |
      resource.name != "firewalld" ||
      (resource.status == "running" && resource.enabled == true)
    docs:
      desc: System firewall should be enabled and running.
      remediation: Enable and start the firewall service.
```

### File Permissions

```yaml
queries:
  - uid: sensitive-file-permissions
    title: Ensure sensitive files have correct permissions
    resource: os_file
    impact: 85
    query: |
      !resource.path.endsWith("/etc/shadow") ||
      resource.permissions == "0640" || resource.permissions == "0600"
    docs:
      desc: Sensitive files like /etc/shadow should have restrictive permissions.
      remediation: |
        Set correct permissions:
        chmod 640 /etc/shadow
```

### Package Compliance

```yaml
queries:
  - uid: package-version-check
    title: Ensure critical packages are up to date
    resource: os_package
    impact: 70
    query: |
      resource.name != "openssl" ||
      resource.version.startsWith("3.")
    docs:
      desc: OpenSSL should be version 3.x for latest security features.
      remediation: Update OpenSSL to version 3.x or later.
```

---

## Platform Support

### Linux (systemd)

The OS provider uses systemd for service discovery on Linux systems:

```bash
# Services are discovered via systemctl
systemctl list-units --type=service
```

### macOS (launchd)

On macOS, the provider uses launchd:

```bash
# Services are discovered via launchctl
launchctl list
```

### Package Managers

Supported package managers:
- **apt** (Debian/Ubuntu)
- **yum/dnf** (RHEL/CentOS/Fedora)
- **brew** (macOS Homebrew)
- **pacman** (Arch Linux)

---

## Troubleshooting

### Permission Denied

```
os: permission denied reading service status
```

**Solutions:**
- Run kspec with elevated privileges (sudo)
- Check file permissions for the scanning user

### Service Not Found

```
os: service not found
```

**Solutions:**
- Verify the service exists on the system
- Check if the service manager is supported
- Use the correct service name for the platform
