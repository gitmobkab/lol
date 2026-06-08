package choice_picker

import "charm.land/lipgloss/v2"

var (
	pickerCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7c3aed")).Bold(true)
	pickerSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)
	pickerNormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa"))
	pickerHintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true)
)
