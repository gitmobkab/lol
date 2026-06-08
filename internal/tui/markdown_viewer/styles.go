package markdown_viewer

import (
	"charm.land/lipgloss/v2"
)

var CenteredStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#535151")).
						Align(lipgloss.Center, lipgloss.Center)