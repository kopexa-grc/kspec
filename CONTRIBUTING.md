# Contributing to kspec

Thank you for your interest in contributing to kspec! This guide covers the most common contribution: adding a new provider.

## Adding a New Provider

kspec uses a self-registration pattern for providers. Adding a new provider requires only two files:

1. `provider/<name>/provider.go` - Provider implementation
2. `provider/<name>/register.go` - Self-registration

### Step 1: Create Provider Directory

```bash
mkdir -p provider/myprovider
```

### Step 2: Implement the Provider

Create `provider/myprovider/provider.go`:

```go
package myprovider

import (
    "context"

    "github.com/kopexa-grc/kspec/core"
)

// Provider implements core.Provider for MyProvider.
type Provider struct {
    config map[string]string
}

// NewProvider creates a new MyProvider provider instance.
func NewProvider() *Provider {
    return &Provider{}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
    return "myprovider"
}

// Configure sets up the provider with the given configuration.
func (p *Provider) Configure(config map[string]string) error {
    p.config = config
    // Validate required configuration here
    return nil
}

// Resources returns the resource specifications this provider can fetch.
func (p *Provider) Resources() []core.ResourceSpec {
    return []core.ResourceSpec{
        // Add your resource implementations here
    }
}
```

### Step 3: Create Registration File

Create `provider/myprovider/register.go`:

```go
package myprovider

import (
    "github.com/kopexa-grc/kspec/core"
    "github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
    registry.Register(&registry.ProviderDefinition{
        Name:        "myprovider",
        Aliases:     []string{"mp"},  // Optional aliases
        Description: "Scan MyProvider resources for compliance",
        Flags: []registry.FlagDefinition{
            {
                Name:        "api-token",
                Description: "API token for authentication",
                EnvVar:      "MYPROVIDER_API_TOKEN",
                Required:    true,
            },
            {
                Name:        "region",
                Description: "Region to scan",
                Default:     "us-east-1",
            },
        },
        AssetTypes: []registry.AssetDefinition{
            {
                Name:        "account",
                Description: "Scan entire account",
                // No args = no positional arguments required
            },
        },
        Factory: func() core.Provider {
            return NewProvider()
        },
    })
}
```

### Step 4: Register in Import File

Add your provider to `provider/all/all.go`:

```go
import (
    // ... existing imports ...
    _ "github.com/kopexa-grc/kspec/provider/myprovider"
)
```

### Step 5: Done!

Your provider is now available:

```bash
kspec scan myprovider -d policies
kspec scan mp -d policies  # Using alias
```

## Registry Types Reference

### ProviderDefinition

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Primary provider name (used in CLI) |
| `Aliases` | `[]string` | Alternative names for the provider |
| `Description` | `string` | Help text shown in CLI |
| `Flags` | `[]FlagDefinition` | Provider-specific CLI flags |
| `AssetTypes` | `[]AssetDefinition` | Supported asset types |
| `ConfigMapping` | `map[string]string` | Maps flag names to config keys |
| `Factory` | `func() core.Provider` | Creates provider instance |

### FlagDefinition

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Flag name (e.g., "api-token" → `--api-token`) |
| `Short` | `string` | Single-char shorthand (e.g., "t" → `-t`) |
| `Description` | `string` | Help text for the flag |
| `Default` | `string` | Default value |
| `Required` | `bool` | Whether flag is required |
| `EnvVar` | `string` | Environment variable fallback |
| `IsBool` | `bool` | Whether flag is a boolean |

### AssetDefinition

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Asset type name (e.g., "org", "repo") |
| `Aliases` | `[]string` | Alternative names |
| `Description` | `string` | Help text |
| `Args` | `[]ArgDefinition` | Positional arguments |
| `DefaultName` | `string` | Default asset name if no args |
| `ScannerKey` | `string` | Override for scanner lookup key |

### ArgDefinition

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Argument name (shown in usage) |
| `Description` | `string` | Help text |
| `Required` | `bool` | Whether argument is required |
| `ConfigKey` | `string` | Key used in config map |

## Examples

### Single Asset Type (like AWS)

Providers with one asset type run directly on the provider command:

```go
AssetTypes: []registry.AssetDefinition{
    {
        Name:        "account",
        Description: "Scan AWS account",
    },
},
```

Usage: `kspec scan aws -d policies`

### Multiple Asset Types (like GitHub)

Providers with multiple asset types get subcommands:

```go
AssetTypes: []registry.AssetDefinition{
    {
        Name:        "org",
        Description: "Scan GitHub organization",
        Args: []registry.ArgDefinition{
            {Name: "owner", Required: true, ConfigKey: "owner"},
        },
    },
    {
        Name:        "repo",
        Description: "Scan GitHub repository",
        Args: []registry.ArgDefinition{
            {Name: "repository", Required: true, ConfigKey: "repository"},
        },
    },
},
```

Usage:
- `kspec scan github org my-org -d policies`
- `kspec scan github repo owner/repo -d policies`

### Custom Scanner Key (like SBOM)

When asset types need different scanner behavior:

```go
AssetTypes: []registry.AssetDefinition{
    {
        Name:        "file",
        Description: "Scan single SBOM file",
        ScannerKey:  "sbom-file",
        Args: []registry.ArgDefinition{
            {Name: "path", Required: true, ConfigKey: "sbom_path"},
        },
    },
    {
        Name:        "dir",
        Description: "Scan directory of SBOMs",
        ScannerKey:  "sbom-directory",
        Args: []registry.ArgDefinition{
            {Name: "path", Required: true, ConfigKey: "sbom_path"},
        },
    },
},
```

## Testing

Run tests for your provider:

```bash
go test ./provider/myprovider/...
```

Run linter:

```bash
golangci-lint run ./provider/myprovider/...
```

## Code Style

- Follow Go conventions and existing code style
- Add comments for exported types and functions
- Use meaningful variable names
- Handle errors appropriately

## Questions?

Open an issue at https://github.com/kopexa-grc/kspec/issues
