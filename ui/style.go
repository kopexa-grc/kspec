package ui

import "github.com/charmbracelet/lipgloss"

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("240")).
			Padding(0, 1)

	PassedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")) // Green

	FailedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("160")) // Red

	PendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")) // Grey

	RunningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")) // Yellow

	SkippedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")) // Grey

	CheckNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	GroupStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")) // Blue
)
