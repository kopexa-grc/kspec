package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CheckStatus int

const (
	StatusPending CheckStatus = iota
	StatusRunning
	StatusPassed
	StatusFailed
	StatusSkipped
)

type CheckResultMsg struct {
	ID       string
	Status   CheckStatus
	ErrorMsg string
	Details  string
}

type CheckItem struct {
	ID       string
	Group    string
	Name     string
	Status   CheckStatus
	ErrorMsg string
}

type Model struct {
	table    table.Model
	checks   []CheckItem
	quitting bool
}

func NewModel(checks []CheckItem) Model {
	columns := []table.Column{
		{Title: "State", Width: 12},
		{Title: "Group", Width: 20},
		{Title: "Check", Width: 50},
		{Title: "Details", Width: 40},
	}

	rows := make([]table.Row, len(checks))
	for i, check := range checks {
		rows[i] = rowFromCheck(check)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return Model{
		table:  t,
		checks: checks,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case CheckResultMsg:
		for i, check := range m.checks {
			if check.ID == msg.ID {
				m.checks[i].Status = msg.Status
				m.checks[i].ErrorMsg = msg.ErrorMsg
				// Update the row
				rows := m.table.Rows()
				rows[i] = rowFromCheck(m.checks[i])
				if msg.Details != "" {
					// Append details if space allows or handle differently
					// For now, simpler is better
				}
				m.table.SetRows(rows)
				break
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Calculate usage stats
	var passed, failed, pending, running, skipped int
	for _, c := range m.checks {
		switch c.Status {
		case StatusPassed:
			passed++
		case StatusFailed:
			failed++
		case StatusPending:
			pending++
		case StatusRunning:
			running++
		case StatusSkipped:
			skipped++
		}
	}

	total := len(m.checks)

	summary := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().MarginRight(2).Render(fmt.Sprintf("Total: %d", total)),
		PassedStyle.MarginRight(2).Render(fmt.Sprintf("Passed: %d", passed)),
		FailedStyle.MarginRight(2).Render(fmt.Sprintf("Failed: %d", failed)),
		PendingStyle.MarginRight(2).Render(fmt.Sprintf("Pending: %d", pending)),
		SkippedStyle.MarginRight(2).Render(fmt.Sprintf("Skipped: %d", skipped)),
	)

	if running > 0 {
		summary = lipgloss.JoinHorizontal(lipgloss.Left,
			summary,
			RunningStyle.MarginRight(2).Render(fmt.Sprintf("Running: %d", running)),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		baseStyle.Render(m.table.View()),
		"\n",
		summary,
		"\n",
	)
}

func rowFromCheck(c CheckItem) table.Row {
	stateStr := " PENDING  "

	switch c.Status {
	case StatusRunning:
		stateStr = RunningStyle.Render(" RUNNING  ")
	case StatusPassed:
		stateStr = PassedStyle.Render("  PASSED  ")
	case StatusFailed:
		stateStr = FailedStyle.Render("  FAILED  ")
	case StatusSkipped:
		stateStr = SkippedStyle.Render(" SKIPPED  ")
	default:
		stateStr = PendingStyle.Render(" PENDING  ")
	}

	return table.Row{
		stateStr,
		c.Group,
		c.Name,
		c.ErrorMsg,
	}
}
