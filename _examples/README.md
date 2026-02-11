# kspec Examples

Working examples that demonstrate how to use kspec as a **Discovery Tool** and as a **Check Engine**, both via CLI and Go API.

## Quick Start

All Go examples live in the main module — no separate `go.mod` needed:

```bash
# From the repository root:
go run ./_examples/discovery/basic/ example.com
```

## Examples

### Discovery (Go API)

| Example | Description | Run |
|---------|-------------|-----|
| [discovery/basic](discovery/basic/) | `discovery.Discover()` with table output | `go run ./_examples/discovery/basic/ example.com` |
| [discovery/inventory](discovery/inventory/) | `discovery.QuickInventory()` → JSON | `go run ./_examples/discovery/inventory/ example.com` |
| [discovery/graph](discovery/graph/) | `discovery.DiscoverWithInstances()` + graph export | `go run ./_examples/discovery/graph/ example.com` |

### Check Engine (Go API)

| Example | Description | Run |
|---------|-------------|-----|
| [check-engine/basic](check-engine/basic/) | End-to-end scan with `scanner.NewScanner()` | `go run ./_examples/check-engine/basic/ example.com` |
| [check-engine/advanced](check-engine/advanced/) | Events, scoring, JSON/CSV report export | `go run ./_examples/check-engine/advanced/ example.com` |

### CI/CD

| File | Description |
|------|-------------|
| [ci-cd/github-actions.yml](ci-cd/github-actions.yml) | GitHub Actions workflow |
| [ci-cd/gitlab-ci.yml](ci-cd/gitlab-ci.yml) | GitLab CI pipeline |
| [ci-cd/scan.sh](ci-cd/scan.sh) | Shell script wrapper |

### Policies

| File | Description |
|------|-------------|
| [policies/minimal-tls.yaml](policies/minimal-tls.yaml) | Minimal TLS policy (3 checks) |
| [policies/minimal-github.yaml](policies/minimal-github.yaml) | Minimal GitHub org policy (3 checks) |

## Provider Registration

Every Go example that uses kspec must import the provider registration package:

```go
import (
    _ "github.com/kopexa-grc/kspec/provider/all"
)
```

This blank import triggers `init()` functions that register all providers (network, GitHub, AWS, Azure, etc.). Without it, `discovery.Discover()` and `scanner.NewScanner()` will fail because no providers are available.

If you only need a specific provider, import it directly:

```go
import (
    _ "github.com/kopexa-grc/kspec/provider/network"
)
```

## Using Examples as Standalone Projects

To use an example outside the kspec repository, create your own module:

```bash
mkdir my-scanner && cd my-scanner
go mod init my-scanner
go get github.com/kopexa-grc/kspec@latest
```

Then copy the example `main.go` and run it:

```bash
go run . example.com
```
