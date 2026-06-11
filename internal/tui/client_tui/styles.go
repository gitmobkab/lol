package client_tui

import "charm.land/lipgloss/v2"

var (
	membersPanelStyle = lipgloss.NewStyle().
				BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder()).
				PaddingLeft(1)
	inputBarStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder())
	inputBarBlurredStyle = lipgloss.NewStyle().
				BorderTop(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#333333"))
	shortIDStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	overlayStyle     = lipgloss.NewStyle().BorderTop(true).BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#444444"))
	overlayUsageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa"))

	// Markdown syntax highlight styles applied to the input view.
	mdBoldStyle    = lipgloss.NewStyle().Bold(true)
	mdItalicStyle  = lipgloss.NewStyle().Italic(true)
	mdCodeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f97316"))
	mdHeadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa")).Bold(true)
)
