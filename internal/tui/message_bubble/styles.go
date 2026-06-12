package message_bubble

import (
	"charm.land/lipgloss/v2"
	"github.com/gitmobkab/lol/internal/tui/theme"
)

var currentGlamourStyle = "dark"

var (
	headerOtherStyle  lipgloss.Style
	headerOwnStyle    lipgloss.Style
	headerSystemStyle lipgloss.Style
	headerDMStyle     lipgloss.Style

	bubbleBodyStyle lipgloss.Style
	bubbleTimeStyle lipgloss.Style

	bubbleBorderStyle    lipgloss.Style
	bubbleBorderOwnStyle lipgloss.Style
	bubbleBorderDMStyle  lipgloss.Style
)

func init() {
	SetTheme(theme.Dark)
}

func SetTheme(t theme.Theme) {
	currentGlamourStyle = t.GlamourStyle

	headerOtherStyle = lipgloss.NewStyle().
		Background(t.HeaderOtherBg).
		Foreground(t.HeaderOtherFg).
		PaddingLeft(1)
	headerOwnStyle = lipgloss.NewStyle().
		Background(t.HeaderOwnBg).
		Foreground(t.HeaderOwnFg).
		PaddingLeft(1)
	headerSystemStyle = lipgloss.NewStyle().
		Background(t.HeaderSystemBg).
		Foreground(t.HeaderSystemFg).
		PaddingLeft(1)
	headerDMStyle = lipgloss.NewStyle().
		Background(t.HeaderDMBg).
		Foreground(t.HeaderDMFg).
		PaddingLeft(1)

	bubbleBodyStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	bubbleTimeStyle = lipgloss.NewStyle().Foreground(t.TimeColor).PaddingRight(1)

	bubbleBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.BorderOther)
	bubbleBorderOwnStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.BorderOwn)
	bubbleBorderDMStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.BorderDM)
}
