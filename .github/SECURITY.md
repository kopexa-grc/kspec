# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

**DO NOT** create a public GitHub issue for security vulnerabilities.

Instead, please report security vulnerabilities through one of these channels:

1. **GitHub Security Advisories** (Preferred):
   - Go to [Security Advisories](https://github.com/kopexa-grc/kspec/security/advisories/new)
   - Click "Report a vulnerability"
   - Fill out the form with details

2. **Email**:
   - Send details to: security@kopexa.com
   - Use subject: `[SECURITY] kspec vulnerability report`

### What to Include

Please include the following information:

- Type of vulnerability (e.g., injection, authentication bypass, etc.)
- Full path to the affected source file(s)
- Steps to reproduce the issue
- Proof-of-concept or exploit code (if available)
- Impact assessment
- Any suggested fixes

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution Target**: Within 30 days (depending on severity)

### Disclosure Policy

- We will acknowledge receipt of your report
- We will confirm the vulnerability and determine its impact
- We will release a fix and publicly disclose the issue
- We will credit you for the discovery (unless you prefer anonymity)

### Safe Harbor

We consider security research conducted in accordance with this policy to be:

- Authorized
- Not a violation of any terms of service
- Helpful to the security of our users

We will not pursue legal action against researchers who:

- Follow this disclosure policy
- Make a good faith effort to avoid privacy violations
- Do not exploit vulnerabilities beyond proof of concept
- Do not disclose vulnerabilities publicly before we've addressed them

## Security Best Practices

When using kspec:

1. **Credentials**: Never commit credentials to source control
2. **Permissions**: Use least-privilege service accounts
3. **Tokens**: Rotate tokens regularly
4. **Policies**: Review policy files for sensitive information before sharing

Thank you for helping keep kspec and its users safe!
