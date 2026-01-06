# Network Provider

The Network provider scans remote hosts for TLS/SSL configuration, certificate security, DNS records, and HTTP security headers.

## Overview

Use the Network provider to validate:
- TLS protocol versions and cipher suites
- X.509 certificate validity and security
- DNS record configuration
- HTTP security headers (HSTS, CSP, etc.)

## Quick Start

```bash
# Scan a host with all network policies
kspec scan host example.com -d policies

# Scan with specific policy
kspec scan host example.com -f policies/tls_security.yaml
```

## Prerequisites

- Network connectivity to the target host
- No authentication required (public endpoints)

## Resources

The Network provider discovers four resource types:

| Resource | Description |
|----------|-------------|
| `tls` | TLS/SSL connection details |
| `certificate` | X.509 certificates in the chain |
| `dns` | DNS records for the domain |
| `http` | HTTP response and security headers |

---

## TLS Resource

The `tls` resource contains information about the TLS connection to the host.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `versions` | `[]string` | Supported TLS versions (`tls1.0`, `tls1.1`, `tls1.2`, `tls1.3`) |
| `ciphers` | `[]string` | Supported cipher suites |
| `certificates` | `[]object` | Certificate chain (deprecated, use `certificate` resource) |

### Example Queries

```yaml
# Require TLS 1.2 or higher only
- uid: no-weak-tls
  title: Avoid weak SSL and TLS versions
  resource: tls
  query: |
    resource.versions.all(v, v == 'tls1.2' || v == 'tls1.3')

# Require AEAD ciphers
- uid: require-aead
  title: Must include AEAD ciphers
  resource: tls
  query: |
    resource.ciphers.exists(c, c.matches('(?i)gcm|chacha20'))

# No RC4 ciphers
- uid: no-rc4
  title: Avoid RC4 ciphers
  resource: tls
  query: |
    resource.ciphers.all(c, !c.matches('(?i)rc4'))
```

---

## Certificate Resource

The `certificate` resource represents X.509 certificates in the chain. Multiple certificates may be discovered (leaf, intermediate, root).

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `serialNumber` | `string` | Certificate serial number |
| `subject.commonName` | `string` | Subject Common Name (CN) |
| `subject.organization` | `[]string` | Subject Organization (O) |
| `issuer.commonName` | `string` | Issuer Common Name |
| `issuer.organization` | `[]string` | Issuer Organization |
| `notBefore` | `timestamp` | Valid from date |
| `notAfter` | `timestamp` | Valid until date |
| `dnsNames` | `[]string` | Subject Alternative Names (SANs) |
| `emailAddresses` | `[]string` | Email addresses in certificate |
| `ipAddresses` | `[]string` | IP addresses in certificate |
| `signatureAlgorithm` | `string` | Signature algorithm (e.g., `SHA256-RSA`) |
| `publicKeyAlgorithm` | `string` | Public key algorithm (e.g., `RSA`, `ECDSA`) |
| `version` | `int` | X.509 version |
| `isCA` | `bool` | Is a CA certificate |
| `keyUsage` | `[]string` | Key usage flags |
| `extKeyUsage` | `[]string` | Extended key usage |
| `fingerprint_sha256` | `string` | SHA-256 fingerprint |

### Computed Fields

| Field | Type | Description |
|-------|------|-------------|
| `expiresIn.days` | `int` | Days until expiration (negative if expired) |
| `validityDays` | `int` | Total validity period in days |
| `isExpired` | `bool` | Certificate has expired |
| `isExpiringSoon` | `bool` | Expires within 30 days |
| `hasLongValidity` | `bool` | Validity exceeds 398 days (CA/B Forum limit) |
| `isVerified` | `bool` | Chain verified against system roots |
| `isSelfSigned` | `bool` | Self-signed certificate |
| `domainMatches` | `bool` | CN/SAN matches target domain |
| `is_leaf` | `bool` | Is the leaf (end-entity) certificate |
| `chain_index` | `int` | Position in certificate chain (0 = leaf) |

### Example Queries

```yaml
# Certificate not expired
- uid: cert-not-expired
  title: Certificate must not be expired
  resource: certificate
  severity: critical
  query: |
    resource.is_leaf == false || resource.isExpired == false

# Certificate not expiring soon
- uid: cert-not-expiring
  title: Certificate should not expire within 30 days
  resource: certificate
  severity: high
  query: |
    resource.is_leaf == false || resource.isExpiringSoon == false

# Strong signature algorithm
- uid: strong-signature
  title: Use strong signature algorithm
  resource: certificate
  query: |
    !resource.signatureAlgorithm.matches('(?i)md2|md5|sha1')

# Domain must match
- uid: domain-match
  title: Certificate must match domain
  resource: certificate
  severity: critical
  query: |
    resource.is_leaf == false || resource.domainMatches == true

# Not self-signed
- uid: not-self-signed
  title: Do not use self-signed certificates
  resource: certificate
  query: |
    resource.is_leaf == false || resource.isSelfSigned == false
```

