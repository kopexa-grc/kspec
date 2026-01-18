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
	detailScroll int // Scroll position in detail panel
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
		detailScroll: 0,
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
			// In check detail view
			if m.viewMode == ViewModeCheckDetail {
				if m.focusedPanel == PanelLeft {
					// Navigate checks
					if m.checkCursor > 0 {
						m.checkCursor--
						m.detailScroll = 0 // Reset scroll when changing checks
					}
				} else {
					// Scroll detail content up
					if m.detailScroll > 0 {
						m.detailScroll--
					}
				}
				return m, nil
			}
			// Normal view: Move cursor up in current panel
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
			// In check detail view
			if m.viewMode == ViewModeCheckDetail {
				if m.focusedPanel == PanelLeft {
					// Navigate checks
					children := m.tree.GetCurrentChildren()
					if m.cursor < len(children) {
						node := children[m.cursor]
						checks := m.getChecksForNode(node)
						if m.checkCursor < len(checks)-1 {
							m.checkCursor++
							m.detailScroll = 0 // Reset scroll when changing checks
						}
					}
				} else {
					// Scroll detail content down
					m.detailScroll++
				}
				return m, nil
			}
			// Normal view: Move cursor down in current panel
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

		case "pgdown", "ctrl+d":
			// Fast scroll detail content down
			if m.viewMode == ViewModeCheckDetail && m.focusedPanel == PanelRight {
				m.detailScroll += 5
				return m, nil
			}

		case "pgup", "ctrl+u":
			// Fast scroll detail content up
			if m.viewMode == ViewModeCheckDetail && m.focusedPanel == PanelRight {
				m.detailScroll -= 5
				if m.detailScroll < 0 {
					m.detailScroll = 0
				}
				return m, nil
			}

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

// renderCheckDetailView renders the split-panel check detail view
func (m Model) renderCheckDetailView() string {
	// Get checks for the current node
	children := m.tree.GetCurrentChildren()
	if m.cursor >= len(children) {
		return "No resource selected"
	}
	node := children[m.cursor]
	checks := m.getChecksForNode(node)

	if len(checks) == 0 {
		return "No checks available"
	}

	// Ensure checkCursor is in bounds
	if m.checkCursor >= len(checks) {
		m.checkCursor = len(checks) - 1
	}
	if m.checkCursor < 0 {
		m.checkCursor = 0
	}

	check := &checks[m.checkCursor]

	// Calculate panel widths
	leftWidth := m.width * 35 / 100 // 35% for check list
	rightWidth := m.width - leftWidth - 3
	panelHeight := m.height - 4

	// Build left panel - check list
	leftPanel := m.renderCheckListPanel(checks, leftWidth, panelHeight)

	// Build right panel - check detail
	rightPanel := m.renderCheckDetailPanel(check, rightWidth, panelHeight)

	// Combine panels with focus indication
	inactiveBorder := lipgloss.Color("#444")
	activeBorder := common.ColorPrimary

	leftBorderColor := inactiveBorder
	rightBorderColor := inactiveBorder
	if m.focusedPanel == PanelLeft {
		leftBorderColor = activeBorder
	} else {
		rightBorderColor = activeBorder
	}

	leftStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Width(leftWidth).
		Height(panelHeight)

	rightStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Width(rightWidth).
		Height(panelHeight)

	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftPanel),
		rightStyle.Render(rightPanel),
	)

	// Help bar
	helpStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
	var helpText string
	if m.focusedPanel == PanelLeft {
		helpText = helpStyle.Render("↑↓ Navigate checks • Tab Switch to details • ESC Back • q Quit")
	} else {
		helpText = helpStyle.Render("↑↓ Scroll details • PgUp/PgDn Fast scroll • Tab Switch to checks • ESC Back")
	}

	return combined + "\n" + helpText
}

// renderCheckListPanel renders the left panel with the list of checks
func (m Model) renderCheckListPanel(checks []common.CheckResult, width, height int) string {
	lines := make([]string, 0, height)

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(common.ColorPrimary)
	lines = append(lines, headerStyle.Render("📋 Checks"), "")

	// Count stats
	passed, failed, skipped := 0, 0, 0
	for _, c := range checks {
		switch c.Status {
		case StatusPassed:
			passed++
		case StatusFailed:
			failed++
		case StatusSkipped:
			skipped++
		}
	}
	statsLine := fmt.Sprintf("%s %d  %s %d  %s %d",
		lipgloss.NewStyle().Foreground(common.ColorSuccess).Render("✓"),
		passed,
		lipgloss.NewStyle().Foreground(common.ColorError).Render("✗"),
		failed,
		lipgloss.NewStyle().Foreground(common.ColorMuted).Render("⊘"),
		skipped,
	)
	lines = append(lines, statsLine, strings.Repeat("─", width-4), "")

	// Calculate visible area
	visibleItems := height - 8
	if visibleItems < 1 {
		visibleItems = 1
	}
	startIdx := 0
	if m.checkCursor >= visibleItems {
		startIdx = m.checkCursor - visibleItems + 1
	}

	// Scroll up indicator
	if startIdx > 0 {
		lines = append(lines, common.MutedStyle.Render(fmt.Sprintf("↑ %d more", startIdx)))
	}

	// Check items
	for i, check := range checks {
		if i < startIdx {
			continue
		}
		if i-startIdx >= visibleItems {
			break
		}

		// Status icon
		var icon string
		var iconStyle lipgloss.Style
		switch check.Status {
		case StatusPassed:
			icon = SymbolCheck
			iconStyle = lipgloss.NewStyle().Foreground(common.ColorSuccess)
		case StatusFailed:
			icon = SymbolCross
			iconStyle = lipgloss.NewStyle().Foreground(common.ColorError)
		case StatusSkipped:
			icon = SymbolSkipped
			iconStyle = lipgloss.NewStyle().Foreground(common.ColorMuted)
		default:
			icon = "?"
			iconStyle = lipgloss.NewStyle()
		}

		// Truncate name to fit
		name := check.Name
		maxNameLen := width - 8
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		line := fmt.Sprintf("%s %s", iconStyle.Render(icon), name)

		// Highlight selected
		if i == m.checkCursor {
			line = lipgloss.NewStyle().
				Background(common.ColorPrimary).
				Foreground(lipgloss.Color("#000")).
				Bold(true).
				Width(width - 4).
				Render(line)
		}

		lines = append(lines, line)
	}

	// Scroll down indicator
	endIdx := startIdx + visibleItems
	if endIdx > len(checks) {
		endIdx = len(checks)
	}
	remaining := len(checks) - endIdx
	if remaining > 0 {
		lines = append(lines, common.MutedStyle.Render(fmt.Sprintf("↓ %d more", remaining)))
	}

	return strings.Join(lines, "\n")
}

