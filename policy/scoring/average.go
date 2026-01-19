// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

import "fmt"

// AverageCalculator implements weighted average scoring.
//
// The score is calculated as:
//
//	score = passed_weight / total_weight * 100
//
// Where weights are based on severity:
//   - Passed checks: weight 1
//   - Critical findings: weight 10
//   - High findings: weight 7
//   - Medium findings: weight 4
//   - Low findings: weight 1
type AverageCalculator struct{}

// NewAverageCalculator creates a new average calculator.
func NewAverageCalculator() *AverageCalculator {
	return &AverageCalculator{}
}

// System returns the scoring system.
func (c *AverageCalculator) System() System {
	return SystemAverage
}

// Calculate computes the score using weighted average.
func (c *AverageCalculator) Calculate(f Findings) Score {
	score := Score{
		Completion: c.completion(f),
	}

	evaluated := f.Evaluated() - f.Accepted
	if evaluated <= 0 {
		score.Value = 100
		score.Explanation = "No checks evaluated"
		score.Finalize()
		return score
	}

	// Calculate weighted totals
	// Passed checks get weight of 1
	passedWeight := float64(f.Passed)
	totalWeight := float64(f.Passed)

	// Failed checks weighted by severity
	// Info findings are excluded (weight 0) as they are purely informational
	totalWeight += float64(f.Critical) * SeverityCritical.Weight()
	totalWeight += float64(f.High) * SeverityHigh.Weight()
	totalWeight += float64(f.Medium) * SeverityMedium.Weight()
	totalWeight += float64(f.Low) * SeverityLow.Weight()
	// Note: Info findings have weight 0, so they don't affect the calculation

	if totalWeight == 0 {
		score.Value = 100
	} else {
		score.Value = clampScore(int(passedWeight / totalWeight * 100))
	}

	score.Explanation = fmt.Sprintf("%d/%d checks passed", f.Passed, evaluated)
	score.Finalize()

	return score
}

func (c *AverageCalculator) completion(f Findings) float64 {
	if f.Total == 0 {
		return 100.0
	}
	evaluated := f.Evaluated()
	return float64(evaluated) / float64(f.Total) * 100
}
