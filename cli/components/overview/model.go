// Package overview provides the asset overview table component for the CLI.
package overview

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kopexa-grc/kspec/cli/components/common"
)

// Model represents the asset overview table
type Model struct {
	table  table.Model
	assets []common.Asset
	keys   common.KeyMap
}

// ShowDetailMsg is sent when user wants to drill down
type ShowDetailMsg struct {
	AssetID string
}

// New creates a new overview model
func New(assets []common.Asset) Model {
	columns := []table.Column{
		{Title: "Type", Width: 25},
		{Title: "Name", Width: 40},
		{Title: "Resources", Width: 12},
		{Title: "Status", Width: 15},
		{Title: "Results", Width: 25},
	}

	rows := make([]table.Row, len(assets))
	for i, asset := range assets {
		status := string(asset.State)
		statusStyled := common.StatusStyle(status).Render(status)

		results := fmt.Sprintf("✓%d ✗%d ⊘%d",
			asset.ChecksPassed,
			asset.ChecksFailed,
			asset.ChecksSkipped,
		)

		rows[i] = table.Row{
			asset.Type,
			asset.Name,
			fmt.Sprintf("%d", asset.ResourceCount),
			statusStyled,
			results,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(assets)+2),
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
		table:  t,
		assets: assets,
		keys:   common.DefaultKeyMap(),
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
		case "enter":
			// Drill down to selected asset
			selectedIdx := m.table.Cursor()
			if selectedIdx < len(m.assets) {
				return m, func() tea.Msg {
					return ShowDetailMsg{AssetID: m.assets[selectedIdx].ID}
				}
			}

		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the overview table
func (m Model) View() string {
	var output string

	output += "\n"
	output += common.TitleStyle.Render("Asset Overview")
	output += "\n\n"
	output += m.table.View()
	output += "\n\n"
	output += common.HelpStyle.Render("↑/↓: navigate • enter: view details • q: quit")
	output += "\n"

	return output
}

// UpdateAsset updates a specific asset's state
func (m *Model) UpdateAsset(assetID string, asset common.Asset) {
	for i, a := range m.assets {
		if a.ID != assetID {
			continue
		}
		m.assets[i] = asset

		// Update table row
		status := string(asset.State)
		statusStyled := common.StatusStyle(status).Render(status)

		results := fmt.Sprintf("✓%d ✗%d ⊘%d",
			asset.ChecksPassed,
			asset.ChecksFailed,
			asset.ChecksSkipped,
		)

		m.table.SetRows([]table.Row{
			{
				asset.Type,
				asset.Name,
				fmt.Sprintf("%d", asset.ResourceCount),
				statusStyled,
				results,
			},
		})
		break
	}
}
