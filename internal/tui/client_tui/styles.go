package client_tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/gitmobkab/lol/internal/tui/theme"
)

var (
	currentBackground color.Color

	membersPanelStyle    lipgloss.Style
	inputBarStyle        lipgloss.Style
	inputBarBlurredStyle lipgloss.Style
	shortIDStyle         lipgloss.Style
	overlayStyle         lipgloss.Style
	overlayUsageStyle    lipgloss.Style

	mdBoldStyle    lipgloss.Style
	mdItalicStyle  lipgloss.Style
	mdCodeStyle    lipgloss.Style
	mdHeadingStyle lipgloss.Style

	helpBarStyle lipgloss.Style
)

func init() {
	SetTheme(theme.Dark)
}

func SetTheme(t theme.Theme) {
	currentBackground = t.Background

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
		BorderForeground(t.OverlayBorder)
	shortIDStyle = lipgloss.NewStyle().Foreground(t.ShortIDColor)
	overlayStyle = lipgloss.NewStyle().
		BorderTop(true).BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.OverlayBorder)
	overlayUsageStyle = lipgloss.NewStyle().Foreground(t.AccentColor)

	mdBoldStyle = lipgloss.NewStyle().Bold(true)
	mdItalicStyle = lipgloss.NewStyle().Italic(true)
	mdCodeStyle = lipgloss.NewStyle().Foreground(t.CodeColor)
	mdHeadingStyle = lipgloss.NewStyle().Foreground(t.AccentColor).Bold(true)

	helpBarStyle = lipgloss.NewStyle().Foreground(t.ShortIDColor)
}
