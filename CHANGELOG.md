# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1](https://github.com/kopexa-grc/kspec/compare/v0.1.0...v0.1.1) (2026-01-06)


### Features

* Add contribution guidelines, issue templates, and security policy ([9fc4454](https://github.com/kopexa-grc/kspec/commit/9fc4454c2f98dbaf48e0a74df247ada0bb4447da))
* Add GitHub Actions workflow for automated releases using Release Please ([ab6924d](https://github.com/kopexa-grc/kspec/commit/ab6924dec680de347d0ee71fa6c6c13fc8883160))
* Add GitHub organization and repository scanning with flexible credential handling. ([e8ce373](https://github.com/kopexa-grc/kspec/commit/e8ce37336f434722b4e6c016a04fe9f68f4ee936))
* Add Makefile, golangci-lint configuration, and Lefthook setup for improved development workflow ([6782d64](https://github.com/kopexa-grc/kspec/commit/6782d64101d22c8511d73ce1e4abd959e115bb9b))
* azure ([05ab3ed](https://github.com/kopexa-grc/kspec/commit/05ab3ed580bc196f7cd8a7b3daa8935822d6c5e1))
* cli ([dfcdf6d](https://github.com/kopexa-grc/kspec/commit/dfcdf6d211e2b9fd28516cd687ce023db45ea3c1))
* Create GoReleaser configuration for building and releasing kspec ([ab6924d](https://github.com/kopexa-grc/kspec/commit/ab6924dec680de347d0ee71fa6c6c13fc8883160))
* first version ([cd3e6f2](https://github.com/kopexa-grc/kspec/commit/cd3e6f264a82721421170f48ab87e16f675b943d))
* introduce CLI with scan command for policy evaluation and TUI results ([e7551ee](https://github.com/kopexa-grc/kspec/commit/e7551ee0bf6aca79a826c22950a5f650705315ce))
* **ms365:** Add Microsoft 365 provider with Teams and Tenant resources ([7845ba4](https://github.com/kopexa-grc/kspec/commit/7845ba4c9fa854269ecb494d87d6cb6dc5382375))
* remove `cnquery` command and `example_policy.yaml`, update `README.md` for `kspec` command, and add `kspec` to `.gitignore` ([2dece45](https://github.com/kopexa-grc/kspec/commit/2dece45efb5663b71bebb0bde1faaa3b07f0ffc6))


### Bug Fixes

* Update Azure provider imports to new repository path ([ab6924d](https://github.com/kopexa-grc/kspec/commit/ab6924dec680de347d0ee71fa6c6c13fc8883160))
* Update golangci-lint configuration and change changelog type to default ([e8b020b](https://github.com/kopexa-grc/kspec/commit/e8b020bd5b63c2ea2246d897444d61b1c0b3542a))
* Update import paths to reflect new repository structure ([ab6924d](https://github.com/kopexa-grc/kspec/commit/ab6924dec680de347d0ee71fa6c6c13fc8883160))
* Update README.md with new repository link and author information ([ab6924d](https://github.com/kopexa-grc/kspec/commit/ab6924dec680de347d0ee71fa6c6c13fc8883160))


### Documentation

* Create CHANGELOG.md to document notable changes ([ab6924d](https://github.com/kopexa-grc/kspec/commit/ab6924dec680de347d0ee71fa6c6c13fc8883160))
* Update project banner image. ([1fb0fe4](https://github.com/kopexa-grc/kspec/commit/1fb0fe4bdcb421de35ac061e5d80271264509657))


### Code Refactoring

* Clean up comments and improve clarity in policy files; add git log permission ([4febaa6](https://github.com/kopexa-grc/kspec/commit/4febaa6b8e52a12909dd02a077a43fdabde4a9e0))
* Update all provider imports to point to the new repository path ([ab6924d](https://github.com/kopexa-grc/kspec/commit/ab6924dec680de347d0ee71fa6c6c13fc8883160))

## [Unreleased]

### Features

- Initial release of kspec policy-as-code engine
- Azure provider with support for storage accounts, SQL servers, Key Vaults, NSGs, VMs, and App Services
- Microsoft 365 provider with support for users, groups, applications, Teams, security policies, and more
- GitHub provider with support for organizations, repositories, branches, and teams
- Network provider for TLS, DNS, and HTTP security scanning
- CEL-based policy evaluation engine
- Interactive TUI for real-time scan progress
- Policy-as-code YAML format with comprehensive documentation

### Documentation

- Azure provider setup guide
- Microsoft 365 provider setup guide
- GitHub provider setup guide
- Example security policies for all providers
