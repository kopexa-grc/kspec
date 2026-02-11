# Integration Guide

kspec can be used in two ways:

1. **Discovery Tool** — Inventory your infrastructure (what resources exist?)
2. **Check Engine** — Evaluate policies against resources (are they compliant?)

Both modes are available via the **CLI** (for CI/CD and shell scripts) and the **Go API** (for programmatic integration).

## Part 1: Discovery

### 1.1 CLI Discovery

The `kspec discover` command lists resources without evaluating policies:

```bash
# Table output (default)
kspec discover network example.com

# JSON output
kspec discover network example.com -o json

# Tree output
kspec discover network example.com -o tree

# With resource instances and graph data
kspec discover network example.com -o json --include-instances --graph

# Export to file
kspec discover network example.com -e inventory.json
```

For providers that require credentials:

```bash
# GitHub (uses GITHUB_TOKEN env var)
export GITHUB_TOKEN=ghp_xxx
kspec discover github org my-org

# AWS (uses standard AWS credential chain)
kspec discover aws account production
```

### 1.2 Go API Discovery

Import the discovery package and at least one provider:

```go
import (
    "context"

    "github.com/kopexa-grc/kspec/discovery"
    _ "github.com/kopexa-grc/kspec/provider/all" // Register providers
)
```

**Basic discovery:**

```go
result, err := discovery.Discover(ctx,
    "network",                             // provider name
    map[string]string{"target": "example.com"}, // config
    "host",                                // asset type
    "example.com",                         // asset name
)
// result.Resources contains []ResourceInfo with Type and Count
```

**Quick inventory** (returns `map[string]int`):

```go
inventory, err := discovery.QuickInventory(ctx,
    "network",
    map[string]string{"target": "example.com"},
    "host",
    "example.com",
)
// inventory: map[string]int{"tls": 1, "dns": 1, "http": 1, ...}
```

**Discovery with graph** (includes resource instances):

```go
result, err := discovery.DiscoverWithInstances(ctx,
    "network",
    map[string]string{"target": "example.com"},
    "host",
    "example.com",
)
// result.Graph contains Nodes and Edges for visualization
```

**Full control with options:**

```go
result, err := discovery.DiscoverWithOptions(ctx, discovery.Config{
    ProviderName:     "network",
    ProviderConfig:   map[string]string{"target": "example.com"},
    Asset:            core.Asset{Type: "host", Name: "example.com", Config: map[string]string{"target": "example.com"}},
    Mode:             discovery.ModeNative,  // or ModeFetch, ModeRegistry
    Concurrency:      4,
    IncludeInstances: true,
})
```

**Export results:**

```go
// To stdout as a table
discovery.Export(result, discovery.FormatTable, os.Stdout)

// To stdout as JSON
discovery.Export(result, discovery.FormatJSON, os.Stdout)

// Export just the graph
discovery.ExportGraph(result.Graph, os.Stdout)
```

See [`_examples/discovery/`](../_examples/discovery/) for complete working examples.

### 1.3 Discovery Result Types

```go
// Result is the top-level discovery result.
type Result struct {
    Provider     string         // "network", "github", "aws", ...
    AssetType    string         // "host", "github-org", ...
    AssetName    string         // "example.com", "my-org", ...
    DiscoveredAt time.Time
    Duration     time.Duration
    Resources    []ResourceInfo // Discovered resource types
    TotalCount   int            // Sum of all counts
    Graph        *Graph         // Only when IncludeInstances=true
    Errors       []Error        // Non-fatal errors
}

// ResourceInfo describes one resource type.
type ResourceInfo struct {
    Type           string             // "tls", "dns", "github_repo", ...
    Count          int                // Number of instances (-1 = unknown)
    SupportsNative bool               // Has native discovery support
    Instances      []ResourceInstance // Only when IncludeInstances=true
}

// Graph for visualization.
type Graph struct {
    Nodes []GraphNode // Resources
    Edges []GraphEdge // Relationships (contains, owns, ...)
}
```

