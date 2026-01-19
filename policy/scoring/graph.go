// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package scoring

import (
	"github.com/kopexa-grc/kspec/cli/components/common"
	"github.com/kopexa-grc/kspec/policy"
)

// NodeScore contains score information for a single node.
type NodeScore struct {
	// LocalScore is the score based only on this node's own checks.
	LocalScore Score `json:"local_score"`

	// LocalFindings are the findings from this node's own checks.
	LocalFindings Findings `json:"local_findings"`

	// SubtreeScore is the score based on this node and all its children.
	SubtreeScore Score `json:"subtree_score"`

	// SubtreeFindings are the aggregated findings from the entire subtree.
	SubtreeFindings Findings `json:"subtree_findings"`

	// Criticality is the asset criticality level (if set).
	// TODO: Asset criticality weighting is not yet implemented in scoring
	// calculations. This field is a placeholder for future functionality
	// where critical assets may receive higher weight in aggregate scores.
	Criticality AssetCriticality `json:"criticality,omitempty"`
}

// GraphScorer calculates scores for a ResourceTree using bottom-up aggregation.
type GraphScorer struct {
	calculator    ScoreCalculator
	scores        map[string]*NodeScore
	criticalities map[string]AssetCriticality
}

// NewGraphScorer creates a new graph scorer with the specified scoring system.
func NewGraphScorer(system System) *GraphScorer {
	return &GraphScorer{
		calculator:    NewCalculator(system),
		scores:        make(map[string]*NodeScore),
		criticalities: make(map[string]AssetCriticality),
	}
}

// SetCriticality sets the asset criticality for a node.
func (g *GraphScorer) SetCriticality(nodeID string, criticality AssetCriticality) {
	g.criticalities[nodeID] = criticality
}

// ScoreTree calculates scores for the entire tree.
// Returns the root node's subtree score.
func (g *GraphScorer) ScoreTree(tree *common.ResourceTree) Score {
	if tree == nil || tree.Root == nil {
		return Score{
			Value:       100,
			Grade:       "A",
			RiskLevel:   "None",
			Completion:  100,
			Explanation: "No resources",
		}
	}

	// Reset scores for fresh calculation
	g.scores = make(map[string]*NodeScore)

	// Bottom-up traversal
	g.scoreNode(tree.Root)

	// Return root subtree score
	if rootScore, ok := g.scores[tree.Root.ID]; ok {
		return rootScore.SubtreeScore
	}

	return Score{
		Value:       100,
		Grade:       "A",
		RiskLevel:   "None",
		Completion:  100,
		Explanation: "No checks",
	}
}

// scoreNode recursively calculates scores for a node and its children.
// Returns the aggregated findings for propagation to parent nodes.
func (g *GraphScorer) scoreNode(node *common.ResourceNode) Findings {
	if node == nil {
		return Findings{}
	}

	nodeScore := &NodeScore{}

	// Set asset criticality if configured
	if crit, ok := g.criticalities[node.ID]; ok {
		nodeScore.Criticality = crit
	}

	// 1. Collect local findings from this node's checks
	for _, check := range node.Checks {
		status := mapPolicyStatus(check.Status)
		severity := Severity(check.Severity)
		nodeScore.LocalFindings.Add(status, severity)
	}

	// Calculate local score
	nodeScore.LocalScore = g.calculator.Calculate(nodeScore.LocalFindings)

	// 2. Aggregate subtree findings (local + all children)
	nodeScore.SubtreeFindings = nodeScore.LocalFindings

	// Traverse children
	for _, child := range node.Children {
		childFindings := g.scoreNode(child)
		nodeScore.SubtreeFindings.Merge(childFindings)
	}

	// Traverse sub-resources
	for _, subResources := range node.SubResources {
		for _, subRes := range subResources {
			subResFindings := g.scoreNode(subRes)
			nodeScore.SubtreeFindings.Merge(subResFindings)
		}
	}

	// Calculate subtree score
	nodeScore.SubtreeScore = g.calculator.Calculate(nodeScore.SubtreeFindings)

	// Store score
	g.scores[node.ID] = nodeScore

	// Return findings for parent aggregation
	return nodeScore.SubtreeFindings
}

// mapPolicyStatus maps policy.Status to scoring.CheckStatus.
func mapPolicyStatus(status policy.Status) CheckStatus {
	switch status {
	case policy.StatusPassed:
		return StatusPassed
	case policy.StatusFailed:
		return StatusFailed
	case policy.StatusSkipped:
		return StatusSkipped
	default:
		return StatusError
	}
}

// GetNodeScore returns the score for a specific node.
func (g *GraphScorer) GetNodeScore(nodeID string) (*NodeScore, bool) {
	score, ok := g.scores[nodeID]
	return score, ok
}

// GetAllScores returns all calculated node scores.
func (g *GraphScorer) GetAllScores() map[string]*NodeScore {
	return g.scores
}

// GetRootFindings returns the aggregated findings from the root node.
func (g *GraphScorer) GetRootFindings(tree *common.ResourceTree) Findings {
	if tree == nil || tree.Root == nil {
		return Findings{}
	}
	if rootScore, ok := g.scores[tree.Root.ID]; ok {
		return rootScore.SubtreeFindings
	}
	return Findings{}
}

// Calculator returns the underlying score calculator.
func (g *GraphScorer) Calculator() ScoreCalculator {
	return g.calculator
}
