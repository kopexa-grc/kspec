# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1](https://github.com/kopexa-grc/kspec/compare/v0.2.0...v0.2.1) (2026-01-19)


### Features

* add discovery command and graph-based resource traversal ([#51](https://github.com/kopexa-grc/kspec/issues/51)) ([c25768a](https://github.com/kopexa-grc/kspec/commit/c25768a592f451a61edca17f1091f55ece4b5663))


### Documentation

* **readme:** update commands to use asset type subcommands ([b20d9e2](https://github.com/kopexa-grc/kspec/commit/b20d9e2602a32d5024472e1624b834a63d7b91be))


### Dependencies

* **deps:** bump github.com/aws/aws-sdk-go-v2/service/autoscaling ([#56](https://github.com/kopexa-grc/kspec/issues/56)) ([2afba51](https://github.com/kopexa-grc/kspec/commit/2afba5176bb91889276ccf48e0f5dd064c8713e2))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudtrail ([#58](https://github.com/kopexa-grc/kspec/issues/58)) ([5787049](https://github.com/kopexa-grc/kspec/commit/5787049b98b0a75a43c5224e1257769184b699c6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ec2 ([#59](https://github.com/kopexa-grc/kspec/issues/59)) ([517a3ca](https://github.com/kopexa-grc/kspec/commit/517a3cade01baa649c217bb22184083e57b3ba02))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecs ([#60](https://github.com/kopexa-grc/kspec/issues/60)) ([97e3bf1](https://github.com/kopexa-grc/kspec/commit/97e3bf1ff7eee1dad7a7c66c8beeaeee841dd497))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/secretsmanager ([#55](https://github.com/kopexa-grc/kspec/issues/55)) ([f539f5c](https://github.com/kopexa-grc/kspec/commit/f539f5c3d163171d2d61d860f1ce0e84daac1f84))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sns ([#61](https://github.com/kopexa-grc/kspec/issues/61)) ([5b305be](https://github.com/kopexa-grc/kspec/commit/5b305be058c55bcdb5ff68566e8bc7d7d696797f))
* **deps:** bump github.com/Azure/azure-sdk-for-go/sdk/azcore ([#54](https://github.com/kopexa-grc/kspec/issues/54)) ([c72851d](https://github.com/kopexa-grc/kspec/commit/c72851d745bcba09025ed75ddc2b6f1c681f04d8))
* **deps:** bump github.com/hetznercloud/hcloud-go/v2 ([#57](https://github.com/kopexa-grc/kspec/issues/57)) ([a4d07b0](https://github.com/kopexa-grc/kspec/commit/a4d07b01c6602a6441358b98b7d4ab2e97cb8926))

## [0.2.0](https://github.com/kopexa-grc/kspec/compare/v0.1.6...v0.2.0) (2026-01-18)


### ⚠ BREAKING CHANGES

* **provider:** Default policy directory

### Features

* concurrent scanning, provider refactoring, and expanded AWS security policies ([#50](https://github.com/kopexa-grc/kspec/issues/50)) ([8907e4a](https://github.com/kopexa-grc/kspec/commit/8907e4a39f36444b90646af2b1f73c77e11eab15))
* **provider:** implement dynamic self-registration pattern ([7ad2807](https://github.com/kopexa-grc/kspec/commit/7ad2807ce592f770ed35c5bcd8d8be262163dde7)), closes [#47](https://github.com/kopexa-grc/kspec/issues/47)
* **report:** add HTML export format ([#48](https://github.com/kopexa-grc/kspec/issues/48)) ([d1cea21](https://github.com/kopexa-grc/kspec/commit/d1cea21422a462d90c0db364b867acdd45fbcbf4))
* **report:** add HTML export format with interactive features ([d1cea21](https://github.com/kopexa-grc/kspec/commit/d1cea21422a462d90c0db364b867acdd45fbcbf4))
* **report:** add report export and non-interactive scan mode ([19aa52d](https://github.com/kopexa-grc/kspec/commit/19aa52d71ffe6825fb9e67c99ee7103075c71770))


### Dependencies

* **deps:** bump github.com/aws/aws-sdk-go-v2/config ([#30](https://github.com/kopexa-grc/kspec/issues/30)) ([af4856b](https://github.com/kopexa-grc/kspec/commit/af4856b9eb1fc428f605d1c7cb8c1a3567bee742))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/acm ([#37](https://github.com/kopexa-grc/kspec/issues/37)) ([8c7a397](https://github.com/kopexa-grc/kspec/commit/8c7a3976fabf1959c20585eace85b2b8b313e78c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigateway ([#27](https://github.com/kopexa-grc/kspec/issues/27)) ([22f10e0](https://github.com/kopexa-grc/kspec/commit/22f10e0a85c8438d715ada08049872d5083c480a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/apigatewayv2 ([#22](https://github.com/kopexa-grc/kspec/issues/22)) ([a2694b5](https://github.com/kopexa-grc/kspec/commit/a2694b5310fa48abf5198c37d1ca6da8516e9c70))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudfront ([#38](https://github.com/kopexa-grc/kspec/issues/38)) ([8fd18af](https://github.com/kopexa-grc/kspec/commit/8fd18af5b4bba81f53c1df181554ab83ac0e3147))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatch ([#39](https://github.com/kopexa-grc/kspec/issues/39)) ([24b0fd8](https://github.com/kopexa-grc/kspec/commit/24b0fd86a2e20bc484bb875b1c942085dde3cc4c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs ([#19](https://github.com/kopexa-grc/kspec/issues/19)) ([7f89f5f](https://github.com/kopexa-grc/kspec/commit/7f89f5fb38b9eec918c412450a7152d39d5cb24a))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/configservice ([#44](https://github.com/kopexa-grc/kspec/issues/44)) ([bba331c](https://github.com/kopexa-grc/kspec/commit/bba331ca74c09ef78ba8e1d0763baab90c2b2607))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/dynamodb ([#45](https://github.com/kopexa-grc/kspec/issues/45)) ([40bdf9b](https://github.com/kopexa-grc/kspec/commit/40bdf9bb09ea3580e915e1560b692ef8ce5b007d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ecr ([#23](https://github.com/kopexa-grc/kspec/issues/23)) ([9c38402](https://github.com/kopexa-grc/kspec/commit/9c384022924f1e59887075edd8320667b58423a6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/eks ([#34](https://github.com/kopexa-grc/kspec/issues/34)) ([be480fa](https://github.com/kopexa-grc/kspec/commit/be480fafda7a591474bd9db7851107ec184c37d3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticache ([#20](https://github.com/kopexa-grc/kspec/issues/20)) ([449bd22](https://github.com/kopexa-grc/kspec/commit/449bd228dafb5643383c5eec63d3cff6f9b4ff5d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 ([#41](https://github.com/kopexa-grc/kspec/issues/41)) ([3a30a91](https://github.com/kopexa-grc/kspec/commit/3a30a916d91fd082414cf390e1e5f0d1445177d3))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/guardduty ([#33](https://github.com/kopexa-grc/kspec/issues/33)) ([8a1cb14](https://github.com/kopexa-grc/kspec/commit/8a1cb1412abc3e83bbea17d46892f0dfaef072ea))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/iam ([#43](https://github.com/kopexa-grc/kspec/issues/43)) ([ad6bd26](https://github.com/kopexa-grc/kspec/commit/ad6bd263e65a9bc106feb50580714f9d17fef96f))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/kms ([#42](https://github.com/kopexa-grc/kspec/issues/42)) ([c76998f](https://github.com/kopexa-grc/kspec/commit/c76998ffd53937601297036a78235b36719f0cb6))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/lambda ([#21](https://github.com/kopexa-grc/kspec/issues/21)) ([ae004e6](https://github.com/kopexa-grc/kspec/commit/ae004e6a026803c436dca7dba35de20a20199aa1))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/organizations ([#32](https://github.com/kopexa-grc/kspec/issues/32)) ([01b6e7b](https://github.com/kopexa-grc/kspec/commit/01b6e7b46b8cddc65c316d22ab19520558d1bd3c))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/rds ([#40](https://github.com/kopexa-grc/kspec/issues/40)) ([5d4c56e](https://github.com/kopexa-grc/kspec/commit/5d4c56e46d2a911d5032553c3325454ae627f896))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/s3 ([#26](https://github.com/kopexa-grc/kspec/issues/26)) ([ee03308](https://github.com/kopexa-grc/kspec/commit/ee033084d23ce4cf00e65dda9ff0132bc00ffc10))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/securityhub ([#36](https://github.com/kopexa-grc/kspec/issues/36)) ([896822a](https://github.com/kopexa-grc/kspec/commit/896822a10724bfe542d747e685e5241d792e4802))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/sqs ([#31](https://github.com/kopexa-grc/kspec/issues/31)) ([dbfac6d](https://github.com/kopexa-grc/kspec/commit/dbfac6d2550fd11cadce8380e1bcff0f8d5fd71b))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/ssm ([#24](https://github.com/kopexa-grc/kspec/issues/24)) ([9c9a446](https://github.com/kopexa-grc/kspec/commit/9c9a446affbe527db995d320dcc73a95375b880d))
* **deps:** bump github.com/aws/aws-sdk-go-v2/service/wafv2 ([#25](https://github.com/kopexa-grc/kspec/issues/25)) ([406f91b](https://github.com/kopexa-grc/kspec/commit/406f91bbfc5d0c17d4c96be6dd4124f92dbd82b6))
* **deps:** bump github.com/hetznercloud/hcloud-go/v2 ([#28](https://github.com/kopexa-grc/kspec/issues/28)) ([35d2eb0](https://github.com/kopexa-grc/kspec/commit/35d2eb009eaeca538155dadf98be7f7105c53164))
* **deps:** bump github.com/microsoftgraph/msgraph-sdk-go ([#14](https://github.com/kopexa-grc/kspec/issues/14)) ([088ba75](https://github.com/kopexa-grc/kspec/commit/088ba75a0be7d981132237791604019f0706edb8))
* **deps:** bump github.com/miekg/dns from 1.1.69 to 1.1.70 ([#29](https://github.com/kopexa-grc/kspec/issues/29)) ([0e435f4](https://github.com/kopexa-grc/kspec/commit/0e435f436fc7b26fc7d1dc299e3168fabab79f37))

## [0.1.6](https://github.com/kopexa-grc/kspec/compare/v0.1.5...v0.1.6) (2026-01-08)


### Features

* add ptr package and optimize AWS provider with fuzz tests ([7ebe687](https://github.com/kopexa-grc/kspec/commit/7ebe6879196be67c0d5492815d714fc1834a8b42))
* **azure:** add comprehensive Azure provider resources ([c8a3216](https://github.com/kopexa-grc/kspec/commit/c8a32162cd8927a7eb8c61b237533678af1d77e9))
* **schema:** add JSON Schema generation for policy validation ([d0f59b9](https://github.com/kopexa-grc/kspec/commit/d0f59b9278ca74de0cda020c542f34cf67619bc8))


### Bug Fixes

* preallocate slices to satisfy prealloc linter ([a90f8af](https://github.com/kopexa-grc/kspec/commit/a90f8afbb4ffa60e8b9b269f76dc573c7e1072c0))
* resolve remaining prealloc linter issues for golangci-lint v2.8.0 ([3b2efd9](https://github.com/kopexa-grc/kspec/commit/3b2efd9d82dcb23354d18ecd2f283decf933532d))
* update license to ELv2 in policy files ([41b7593](https://github.com/kopexa-grc/kspec/commit/41b75934bf83d85be088d7a6a47a04f77eb4119e))


### Code Refactoring

* add core helpers and improve error handling ([ea357ff](https://github.com/kopexa-grc/kspec/commit/ea357ffc014911e68628272f98c22945653ea12f))

## [0.1.5](https://github.com/kopexa-grc/kspec/compare/v0.1.4...v0.1.5) (2026-01-07)


### Features

* **aws:** add comprehensive AWS provider for security scanning ([51e4383](https://github.com/kopexa-grc/kspec/commit/51e4383a9b87eeda52863c3340d09e14c1d3a07b))


### Bug Fixes

* resolve linter issues and add provider documentation ([ae67317](https://github.com/kopexa-grc/kspec/commit/ae67317f36f08a185d6ecff52b3714181f9dc6ee))

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
