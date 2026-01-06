package discovery

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kopexa-grc/kspec/cli/components/common"
)

// Model represents the discovery progress view
type Model struct {
	spinner   spinner.Model
	resources map[string]int
	stage     string
	complete  bool
}

// DiscoveryUpdate is sent when a resource type is discovered
type DiscoveryUpdate struct {
	ResourceType string
	Count        int
}

// DiscoveryComplete is sent when discovery is finished
type DiscoveryComplete struct {
	TotalResources int
	TotalTypes     int
}

// New creates a new discovery model
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(common.ColorPrimary)

	return Model{
		spinner:   s,
		resources: make(map[string]int),
		stage:     "Initializing...",
		complete:  false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DiscoveryUpdate:
		m.resources[msg.ResourceType] = msg.Count
		return m, nil

	case DiscoveryComplete:
		m.complete = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the discovery progress
func (m Model) View() string {
	if m.complete {
		return m.completeView()
	}

	var output strings.Builder

	// Title with spinner
	output.WriteString("\n")
	output.WriteString(m.spinner.View())
	output.WriteString(" ")
	output.WriteString(common.TitleStyle.Render("Discovering Azure Resources"))
	output.WriteString("\n\n")

	// Resource counts
	if len(m.resources) == 0 {
		output.WriteString(common.MutedStyle.Render("  Scanning subscription..."))
		output.WriteString("\n")
	} else {
		for resourceType, count := range m.resources {
			line := fmt.Sprintf("  ✓ %s: %d", resourceType, count)
			output.WriteString(common.SuccessStyle.Render(line))
			output.WriteString("\n")
		}
	}

	output.WriteString("\n")
	output.WriteString(common.HelpStyle.Render("Press q to quit"))
	output.WriteString("\n")

	return output.String()
}

// completeView renders the completion summary
func (m Model) completeView() string {
	var output strings.Builder

	output.WriteString("\n")
	output.WriteString(common.SuccessStyle.Render("✓ Discovery Complete"))
	output.WriteString("\n\n")

	totalResources := 0
	for _, count := range m.resources {
		totalResources += count
	}

	summary := fmt.Sprintf("📊 Total: %d resources across %d types", totalResources, len(m.resources))
	output.WriteString(common.TitleStyle.Render(summary))
	output.WriteString("\n\n")

	for resourceType, count := range m.resources {
		line := fmt.Sprintf("  ✓ %s: %d", resourceType, count)
		output.WriteString(line)
		output.WriteString("\n")
	}

	output.WriteString("\n")

	return output.String()
}
