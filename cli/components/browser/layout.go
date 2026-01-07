// Package browser provides a hierarchical browser component for navigating
// scan results and policy checks.
package browser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kopexa-grc/kspec/cli/components/common"
)

// Layout constants
const (
	MinLeftPanelWidth  = 35
	MinRightPanelWidth = 50
	StatusBarHeight    = 3
	HeaderHeight       = 3
)

// Status symbols
const (
	SymbolPending = "○"
	SymbolCheck   = "✓"
	SymbolCross   = "✗"
	SymbolSkipped = "⊘"
)

// Status strings
const (
	StatusPassed  = "passed"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Panel represents which panel is focused
type Panel int

// Panel constants represent the available panels in the split view.
const (
	PanelLeft Panel = iota
	PanelRight
)

// LayoutStyles holds all the styles for the layout
type LayoutStyles struct {
	// Panel styles
	LeftPanel       lipgloss.Style
	RightPanel      lipgloss.Style
	LeftPanelFocus  lipgloss.Style
	RightPanelFocus lipgloss.Style

	// Header styles
	Header      lipgloss.Style
	HeaderTitle lipgloss.Style

	// Status bar styles
	StatusBar     lipgloss.Style
	StatusSuccess lipgloss.Style
	StatusError   lipgloss.Style
	StatusPending lipgloss.Style

	// Content styles
	TreeItem         lipgloss.Style
	TreeItemSelected lipgloss.Style
	TreeItemScanning lipgloss.Style
	CheckItem        lipgloss.Style
	CheckPassed      lipgloss.Style
	CheckFailed      lipgloss.Style
	CheckSkipped     lipgloss.Style
}

// NewLayoutStyles creates the default layout styles
func NewLayoutStyles() LayoutStyles {
	borderColor := lipgloss.Color("#444")
	focusBorderColor := common.ColorPrimary

	basePanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	focusPanelStyle := basePanelStyle.
		BorderForeground(focusBorderColor)

	return LayoutStyles{
		LeftPanel:       basePanelStyle,
		RightPanel:      basePanelStyle,
		LeftPanelFocus:  focusPanelStyle,
		RightPanelFocus: focusPanelStyle,

		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(common.ColorPrimary).
			Padding(0, 1).
			MarginBottom(1),

		HeaderTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fff")),

		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("#333")).
			Foreground(lipgloss.Color("#fff")).
			Padding(0, 2),

		StatusSuccess: lipgloss.NewStyle().
			Foreground(common.ColorSuccess).
			Bold(true),

		StatusError: lipgloss.NewStyle().
			Foreground(common.ColorError).
			Bold(true),

		StatusPending: lipgloss.NewStyle().
			Foreground(common.ColorWarning),

		TreeItem: lipgloss.NewStyle().
			Padding(0, 1),

		TreeItemSelected: lipgloss.NewStyle().
			Background(common.ColorPrimary).
			Foreground(lipgloss.Color("#fff")).
			Bold(true).
			Padding(0, 1),

		TreeItemScanning: lipgloss.NewStyle().
			Foreground(common.ColorWarning).
			Padding(0, 1),

		CheckItem: lipgloss.NewStyle().
			Padding(0, 1),

		CheckPassed: lipgloss.NewStyle().
			Foreground(common.ColorSuccess),

		CheckFailed: lipgloss.NewStyle().
			Foreground(common.ColorError),

		CheckSkipped: lipgloss.NewStyle().
			Foreground(common.ColorMuted),
	}
}

// renderLayout renders the split view layout
func (m Model) renderLayout() string {
	// Calculate panel dimensions
	totalWidth := m.width
	if totalWidth < MinLeftPanelWidth+MinRightPanelWidth {
		totalWidth = MinLeftPanelWidth + MinRightPanelWidth
	}

	leftWidth := totalWidth * 35 / 100 // 35% for left panel
	if leftWidth < MinLeftPanelWidth {
		leftWidth = MinLeftPanelWidth
	}
	rightWidth := totalWidth - leftWidth - 4 // Account for borders

	contentHeight := m.height - StatusBarHeight - HeaderHeight - 2
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Build header
	header := m.renderHeader()

	// Build left panel (resource tree)
	leftContent := m.renderResourceTree(leftWidth-4, contentHeight-2)
	leftPanel := m.styles.LeftPanel
	if m.focusedPanel == PanelLeft {
		leftPanel = m.styles.LeftPanelFocus
	}
	leftPanel = leftPanel.Width(leftWidth).Height(contentHeight)
	leftRendered := leftPanel.Render(leftContent)

	// Build right panel (check results)
	rightContent := m.renderCheckPanel(rightWidth-4, contentHeight-2)
	rightPanel := m.styles.RightPanel
	if m.focusedPanel == PanelRight {
		rightPanel = m.styles.RightPanelFocus
	}
	rightPanel = rightPanel.Width(rightWidth).Height(contentHeight)
	rightRendered := rightPanel.Render(rightContent)

	// Join panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, rightRendered)

	// Build status bar
	statusBar := m.renderStatusBar(totalWidth)

	// Join everything vertically
	return lipgloss.JoinVertical(lipgloss.Left, header, panels, statusBar)
}

