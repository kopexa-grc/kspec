// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

// BandedCalculator implements banded scoring.
//
// Scores are calculated based on severity bands:
//   - Critical findings: max score 40, -8 per additional finding
//   - High findings: max score 70, -5 per additional finding
//   - Medium findings: max score 90, -3 per additional finding
//   - Low findings: max score 99, -1 per additional finding
//   - No findings: score 100
//
// The highest severity band with findings determines the score.
type BandedCalculator struct{}

// NewBandedCalculator creates a new banded calculator.
func NewBandedCalculator() *BandedCalculator {
	return &BandedCalculator{}
}

// System returns the scoring system.
func (c *BandedCalculator) System() System {
	return SystemBanded
}

// Calculate computes the score using banded logic.
func (c *BandedCalculator) Calculate(f Findings) Score {
	score := Score{
		Completion: c.completion(f),
	}

	switch {
	case f.Critical > 0:
		// Critical: max 40, -8 per additional
		score.Value = clampScore(max(0, 40-(f.Critical-1)*8))
	case f.High > 0:
		// High: max 70, -5 per additional
		score.Value = clampScore(max(40, 70-(f.High-1)*5))
	case f.Medium > 0:
		// Medium: max 90, -3 per additional
		score.Value = clampScore(max(70, 90-(f.Medium-1)*3))
	case f.Low > 0:
		// Low: max 99, -1 per additional
		score.Value = clampScore(max(90, 99-(f.Low-1)*1))
	default:
		score.Value = 100
	}

	score.Explanation = f.Explain()
	score.Finalize()

	return score
}

func (c *BandedCalculator) completion(f Findings) float64 {
	if f.Total == 0 {
		return 100.0
	}
	evaluated := f.Evaluated()
	return float64(evaluated) / float64(f.Total) * 100
}
