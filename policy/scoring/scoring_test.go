// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

import (
	"testing"

	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/policy"
)

// TestFindings tests the Findings type.
func TestFindings_Add(t *testing.T) {
	tests := []struct {
		name     string
		statuses []struct {
			status   CheckStatus
			severity Severity
		}
		want Findings
	}{
		{
			name:     "empty",
			statuses: nil,
			want:     Findings{},
		},
		{
			name: "all passed",
			statuses: []struct {
				status   CheckStatus
				severity Severity
			}{
				{StatusPassed, SeverityMedium},
				{StatusPassed, SeverityHigh},
			},
			want: Findings{Total: 2, Passed: 2},
		},
		{
			name: "mixed",
			statuses: []struct {
				status   CheckStatus
				severity Severity
			}{
				{StatusPassed, SeverityMedium},
				{StatusFailed, SeverityCritical},
				{StatusFailed, SeverityHigh},
				{StatusSkipped, SeverityMedium},
			},
			want: Findings{
				Total: 4, Passed: 1, Failed: 2, Skipped: 1,
				Critical: 1, High: 1,
			},
		},
		{
			name: "accepted",
			statuses: []struct {
				status   CheckStatus
				severity Severity
			}{
				{StatusAccepted, SeverityCritical},
				{StatusPassed, SeverityMedium},
			},
			want: Findings{Total: 2, Passed: 1, Accepted: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Findings
			for _, s := range tt.statuses {
				f.Add(s.status, s.severity)
			}

			if f.Total != tt.want.Total {
				t.Errorf("Total = %d, want %d", f.Total, tt.want.Total)
			}
			if f.Passed != tt.want.Passed {
				t.Errorf("Passed = %d, want %d", f.Passed, tt.want.Passed)
			}
			if f.Failed != tt.want.Failed {
				t.Errorf("Failed = %d, want %d", f.Failed, tt.want.Failed)
			}
			if f.Skipped != tt.want.Skipped {
				t.Errorf("Skipped = %d, want %d", f.Skipped, tt.want.Skipped)
			}
			if f.Critical != tt.want.Critical {
				t.Errorf("Critical = %d, want %d", f.Critical, tt.want.Critical)
			}
			if f.High != tt.want.High {
				t.Errorf("High = %d, want %d", f.High, tt.want.High)
			}
		})
	}
}

func TestFindings_Merge(t *testing.T) {
	f1 := Findings{Total: 5, Passed: 3, Failed: 2, Critical: 1, High: 1}
	f2 := Findings{Total: 3, Passed: 1, Failed: 2, Medium: 2}

	f1.Merge(f2)

	if f1.Total != 8 {
		t.Errorf("Total = %d, want 8", f1.Total)
	}
	if f1.Passed != 4 {
		t.Errorf("Passed = %d, want 4", f1.Passed)
	}
	if f1.Failed != 4 {
		t.Errorf("Failed = %d, want 4", f1.Failed)
	}
	if f1.Critical != 1 {
		t.Errorf("Critical = %d, want 1", f1.Critical)
	}
	if f1.Medium != 2 {
		t.Errorf("Medium = %d, want 2", f1.Medium)
	}
}

// TestBandedCalculator tests the banded scoring algorithm.
func TestBandedCalculator(t *testing.T) {
	tests := []struct {
		name      string
		findings  Findings
		wantScore uint32
		wantGrade string
	}{
		{
			name:      "no findings",
			findings:  Findings{Total: 10, Passed: 10},
			wantScore: 100,
			wantGrade: "A",
		},
		{
			name:      "one critical",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Critical: 1},
			wantScore: 40,
			wantGrade: "F",
		},
		{
			name:      "two critical",
			findings:  Findings{Total: 10, Passed: 8, Failed: 2, Critical: 2},
			wantScore: 32,
			wantGrade: "F",
		},
		{
			name:      "five critical",
			findings:  Findings{Total: 10, Passed: 5, Failed: 5, Critical: 5},
			wantScore: 8,
			wantGrade: "F",
		},
		{
			name:      "six critical hits zero",
			findings:  Findings{Total: 10, Passed: 4, Failed: 6, Critical: 6},
			wantScore: 0,
			wantGrade: "F",
		},
		{
			name:      "one high",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, High: 1},
			wantScore: 70,
			wantGrade: "C",
		},
		{
			name:      "three high",
			findings:  Findings{Total: 10, Passed: 7, Failed: 3, High: 3},
			wantScore: 60,
			wantGrade: "D",
		},
		{
			name:      "one medium",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Medium: 1},
			wantScore: 90,
			wantGrade: "A",
		},
		{
			name:      "five medium",
			findings:  Findings{Total: 10, Passed: 5, Failed: 5, Medium: 5},
			wantScore: 78,
			wantGrade: "C",
		},
		{
			name:      "one low",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Low: 1},
			wantScore: 99,
			wantGrade: "A",
		},
		{
			name:      "critical dominates high",
			findings:  Findings{Total: 10, Passed: 5, Failed: 5, Critical: 1, High: 4},
			wantScore: 40,
			wantGrade: "F",
		},
		{
			name:      "high dominates medium",
			findings:  Findings{Total: 10, Passed: 5, Failed: 5, High: 1, Medium: 4},
			wantScore: 70,
			wantGrade: "C",
		},
	}

	calc := NewBandedCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calc.Calculate(tt.findings)

			if score.Value != tt.wantScore {
				t.Errorf("Score = %d, want %d", score.Value, tt.wantScore)
			}
			if score.Grade != tt.wantGrade {
				t.Errorf("Grade = %s, want %s", score.Grade, tt.wantGrade)
			}
		})
	}
}

