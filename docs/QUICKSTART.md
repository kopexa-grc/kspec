# Quickstart: Scan a Host with kspec

This guide walks you through scanning a website's TLS/SSL and certificate configuration in under 5 minutes.

## 1. Build kspec

```bash
git clone https://github.com/kopexa-grc/kspec.git
cd kspec
go build -o kspec ./cmd/kspec
```

## 2. Run Your First Scan

Scan a host with the built-in TLS and certificate policies:

```bash
./kspec scan network host example.com -d policies
```

This scans `example.com` and evaluates all policies in the `policies/` directory.

### What Gets Checked

The host scan discovers these resources:

| Resource | Description |
|----------|-------------|
| `tls` | TLS versions, cipher suites, protocol configuration |
| `certificate` | X.509 certificates in the chain |
| `dns` | DNS records (A, AAAA, MX, TXT, NS, etc.) |
| `http` | HTTP headers, redirects, security headers |

### Included Policies

| Policy | Checks |
|--------|--------|
| `tls_security.yaml` | TLS versions, cipher suites, PFS, AEAD |
| `certificate_security.yaml` | Expiration, validity, signature algorithms |
| `dns_security.yaml` | DNS configuration |
| `http_security.yaml` | Security headers (HSTS, CSP, etc.) |

## 3. Scan a Specific Policy

To run only specific checks:

```bash
# TLS checks only
./kspec scan network host example.com -f policies/tls_security.yaml

# Certificate checks only
./kspec scan network host example.com -f policies/certificate_security.yaml

# Multiple policies
./kspec scan network host example.com -f policies/tls_security.yaml -f policies/certificate_security.yaml
```

## 4. Navigate the Results

kspec displays results in an interactive TUI:

```
kspec │ host > example.com                                    ✓ Complete

╭─ Resources ─────────────────────╮╭─ Checks ─────────────────────────────╮
│ ● tls (1)          18✓ 0✗      ││ Resource: tls                        │
│ ● certificate (2)  18✓ 0✗      ││ Total: 18  ✓ 18  ✗ 0  ⊘ 0            │
│ ● dns              8✓ 0✗       ││ ─────────────────────────────────────│
│ ● http             5✓ 0✗       ││ ✓ Avoid weak TLS versions [high]     │
╰─────────────────────────────────╯│ ✓ Include AEAD ciphers [medium]     │
                                   │ ✓ Include PFS ciphers [medium]      │
                                   │ ✓ Certificate not expired [critical]│
                                   ╰──────────────────────────────────────╯
```

### Keyboard Controls

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate resources/checks |
| `Tab` | Switch between panels |
| `Enter` | Drill into resource / View check details |
| `Esc` | Go back |
| `q` | Quit |

### Check Detail View

Press `Enter` on a check to see full details:

```
╭─ Checks ────────────────╮╭─ Details ───────────────────────────────────╮
│ ✓ Avoid weak TLS...     ││ Certificate must not be expired             │
│ ✓ Include AEAD...       ││                                             │
│ ✗ Avoid CBC mode        ││ Status: ✓ PASSED                            │
│ ✓ Certificate valid     ││ Severity: critical                          │
│                         ││                                             │
│                         ││ Description                                 │
│                         ││ Expired certificates cause browser warnings │
│                         ││ and prevent users from accessing your site. │
│                         ││                                             │
│                         ││ Remediation                                 │
│                         ││ Renew the certificate immediately.          │
╰─────────────────────────╯╰──────────────────────────────────────────────╯
```

Use `Tab` to switch focus between check list and details, then `↑` `↓` to navigate or scroll.

## 5. Example Output Interpretation

### Passed Check
```
✓ Avoid weak SSL and TLS versions [high]
```
The server only supports TLS 1.2 and TLS 1.3.

### Failed Check
```
✗ Avoid weak block cipher modes [medium]
```
The server supports CBC cipher suites. View details for remediation steps.

## 6. Write Custom Policies

Create `my-policy.yaml`:

```yaml
apiVersion: kopexa.io/v1alpha1
kind: Policy
metadata:
  name: my-tls-policy
  title: My TLS Policy
  version: 1.0.0

groups:
  - title: TLS Security
    filter: asset.type == 'host'
    checks:
      - uid: require-tls-1-3

queries:
  - uid: require-tls-1-3
    title: Require TLS 1.3 support
    resource: tls
    severity: high
    query: |
      resource.versions.exists(v, v == 'tls1.3')
    docs: Server must support TLS 1.3 for optimal security.
    remediation: Enable TLS 1.3 in your server configuration.
```

Run your custom policy:

```bash
./kspec scan network host example.com -f my-policy.yaml
```

## 7. Available Resources & Fields

### TLS Resource (`resource: tls`)

| Field | Type | Description |
|-------|------|-------------|
| `versions` | `[]string` | Supported TLS versions (`tls1.0`, `tls1.1`, `tls1.2`, `tls1.3`) |
| `ciphers` | `[]string` | Supported cipher suites |
| `certificates` | `[]object` | Certificate chain (see certificate fields) |

### Certificate Resource (`resource: certificate`)

| Field | Type | Description |
|-------|------|-------------|
| `subject.commonName` | `string` | Certificate CN |
| `issuer.commonName` | `string` | Issuer CN |
| `dnsNames` | `[]string` | Subject Alternative Names |
| `notBefore` | `time` | Valid from |
| `notAfter` | `time` | Valid until |
| `expiresIn.days` | `int` | Days until expiration |
| `validityDays` | `int` | Total validity period |
| `isExpired` | `bool` | Certificate has expired |
| `isExpiringSoon` | `bool` | Expires within 30 days |
| `isVerified` | `bool` | Chain verified against system roots |
| `isSelfSigned` | `bool` | Self-signed certificate |
| `domainMatches` | `bool` | CN/SAN matches target domain |
| `signatureAlgorithm` | `string` | e.g., `SHA256-RSA` |
| `publicKeyAlgorithm` | `string` | e.g., `RSA`, `ECDSA` |
| `isCA` | `bool` | Is a CA certificate |
| `is_leaf` | `bool` | Is the leaf certificate |

### DNS Resource (`resource: dns`)

| Field | Type | Description |
|-------|------|-------------|
| `records` | `map` | DNS records by type (A, AAAA, MX, TXT, NS, CNAME) |

### HTTP Resource (`resource: http`)

| Field | Type | Description |
|-------|------|-------------|
| `statusCode` | `int` | HTTP response status |
| `headers` | `map` | Response headers |
| `redirectsToHttps` | `bool` | HTTP redirects to HTTPS |

## Next Steps

- Explore other providers: [GitHub](github-setup.md), [Azure](azure-setup.md), [Hetzner](hetzner-setup.md)
- Read about [policy structure](../README.md#writing-policies)
- Check the [built-in policies](../policies/) for examples
