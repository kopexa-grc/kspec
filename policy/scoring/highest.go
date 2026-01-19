// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

import "fmt"

// HighestImpactCalculator implements highest impact scoring.
//
// Only the highest severity level with findings matters:
//   - Any critical finding: score 0
//   - Any high finding (no critical): score 30
//   - Any medium finding (no critical/high): score 60
//   - Any low finding (no critical/high/medium): score 80
//   - No findings: score 100
//
// This is a zero-tolerance scoring system.
type HighestImpactCalculator struct{}

// NewHighestImpactCalculator creates a new highest impact calculator.
func NewHighestImpactCalculator() *HighestImpactCalculator {
	return &HighestImpactCalculator{}
}

// System returns the scoring system.
func (c *HighestImpactCalculator) System() System {
	return SystemHighest
}

// Calculate computes the score based on highest severity finding.
func (c *HighestImpactCalculator) Calculate(f Findings) Score {
	score := Score{
		Completion: c.completion(f),
	}

	switch {
	case f.Critical > 0:
		score.Value = 0
		score.Explanation = fmt.Sprintf("%d critical findings", f.Critical)
	case f.High > 0:
		score.Value = 30
		score.Explanation = fmt.Sprintf("%d high findings", f.High)
	case f.Medium > 0:
		score.Value = 60
		score.Explanation = fmt.Sprintf("%d medium findings", f.Medium)
	case f.Low > 0:
		score.Value = 80
		score.Explanation = fmt.Sprintf("%d low findings", f.Low)
	case f.Info > 0:
		score.Value = 95
		score.Explanation = fmt.Sprintf("%d info findings", f.Info)
	default:
		score.Value = 100
		score.Explanation = explanationNoFindings
	}

	score.Finalize()

	return score
}

func (c *HighestImpactCalculator) completion(f Findings) float64 {
	if f.Total == 0 {
		return 100.0
	}
	evaluated := f.Evaluated()
	return float64(evaluated) / float64(f.Total) * 100
}