// TestAverageCalculator tests the weighted average scoring algorithm.
func TestAverageCalculator(t *testing.T) {
	tests := []struct {
		name      string
		findings  Findings
		wantScore uint32
	}{
		{
			name:      "no findings",
			findings:  Findings{Total: 10, Passed: 10},
			wantScore: 100,
		},
		{
			name:      "all passed",
			findings:  Findings{Total: 5, Passed: 5},
			wantScore: 100,
		},
		{
			name: "one critical vs many passed",
			// 10 passed (weight 10) + 1 critical (weight 10) = 20 total weight
			// passed weight = 10, score = 10/20*100 = 50
			findings:  Findings{Total: 11, Passed: 10, Failed: 1, Critical: 1},
			wantScore: 50,
		},
		{
			name: "one low vs one passed",
			// 1 passed (weight 1) + 1 low (weight 1) = 2 total weight
			// passed weight = 1, score = 1/2*100 = 50
			findings:  Findings{Total: 2, Passed: 1, Failed: 1, Low: 1},
			wantScore: 50,
		},
	}

	calc := NewAverageCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calc.Calculate(tt.findings)

			if score.Value != tt.wantScore {
				t.Errorf("Score = %d, want %d", score.Value, tt.wantScore)
			}
		})
	}
}

// TestDecayedCalculator tests the exponential decay scoring algorithm.
func TestDecayedCalculator(t *testing.T) {
	tests := []struct {
		name      string
		findings  Findings
		wantScore uint32
	}{
		{
			name:      "no findings",
			findings:  Findings{Total: 10, Passed: 10},
			wantScore: 100,
		},
		{
			name:      "one critical", // decay 40% -> 60
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Critical: 1},
			wantScore: 60,
		},
		{
			name:      "two critical", // decay 40% twice -> 36
			findings:  Findings{Total: 10, Passed: 8, Failed: 2, Critical: 2},
			wantScore: 36,
		},
		{
			name:      "one high", // decay 20% -> 80
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, High: 1},
			wantScore: 80,
		},
		{
			name:      "one medium", // decay 10% -> 90
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Medium: 1},
			wantScore: 90,
		},
		{
			name:      "one low", // decay 5% -> 95
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Low: 1},
			wantScore: 95,
		},
		{
			name: "mixed findings",
			// 100 * 0.60 * 0.80 * 0.90 = 43.2 ≈ 43
			findings:  Findings{Total: 10, Passed: 7, Failed: 3, Critical: 1, High: 1, Medium: 1},
			wantScore: 43,
		},
	}

	calc := NewDecayedCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calc.Calculate(tt.findings)

			if score.Value != tt.wantScore {
				t.Errorf("Score = %d, want %d", score.Value, tt.wantScore)
			}
		})
	}
}

