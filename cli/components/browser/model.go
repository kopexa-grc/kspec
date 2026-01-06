package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kopexa-grc/kspec/cli/components/common"
)

// ViewMode represents the current view mode
type ViewMode string

// ViewMode constants represent the available view modes for the browser.
const (
	ViewModeList        ViewMode = "list"         // Show children/sub-resources
	ViewModeChecks      ViewMode = "checks"       // Show check results
	ViewModeCheckDetail ViewMode = "check_detail" // Show single check detail
)

// Model represents the hierarchical resource browser
type Model struct {
	tree          *common.ResourceTree
	table         table.Model
	spinner       spinner.Model
	styles        LayoutStyles
	viewMode      ViewMode
	selectedCheck *common.CheckResult
	keys          common.KeyMap
	width         int
	height        int

	// Split view state
	focusedPanel Panel
	cursor       int // Cursor position in left panel
	checkCursor  int // Cursor position in right panel (checks)
}

// NavigateDownMsg requests navigation into a child node.
type NavigateDownMsg struct{ NodeID string }

// NavigateUpMsg requests navigation to the parent node.
type NavigateUpMsg struct{}

// UpdateTreeMsg updates the browser with a new resource tree.
type UpdateTreeMsg struct{ Tree *common.ResourceTree }

// New creates a new browser model
func New(tree *common.ResourceTree) Model {
	// Initialize spinner with a nice style
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(common.ColorPrimary)

	m := Model{
		tree:         tree,
		spinner:      s,
		styles:       NewLayoutStyles(),
		viewMode:     ViewModeList,
		keys:         common.DefaultKeyMap(),
		width:        120,
		height:       30,
		focusedPanel: PanelLeft,
		cursor:       0,
		checkCursor:  0,
	}

	m.rebuildTable()
	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case UpdateTreeMsg:
		m.tree = msg.Tree
		m.rebuildTable()
		// Keep spinner running during discovery/scanning
		if m.tree.Root.State == common.AssetStateDiscovery || m.tree.Root.State == common.AssetStateScanning {
			cmds = append(cmds, m.spinner.Tick)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			// Move cursor up in current panel
			if m.focusedPanel == PanelLeft {
				if m.cursor > 0 {
					m.cursor--
				}
			} else {
				if m.checkCursor > 0 {
					m.checkCursor--
				}
			}
			return m, nil

		case "down", "j":
			// Move cursor down in current panel
			if m.focusedPanel == PanelLeft {
				children := m.tree.GetCurrentChildren()
				if m.cursor < len(children)-1 {
					m.cursor++
				}
			} else {
				children := m.tree.GetCurrentChildren()
				if m.cursor < len(children) {
					node := children[m.cursor]
					checks := m.getChecksForNode(node)
					if m.checkCursor < len(checks)-1 {
						m.checkCursor++
					}
				}
			}
			return m, nil

		case "tab":
			// Switch between panels
			if m.focusedPanel == PanelLeft {
				m.focusedPanel = PanelRight
			} else {
				m.focusedPanel = PanelLeft
			}
			return m, nil

		case "enter":
			if m.focusedPanel == PanelLeft {
				// Navigate down into selected item
				children := m.tree.GetCurrentChildren()
				if m.cursor < len(children) {
					selectedNode := children[m.cursor]

					// Check if this is a sub-resource type group
					if selectedNode.Type == "sub_resource_type_group" {
						instances := m.tree.GetSubResourceInstances(selectedNode.ResourceType)
						if len(instances) > 0 {
							m.tree.Current.Children = instances
							m.cursor = 0
							m.rebuildTable()
						}
					} else {
						// Regular node navigation
						err := m.tree.NavigateDown(selectedNode.ID)
						if err == nil {
							m.cursor = 0
							m.checkCursor = 0
							m.rebuildTable()
						}
					}
				}
			} else {
				// In right panel - show check detail
				children := m.tree.GetCurrentChildren()
				if m.cursor < len(children) {
					node := children[m.cursor]
					checks := m.getChecksForNode(node)
					if m.checkCursor < len(checks) {
						m.selectedCheck = &checks[m.checkCursor]
						m.viewMode = ViewModeCheckDetail
					}
				}
			}
			return m, nil

		case "backspace", "esc":
			// Handle based on current view mode
			if m.viewMode == ViewModeCheckDetail {
				// Back to split view
				m.viewMode = ViewModeList
				m.selectedCheck = nil
				return m, nil
			}

			// Navigate up to parent
			err := m.tree.NavigateUp()
			if err == nil {
				m.cursor = 0
				m.checkCursor = 0
				m.rebuildTable()
			}
			return m, nil

		case "r":
			// Navigate to root
			m.tree.NavigateToRoot()
			m.cursor = 0
			m.checkCursor = 0
			m.rebuildTable()
			return m, nil

		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	// Update table
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the browser view
func (m Model) View() string {
	// Use the new split layout for list view
	if m.viewMode == ViewModeCheckDetail {
		return m.renderCheckDetailView()
	}

	return m.renderLayout()
}

// renderCheckDetailView renders the full-screen check detail view
func (m Model) renderCheckDetailView() string {
	if m.selectedCheck == nil {
		return "No check selected"
	}

	check := m.selectedCheck
	contentWidth := min(m.width-8, 100)

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorPrimary)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#fff")).
		Background(lipgloss.Color("#333")).
		Padding(0, 2).
		MarginBottom(1)

	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorPrimary).
		MarginTop(1)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#888"))

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ccc")).
		Width(contentWidth)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444")).
		Padding(1, 2).
		Width(contentWidth + 6)

	var output strings.Builder

	// Header
	output.WriteString(headerStyle.Render("🔍 Check Detail"))
	output.WriteString("\n\n")

	// Title
	output.WriteString(titleStyle.Render(check.Name))
	output.WriteString("\n\n")

	// Status and Severity in a row
	var statusIcon, statusText string
	var statusStyle lipgloss.Style
	switch check.Status {
	case StatusPassed:
		statusIcon = SymbolCheck
		statusText = "PASSED"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorSuccess).Bold(true)
	case StatusFailed:
		statusIcon = SymbolCross
		statusText = "FAILED"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorError).Bold(true)
	case StatusSkipped:
		statusIcon = SymbolSkipped
		statusText = "SKIPPED"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorMuted)
	default:
		statusIcon = "?"
		statusText = check.Status
		statusStyle = lipgloss.NewStyle()
	}

	severityStyle := lipgloss.NewStyle()
	switch check.Severity {
	case "critical":
		severityStyle = severityStyle.Background(common.ColorError).Foreground(lipgloss.Color("#fff")).Bold(true).Padding(0, 1)
	case "high":
		severityStyle = severityStyle.Foreground(common.ColorError).Bold(true)
	case "medium":
		severityStyle = severityStyle.Foreground(common.ColorWarning)
	case "low":
		severityStyle = severityStyle.Foreground(common.ColorMuted)
	}

	statusLine := fmt.Sprintf("%s %s    %s %s    %s %s",
		labelStyle.Render("Status:"),
		statusStyle.Render(statusIcon+" "+statusText),
		labelStyle.Render("Severity:"),
		severityStyle.Render(check.Severity),
		labelStyle.Render("Group:"),
		check.Group,
	)
	output.WriteString(statusLine)
	output.WriteString("\n")

	// Check ID
	if check.ID != "" {
		output.WriteString(labelStyle.Render("ID: "))
		output.WriteString(lipgloss.NewStyle().Foreground(common.ColorMuted).Render(check.ID))
		output.WriteString("\n")
	}

	output.WriteString("\n")

	// Description (docs)
	if check.Docs != "" {
		output.WriteString(sectionTitleStyle.Render("📖 Description"))
		output.WriteString("\n")
		docsBox := boxStyle.Render(contentStyle.Render(check.Docs))
		output.WriteString(docsBox)
		output.WriteString("\n\n")
	}

	// Remediation
	if check.Details != "" {
		output.WriteString(sectionTitleStyle.Render("🔧 Remediation"))
		output.WriteString("\n")
		remediationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffa500")).
			Width(contentWidth)
		remediationBox := boxStyle.BorderForeground(common.ColorWarning).Render(remediationStyle.Render(check.Details))
		output.WriteString(remediationBox)
		output.WriteString("\n\n")
	}

	// Audit
	if check.Audit != "" {
		output.WriteString(sectionTitleStyle.Render("🔎 Audit"))
		output.WriteString("\n")
		auditStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88ccff")).
			Width(contentWidth)
		auditBox := boxStyle.BorderForeground(lipgloss.Color("#88ccff")).Render(auditStyle.Render(check.Audit))
		output.WriteString(auditBox)
		output.WriteString("\n\n")
	}

	// Help
	helpStyle := lipgloss.NewStyle().Foreground(common.ColorMuted).MarginTop(1)
	output.WriteString(helpStyle.Render("Press ESC to go back"))

	return output.String()
}

