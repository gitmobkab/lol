package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	Name         string
	GlamourStyle string

	HeaderOtherBg  color.Color
	HeaderOtherFg  color.Color
	HeaderOwnBg    color.Color
	HeaderOwnFg    color.Color
	HeaderSystemBg color.Color
	HeaderSystemFg color.Color
	HeaderDMBg     color.Color
	HeaderDMFg     color.Color

	BorderOther color.Color
	BorderOwn   color.Color
	BorderDM    color.Color

	TimeColor     color.Color
	ShortIDColor  color.Color
	OverlayBorder color.Color
	AccentColor   color.Color
	CodeColor     color.Color
}

var (
	Dark = Theme{
		Name:         "dark",
		GlamourStyle: "dark",

		HeaderOtherBg:  lipgloss.Color("#2c2c2c"),
		HeaderOtherFg:  lipgloss.Color("#aaaaaa"),
		HeaderOwnBg:    lipgloss.Color("#1e1435"),
		HeaderOwnFg:    lipgloss.Color("#a78bfa"),
		HeaderSystemBg: lipgloss.Color("#1c1c1c"),
		HeaderSystemFg: lipgloss.Color("#666666"),
		HeaderDMBg:     lipgloss.Color("#1e0e2e"),
		HeaderDMFg:     lipgloss.Color("#a855f7"),

		BorderOther: lipgloss.Color("#333333"),
		BorderOwn:   lipgloss.Color("#6d28d9"),
		BorderDM:    lipgloss.Color("#7e22ce"),

		TimeColor:     lipgloss.Color("#444444"),
		ShortIDColor:  lipgloss.Color("#555555"),
		OverlayBorder: lipgloss.Color("#444444"),
		AccentColor:   lipgloss.Color("#a78bfa"),
		CodeColor:     lipgloss.Color("#f97316"),
	}

	Dracula = Theme{
		Name:         "dracula",
		GlamourStyle: "dracula",

		HeaderOtherBg:  lipgloss.Color("#44475a"),
		HeaderOtherFg:  lipgloss.Color("#f8f8f2"),
		HeaderOwnBg:    lipgloss.Color("#282a36"),
		HeaderOwnFg:    lipgloss.Color("#bd93f9"),
		HeaderSystemBg: lipgloss.Color("#21222c"),
		HeaderSystemFg: lipgloss.Color("#6272a4"),
		HeaderDMBg:     lipgloss.Color("#2d1b2e"),
		HeaderDMFg:     lipgloss.Color("#ff79c6"),

		BorderOther: lipgloss.Color("#6272a4"),
		BorderOwn:   lipgloss.Color("#bd93f9"),
		BorderDM:    lipgloss.Color("#ff79c6"),

		TimeColor:     lipgloss.Color("#6272a4"),
		ShortIDColor:  lipgloss.Color("#6272a4"),
		OverlayBorder: lipgloss.Color("#6272a4"),
		AccentColor:   lipgloss.Color("#bd93f9"),
		CodeColor:     lipgloss.Color("#50fa7b"),
	}

	Nord = Theme{
		Name:         "nord",
		GlamourStyle: "dark",

		HeaderOtherBg:  lipgloss.Color("#3b4252"),
		HeaderOtherFg:  lipgloss.Color("#d8dee9"),
		HeaderOwnBg:    lipgloss.Color("#2e3440"),
		HeaderOwnFg:    lipgloss.Color("#88c0d0"),
		HeaderSystemBg: lipgloss.Color("#2e3440"),
		HeaderSystemFg: lipgloss.Color("#4c566a"),
		HeaderDMBg:     lipgloss.Color("#2e3440"),
		HeaderDMFg:     lipgloss.Color("#81a1c1"),

		BorderOther: lipgloss.Color("#4c566a"),
		BorderOwn:   lipgloss.Color("#88c0d0"),
		BorderDM:    lipgloss.Color("#81a1c1"),

		TimeColor:     lipgloss.Color("#4c566a"),
		ShortIDColor:  lipgloss.Color("#4c566a"),
		OverlayBorder: lipgloss.Color("#4c566a"),
		AccentColor:   lipgloss.Color("#88c0d0"),
		CodeColor:     lipgloss.Color("#a3be8c"),
	}

	All = map[string]Theme{
		Dark.Name:    Dark,
		Dracula.Name: Dracula,
		Nord.Name:    Nord,
	}
)
