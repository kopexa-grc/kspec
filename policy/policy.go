// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package policy provides types and utilities for defining and loading security policies.
package policy

// Policy represents a collection of checks and groups.
type Policy struct {
	APIVersion    string        `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty" jsonschema:"description=API version for the policy format"`
	Kind          string        `yaml:"kind,omitempty" json:"kind,omitempty" jsonschema:"enum=Policy,description=Must be 'Policy'"`
	Metadata      Metadata      `yaml:"metadata,omitempty" json:"metadata,omitempty" jsonschema:"description=Policy metadata"`
	Imports       []string      `yaml:"imports,omitempty" json:"imports,omitempty" jsonschema:"description=Import other policies by path, URL, or glob pattern"`
	Groups        []Group       `yaml:"groups,omitempty" json:"groups,omitempty" jsonschema:"description=Groups of related checks"`
	Queries       []Check       `yaml:"queries,omitempty" json:"queries,omitempty" jsonschema:"description=Check definitions that can be referenced by groups"`
	Require       []Requirement `yaml:"require,omitempty" json:"require,omitempty" jsonschema:"description=Provider dependencies required by this policy"`
	Authors       []Author      `yaml:"authors,omitempty" json:"authors,omitempty" jsonschema:"description=Policy authors"`
	ScoringSystem string        `yaml:"scoring_system,omitempty" json:"scoring_system,omitempty" jsonschema:"description=Scoring system to use for this policy"`
}

// Metadata contains policy metadata like name, version, and tags.
type Metadata struct {
	Name    string            `yaml:"name" json:"name" jsonschema:"required,description=Unique identifier for the policy"`
	Version string            `yaml:"version,omitempty" json:"version,omitempty" jsonschema:"description=Semantic version of the policy"`
	Title   string            `yaml:"title,omitempty" json:"title,omitempty" jsonschema:"description=Human-readable title"`
	License string            `yaml:"license,omitempty" json:"license,omitempty" jsonschema:"description=License identifier (e.g. MIT or Apache-2.0)"`
	Tags    map[string]string `yaml:"tags,omitempty" json:"tags,omitempty" jsonschema:"description=Key-value tags for categorization"`
}

// Requirement specifies a provider dependency for the policy.
type Requirement struct {
	Provider string `yaml:"provider" json:"provider" jsonschema:"required,description=Provider name (e.g. aws | azure | github)"`
}

// Author represents a policy author.
type Author struct {
	Name  string `yaml:"name" json:"name" jsonschema:"required,description=Author name"`
	Email string `yaml:"email,omitempty" json:"email,omitempty" jsonschema:"format=email,description=Author email address"`
}

// Group represents a collection of related checks with optional filtering.
type Group struct {
	Title  string  `yaml:"title" json:"title" jsonschema:"required,description=Group title"`
	Filter string  `yaml:"filter,omitempty" json:"filter,omitempty" jsonschema:"description=CEL expression to filter assets for this group"`
	Checks []Check `yaml:"checks" json:"checks" jsonschema:"description=Checks in this group (can reference queries by uid)"`
}

// Check represents a security check to evaluate against resources.
type Check struct {
	UID      string            `yaml:"uid,omitempty" json:"uid,omitempty" jsonschema:"description=Unique identifier for this check"`
	ID       string            `yaml:"id,omitempty" json:"id,omitempty" jsonschema:"description=Alternative identifier"`
	Title    string            `yaml:"title,omitempty" json:"title,omitempty" jsonschema:"description=Human-readable check title"`
	Resource string            `yaml:"resource,omitempty" json:"resource,omitempty" jsonschema:"description=Resource type to query (e.g. tls | dns | github_repo)"`
	Config   map[string]string `yaml:"config,omitempty" json:"config,omitempty" jsonschema:"description=Resource-specific configuration"`

	Query       string `yaml:"query,omitempty" json:"query,omitempty" jsonschema:"description=CEL expression that must evaluate to true for the check to pass"`
	Remediation string `yaml:"remediation,omitempty" json:"remediation,omitempty" jsonschema:"description=Remediation guidance when check fails"`
	Severity    string `yaml:"severity,omitempty" json:"severity,omitempty" jsonschema:"enum=critical,enum=high,enum=medium,enum=low,enum=info,description=Check severity level"`

	Docs  string `yaml:"docs,omitempty" json:"docs,omitempty" jsonschema:"description=Documentation or explanation"`
	Audit string `yaml:"audit,omitempty" json:"audit,omitempty" jsonschema:"description=Manual audit instructions"`
	Props []Prop `yaml:"props,omitempty" json:"props,omitempty" jsonschema:"description=Properties to extract from resources"`
}

// Prop represents a named property extraction for check results.
type Prop struct {
	UID   string `yaml:"uid,omitempty" json:"uid,omitempty" jsonschema:"description=Property identifier"`
	Title string `yaml:"title,omitempty" json:"title,omitempty" jsonschema:"description=Property title"`
	MQL   any    `yaml:"mql,omitempty" json:"mql,omitempty" jsonschema:"description=MQL expression for property extraction"`
}

// Severity constants.
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)
