// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

//go:build fast

// Package suites provides policy integration test fixtures for GitHub resources.
package suites

import "github.com/kopexa-grc/kspec/testing/suites"

// GitHub Organization Fixtures

// SecureOrganization represents a compliant GitHub organization.
var SecureOrganization = suites.NewFixture().
	Set("id", "org-secure-001").
	Set("login", "secure-org").
	Set("name", "Secure Organization").
	Set("two_factor_requirement_enabled", true).
	Set("is_verified", true).
	Set("default_repository_permission", "read")

// InsecureOrganization represents a non-compliant GitHub organization.
var InsecureOrganization = suites.NewFixture().
	Set("id", "org-insecure-001").
	Set("login", "insecure-org").
	Set("name", "Insecure Organization").
	Set("two_factor_requirement_enabled", false).
	Set("is_verified", false).
	Set("default_repository_permission", "write")

// GitHub Branch Fixtures

// SecureDefaultBranch represents a protected default branch with all security features.
var SecureDefaultBranch = suites.NewFixture().
	Set("name", "main").
	Set("is_default", true).
	Set("protected", true).
	Set("allow_force_pushes", map[string]interface{}{"enabled": false}).
	Set("required_conversation_resolution", map[string]interface{}{"enabled": true}).
	Set("required_status_checks", []string{"ci/build", "ci/test"}).
	Set("required_signatures", true).
	Set("enforce_admins", map[string]interface{}{"enabled": true})

// InsecureDefaultBranch represents an unprotected default branch.
var InsecureDefaultBranch = suites.NewFixture().
	Set("name", "main").
	Set("is_default", true).
	Set("protected", false)

// PartiallyProtectedDefaultBranch represents a protected branch with some security gaps.
var PartiallyProtectedDefaultBranch = suites.NewFixture().
	Set("name", "main").
	Set("is_default", true).
	Set("protected", true).
	Set("allow_force_pushes", map[string]interface{}{"enabled": true}).
	Set("required_conversation_resolution", map[string]interface{}{"enabled": false}).
	Set("required_status_checks", []string{}).
	Set("required_signatures", false).
	Set("enforce_admins", map[string]interface{}{"enabled": false})

// SecureReleaseBranch represents a protected release branch.
var SecureReleaseBranch = suites.NewFixture().
	Set("name", "release/v1.0").
	Set("is_default", false).
	Set("protected", true).
	Set("allow_force_pushes", map[string]interface{}{"enabled": false})

// InsecureReleaseBranch represents an unprotected release branch.
var InsecureReleaseBranch = suites.NewFixture().
	Set("name", "release/v2.0").
	Set("is_default", false).
	Set("protected", false)

// NonDefaultNonReleaseBranch represents a feature branch (should pass all checks).
var NonDefaultNonReleaseBranch = suites.NewFixture().
	Set("name", "feature/new-feature").
	Set("is_default", false).
	Set("protected", false)

// GitHub Repository Fixtures

// RepoWithDependabot represents a repository with Dependabot configured.
var RepoWithDependabot = suites.NewFixture().
	Set("id", "repo-001").
	Set("name", "secure-repo").
	Set("full_name", "org/secure-repo").
	Set("private", true).
	Set("files", []map[string]interface{}{
		{"path": ".github/dependabot.yml", "content": "version: 2"},
		{"path": "README.md", "content": "# Project"},
	})

// RepoWithoutDependabot represents a repository without Dependabot configured.
var RepoWithoutDependabot = suites.NewFixture().
	Set("id", "repo-002").
	Set("name", "no-dependabot-repo").
	Set("full_name", "org/no-dependabot-repo").
	Set("private", true).
	Set("files", []map[string]interface{}{
		{"path": "README.md", "content": "# Project"},
	})
