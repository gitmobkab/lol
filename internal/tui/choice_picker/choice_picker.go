package choice_picker

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/gitmobkab/lol/internal/utils"
)

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Accept  key.Binding
	Quit  key.Binding
}


var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),

	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Accept: key.NewBinding(
		key.WithKeys("enter", "e"),
		key.WithHelp("<ENTER>", "select the choice"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type PickerModel struct {
	choices []string
	cursor  int
	Chosen  string
	Keys keyMap
}

func NewPickerModel(choices []string) PickerModel {
	return PickerModel{
		choices: choices, 
		cursor: 0,
		Keys: keys,
	}
}

func (picker PickerModel) Init() tea.Cmd { 
	return nil 
}

func (picker PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, picker.Keys.Quit):
			return picker, tea.Quit

		case key.Matches(msg, picker.Keys.Up):
			picker.cursor--

		case key.Matches(msg, picker.Keys.Down):
			picker.cursor++

		case key.Matches(msg, picker.Keys.Accept):
			picker.Chosen = picker.choices[picker.cursor]
			return picker, tea.Quit
		}
		
		last_index := len(picker.choices) - 1
		picker.cursor = utils.Warp(picker.cursor, 0, last_index)

	}
	return picker, nil
}

func (picker PickerModel) View() tea.View {
	var view tea.View
	view.AltScreen = true

	var sb strings.Builder
	sb.WriteString("Multiple network interfaces found. Select one:\n\n")
	for i, ip := range picker.choices {
		var out string
		if i == picker.cursor {
			out = pickerCursorStyle.Render("> ") + pickerSelectedStyle.Render(ip) + "\n"
		} else {
			out = "  " + pickerNormalStyle.Render(ip) + "\n"
		}
		sb.WriteString(out)
	}
	hint := "\n" + pickerHintStyle.Render("↑/↓  navigate   enter  select   q  quit")
	sb.WriteString(hint)

	view.SetContent(sb.String())
	return view
}
