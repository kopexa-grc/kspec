// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

import (
	"fmt"
	"math"
)

// DecayedCalculator implements exponential decay scoring.
//
// Starting from 100, each finding reduces the score proportionally
// to the current value:
//   - Critical: 40% decay per finding
//   - High: 20% decay per finding
//   - Medium: 10% decay per finding
//   - Low: 5% decay per finding
//
// Formula: score = 100 * (1-decay)^count for each severity level.
type DecayedCalculator struct {
	CriticalDecay float64
	HighDecay     float64
	MediumDecay   float64
	LowDecay      float64
}

// NewDecayedCalculator creates a new decayed calculator with default decay rates.
func NewDecayedCalculator() *DecayedCalculator {
	return &DecayedCalculator{
		CriticalDecay: 0.40,
		HighDecay:     0.20,
		MediumDecay:   0.10,
		LowDecay:      0.05,
	}
}

// NewDecayedCalculatorWithRates creates a calculator with custom decay rates.
func NewDecayedCalculatorWithRates(critical, high, medium, low float64) *DecayedCalculator {
	return &DecayedCalculator{
		CriticalDecay: critical,
		HighDecay:     high,
		MediumDecay:   medium,
		LowDecay:      low,
	}
}

// System returns the scoring system.
func (c *DecayedCalculator) System() System {
	return SystemDecayed
}

// Calculate computes the score using exponential decay.
func (c *DecayedCalculator) Calculate(f Findings) Score {
	score := Score{
		Completion: c.completion(f),
	}

	if f.Failed == 0 {
		score.Value = 100
		score.Explanation = explanationNoFindings
		score.Finalize()
		return score
	}

	// Start at 100 and apply decay for each finding
	value := 100.0

	// Apply decay in order of severity (most severe first)
	value *= math.Pow(1-c.CriticalDecay, float64(f.Critical))
	value *= math.Pow(1-c.HighDecay, float64(f.High))
	value *= math.Pow(1-c.MediumDecay, float64(f.Medium))
	value *= math.Pow(1-c.LowDecay, float64(f.Low))

	score.Value = uint32(math.Round(value))
	score.Explanation = fmt.Sprintf("Decayed from %d findings", f.Failed)
	score.Finalize()

	return score
}

func (c *DecayedCalculator) completion(f Findings) float64 {
	if f.Total == 0 {
		return 100.0
	}
	evaluated := f.Evaluated()
	return float64(evaluated) / float64(f.Total) * 100
}
