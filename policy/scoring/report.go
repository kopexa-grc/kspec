// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

import "time"

// ScoredReport is the complete scoring result for a scan.
type ScoredReport struct {
	// Timestamp is when the score was calculated.
	Timestamp time.Time `json:"timestamp"`

	// Score is the overall score for the scanned asset.
	Score Score `json:"score"`

	// Findings is the aggregated findings from all checks.
	Findings Findings `json:"findings"`

	// System is the scoring system used.
	System System `json:"scoring_system"`

	// NodeScores contains per-node scores for drill-down.
	// Key is the node ID.
	NodeScores map[string]*NodeScore `json:"node_scores,omitempty"`
}

// NewScoredReport creates a new scored report from a GraphScorer result.
func NewScoredReport(scorer *GraphScorer, score Score, findings Findings) *ScoredReport {
	return &ScoredReport{
		Timestamp:  time.Now(),
		Score:      score,
		Findings:   findings,
		System:     scorer.Calculator().System(),
		NodeScores: scorer.GetAllScores(),
	}
}

// Summary returns a human-readable summary of the report.
func (r *ScoredReport) Summary() string {
	return r.Score.Explanation
}

// HasFindings returns true if there are any failed checks.
func (r *ScoredReport) HasFindings() bool {
	return r.Findings.Failed > 0
}

// IsPassing returns true if the score is above the passing threshold (70).
func (r *ScoredReport) IsPassing() bool {
	return r.Score.Value >= 70
}

// IsHealthy returns true if the score is 90 or above.
func (r *ScoredReport) IsHealthy() bool {
	return r.Score.Value >= 90
}
