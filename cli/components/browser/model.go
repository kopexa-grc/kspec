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

// Messages
type NavigateDownMsg struct{ NodeID string }
type NavigateUpMsg struct{}
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
	case "passed":
		statusIcon = "✓"
		statusText = "PASSED"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorSuccess).Bold(true)
	case "failed":
		statusIcon = "✗"
		statusText = "FAILED"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorError).Bold(true)
	case "skipped":
		statusIcon = "⊘"
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

// renderBreadcrumb renders the breadcrumb navigation
func (m Model) renderBreadcrumb() string {
	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		Bold(true)

	return breadcrumbStyle.Render(m.tree.GetBreadcrumbString())
}

// renderNodeSummary renders the summary of the current node
func (m Model) renderNodeSummary() string {
	node := m.tree.Current

	summary := fmt.Sprintf(
		"%s: %s | State: %s | Resources: %d | Checks: %s",
		node.Type,
		node.Name,
		node.StatusString(),
		node.ResourceCount,
		node.ChecksString(),
	)

	return common.TitleStyle.Render(summary)
}

// renderList renders the list view (children and sub-resources)
func (m Model) renderList() string {
	node := m.tree.Current

	// Show progress during discovery or scanning
	if node.State == common.AssetStateDiscovery || node.State == common.AssetStateScanning {
		return m.renderDiscoveryProgress()
	}

	// After scanning completes, if no children
	if !m.tree.HasChildren() {
		noChildrenStyle := lipgloss.NewStyle().
			Foreground(common.ColorMuted).
			Italic(true)
		return noChildrenStyle.Render("No sub-resources available.")
	}

	return m.table.View()
}

// renderDiscoveryProgress renders a nice discovery/scanning progress view
func (m Model) renderDiscoveryProgress() string {
	var output strings.Builder
	node := m.tree.Current

	// Box style for the progress panel
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorPrimary).
		Padding(1, 2).
		Width(60)

	// Header with spinner
	var headerText string
	var headerIcon string
	if node.State == common.AssetStateDiscovery {
		headerIcon = "🔍"
		headerText = "Discovering Resources"
	} else {
		headerIcon = "⚙️"
		headerText = "Scanning Resources"
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorPrimary)

	header := fmt.Sprintf("%s %s %s", m.spinner.View(), headerIcon, headerText)
	output.WriteString(headerStyle.Render(header))
	output.WriteString("\n\n")

	// Show discovered resources
	children := m.tree.GetCurrentChildren()
	if len(children) > 0 {
		// Resource discovery table
		resourceStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
		countStyle := lipgloss.NewStyle().Foreground(common.ColorSuccess).Bold(true)
		scanningStyle := lipgloss.NewStyle().Foreground(common.ColorWarning)
		completeStyle := lipgloss.NewStyle().Foreground(common.ColorSuccess)
		errorStyle := lipgloss.NewStyle().Foreground(common.ColorError)

		output.WriteString(resourceStyle.Render("Discovered resources:"))
		output.WriteString("\n\n")

		for _, child := range children {
			// Status indicator
			var statusIcon string
			var statusStyle lipgloss.Style

			switch child.State {
			case common.AssetStatePending:
				statusIcon = "○"
				statusStyle = resourceStyle
			case common.AssetStateScanning:
				statusIcon = "◐"
				statusStyle = scanningStyle
			case common.AssetStateComplete:
				statusIcon = "●"
				statusStyle = completeStyle
			case common.AssetStateError:
				statusIcon = "✗"
				statusStyle = errorStyle
			default:
				statusIcon = "○"
				statusStyle = resourceStyle
			}

			// Resource name and count
			resourceName := child.ResourceType
			if resourceName == "" {
				resourceName = child.Name
			}

			// Format: ● resource_type (count)
			line := fmt.Sprintf("  %s %s",
				statusStyle.Render(statusIcon),
				resourceStyle.Render(resourceName),
			)

			if child.ResourceCount > 0 {
				line += " " + countStyle.Render(fmt.Sprintf("(%d)", child.ResourceCount))
			}

			// Show check progress if scanning
			if child.State == common.AssetStateScanning || child.State == common.AssetStateComplete {
				total := child.ChecksPassed + child.ChecksFailed + child.ChecksSkipped
				if total > 0 {
					checkInfo := fmt.Sprintf(" [%d✓ %d✗]",
						child.ChecksPassed,
						child.ChecksFailed,
					)
					line += resourceStyle.Render(checkInfo)
				}
			}

			output.WriteString(line)
			output.WriteString("\n")
		}

		// Summary line
		output.WriteString("\n")
		totalResources := 0
		for _, child := range children {
			totalResources += child.ResourceCount
		}

		summaryStyle := lipgloss.NewStyle().Foreground(common.ColorMuted).Italic(true)
		summary := fmt.Sprintf("Total: %d resource types, %d resources",
			len(children),
			totalResources,
		)
		output.WriteString(summaryStyle.Render(summary))
	} else {
		// Still discovering, show placeholder
		waitingStyle := lipgloss.NewStyle().
			Foreground(common.ColorMuted).
			Italic(true)
		output.WriteString(waitingStyle.Render("Connecting to provider..."))
	}

	return boxStyle.Render(output.String())
}