// renderHeader renders the header with breadcrumb and scan status
func (m Model) renderHeader() string {
	parts := make([]string, 0, 4)

	// Logo/Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorPrimary)
	parts = append(parts, titleStyle.Render("kspec"))

	// Separator
	sepStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
	parts = append(parts, sepStyle.Render(" │ "))

	// Breadcrumb
	breadcrumb := m.tree.GetBreadcrumbString()
	breadcrumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
	parts = append(parts, breadcrumbStyle.Render(breadcrumb))

	// Status indicator (right aligned)
	statusIcon := ""
	statusText := ""
	var statusStyle lipgloss.Style

	switch m.tree.Root.State {
	case common.AssetStatePending:
		statusIcon = SymbolPending
		statusText = "Ready"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorMuted)
	case common.AssetStateDiscovery:
		statusIcon = m.spinner.View()
		statusText = "Discovering..."
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorWarning)
	case common.AssetStateScanning:
		statusIcon = m.spinner.View()
		statusText = "Scanning..."
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorWarning)
	case common.AssetStateComplete:
		statusIcon = SymbolCheck
		statusText = "Complete"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorSuccess)
	case common.AssetStateError:
		statusIcon = SymbolCross
		statusText = "Error"
		statusStyle = lipgloss.NewStyle().Foreground(common.ColorError)
	}

	status := fmt.Sprintf("%s %s", statusIcon, statusText)
	parts = append(parts, "  "+statusStyle.Render(status))

	return strings.Join(parts, "") + "\n"
}

// renderResourceTree renders the left panel with resource tree
func (m Model) renderResourceTree(width, height int) string {
	var lines []string

	// Panel title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorPrimary).
		MarginBottom(1)
	lines = append(lines, titleStyle.Render("📁 Resources"), "")

	// Get children to display
	children := m.tree.GetCurrentChildren()

	if len(children) == 0 {
		// Show discovery progress or empty state
		if m.tree.Root.State == common.AssetStateDiscovery {
			lines = append(lines, m.styles.StatusPending.Render(m.spinner.View()+" Connecting..."))
		} else {
			lines = append(lines, common.MutedStyle.Render("No resources found"))
		}
	} else {
		// Render each resource
		visibleItems := height - 4 // Account for title and margins
		startIdx := 0
		if m.cursor >= visibleItems {
			startIdx = m.cursor - visibleItems + 1
		}

		for i, child := range children {
			if i < startIdx {
				continue
			}
			if i-startIdx >= visibleItems {
				break
			}

			line := m.renderTreeItem(child, i == m.cursor, width-2)
			lines = append(lines, line)
		}

		// Show scroll indicator if needed
		if len(children) > visibleItems {
			scrollInfo := fmt.Sprintf("(%d/%d)", m.cursor+1, len(children))
			lines = append(lines, common.MutedStyle.Render(scrollInfo))
		}
	}

	return strings.Join(lines, "\n")
}

