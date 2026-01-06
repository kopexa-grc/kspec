# Installation

This guide covers different ways to install kspec on your system.

## From Source

### Prerequisites

- Go 1.21 or later
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/kopexa-grc/kspec.git
cd kspec

# Build the binary
go build -o kspec ./cmd/kspec

# Verify installation
./kspec --version

# Optional: Move to PATH
sudo mv kspec /usr/local/bin/
```

### Install with go install

```bash
go install github.com/kopexa-grc/kspec/cmd/kspec@latest
```

## From Releases

Download pre-built binaries from the [GitHub Releases](https://github.com/kopexa-grc/kspec/releases) page.

### Linux (amd64)

```bash
curl -LO https://github.com/kopexa-grc/kspec/releases/latest/download/kspec_Linux_x86_64.tar.gz
tar -xzf kspec_Linux_x86_64.tar.gz
sudo mv kspec /usr/local/bin/
```

### Linux (arm64)

```bash
curl -LO https://github.com/kopexa-grc/kspec/releases/latest/download/kspec_Linux_arm64.tar.gz
tar -xzf kspec_Linux_arm64.tar.gz
sudo mv kspec /usr/local/bin/
```

### macOS (Apple Silicon)

```bash
curl -LO https://github.com/kopexa-grc/kspec/releases/latest/download/kspec_Darwin_arm64.tar.gz
tar -xzf kspec_Darwin_arm64.tar.gz
sudo mv kspec /usr/local/bin/
```

### macOS (Intel)

```bash
curl -LO https://github.com/kopexa-grc/kspec/releases/latest/download/kspec_Darwin_x86_64.tar.gz
tar -xzf kspec_Darwin_x86_64.tar.gz
sudo mv kspec /usr/local/bin/
```

### Windows

Download `kspec_Windows_x86_64.zip` from the releases page and extract to a directory in your PATH.

## Verify Installation

```bash
kspec --version
kspec --help
```

## Verify Release Signatures

All releases are signed using [Sigstore Cosign](https://www.sigstore.dev/). To verify:

```bash
# Install cosign
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# Download signature files
curl -LO https://github.com/kopexa-grc/kspec/releases/latest/download/checksums.txt
curl -LO https://github.com/kopexa-grc/kspec/releases/latest/download/checksums.txt.sig
curl -LO https://github.com/kopexa-grc/kspec/releases/latest/download/checksums.txt.pem

# Verify signature
cosign verify-blob \
  --signature checksums.txt.sig \
  --certificate checksums.txt.pem \
  --certificate-identity-regexp "https://github.com/kopexa-grc/kspec" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  checksums.txt

# Verify checksum
sha256sum -c checksums.txt --ignore-missing
```

## SLSA Provenance

Releases include [SLSA Level 3](https://slsa.dev/) provenance attestations. Download the `.intoto.jsonl` file from releases to verify build provenance.

## Development Setup

For contributing to kspec:

```bash
# Clone repository
git clone https://github.com/kopexa-grc/kspec.git
cd kspec

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o kspec ./cmd/kspec

# Run linter
golangci-lint run
```

## Docker (Coming Soon)

Docker images will be available at `ghcr.io/kopexa-grc/kspec`.

## Next Steps

- [Quickstart Guide](QUICKSTART.md) - Run your first scan
- [Writing Policies](policies.md) - Create custom security policies
- [CLI Reference](reference/cli.md) - Full command reference
