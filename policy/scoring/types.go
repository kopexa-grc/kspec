// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package scoring provides algorithms for calculating security scores from findings.
// It supports multiple scoring systems including banded, average, decayed, and highest impact.
package scoring

import (
	"fmt"
	"strings"
)

// System defines the scoring algorithm to use.
type System string

const (
	// SystemBanded uses severity bands to calculate scores.
	// Critical findings cap the score at 40, high at 70, medium at 90.
	SystemBanded System = "banded"

	// SystemAverage calculates a weighted average based on severity.
	SystemAverage System = "average"

	// SystemDecayed applies exponential decay for each finding.
	SystemDecayed System = "decayed"

	// SystemHighest only considers the highest severity finding.
	SystemHighest System = "highest_impact"
)

// Explanation constants.
const (
	explanationNoFindings = "No findings"
)

// clampScore safely converts an int to a score value (0-100).
func clampScore(v int) uint32 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return uint32(v)
}

// ParseSystem parses a string to System.
func ParseSystem(s string) System {
	switch strings.ToLower(s) {
	case "banded":
		return SystemBanded
	case "average":
		return SystemAverage
	case "decayed":
		return SystemDecayed
	case "highest_impact", "highest":
		return SystemHighest
	default:
		return SystemBanded
	}
}

// String returns the string representation.
func (s System) String() string {
	return string(s)
}

// Severity represents the severity level of a finding.
type Severity string

// Severity level constants.
const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Weight returns the impact weight for the severity.
func (s Severity) Weight() float64 {
	switch s {
	case SeverityCritical:
		return 10.0
	case SeverityHigh:
		return 7.0
	case SeverityMedium:
		return 4.0
	case SeverityLow:
		return 1.0
	case SeverityInfo:
		return 0.0
	default:
		return 0.0
	}
}

// CheckStatus represents the status of a check result.
type CheckStatus string

// CheckStatus constants.
const (
	StatusPassed   CheckStatus = "passed"
	StatusFailed   CheckStatus = "failed"
	StatusSkipped  CheckStatus = "skipped"
	StatusError    CheckStatus = "error"
	StatusAccepted CheckStatus = "accepted" // Risk accepted, does not count as failed
)

// Score represents the result of a scoring calculation.
type Score struct {
	Value       uint32  `json:"value"`       // 0-100, higher is better
	Grade       string  `json:"grade"`       // A, B, C, D, F
	RiskLevel   string  `json:"risk_level"`  // Critical, High, Medium, Low, None
	Completion  float64 `json:"completion"`  // Percentage of checks evaluated
	Explanation string  `json:"explanation"` // Human-readable explanation
}

// ComputeGrade calculates the grade based on the score value.
func (s *Score) ComputeGrade() {
	switch {
	case s.Value >= 90:
		s.Grade = "A"
	case s.Value >= 80:
		s.Grade = "B"
	case s.Value >= 70:
		s.Grade = "C"
	case s.Value >= 60:
		s.Grade = "D"
	default:
		s.Grade = "F"
	}
}

// ComputeRiskLevel calculates the risk level based on the score value.
func (s *Score) ComputeRiskLevel() {
	switch {
	case s.Value >= 90:
		s.RiskLevel = "None"
	case s.Value >= 70:
		s.RiskLevel = "Low"
	case s.Value >= 40:
		s.RiskLevel = "Medium"
	case s.Value >= 1:
		s.RiskLevel = "High"
	default:
		s.RiskLevel = "Critical"
	}
}

// Finalize computes derived fields (Grade and RiskLevel).
func (s *Score) Finalize() {
	s.ComputeGrade()
	s.ComputeRiskLevel()
}

// Findings aggregates check results by status and severity.
type Findings struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
	Errors   int `json:"errors"`
	Accepted int `json:"accepted"` // Risk accepted, not counted in score

	// Failed checks by severity
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
}

// Add adds a check result to the findings.
func (f *Findings) Add(status CheckStatus, severity Severity) {
	f.Total++

	switch status {
	case StatusPassed:
		f.Passed++
	case StatusFailed:
		f.Failed++
		f.addSeverity(severity)
	case StatusSkipped:
		f.Skipped++
	case StatusError:
		f.Errors++
	case StatusAccepted:
		f.Accepted++
	}
}

func (f *Findings) addSeverity(severity Severity) {
	switch severity {
	case SeverityCritical:
		f.Critical++
	case SeverityHigh:
		f.High++
	case SeverityMedium:
		f.Medium++
	case SeverityLow:
		f.Low++
	case SeverityInfo:
		f.Info++
	}
}

// Merge combines two Findings structs.
func (f *Findings) Merge(other Findings) {
	f.Total += other.Total
	f.Passed += other.Passed
	f.Failed += other.Failed
	f.Skipped += other.Skipped
	f.Errors += other.Errors
	f.Accepted += other.Accepted
	f.Critical += other.Critical
	f.High += other.High
	f.Medium += other.Medium
	f.Low += other.Low
	f.Info += other.Info
}

// IsEmpty returns true if there are no findings.
func (f *Findings) IsEmpty() bool {
	return f.Total == 0
}

// FailedCount returns the total number of failed checks.
func (f *Findings) FailedCount() int {
	return f.Critical + f.High + f.Medium + f.Low + f.Info
}

// Evaluated returns the number of checks that were actually evaluated.
func (f *Findings) Evaluated() int {
	return f.Total - f.Skipped - f.Errors
}

// Explain returns a human-readable explanation of the findings.
func (f *Findings) Explain() string {
	if f.Failed == 0 {
		return "No findings"
	}

	parts := []string{}
	if f.Critical > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", f.Critical))
	}
	if f.High > 0 {
		parts = append(parts, fmt.Sprintf("%d high", f.High))
	}
	if f.Medium > 0 {
		parts = append(parts, fmt.Sprintf("%d medium", f.Medium))
	}
	if f.Low > 0 {
		parts = append(parts, fmt.Sprintf("%d low", f.Low))
	}
	if f.Info > 0 {
		parts = append(parts, fmt.Sprintf("%d info", f.Info))
	}

	if len(parts) == 0 {
		return "No findings"
	}

	return strings.Join(parts, ", ") + " findings"
}

// AssetCriticality defines the criticality level of an asset.
type AssetCriticality string

// AssetCriticality constants.
const (
	CriticalityMissionCritical AssetCriticality = "mission_critical"
	CriticalityHigh            AssetCriticality = "high"
	CriticalityMedium          AssetCriticality = "medium"
	CriticalityLow             AssetCriticality = "low"
)

// Weight returns the weighting factor for the criticality.
func (c AssetCriticality) Weight() float64 {
	switch c {
	case CriticalityMissionCritical:
		return 3.0
	case CriticalityHigh:
		return 2.0
	case CriticalityMedium:
		return 1.0
	case CriticalityLow:
		return 0.5
	default:
		return 1.0
	}
}
