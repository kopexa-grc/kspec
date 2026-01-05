package common

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSuccess   = lipgloss.Color("#04B575")
	ColorWarning   = lipgloss.Color("#FFAA00")
	ColorError     = lipgloss.Color("#FF0000")
	ColorMuted     = lipgloss.Color("#626262")
	ColorHighlight = lipgloss.Color("#F780E2")

	// Base styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	// Table styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorMuted)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	CellStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)
)

// StatusStyle returns a styled status string
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "passed", "complete":
		return SuccessStyle
	case "failed", "error":
		return ErrorStyle
	case "pending", "scanning":
		return WarningStyle
	default:
		return MutedStyle
	}
}