// TestHighestImpactCalculator tests the highest impact scoring algorithm.
func TestHighestImpactCalculator(t *testing.T) {
	tests := []struct {
		name      string
		findings  Findings
		wantScore uint32
		wantGrade string
	}{
		{
			name:      "no findings",
			findings:  Findings{Total: 10, Passed: 10},
			wantScore: 100,
			wantGrade: "A",
		},
		{
			name:      "one critical",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Critical: 1},
			wantScore: 0,
			wantGrade: "F",
		},
		{
			name:      "many critical",
			findings:  Findings{Total: 10, Passed: 5, Failed: 5, Critical: 5},
			wantScore: 0,
			wantGrade: "F",
		},
		{
			name:      "one high",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, High: 1},
			wantScore: 30,
			wantGrade: "F",
		},
		{
			name:      "one medium",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Medium: 1},
			wantScore: 60,
			wantGrade: "D",
		},
		{
			name:      "one low",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Low: 1},
			wantScore: 80,
			wantGrade: "B",
		},
		{
			name:      "one info",
			findings:  Findings{Total: 10, Passed: 9, Failed: 1, Info: 1},
			wantScore: 95,
			wantGrade: "A",
		},
		{
			name:      "critical dominates all",
			findings:  Findings{Total: 10, Passed: 0, Failed: 10, Critical: 1, High: 3, Medium: 3, Low: 3},
			wantScore: 0,
			wantGrade: "F",
		},
	}

	calc := NewHighestImpactCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calc.Calculate(tt.findings)

			if score.Value != tt.wantScore {
				t.Errorf("Score = %d, want %d", score.Value, tt.wantScore)
			}
			if score.Grade != tt.wantGrade {
				t.Errorf("Grade = %s, want %s", score.Grade, tt.wantGrade)
			}
		})
	}
}

// TestNewCalculator tests the calculator factory.
func TestNewCalculator(t *testing.T) {
	tests := []struct {
		system System
		want   System
	}{
		{SystemBanded, SystemBanded},
		{SystemAverage, SystemAverage},
		{SystemDecayed, SystemDecayed},
		{SystemHighest, SystemHighest},
		{"unknown", SystemBanded}, // Default to banded
	}

	for _, tt := range tests {
		t.Run(string(tt.system), func(t *testing.T) {
			calc := NewCalculator(tt.system)
			if calc.System() != tt.want {
				t.Errorf("System() = %s, want %s", calc.System(), tt.want)
			}
		})
	}
}

