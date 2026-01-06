# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.4](https://github.com/kopexa-grc/kspec/compare/v0.1.3...v0.1.4) (2026-01-06)


### Features

* add enterprise security features ([e53fc94](https://github.com/kopexa-grc/kspec/commit/e53fc949d50a1fed53810b178423e8e004f19992))
* add signed releases and SLSA provenance ([cd290a9](https://github.com/kopexa-grc/kspec/commit/cd290a9714385b6f503ebcfe6a652d4762079306))
* **factorial:** add Factorial HR provider for compliance scanning ([31972b9](https://github.com/kopexa-grc/kspec/commit/31972b99edf3f5c210ea039cec8ec80b2f80287e))


### Bug Fixes

* resolve linter issues and add test coverage ([4b986db](https://github.com/kopexa-grc/kspec/commit/4b986db0cf5ac07c0023dc379ce3f1221b9c774e))
* standardize policy YAML files to match Go struct ([b36b00e](https://github.com/kopexa-grc/kspec/commit/b36b00e546542ea1b858629770fac5d6ffc2cff1))


### Documentation

* add comprehensive provider documentation ([9ed122f](https://github.com/kopexa-grc/kspec/commit/9ed122fc3b356c0bbce1e66111a5494e898f3817))
* reorganize and complete documentation structure ([1287a4e](https://github.com/kopexa-grc/kspec/commit/1287a4e990dbd4fea8482e4b7fed4020591140de))

## [0.1.3](https://github.com/kopexa-grc/kspec/compare/v0.1.2...v0.1.3) (2026-01-06)


### Bug Fixes

* **ci:** fix archive upload and checksum generation ([1c252f1](https://github.com/kopexa-grc/kspec/commit/1c252f159f2cfb36332063d7989ef12510dd789b))

## [0.1.2](https://github.com/kopexa-grc/kspec/compare/v0.1.1...v0.1.2) (2026-01-06)


### Bug Fixes

* use default font in demo GIF to fix letter spacing ([27dd58d](https://github.com/kopexa-grc/kspec/commit/27dd58dd193217ab250dd17f0ff286a162f70131))


### Documentation

* add demo GIF showcasing host scanning ([016cd41](https://github.com/kopexa-grc/kspec/commit/016cd41ebcdfcc3676528f3f8d57fce308ff27d9))

## [0.1.1](https://github.com/kopexa-grc/kspec/compare/v0.1.0...v0.1.1) (2026-01-06)


### Features

* Add contribution guidelines, issue templates, and security policy ([4366e66](https://github.com/kopexa-grc/kspec/commit/4366e660e872934f2af8203b4a52ee96ee8b53e5))
* Add GitHub Actions workflow for automated releases using Release Please ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Add GitHub organization and repository scanning with flexible credential handling. ([07359af](https://github.com/kopexa-grc/kspec/commit/07359af686e2d0ecda1504fea665b4810728d6fc))
* Add Makefile, golangci-lint configuration, and Lefthook setup for improved development workflow ([6fa8896](https://github.com/kopexa-grc/kspec/commit/6fa8896e4b06b4a9405ac456792acbd83698b494))
* Add SBOM component and vulnerability resources with tests ([1dfecfb](https://github.com/kopexa-grc/kspec/commit/1dfecfb2e337ac0761610b0d234456cab02d76c2))
* **atlassian:** add Jira security scheme and user resources ([79ed2ac](https://github.com/kopexa-grc/kspec/commit/79ed2acf2e079569730e793acedcde8628e4d02f))
* azure ([c8dcd76](https://github.com/kopexa-grc/kspec/commit/c8dcd762f21fe671194c487072f815a7f769a50c))
* cli ([42d5d63](https://github.com/kopexa-grc/kspec/commit/42d5d63c09d18a6959489c4c45b9ac1e0eaf3a89))
* **cloudflare:** add support for various Cloudflare resources ([7518757](https://github.com/kopexa-grc/kspec/commit/75187578f795a055435a26eea76dd3c5a1adca28))
* Create GoReleaser configuration for building and releasing kspec ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* enhanced TUI, certificate scanning, and quickstart docs ([acf87f1](https://github.com/kopexa-grc/kspec/commit/acf87f1ad679629216ba096cfbc562f6092b93eb))
* first version ([2da72ea](https://github.com/kopexa-grc/kspec/commit/2da72eac7c700be9fd07a4ea361924dcf51b49c7))
* **hetzner:** add Hetzner Cloud provider for infrastructure security scanning ([b142ceb](https://github.com/kopexa-grc/kspec/commit/b142cebb2f510faf0d24e1c151b3e2b5f304cc8b))
* introduce CLI with scan command for policy evaluation and TUI results ([4b25063](https://github.com/kopexa-grc/kspec/commit/4b25063cf717f1b0779e3e63f18728fe00785ae3))
* **ms365:** Add Microsoft 365 provider with Teams and Tenant resources ([4fb91b2](https://github.com/kopexa-grc/kspec/commit/4fb91b298519873a88bf84b84e20cda832e56bab))
* remove `cnquery` command and `example_policy.yaml`, update `README.md` for `kspec` command, and add `kspec` to `.gitignore` ([8214b0e](https://github.com/kopexa-grc/kspec/commit/8214b0ea8bcf2fdf9a0f261177a74c9a4bf03284))


### Bug Fixes

* address linting issues across the codebase ([b28da0c](https://github.com/kopexa-grc/kspec/commit/b28da0c90e055a5e06af0e792f61bc9cb0d00ef3))
* resolve all remaining linter issues ([31ad096](https://github.com/kopexa-grc/kspec/commit/31ad0965aa0be99a4ac6c7690c279023c07b5b57))
* resolve revive linter stuttering and package-comments issues ([c88351b](https://github.com/kopexa-grc/kspec/commit/c88351bcc2d87d5315d8698545ee7066cf11040a))
* Update Azure provider imports to new repository path ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Update golangci-lint configuration and change changelog type to default ([8754798](https://github.com/kopexa-grc/kspec/commit/8754798a47b3e8fd54aea63f5215ca49c0aed500))
* Update import paths to reflect new repository structure ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Update README.md with new repository link and author information ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))


### Documentation

* Create CHANGELOG.md to document notable changes ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))
* Update project banner image. ([bec4729](https://github.com/kopexa-grc/kspec/commit/bec4729b9fc61fc094c868570887c7ce8182c7a4))


### Code Refactoring

* Clean up comments and improve clarity in policy files; add git log permission ([e218ea9](https://github.com/kopexa-grc/kspec/commit/e218ea9662d752596e05105154948b82518e5c03))
* Update all provider imports to point to the new repository path ([cb7de72](https://github.com/kopexa-grc/kspec/commit/cb7de729b3eb99f1915bf872dd750f01cc3388a9))

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
