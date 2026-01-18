// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package cli provides the terminal user interface for kspec security scanning.
package cli

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kopexa-grc/kspec/cli/components/browser"
	"github.com/kopexa-grc/kspec/cli/components/common"
)

// Model is the main UI model that uses the hierarchical browser
type Model struct {
	browser browser.Model
	tree    *common.ResourceTree
}

// InitialModel creates the initial UI model with hierarchical browser
func InitialModel() Model {
	// Start with an empty tree - will be populated during scan
	tree := common.NewResourceTree("Initializing...", "root")

	return Model{
		browser: browser.New(tree),
		tree:    tree,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.browser.Init()
}

// Update handles messages and updates the browser
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case UpdateTreeMsg:
		// Update tree reference and forward to browser
		m.tree = msg.Tree
		updated, cmd := m.browser.Update(browser.UpdateTreeMsg{Tree: msg.Tree})
		if b, ok := updated.(browser.Model); ok {
			m.browser = b
		}
		return m, cmd

	case SetTreeMsg:
		// Initialize/replace tree
		m.tree = msg.Tree
		m.browser = browser.New(msg.Tree)
		// Start spinner for new tree
		return m, m.browser.Init()
	}

	// Route to browser
	updated, cmd := m.browser.Update(msg)
	if b, ok := updated.(browser.Model); ok {
		m.browser = b
	}
	return m, cmd
}

// View renders the current view
func (m Model) View() string {
	return m.browser.View()
}

// UpdateTreeMsg updates the tree
type UpdateTreeMsg struct {
	Tree *common.ResourceTree
}

// SetTreeMsg sets/replaces the tree
type SetTreeMsg struct {
	Tree *common.ResourceTree
}

// AddCheckResultMsg adds a check result to a specific node
type AddCheckResultMsg struct {
	NodeID string
	Check  common.CheckResult
}

// UpdateNodeStateMsg updates a node's state
type UpdateNodeStateMsg struct {
	NodeID        string
	State         common.AssetState
	ChecksPassed  int
	ChecksFailed  int
	ChecksSkipped int
}

// ScanErrorMsg reports an error that occurred during scanning
type ScanErrorMsg struct {
	Error error
}

// GetTree returns the current resource tree from the model
func (m Model) GetTree() *common.ResourceTree {
	return m.tree
}
