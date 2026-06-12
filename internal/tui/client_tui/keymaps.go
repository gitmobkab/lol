package client_tui

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	ToggleScroll key.Binding
	QuitScroll   key.Binding
	Send         key.Binding
	NewLine      key.Binding
	HistoryPrev  key.Binding
	HistoryNext  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		ToggleScroll: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle scroll"),
		),
		QuitScroll: key.NewBinding(
			key.WithKeys("esc", "q", "ctrl+c", "ctrl+q"),
			key.WithHelp("esc/q/ctrl+c", "quit"),
		),
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		NewLine: key.NewBinding(
			key.WithKeys("shift+enter"),
			key.WithHelp("shift+enter", "newline"),
		),
		HistoryPrev: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "prev"),
		),
		HistoryNext: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "next"),
		),
	}
}