// UpdateTree updates the tree reference
func (m *Model) UpdateTree(tree *common.ResourceTree) {
	m.tree = tree
	m.rebuildTable()
}

// rebuildTable rebuilds the table based on current view mode and tree state
func (m *Model) rebuildTable() {
	if m.viewMode == ViewModeList {
		m.buildListTable()
	} else {
		m.buildChecksTable()
	}
}

// buildListTable builds the table for list view
func (m *Model) buildListTable() {
	columns := []table.Column{
		{Title: "Type", Width: 25},
		{Title: "Name", Width: 35},
		{Title: "State", Width: 15},
		{Title: "Count", Width: 10},
		{Title: "Checks", Width: 20},
	}

	children := m.getDisplayItems()
	rows := make([]table.Row, len(children))

	for i, child := range children {
		var typeStr, nameStr, stateStr, countStr, checksStr string

		if child.Type == "sub_resource_type_group" {
			// Sub-resource type group
			typeStr = "Sub-Resources"
			nameStr = child.Name
			stateStr = ""
			countStr = fmt.Sprintf("%d", child.ResourceCount)
			checksStr = ""
		} else {
			// Regular node
			typeStr = child.ResourceType
			if typeStr == "" {
				typeStr = child.Type
			}
			nameStr = child.Name
			stateStr = common.StatusStyle(child.StatusString()).Render(child.StatusString())

			if child.ResourceCount > 0 {
				countStr = fmt.Sprintf("%d", child.ResourceCount)
			} else {
				countStr = "-"
			}

			if child.HasChecks() || child.TotalChecks() > 0 {
				checksStr = child.ChecksString()
			} else {
				checksStr = "-"
			}
		}

		rows[i] = table.Row{
			typeStr,
			nameStr,
			stateStr,
			countStr,
			checksStr,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+2, 20)),
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
	m.table = t
}

// buildChecksTable builds the table for checks view
func (m *Model) buildChecksTable() {
	columns := []table.Column{
		{Title: "Group", Width: 20},
		{Title: "Check", Width: 50},
		{Title: "Status", Width: 15},
		{Title: "Severity", Width: 12},
	}

	checks := m.tree.Current.Checks
	rows := make([]table.Row, len(checks))

	for i, check := range checks {
		// Status with symbol and color
		var statusDisplay string

		switch check.Status {
		case StatusPassed:
			statusDisplay = SymbolCheck + " " + StatusPassed
		case StatusFailed:
			statusDisplay = SymbolCross + " " + StatusFailed
		case StatusSkipped:
			statusDisplay = SymbolSkipped + " " + StatusSkipped
		default:
			statusDisplay = check.Status
		}

		rows[i] = table.Row{
			check.Group,
			check.Name,
			statusDisplay,
			check.Severity,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+2, 20)),
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
	m.table = t
}

// getDisplayItems returns the items to display in the list view
func (m *Model) getDisplayItems() []*common.ResourceNode {
	return m.tree.GetCurrentChildren()
}
