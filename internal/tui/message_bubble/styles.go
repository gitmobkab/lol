package message_bubble

import "charm.land/lipgloss/v2"

var (
	headerOtherStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2c2c2c")).
				Foreground(lipgloss.Color("#aaaaaa")).
				PaddingLeft(1)
	headerOwnStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1e1435")).
				Foreground(lipgloss.Color("#a78bfa")).
				PaddingLeft(1)
	headerSystemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1c1c1c")).
				Foreground(lipgloss.Color("#666666")).
				PaddingLeft(1)
	headerDMStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1e0e2e")).
				Foreground(lipgloss.Color("#a855f7")).
				PaddingLeft(1)

	bubbleBodyStyle = lipgloss.NewStyle().
				PaddingLeft(1).PaddingRight(1)
	bubbleTimeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#444444")).
				PaddingRight(1)

	bubbleBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#333333"))
	bubbleBorderOwnStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#6d28d9"))
	bubbleBorderDMStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#7e22ce"))
)
