package server_tui

import (
	"charm.land/lipgloss/v2"
)

var (
	serverStatusStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#383838")).Foreground(lipgloss.Color("#ffffff"))
	connectedLineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80"))
	disconnectedLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ac1e1e"))
)