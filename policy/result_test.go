// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package policy

import "testing"

func TestResult_IsPassed(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"passed status", StatusPassed, true},
		{"failed status", StatusFailed, false},
		{"skipped status", StatusSkipped, false},
		{"empty status", "", false},
		{"random status", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Status: tt.status}
			if got := r.IsPassed(); got != tt.want {
				t.Errorf("Result.IsPassed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_IsFailed(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"failed status", StatusFailed, true},
		{"passed status", StatusPassed, false},
		{"skipped status", StatusSkipped, false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Status: tt.status}
			if got := r.IsFailed(); got != tt.want {
				t.Errorf("Result.IsFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_IsSkipped(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"skipped status", StatusSkipped, true},
		{"passed status", StatusPassed, false},
		{"failed status", StatusFailed, false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Status: tt.status}
			if got := r.IsSkipped(); got != tt.want {
				t.Errorf("Result.IsSkipped() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusConstants(t *testing.T) {
	// Verify status constants have expected values
	if StatusPassed != "passed" {
		t.Errorf("StatusPassed = %q, want %q", StatusPassed, "passed")
	}
	if StatusFailed != "failed" {
		t.Errorf("StatusFailed = %q, want %q", StatusFailed, "failed")
	}
	if StatusSkipped != "skipped" {
		t.Errorf("StatusSkipped = %q, want %q", StatusSkipped, "skipped")
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPassed, "passed"},
		{StatusFailed, "failed"},
		{StatusSkipped, "skipped"},
		{Status("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

