
<div align="center">
  <img src="docs/banner.png" alt="kspec banner" width="100%" />

  # kspec
  
  **The Enterprise-Grade Policy-as-Code Engine.**

  [![License](https://img.shields.io/badge/License-PolyForm%20Noncommercial-blue.svg)](LICENSE)
  [![Go Report Card](https://goreportcard.com/badge/github.com/cnspec/kspec)](https://goreportcard.com/report/github.com/cnspec/kspec)
  
  <p align="center">
    <b>Validate. Secure. Comply.</b><br />
    A modern, extensible framework for defining and enforcing security policies across your digital infrastructure.
  </p>
</div>

---

## 🚀 Overview

**kspec** is a powerful policy engine designed to bridge the gap between complex security requirements and automated validation. specific for cloud-native environments, it allows organizations to define security posture as code, ensuring consistent enforcement across network, cloud, and application layers.

Whether you are auditing TLS configurations, verifying email security standards (SPF/DMARC), or enforcing corporate compliance, **kspec** provides the primitives to build, test, and run policies at scale.

## ✨ Key Features

-   **🔒 Context-Aware Security**: Go beyond simple checks. Validate TLS negotiation parameters, certificate chains, DNS records, and HTTP headers with deep, structure-aware providers.
-   **📄 Policy-as-Code**: Define your security expectations in clear, version-controlled YAML. Treat compliance as code.
-   **🔌 Extensible Provider Architecture**: Modular design allows for easy addition of new asset types and verification logic (Network, HTTP, Cloud, etc.).
-   **⚡ High Performance**: Built in Go for speed, portability, and minimal overhead.

## 🛠️ Usage

### Quick Scan
Validate a target host against a specific policy file:

```bash
# Check Google for Email Security Compliance (SPF, DMARC)
go run ./cmd/cnspec scan host google.com -f policies/email-security.policy.yaml
```

### Protocol Validation
Ensure your web servers meet modern TLS standards:

```bash
# Check TLS configuration
go run ./cmd/cnspec scan host example.com -f policies/tls_security.yaml
```

## 📦 Policy Library

The repository comes pre-loaded with essential security policies:

| Policy | Description |
| :--- | :--- |
| **[TLS Security](policies/tls_security.yaml)** | Validates Protocol Versions (TLS 1.2+), Cipher Suites, and Certificate validity. |
| **[Email Security](policies/email-security.policy.yaml)** | Checks for generic SPF records, DMARC enforcement, and DNS hygiene. |
| **[HTTP Security](policies/http_security.yaml)** | Verifies presence of security headers (HSTS, CSP, X-Frame-Options). |
| **[DNS Security](policies/dns_security.yaml)** | Validates DNSSEC and name server configurations. |

## 🏗️ Architecture

kspec operates on a **Provider-Resource-Query** model:
1.  **Providers** (e.g., `network`, `http`) connect to the target asset.
2.  **Resources** (e.g., `tls`, `headers`) expose structured data.
3.  **Policies** (MQL/YAML) define the expected state of those resources.

## 📄 License

This project is licensed under the **PolyForm Noncommercial License 1.0.0**.

-   **Non-commercial use:** You are free to use, modify, and distribute this software for non-commercial purposes.
-   **Commercial use:** Commercial use is strictly prohibited without a separate commercial license.

### Attribution
If you use this software in your project or product (where permitted), you must include a backlink or note referencing this repository.