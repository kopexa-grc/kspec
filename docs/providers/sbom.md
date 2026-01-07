# SBOM Provider

The SBOM (Software Bill of Materials) provider scans SBOM files for security compliance, vulnerability analysis, and software composition validation. It supports CycloneDX and SPDX formats.

## Overview

Use the SBOM provider to validate:
- Software component inventory
- Known vulnerabilities in dependencies
- License compliance
- Dependency relationships
- Component versions and origins

## Quick Start

```bash
# Set SBOM file path
export SBOM_PATH="./sbom.json"

# Scan SBOM
kspec scan sbom file -f policies/sbom-security.yml

# Scan directory of SBOMs
export SBOM_PATH="./sboms/"
kspec scan sbom file -f policies/sbom-security.yml
```

## Prerequisites

- SBOM file(s) in CycloneDX or SPDX format
- JSON format supported

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SBOM_PATH` | Path to SBOM file or directory |

### Configuration Options

| Option | Description |
|--------|-------------|
| `sbom_path` | Path to SBOM file or directory containing SBOMs |

**Environment Variable:**
```bash
export SBOM_PATH="./sbom.json"
kspec scan sbom file -f policy.yml
```

**Directory Scanning:**
```bash
export SBOM_PATH="./artifacts/sboms/"
kspec scan sbom file -f policy.yml
```

---

## Supported Formats

### CycloneDX

CycloneDX is a lightweight SBOM standard designed for use in application security contexts and supply chain component analysis.

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "components": [...]
}
```

### SPDX

SPDX (Software Package Data Exchange) is an open standard for communicating software bill of materials information.

```json
{
  "spdxVersion": "SPDX-2.3",
  "packages": [...]
}
```

---

## Resources

The SBOM provider discovers the following resources:

| Resource | Description |
|----------|-------------|
| `sbom_document` | SBOM document metadata |
| `sbom_component` | Software components/packages |
| `sbom_vulnerability` | Known vulnerabilities |
| `sbom_dependency` | Dependency relationships |

---

## Resource Fields

### sbom_document

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Document identifier |
| `format` | string | SBOM format (cyclonedx, spdx) |
| `spec_version` | string | Specification version |
| `serial_number` | string | Document serial number |
| `created` | string | Creation timestamp |
| `tool_name` | string | Tool that generated the SBOM |
| `component_count` | number | Number of components |

### sbom_component

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Component identifier |
| `name` | string | Component/package name |
| `version` | string | Component version |
| `type` | string | Component type (library, application, etc.) |
| `purl` | string | Package URL (purl) |
| `license` | string | License identifier |
| `supplier` | string | Component supplier |
| `has_vulnerabilities` | bool | Whether component has known vulnerabilities |
| `vulnerability_count` | number | Number of known vulnerabilities |

### sbom_vulnerability

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Vulnerability ID (CVE, etc.) |
| `source` | string | Vulnerability source |
| `severity` | string | Severity level (critical, high, medium, low) |
| `cvss_score` | number | CVSS score |
| `component_name` | string | Affected component |
| `component_version` | string | Affected version |
| `description` | string | Vulnerability description |
| `recommendation` | string | Remediation recommendation |

### sbom_dependency

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Dependency identifier |
| `ref` | string | Component reference |
| `depends_on` | array | List of dependencies |
| `is_direct` | bool | Whether it's a direct dependency |
| `depth` | number | Dependency depth in tree |

---

## Example Policies

### Vulnerability Management

```yaml
policies:
  - uid: sbom-security
    name: SBOM Security Policy
    version: 1.0.0
    require:
      - provider: sbom
    groups:
      - title: Vulnerability Management
        checks:
          - uid: no-critical-vulnerabilities
          - uid: no-high-vulnerabilities

queries:
  - uid: no-critical-vulnerabilities
    title: Ensure no critical vulnerabilities exist
    resource: sbom_vulnerability
    impact: 100
    query: |
      resource.severity != "critical"
    docs:
      desc: Critical vulnerabilities must be remediated immediately.
      remediation: Update the affected component to a patched version.

  - uid: no-high-vulnerabilities
    title: Ensure no high severity vulnerabilities
    resource: sbom_vulnerability
    impact: 90
    query: |
      resource.severity != "high"
    docs:
      desc: High severity vulnerabilities should be addressed promptly.
      remediation: Update the affected component or apply compensating controls.
```

### License Compliance

```yaml
queries:
  - uid: approved-licenses
    title: Ensure components use approved licenses
    resource: sbom_component
    impact: 70
    query: |
      resource.license == "" ||
      resource.license in ["MIT", "Apache-2.0", "BSD-3-Clause", "ISC"]
    docs:
      desc: Components should use approved open source licenses.
      remediation: Replace the component with an alternative using an approved license.
```

### Component Hygiene

```yaml
queries:
  - uid: components-have-version
    title: Ensure all components have versions
    resource: sbom_component
    impact: 60
    query: |
      has(resource.version) && resource.version != ""
    docs:
      desc: All components should have explicit versions for reproducibility.
      remediation: Pin component versions in dependency manifests.

  - uid: components-have-purl
    title: Ensure components have Package URLs
    resource: sbom_component
    impact: 50
    query: |
      has(resource.purl) && resource.purl != ""
    docs:
      desc: Components should have Package URLs for identification.
      remediation: Generate SBOM with a tool that produces Package URLs.
```

---

## Generating SBOMs

### Using Syft

```bash
# Generate CycloneDX SBOM for a container image
syft alpine:latest -o cyclonedx-json > sbom.json

# Generate SBOM for a directory
syft dir:./my-app -o cyclonedx-json > sbom.json
```

### Using Trivy

```bash
# Generate CycloneDX SBOM
trivy image --format cyclonedx alpine:latest > sbom.json

# Generate SPDX SBOM
trivy image --format spdx-json alpine:latest > sbom.json
```

### Using cdxgen

```bash
# Generate CycloneDX SBOM for a Node.js project
cdxgen -o sbom.json

# Generate for specific language
cdxgen -t python -o sbom.json
```

---

## Troubleshooting

### No SBOM Path

```
sbom: no SBOM file path provided
```

**Solutions:**
- Set `SBOM_PATH` environment variable
- Provide `sbom_path` in configuration

### Invalid Format

```
sbom: unknown format
```

**Solutions:**
- Verify SBOM is in CycloneDX or SPDX format
- Check JSON is valid
- Regenerate SBOM with a supported tool

### Empty Results

```
No resources found
```

**Solutions:**
- Verify SBOM file contains components
- Check file path is correct
- Ensure SBOM is not empty
