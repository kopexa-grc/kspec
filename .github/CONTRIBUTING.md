# Contributing to kspec

Thank you for your interest in contributing to kspec! This document provides guidelines and information for contributors.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/kopexa-grc/kspec/issues)
2. If not, create a new issue using the **Bug Report** template
3. Provide as much detail as possible, including:
   - kspec version (`kspec --version`)
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior

### Suggesting Features

1. Check [Issues](https://github.com/kopexa-grc/kspec/issues) and [Discussions](https://github.com/kopexa-grc/kspec/discussions) for existing proposals
2. Create a new issue using the **Feature Request** template
3. Clearly describe the problem and proposed solution

### Requesting New Policies

1. Use the **Policy Request** template
2. Include the security rationale and compliance framework references
3. If possible, provide a draft CEL query

### Submitting Changes

1. Fork the repository
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
3. Make your changes following our coding standards
4. Write or update tests as needed
5. Commit using [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat(provider): add new resource type"
   git commit -m "fix(scanner): resolve nil pointer error"
   git commit -m "docs: update Azure setup guide"
   ```
6. Push to your fork and create a Pull Request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Building

```bash
# Clone the repository
git clone https://github.com/kopexa-grc/kspec.git
cd kspec

# Install dependencies
go mod download

# Build
go build -o kspec ./cmd/kspec

# Run tests
go test -v ./...
```

### Running Locally

```bash
# Build with version info
go build -ldflags="-X main.version=dev -X main.commit=$(git rev-parse --short HEAD)" -o kspec ./cmd/kspec

# Run a scan
./kspec scan github org <org-name> -f policies/github-security.yml
```

## Coding Standards

### Go Code

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Handle errors explicitly
- Write unit tests for new functionality

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation changes |
| `refactor` | Code refactoring |
| `test` | Adding or updating tests |
| `perf` | Performance improvements |
| `ci` | CI/CD changes |
| `deps` | Dependency updates |
| `chore` | Other changes |

**Scopes** (optional):
- `provider` - Provider-related changes
- `scanner` - Scanner engine changes
- `cli` - CLI changes
- `tui` - Terminal UI changes
- `policy` - Policy-related changes

**Examples:**
```
feat(provider): add AWS provider support
fix(scanner): handle nil resource gracefully
docs: add MS365 setup guide
refactor(cli): simplify scan command
```

### Policy Files

- Use descriptive UIDs: `kspec-<provider>-<category>-<check>`
- Include comprehensive `docs.desc` and `docs.remediation`
- Reference compliance frameworks where applicable
- Test policies against real resources when possible

## Adding a New Provider

1. Create a new package under `provider/<name>/`
2. Implement the `core.Provider` interface
3. Implement `core.ResourceSpec` for each resource type
4. Implement `core.DiscoveryResource` for auto-discovery
5. Register the provider in `provider/registry.go`
6. Update the CLI in `cmd/kspec/cmd/scan.go`
7. Add documentation in `docs/<name>-setup.md`
8. Create example policies in `policies/<name>-security.yml`

## Adding a New Resource Type

1. Create a new file in the provider package
2. Implement `core.ResourceSpec` interface:
   - `Name()` - Return resource type name
   - `Fetch()` - Fetch and return resources
3. Add to the provider's `Resources()` method
4. Update discovery if applicable
5. Add example policies

## Testing

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Run specific package tests
go test -v ./provider/azure/...

# Run with race detection
go test -v -race ./...
```

## Documentation

- Update relevant docs when changing functionality
- Use clear, concise language
- Include code examples where helpful
- Keep setup guides up to date

## Questions?

- Open a [Discussion](https://github.com/kopexa-grc/kspec/discussions)
- Check existing [Issues](https://github.com/kopexa-grc/kspec/issues)

Thank you for contributing!
