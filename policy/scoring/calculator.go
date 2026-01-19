// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

// ScoreCalculator calculates scores from findings.
type ScoreCalculator interface {
	// Calculate computes the score from findings.
	Calculate(findings Findings) Score

	// System returns the scoring system used.
	System() System
}

// NewCalculator creates a calculator for the given scoring system.
func NewCalculator(system System) ScoreCalculator {
	switch system {
	case SystemBanded:
		return NewBandedCalculator()
	case SystemAverage:
		return NewAverageCalculator()
	case SystemDecayed:
		return NewDecayedCalculator()
	case SystemHighest:
		return NewHighestImpactCalculator()
	default:
		return NewBandedCalculator()
	}
}

// Calculate is a convenience function that creates a calculator and computes the score.
func Calculate(system System, findings Findings) Score {
	return NewCalculator(system).Calculate(findings)
}
