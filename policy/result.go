// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

// Check status constants.
const (
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Result represents the result of evaluating a single policy check.
type Result struct {
	ID          string `json:"id"`
	Group       string `json:"group"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Details     string `json:"details,omitempty"`
	Severity    string `json:"severity"`
	Remediation string `json:"remediation,omitempty"`
	Docs        string `json:"docs,omitempty"`
	Audit       string `json:"audit,omitempty"`
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