// renderTreeItem renders a single tree item
func (m Model) renderTreeItem(node *common.ResourceNode, selected bool, maxWidth int) string {
	// Status icon
	var icon string
	var iconStyle lipgloss.Style

	switch node.State {
	case common.AssetStatePending:
		icon = SymbolPending
		iconStyle = lipgloss.NewStyle().Foreground(common.ColorMuted)
	case common.AssetStateDiscovery:
		icon = "◐"
		iconStyle = lipgloss.NewStyle().Foreground(common.ColorWarning)
	case common.AssetStateScanning:
		icon = "◐"
		iconStyle = lipgloss.NewStyle().Foreground(common.ColorWarning)
	case common.AssetStateComplete:
		icon = "●"
		iconStyle = lipgloss.NewStyle().Foreground(common.ColorSuccess)
	case common.AssetStateError:
		icon = SymbolCross
		iconStyle = lipgloss.NewStyle().Foreground(common.ColorError)
	}

	// Resource name - prefer actual name over type for instances
	name := node.Name
	if name == "" || name == node.ResourceType {
		// For resource type nodes, show the type
		name = node.ResourceType
	}

	// For resource instances and sub-resource instances, try to get a better display name
	if node.Type == "resource_instance" || node.Type == "sub_resource_instance" {
		// Try to get name from metadata
		if displayName, ok := node.Metadata["name"].(string); ok && displayName != "" {
			name = displayName
		} else if fullName, ok := node.Metadata["full_name"].(string); ok && fullName != "" {
			name = fullName
		} else if login, ok := node.Metadata["login"].(string); ok && login != "" {
			name = login
		}
	}

	// Truncate if needed
	maxNameLen := maxWidth - 18
	if len(name) > maxNameLen && maxNameLen > 3 {
		name = name[:maxNameLen-3] + "..."
	}

	// Count (only for resource type nodes, not instances)
	countStr := ""
	if node.Type == "resource_type" && node.ResourceCount > 0 {
		countStr = fmt.Sprintf("(%d)", node.ResourceCount)
	}

	// Check summary
	checkStr := ""
	total := node.ChecksPassed + node.ChecksFailed + node.ChecksSkipped
	if total > 0 {
		checkStr = fmt.Sprintf(" %d✓ %d✗", node.ChecksPassed, node.ChecksFailed)
	}

	// Build line
	line := fmt.Sprintf("%s %s %s%s", iconStyle.Render(icon), name, countStr, checkStr)

	// Apply selection style
	if selected {
		return m.styles.TreeItemSelected.Width(maxWidth).Render(line)
	}

	return m.styles.TreeItem.Render(line)
}

// renderCheckPanel renders the right panel with check results
func (m Model) renderCheckPanel(width, height int) string {
	lines := make([]string, 0, height)

	// Panel title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.ColorPrimary).
		MarginBottom(1)
	lines = append(lines, titleStyle.Render("🔍 Checks"), "")

	// Get selected node
	children := m.tree.GetCurrentChildren()
	if len(children) == 0 || m.cursor >= len(children) {
		lines = append(lines, common.MutedStyle.Render("Select a resource to view checks"))
		return strings.Join(lines, "\n")
	}

	selectedNode := children[m.cursor]

	// Show node info
	infoStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
	lines = append(lines, infoStyle.Render(fmt.Sprintf("Resource: %s", selectedNode.Name)))

	if selectedNode.State == common.AssetStateScanning {
		lines = append(lines, "", m.styles.StatusPending.Render(m.spinner.View()+" Scanning..."))
		return strings.Join(lines, "\n")
	}

	// Get checks for this node (either direct or from children)
	checks := m.getChecksForNode(selectedNode)

	if len(checks) == 0 {
		lines = append(lines, "")
		if selectedNode.State == common.AssetStatePending {
			lines = append(lines, common.MutedStyle.Render("Waiting to scan..."))
		} else {
			lines = append(lines, common.MutedStyle.Render("No checks for this resource"))
		}
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "")

	// Summary
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

	summaryLine := fmt.Sprintf("Total: %d  ", len(checks))
	summaryLine += m.styles.CheckPassed.Render(fmt.Sprintf("✓ %d", passed)) + "  "
	summaryLine += m.styles.CheckFailed.Render(fmt.Sprintf("✗ %d", failed)) + "  "
	summaryLine += m.styles.CheckSkipped.Render(fmt.Sprintf("⊘ %d", skipped))
	lines = append(lines, summaryLine, strings.Repeat("─", width-2), "")

	// Check list with scroll indicator
	visibleChecks := height - 10
	if visibleChecks < 1 {
		visibleChecks = 1
	}
	startIdx := 0
	if m.checkCursor >= visibleChecks {
		startIdx = m.checkCursor - visibleChecks + 1
	}

	// Calculate end index for display
	endIdx := startIdx + visibleChecks
	if endIdx > len(checks) {
		endIdx = len(checks)
	}

	// Show scroll up indicator if not at top
	if startIdx > 0 {
		scrollUpHint := common.MutedStyle.Render(fmt.Sprintf("  ↑ %d more above", startIdx))
		lines = append(lines, scrollUpHint)
	}

	for i, check := range checks {
		if i < startIdx {
			continue
		}
		if i-startIdx >= visibleChecks {
			break
		}

		line := m.renderCheckItem(check, i == m.checkCursor && m.focusedPanel == PanelRight, width-4)
		lines = append(lines, line)
	}

	// Show scroll down indicator if more items below
	remainingBelow := len(checks) - endIdx
	if remainingBelow > 0 {
		scrollDownHint := common.MutedStyle.Render(fmt.Sprintf("  ↓ %d more below", remainingBelow))
		lines = append(lines, scrollDownHint)
	}

	// Show position indicator if there are more checks than visible
	if len(checks) > visibleChecks {
		positionHint := common.MutedStyle.Render(fmt.Sprintf("\n  Showing %d-%d of %d (use ↑↓ to scroll)", startIdx+1, endIdx, len(checks)))
		lines = append(lines, positionHint)
	}

	return strings.Join(lines, "\n")
}