## Part 2: Check Engine

### 2.1 CLI Scanning

The `kspec scan` command discovers resources and evaluates policies:

```bash
# Scan with a policy file
kspec scan network example.com -f policy.yaml

# Scan with a policy directory
kspec scan network example.com -d ./policies

# Non-interactive mode (for CI/CD)
kspec scan network example.com -f policy.yaml --no-ui

# Export results
kspec scan network example.com -f policy.yaml --no-ui -o report.json
kspec scan network example.com -f policy.yaml --no-ui -o report.csv
kspec scan network example.com -f policy.yaml --no-ui -o report.html
kspec scan network example.com -f policy.yaml --no-ui -o report.xlsx

# Override scoring system
kspec scan network example.com -f policy.yaml --no-ui --scoring average
```

The `--no-ui` flag disables the interactive TUI and outputs structured logs — essential for CI/CD environments.

### 2.2 Go API Scanning

```go
import (
    "context"

    "github.com/kopexa-grc/kspec/core"
    "github.com/kopexa-grc/kspec/policy"
    "github.com/kopexa-grc/kspec/provider/scanner"
    _ "github.com/kopexa-grc/kspec/provider/all"
)
```

**End-to-end scan:**

```go
// 1. Load policies
policies, err := policy.Load("policy.yaml", "")
policies = policy.FilterByProvider(policies, "network", "network")

// 2. Create scanner
s := scanner.NewScanner(scanner.ScanConfig{
    ProviderName:   "network",
    ProviderConfig: map[string]string{"target": "example.com"},
    Asset: core.Asset{
        Type:   "host",
        Name:   "example.com",
        Config: map[string]string{"target": "example.com"},
    },
    Policies: policies,
})

// 3. Initialize
if err := s.Initialize(ctx); err != nil {
    log.Fatal(err)
}

// 4. Run
result := s.Run(ctx)

// result.Tree contains the resource tree with check results
// result.Errors contains any scan errors
```

**With event handling** (progress tracking):

```go
s.OnEvent(func(event scanner.ScanEvent) {
    switch event.Type {
    case scanner.EventDiscoveryStarted:
        log.Println("Discovery started")
    case scanner.EventResourceScanning:
        log.Printf("Scanning: %s", event.ResourceType)
    case scanner.EventScanComplete:
        log.Println("Scan complete")
    case scanner.EventError:
        log.Printf("Error: %v", event.Error)
    }
})
```

**Loading policies from different sources:**

```go
// From a single file
policies, err := policy.Load("policy.yaml", "")

// From a directory (loads all .yaml/.yml files)
policies, err := policy.Load("", "./policies")

// From bytes (e.g., embedded policies)
p, err := policy.LoadFromBytes(yamlData)

// From a specific file path
p, err := policy.LoadFromFile("path/to/policy.yaml")
```

See [`_examples/check-engine/`](../_examples/check-engine/) for complete working examples.

### 2.3 Scoring

kspec calculates a security score (0–100) from scan results using the `scoring` package:

```go
import "github.com/kopexa-grc/kspec/policy/scoring"

scorer := scoring.NewGraphScorer(scoring.SystemBanded)
score := scorer.ScoreTree(result.Tree)

fmt.Printf("Score: %d/100 (Grade: %s)\n", score.Value, score.Grade)
fmt.Printf("Risk Level: %s\n", score.RiskLevel)
```

**Available scoring systems:**

| System | Description |
|--------|-------------|
| `scoring.SystemBanded` | Severity bands cap the max score (default) |
| `scoring.SystemAverage` | Weighted average by severity |
| `scoring.SystemDecayed` | Exponential decay per finding |
| `scoring.SystemHighest` | Only the highest severity matters |

**Accessing detailed findings:**

```go
findings := scorer.GetRootFindings(result.Tree)
fmt.Printf("Critical: %d, High: %d, Medium: %d, Low: %d\n",
    findings.Critical, findings.High, findings.Medium, findings.Low)
```

