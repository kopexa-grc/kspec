// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

// Status represents the outcome of a policy check evaluation.
// It is a typed string for type safety and better API ergonomics.
type Status string

// Check status constants.
const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// Result represents the result of evaluating a single policy check.
// It contains all information needed to understand what was checked,
// whether it passed, and how to remediate if it failed.
type Result struct {
	// ID is the unique identifier of the check (from check.uid or check.id).
	ID string `json:"id"`

	// Group is the title of the group this check belongs to.
	Group string `json:"group"`

	// Name is the human-readable title of the check.
	Name string `json:"name"`

	// Status indicates the outcome: StatusPassed, StatusFailed, or StatusSkipped.
	Status Status `json:"status"`

	// Details contains additional information about the check result.
	// For failed checks, this may contain error messages or context.
	Details string `json:"details,omitempty"`

	// Severity indicates the impact level: "critical", "high", "medium", "low", or "info".
	// Defaults to "medium" if not specified in the policy.
	Severity string `json:"severity"`

	// Remediation contains markdown-formatted guidance on how to fix a failed check.
	Remediation string `json:"remediation,omitempty"`

	// Docs contains markdown-formatted documentation explaining the check.
	Docs string `json:"docs,omitempty"`

	// Audit contains markdown-formatted manual audit instructions.
	Audit string `json:"audit,omitempty"`
}

// IsPassed returns true if the check passed.
func (r Result) IsPassed() bool {
	return r.Status == StatusPassed
}

// IsFailed returns true if the check failed.
func (r Result) IsFailed() bool {
	return r.Status == StatusFailed
}

// IsSkipped returns true if the check was skipped.
func (r Result) IsSkipped() bool {
	return r.Status == StatusSkipped
}

// EvaluateResult is an alias for Result for backwards compatibility.
// Deprecated: Use Result instead.
type EvaluateResult = Result