// TestParseSystem tests parsing scoring system strings.
func TestParseSystem(t *testing.T) {
	tests := []struct {
		input string
		want  System
	}{
		{"banded", SystemBanded},
		{"BANDED", SystemBanded},
		{"average", SystemAverage},
		{"decayed", SystemDecayed},
		{"highest_impact", SystemHighest},
		{"highest", SystemHighest},
		{"unknown", SystemBanded},
		{"", SystemBanded},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseSystem(tt.input)
			if got != tt.want {
				t.Errorf("ParseSystem(%q) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

// TestGraphScorer tests graph-based scoring.
func TestGraphScorer_LeafNode(t *testing.T) {
	// Create a simple tree with one node
	tree := common.NewResourceTree("test-org", "organization")

	// Add a check result directly to root
	tree.Root.Checks = []common.CheckResult{
		{Status: policy.StatusPassed, Severity: "medium"},
		{Status: policy.StatusFailed, Severity: "high"},
	}

	scorer := NewGraphScorer(SystemBanded)
	score := scorer.ScoreTree(tree)

	// One high finding = score 70
	if score.Value != 70 {
		t.Errorf("Score = %d, want 70", score.Value)
	}
	if score.Grade != "C" {
		t.Errorf("Grade = %s, want C", score.Grade)
	}
}

func TestGraphScorer_Aggregation(t *testing.T) {
	// Create a tree with root and two children
	tree := common.NewResourceTree("test-org", "organization")

	// Add child1 with a critical finding
	child1 := &common.ResourceNode{
		ID:   "repo-1",
		Name: "repo-1",
		Type: "resource_instance",
		Checks: []common.CheckResult{
			{Status: policy.StatusFailed, Severity: "critical"},
		},
	}
	_ = tree.AddNode(tree.Root.ID, child1)

	// Add child2 with all passed
	child2 := &common.ResourceNode{
		ID:   "repo-2",
		Name: "repo-2",
		Type: "resource_instance",
		Checks: []common.CheckResult{
			{Status: policy.StatusPassed, Severity: "medium"},
			{Status: policy.StatusPassed, Severity: "medium"},
		},
	}
	_ = tree.AddNode(tree.Root.ID, child2)

	scorer := NewGraphScorer(SystemBanded)
	score := scorer.ScoreTree(tree)

	// One critical finding in subtree = score 40
	if score.Value != 40 {
		t.Errorf("Root score = %d, want 40", score.Value)
	}

	// Check individual node scores
	child1Score, ok := scorer.GetNodeScore("repo-1")
	if !ok {
		t.Fatal("repo-1 score not found")
	}
	if child1Score.LocalScore.Value != 40 {
		t.Errorf("child1 local score = %d, want 40", child1Score.LocalScore.Value)
	}

	child2Score, ok := scorer.GetNodeScore("repo-2")
	if !ok {
		t.Fatal("repo-2 score not found")
	}
	if child2Score.LocalScore.Value != 100 {
		t.Errorf("child2 local score = %d, want 100", child2Score.LocalScore.Value)
	}
}

func TestGraphScorer_DeepTree(t *testing.T) {
	// Create a deep tree: root -> child -> grandchild
	tree := common.NewResourceTree("org", "organization")

	child := &common.ResourceNode{
		ID:   "team-1",
		Name: "team-1",
		Type: "resource_instance",
	}
	_ = tree.AddNode(tree.Root.ID, child)

	grandchild := &common.ResourceNode{
		ID:   "member-1",
		Name: "member-1",
		Type: "resource_instance",
		Checks: []common.CheckResult{
			{Status: policy.StatusFailed, Severity: "critical"},
		},
	}
	_ = tree.AddNode(child.ID, grandchild)

	scorer := NewGraphScorer(SystemBanded)
	score := scorer.ScoreTree(tree)

	// Critical at leaf should propagate to root
	if score.Value != 40 {
		t.Errorf("Root score = %d, want 40", score.Value)
	}

	// Root findings should include grandchild's findings
	rootFindings := scorer.GetRootFindings(tree)
	if rootFindings.Critical != 1 {
		t.Errorf("Root critical = %d, want 1", rootFindings.Critical)
	}
}

func TestGraphScorer_EmptyTree(t *testing.T) {
	tree := common.NewResourceTree("test", "test")

	scorer := NewGraphScorer(SystemBanded)
	score := scorer.ScoreTree(tree)

	// Empty tree = score 100
	if score.Value != 100 {
		t.Errorf("Score = %d, want 100", score.Value)
	}
}

func TestGraphScorer_NilTree(t *testing.T) {
	scorer := NewGraphScorer(SystemBanded)
	score := scorer.ScoreTree(nil)

	if score.Value != 100 {
		t.Errorf("Score = %d, want 100", score.Value)
	}
}

// TestScore tests Score methods.
func TestScore_ComputeGrade(t *testing.T) {
	tests := []struct {
		value uint32
		grade string
	}{
		{100, "A"},
		{90, "A"},
		{89, "B"},
		{80, "B"},
		{79, "C"},
		{70, "C"},
		{69, "D"},
		{60, "D"},
		{59, "F"},
		{0, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.grade, func(t *testing.T) {
			s := Score{Value: tt.value}
			s.ComputeGrade()
			if s.Grade != tt.grade {
				t.Errorf("Score %d: Grade = %s, want %s", tt.value, s.Grade, tt.grade)
			}
		})
	}
}

func TestScore_ComputeRiskLevel(t *testing.T) {
	tests := []struct {
		value uint32
		risk  string
	}{
		{100, "None"},
		{90, "None"},
		{89, "Low"},
		{70, "Low"},
		{69, "Medium"},
		{40, "Medium"},
		{39, "High"},
		{1, "High"},
		{0, "Critical"},
	}

	for _, tt := range tests {
		t.Run(tt.risk, func(t *testing.T) {
			s := Score{Value: tt.value}
			s.ComputeRiskLevel()
			if s.RiskLevel != tt.risk {
				t.Errorf("Score %d: RiskLevel = %s, want %s", tt.value, s.RiskLevel, tt.risk)
			}
		})
	}
}

// TestAssetCriticality tests asset criticality weights.
func TestAssetCriticality_Weight(t *testing.T) {
	tests := []struct {
		crit   AssetCriticality
		weight float64
	}{
		{CriticalityMissionCritical, 3.0},
		{CriticalityHigh, 2.0},
		{CriticalityMedium, 1.0},
		{CriticalityLow, 0.5},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(string(tt.crit), func(t *testing.T) {
			if tt.crit.Weight() != tt.weight {
				t.Errorf("Weight() = %f, want %f", tt.crit.Weight(), tt.weight)
			}
		})
	}
}

// TestScoredReport tests the scored report.
func TestScoredReport(t *testing.T) {
	tree := common.NewResourceTree("test", "test")
	tree.Root.Checks = []common.CheckResult{
		{Status: policy.StatusFailed, Severity: "high"},
		{Status: policy.StatusFailed, Severity: "high"},
	}

	scorer := NewGraphScorer(SystemBanded)
	score := scorer.ScoreTree(tree)
	findings := scorer.GetRootFindings(tree)

	report := NewScoredReport(scorer, score, findings)

	if !report.HasFindings() {
		t.Error("HasFindings() = false, want true")
	}
	if report.IsPassing() {
		t.Error("IsPassing() = true, want false (score 65)")
	}
	if report.IsHealthy() {
		t.Error("IsHealthy() = true, want false")
	}
	if report.System != SystemBanded {
		t.Errorf("System = %s, want %s", report.System, SystemBanded)
	}
}