The scoring system can also be set in a policy file:

```yaml
apiVersion: kspec/v1
kind: Policy
metadata:
  name: my-policy
scoring_system: banded  # or: average, decayed, highest_impact
```

### 2.4 Exporting Reports

The `report` package converts scan results to various formats:

```go
import "github.com/kopexa-grc/kspec/report"

// Create a report from the resource tree
rep := report.FromResourceTreeWithScoring(result.Tree, "network", scoring.SystemBanded)

// JSON
report.NewJSONExporter(true).Export(rep, "report.json")

// CSV
report.NewCSVExporter().Export(rep, "report.csv")

// HTML (self-contained, dark-theme dashboard)
report.NewHTMLExporter().Export(rep, "report.html")

// Excel
report.NewXLSXExporter().Export(rep, "report.xlsx")
```

**Writing to an `io.Writer`** (e.g., HTTP response):

```go
report.NewJSONExporter(true).ExportToWriter(rep, w)
report.NewCSVExporter().ExportToWriter(rep, w)
report.NewHTMLExporter().ExportToWriter(rep, w)
```

**Report metadata:**

```go
rep.Metadata.Score         // uint32: 0-100
rep.Metadata.Grade         // string: A, B, C, D, F
rep.Metadata.RiskLevel     // string: None, Low, Medium, High, Critical
rep.Metadata.TotalChecks   // int
rep.Metadata.PassedChecks  // int
rep.Metadata.FailedChecks  // int
```

## Part 3: CI/CD Integration

### GitHub Actions

```yaml
- name: Install kspec
  run: go install github.com/kopexa-grc/kspec/cmd/kspec@latest

- name: Run scan
  run: |
    kspec scan network example.com \
      -d ./policies \
      --no-ui \
      -o report.json

- name: Upload report
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: kspec-report
    path: report.json
```

See [`_examples/ci-cd/github-actions.yml`](../_examples/ci-cd/github-actions.yml) for a complete workflow.

### GitLab CI

```yaml
kspec-scan:
  stage: scan
  image: golang:1.24
  script:
    - go install github.com/kopexa-grc/kspec/cmd/kspec@latest
    - kspec scan network example.com -d ./policies --no-ui -o report.json
  artifacts:
    when: always
    paths:
      - report.json
```

See [`_examples/ci-cd/gitlab-ci.yml`](../_examples/ci-cd/gitlab-ci.yml) for a complete pipeline.

### Exit Codes & Error Handling

| Exit Code | Meaning |
|-----------|---------|
| `0` | Scan completed successfully |
| `1` | Scan failed (configuration error, provider unavailable, etc.) |

To fail a pipeline based on score, check the report output:

```bash
score=$(jq '.metadata.score' report.json)
if [ "$score" -lt 40 ]; then
  echo "Score $score is below threshold"
  exit 1
fi
```

See [`_examples/ci-cd/scan.sh`](../_examples/ci-cd/scan.sh) for a reusable wrapper script.

## Part 4: Provider Registration

kspec uses Go's `init()` pattern for provider registration. Every program that uses kspec must import at least one provider package:

```go
// Import ALL providers (recommended for CLI tools):
import _ "github.com/kopexa-grc/kspec/provider/all"

// Or import only what you need:
import (
    _ "github.com/kopexa-grc/kspec/provider/network"
    _ "github.com/kopexa-grc/kspec/provider/github"
)
```

The `provider/all` package imports all available providers:

- `network` — TLS, DNS, HTTP, certificates (no credentials needed)
- `github` — Organizations, repositories, teams
- `aws` — S3, EC2, IAM, Lambda, and more
- `azure` — Storage, SQL, KeyVault, VMs, and more
- `ms365` — Users, groups, policies, security settings
- `hetzner` — Servers, firewalls, networks
- `cloudflare` — DNS, WAF, zones
- `atlassian` — Jira, Confluence

Without a provider import, the registry is empty and all operations will fail with a "provider not found" error.
