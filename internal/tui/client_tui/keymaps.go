package client_tui

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Quit        key.Binding
	Send        key.Binding
	NewLine     key.Binding
	HistoryPrev key.Binding
	HistoryNext key.Binding
	ScrollMode  key.Binding
	FocusInput  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:        key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "quit")),
		Send:        key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send message")),
		NewLine:     key.NewBinding(key.WithKeys("shift+enter"), key.WithHelp("shift+enter", "new line")),
		HistoryPrev: key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "previous message")),
		HistoryNext: key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "next message")),
		ScrollMode:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "scroll messages")),
		FocusInput:  key.NewBinding(key.WithKeys("esc", "tab"), key.WithHelp("esc/tab", "back to input")),
	}
}
