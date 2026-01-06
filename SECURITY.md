# Security Policy

## Supported Versions

We release patches for security vulnerabilities. The following versions are currently being supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of kspec seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to:

**security@kopexa.com**

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

### What to Include

Please include the following information in your report:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit the issue

### What to Expect

- A confirmation of your report within 48 hours
- An initial assessment of the vulnerability within 7 days
- Regular updates on the progress of fixing the vulnerability
- Credit in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

When using kspec, we recommend:

1. **Protect API Tokens**: Store tokens in environment variables or secret management systems, never in code
2. **Use Least Privilege**: Create API tokens with minimum required permissions
3. **Rotate Credentials**: Regularly rotate API tokens and credentials
4. **Audit Access**: Monitor and log who is running security scans
5. **Verify Releases**: Use our signed releases and verify checksums

## Release Verification

All releases are signed using [Sigstore Cosign](https://www.sigstore.dev/). To verify:

```bash
# Download signature files from release
cosign verify-blob \
  --signature checksums.txt.sig \
  --certificate checksums.txt.pem \
  --certificate-identity-regexp "https://github.com/kopexa-grc/kspec" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  checksums.txt
```

## SLSA Provenance

Releases include [SLSA Level 3](https://slsa.dev/) provenance attestations. Download the `.intoto.jsonl` file from releases to verify build provenance.

## Disclosure Policy

When we receive a security bug report, we will:

1. Confirm the problem and determine the affected versions
2. Audit code to find any similar problems
3. Prepare fixes for all supported versions
4. Release patches as soon as possible
5. Publish a security advisory on GitHub

## Comments on this Policy

If you have suggestions on how this process could be improved, please submit a pull request or open an issue.