### Certificate Validity Timeline

Per CA/Browser Forum Ballot SC-081v3, maximum certificate validity is being reduced:

| Date | Maximum Validity |
|------|-----------------|
| Until March 2026 | 398 days |
| From March 2026 | 200 days |
| From March 2027 | 100 days |
| From March 2029 | 47 days |

---

## DNS Resource

The `dns` resource contains DNS records for the target domain.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `records.A` | `[]string` | IPv4 addresses |
| `records.AAAA` | `[]string` | IPv6 addresses |
| `records.MX` | `[]object` | Mail exchange records |
| `records.TXT` | `[]string` | TXT records (SPF, DKIM, etc.) |
| `records.NS` | `[]string` | Nameservers |
| `records.CNAME` | `[]string` | Canonical name records |
| `records.CAA` | `[]object` | Certificate Authority Authorization |

### Example Queries

```yaml
# Has SPF record
- uid: has-spf
  title: Domain should have SPF record
  resource: dns
  query: |
    resource.records.TXT.exists(t, t.startsWith('v=spf1'))

# Has CAA record
- uid: has-caa
  title: Domain should have CAA record
  resource: dns
  query: |
    size(resource.records.CAA) > 0
```

---

## HTTP Resource

The `http` resource contains HTTP response information and security headers.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `statusCode` | `int` | HTTP response status code |
| `headers` | `map[string]string` | Response headers |
| `redirectsToHttps` | `bool` | HTTP redirects to HTTPS |

### Example Queries

```yaml
# HSTS enabled
- uid: has-hsts
  title: Enable HTTP Strict Transport Security
  resource: http
  query: |
    has(resource.headers['Strict-Transport-Security'])

# X-Frame-Options set
- uid: has-x-frame-options
  title: Set X-Frame-Options header
  resource: http
  query: |
    has(resource.headers['X-Frame-Options'])

# Content-Security-Policy set
- uid: has-csp
  title: Set Content-Security-Policy header
  resource: http
  query: |
    has(resource.headers['Content-Security-Policy'])

# HTTP redirects to HTTPS
- uid: redirects-https
  title: HTTP should redirect to HTTPS
  resource: http
  query: |
    resource.redirectsToHttps == true
```

---

## Built-in Policies

kspec includes several network security policies:

| Policy | File | Description |
|--------|------|-------------|
| TLS Security | `policies/tls_security.yaml` | TLS versions, ciphers, PFS |
| Certificate Security | `policies/certificate_security.yaml` | Expiration, validity, signatures |
| HTTP Security | `policies/http_security.yaml` | Security headers |
| DNS Security | `policies/dns_security.yaml` | DNS configuration |
| Email Security | `policies/email-security.policy.yaml` | SPF, DMARC, DKIM |

---

## CLI Reference

```bash
# Basic host scan
kspec scan host <hostname> -f <policy-file>

# Scan with policy directory
kspec scan host <hostname> -d <policy-directory>

# Examples
kspec scan host example.com -f policies/tls_security.yaml
kspec scan host api.example.com -d policies
kspec scan host 192.168.1.1 -f my-policy.yaml
```

### Options

| Flag | Description |
|------|-------------|
| `-f, --policy` | Policy file to use |
| `-d, --policy-dir` | Directory containing policy files |

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Security Scan

on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6 AM

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

      - name: Scan TLS Configuration
        run: ./kspec scan host ${{ vars.TARGET_HOST }} -f policies/tls_security.yaml

      - name: Scan Certificates
        run: ./kspec scan host ${{ vars.TARGET_HOST }} -f policies/certificate_security.yaml
```

### GitLab CI

```yaml
security-scan:
  image: golang:1.21
  script:
    - go build -o kspec ./cmd/kspec
    - ./kspec scan host ${TARGET_HOST} -d policies
```

---

## Troubleshooting

### Connection Refused

```
Error: failed to connect to example.com:443
```

**Causes:**
- Host is not reachable
- Firewall blocking connection
- Wrong port

**Solutions:**
- Verify network connectivity: `curl -v https://example.com`
- Check firewall rules
- Ensure the host has TLS enabled

### Certificate Verification Failed

```
isVerified: false
```

**Causes:**
- Self-signed certificate
- Expired certificate
- Missing intermediate certificates

**Solutions:**
- Use a certificate from a trusted CA
- Ensure full certificate chain is served
- Renew expired certificates

### No DNS Records

```
Error: no DNS records found
```

**Causes:**
- Domain doesn't exist
- DNS resolution failed

**Solutions:**
- Verify domain name spelling
- Check DNS configuration: `dig example.com`