// renderCheckDetailPanel renders the right panel with check details (scrollable)
func (m Model) renderCheckDetailPanel(check *common.CheckResult, width, height int) string {
	if check == nil {
		return "No check selected"
	}

	contentWidth := width - 4

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(common.ColorPrimary)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#fff"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(common.ColorPrimary)

	// Build ALL content lines first (for scrolling)
	allLines := make([]string, 0, height*2)

	// Header (fixed, not scrolled)
	header := headerStyle.Render("🔍 Details")

	// Title
	title := check.Name
	wrappedTitle := wrapText(title, contentWidth)
	for _, t := range wrappedTitle {
		allLines = append(allLines, titleStyle.Render(t))
	}
	allLines = append(allLines, "")

	// Status
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

	// Severity styling
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

	allLines = append(allLines,
		fmt.Sprintf("%s %s", labelStyle.Render("Status:"), statusStyle.Render(statusIcon+" "+statusText)),
		fmt.Sprintf("%s %s", labelStyle.Render("Severity:"), severityStyle.Render(check.Severity)),
	)

	if check.Group != "" {
		allLines = append(allLines, fmt.Sprintf("%s %s", labelStyle.Render("Group:"), check.Group))
	}
	if check.ID != "" {
		allLines = append(allLines, fmt.Sprintf("%s %s", labelStyle.Render("ID:"), common.MutedStyle.Render(check.ID)))
	}

	allLines = append(allLines, "")

	// Description
	if check.Docs != "" {
		allLines = append(allLines, sectionStyle.Render("📖 Description"))
		docLines := wrapText(check.Docs, contentWidth)
		allLines = append(allLines, docLines...)
		allLines = append(allLines, "")
	}

	// Remediation (shown for failed checks)
	if check.Details != "" {
		remediationStyle := lipgloss.NewStyle().Foreground(common.ColorWarning)
		allLines = append(allLines, sectionStyle.Render("🔧 Remediation"))
		remLines := wrapText(check.Details, contentWidth)
		for _, l := range remLines {
			allLines = append(allLines, remediationStyle.Render(l))
		}
		allLines = append(allLines, "")
	}

	// Audit
	if check.Audit != "" {
		auditStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#88ccff"))
		allLines = append(allLines, sectionStyle.Render("🔎 Audit"))
		auditLines := wrapText(check.Audit, contentWidth)
		for _, l := range auditLines {
			allLines = append(allLines, auditStyle.Render(l))
		}
	}

	// Apply scrolling
	visibleHeight := height - 4 // Reserve space for header and scroll indicator
	totalLines := len(allLines)

	// Clamp scroll position
	maxScroll := totalLines - visibleHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	scrollPos := m.detailScroll
	if scrollPos > maxScroll {
		scrollPos = maxScroll
	}
	if scrollPos < 0 {
		scrollPos = 0
	}

	// Build output with scroll
	var output []string
	output = append(output, header, "")

	// Scroll up indicator
	if scrollPos > 0 {
		output = append(output, common.MutedStyle.Render(fmt.Sprintf("↑ %d lines above", scrollPos)))
	}

	// Visible content
	endIdx := scrollPos + visibleHeight
	if endIdx > totalLines {
		endIdx = totalLines
	}

	for i := scrollPos; i < endIdx; i++ {
		output = append(output, allLines[i])
	}

	// Scroll down indicator
	remaining := totalLines - endIdx
	if remaining > 0 {
		output = append(output, common.MutedStyle.Render(fmt.Sprintf("↓ %d lines below", remaining)))
	}

	return strings.Join(output, "\n")
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 40
	}

	var lines []string
	// Split by newlines first
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			lines = append(lines, "")
			continue
		}

		// Wrap long lines
		for len(para) > maxWidth {
			// Find last space before maxWidth
			breakPoint := maxWidth
			for i := maxWidth; i > 0; i-- {
				if para[i] == ' ' {
					breakPoint = i
					break
				}
			}
			lines = append(lines, para[:breakPoint])
			para = strings.TrimSpace(para[breakPoint:])
		}
		if para != "" {
			lines = append(lines, para)
		}
	}

	return lines
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

			// Show checks based on node state
			switch child.State {
			case common.AssetStatePending, common.AssetStateDiscovery, common.AssetStateScanning:
				// Still processing - show spinner
				checksStr = "⏳ scanning..."
			case common.AssetStateComplete:
				if child.HasChecks() || child.TotalChecks() > 0 {
					checksStr = child.ChecksString()
				} else {
					checksStr = "no policies"
				}
			case common.AssetStateError:
				checksStr = "⚠ error"
			default:
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