// renderChecks renders the checks view
func (m Model) renderChecks() string {
	node := m.tree.Current

	if !node.HasChecks() {
		noChecksStyle := lipgloss.NewStyle().
			Foreground(common.ColorMuted).
			Italic(true)
		return noChecksStyle.Render("No checks available for this resource. Press 'l' to return to list view.")
	}

	return m.table.View()
}

// renderHelp renders the help text
func (m Model) renderHelp() string {
	var help string

	switch m.viewMode {
	case ViewModeList:
		help = "↑/↓: navigate • enter: drill down • esc: back • c: view checks • r: root • q: quit"
	case ViewModeChecks:
		help = "↑/↓: navigate • enter: view detail • esc: back • l: list view • r: root • q: quit"
	case ViewModeCheckDetail:
		help = "esc: back to checks • q: quit"
	}

	return common.HelpStyle.Render(help)
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
		case "passed":
			// statusStyle = lipgloss.NewStyle().Foreground(common.ColorSuccess)
			statusDisplay = "✓ passed"
		case "failed":
			// statusStyle = lipgloss.NewStyle().Foreground(common.ColorError)
			statusDisplay = "✗ failed"
		case "skipped":
			// statusStyle = lipgloss.NewStyle().Foreground(common.ColorMuted)
			statusDisplay = "⊘ skipped"
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderCheckDetail renders the detailed view of a single check
func (m Model) renderCheckDetail() string {
	if m.selectedCheck == nil {
		return "No check selected"
	}

	var output strings.Builder

	check := m.selectedCheck

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorPrimary).
		MarginBottom(1)
	output.WriteString(titleStyle.Render(check.Name))
	output.WriteString("\n\n")

	// Status with color
	statusLabel := lipgloss.NewStyle().Bold(true).Render("Status: ")
	statusValue := common.StatusStyle(check.Status).Render(check.Status)
	output.WriteString(statusLabel + statusValue)
	output.WriteString("\n\n")

	// Severity
	severityLabel := lipgloss.NewStyle().Bold(true).Render("Severity: ")
	severityStyle := lipgloss.NewStyle()
	switch check.Severity {
	case "critical", "high":
		severityStyle = severityStyle.Foreground(common.ColorError)
	case "medium":
		severityStyle = severityStyle.Foreground(common.ColorWarning)
	case "low":
		severityStyle = severityStyle.Foreground(common.ColorMuted)
	}
	output.WriteString(severityLabel + severityStyle.Render(check.Severity))
	output.WriteString("\n\n")

	// Group
	if check.Group != "" {
		groupLabel := lipgloss.NewStyle().Bold(true).Render("Group: ")
		output.WriteString(groupLabel + check.Group)
		output.WriteString("\n\n")
	}

	// ID
	if check.ID != "" {
		idLabel := lipgloss.NewStyle().Bold(true).Render("Check ID: ")
		idStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
		output.WriteString(idLabel + idStyle.Render(check.ID))
		output.WriteString("\n\n")
	}

	// Details/Remediation
	if check.Details != "" {
		detailsLabel := lipgloss.NewStyle().Bold(true).Render("Details:")
		output.WriteString(detailsLabel)
		output.WriteString("\n")

		detailsStyle := lipgloss.NewStyle().
			Foreground(common.ColorMuted).
			Width(min(m.width-4, 100)).
			MarginTop(1)
		output.WriteString(detailsStyle.Render(check.Details))
	}

	return output.String()
}