// renderCheckItem renders a single check item
func (m Model) renderCheckItem(check common.CheckResult, selected bool, maxWidth int) string {
	// Status icon and style
	var icon string
	var statusStyle lipgloss.Style

	switch check.Status {
	case StatusPassed:
		icon = SymbolCheck
		statusStyle = m.styles.CheckPassed
	case StatusFailed:
		icon = SymbolCross
		statusStyle = m.styles.CheckFailed
	case StatusSkipped:
		icon = SymbolSkipped
		statusStyle = m.styles.CheckSkipped
	default:
		icon = "?"
		statusStyle = common.MutedStyle
	}

	// Severity badge
	severityStyle := lipgloss.NewStyle()
	switch check.Severity {
	case "critical":
		severityStyle = severityStyle.Foreground(common.ColorError).Bold(true)
	case "high":
		severityStyle = severityStyle.Foreground(common.ColorError)
	case "medium":
		severityStyle = severityStyle.Foreground(common.ColorWarning)
	case "low":
		severityStyle = severityStyle.Foreground(common.ColorMuted)
	}

	// Truncate name if needed
	name := check.Name
	maxNameLen := maxWidth - 20
	if len(name) > maxNameLen && maxNameLen > 3 {
		name = name[:maxNameLen-3] + "..."
	}

	// Build line
	line := fmt.Sprintf("%s %s [%s]",
		statusStyle.Render(icon),
		name,
		severityStyle.Render(check.Severity),
	)

	if selected {
		selectStyle := lipgloss.NewStyle().
			Background(common.ColorPrimary).
			Foreground(lipgloss.Color("#fff"))
		return selectStyle.Width(maxWidth).Render(line)
	}

	return line
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar(width int) string {
	// Left side: check summary
	root := m.tree.Root
	passed := root.ChecksPassed
	failed := root.ChecksFailed
	skipped := root.ChecksSkipped
	total := passed + failed + skipped

	leftParts := []string{}

	if total > 0 {
		leftParts = append(leftParts,
			fmt.Sprintf("Checks: %d", total),
			m.styles.StatusSuccess.Render(fmt.Sprintf("✓ %d", passed)),
			m.styles.StatusError.Render(fmt.Sprintf("✗ %d", failed)),
		)
		if skipped > 0 {
			leftParts = append(leftParts, common.MutedStyle.Render(fmt.Sprintf("⊘ %d", skipped)))
		}
	}

	leftContent := strings.Join(leftParts, "  ")

	// Right side: keyboard shortcuts
	helpStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
	shortcuts := []string{
		"↑↓ navigate",
		"tab switch panel",
		"enter drill down",
		"esc back",
		"q quit",
	}
	rightContent := helpStyle.Render(strings.Join(shortcuts, " • "))

	// Calculate padding
	leftLen := lipgloss.Width(leftContent)
	rightLen := lipgloss.Width(rightContent)
	padding := width - leftLen - rightLen - 4
	if padding < 1 {
		padding = 1
	}

	statusLine := leftContent + strings.Repeat(" ", padding) + rightContent

	return m.styles.StatusBar.Width(width).Render(statusLine)
}

// getChecksForNode returns checks for a node, aggregating from children and sub-resources if needed
func (m Model) getChecksForNode(node *common.ResourceNode) []common.CheckResult {
	// If node has direct checks, return them
	if len(node.Checks) > 0 {
		return node.Checks
	}

	// Otherwise aggregate from children
	var checks []common.CheckResult
	for _, child := range node.Children {
		checks = append(checks, child.Checks...)
	}

	// Also aggregate from sub-resources
	for _, subResources := range node.SubResources {
		for _, subRes := range subResources {
			checks = append(checks, subRes.Checks...)
		}
	}

	return checks
}
