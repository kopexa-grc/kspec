// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package common

import "github.com/charmbracelet/lipgloss"

// Color palette for the CLI.
var (
	// ColorPrimary is the primary accent color.
	ColorPrimary = lipgloss.Color("#7D56F4")
	// ColorSuccess indicates success states.
	ColorSuccess = lipgloss.Color("#04B575")
	// ColorWarning indicates warning states.
	ColorWarning = lipgloss.Color("#FFAA00")
	// ColorError indicates error states.
	ColorError = lipgloss.Color("#FF0000")
	// ColorMuted is used for de-emphasized text.
	ColorMuted = lipgloss.Color("#626262")
	// ColorHighlight is used for highlighted elements.
	ColorHighlight = lipgloss.Color("#F780E2")
)

// Base text styles.
var (
	// TitleStyle is used for main headings.
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
)

// Table styles.
var (
	// HeaderStyle is used for table headers.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorMuted)

	// SelectedStyle is used for selected items.
	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	// CellStyle is used for table cells.
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
