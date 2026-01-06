// Package detail provides the asset detail view component for the CLI.
package detail

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kopexa-grc/kspec/cli/components/common"
)

// Model represents the asset detail view
type Model struct {
	asset       common.Asset
	checksTable table.Model
	checks      []common.CheckResult
	keys        common.KeyMap
}

// BackToOverviewMsg is sent when user wants to go back
type BackToOverviewMsg struct{}

// New creates a new detail model
func New(asset common.Asset, checks []common.CheckResult) Model {
	columns := []table.Column{
		{Title: "Group", Width: 20},
		{Title: "Check", Width: 50},
		{Title: "Status", Width: 15},
		{Title: "Severity", Width: 12},
	}

	rows := make([]table.Row, len(checks))
	for i, check := range checks {
		statusStyled := common.StatusStyle(check.Status).Render(check.Status)

		rows[i] = table.Row{
			check.Group,
			check.Name,
			statusStyled,
			check.Severity,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(common.ColorMuted).
		BorderBottom(true).
		Bold(true).
		Foreground(common.ColorPrimary)

	s.Selected = s.Selected.
		Foreground(common.ColorHighlight).
		Bold(true)

	t.SetStyles(s)

	return Model{
		asset:       asset,
		checksTable: t,
		checks:      checks,
		keys:        common.DefaultKeyMap(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "backspace":
			return m, func() tea.Msg {
				return BackToOverviewMsg{}
			}

		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.checksTable, cmd = m.checksTable.Update(msg)
	return m, cmd
}

// View renders the detail view
func (m Model) View() string {
	var output string

	// Header
	output += "\n"
	output += common.TitleStyle.Render(m.asset.Name)
	output += "\n"
	output += common.SubtitleStyle.Render(m.asset.Type)
	output += "\n\n"

	// Summary
	summary := fmt.Sprintf(
		"Resources: %d  |  ✓ %d  |  ✗ %d  |  ⊘ %d",
		m.asset.ResourceCount,
		m.asset.ChecksPassed,
		m.asset.ChecksFailed,
		m.asset.ChecksSkipped,
	)
	output += summary
	output += "\n\n"

	// Checks table
	output += m.checksTable.View()
	output += "\n\n"

	// Footer
	output += common.HelpStyle.Render("↑/↓: navigate • esc: back • q: quit")
	output += "\n"

	return output
}
