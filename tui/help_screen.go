package tui

import (
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type HelpModel struct {
	viewport viewport.Model
	content  string
	ready    bool
}


func NewHelpModel(content string) HelpModel {
    return HelpModel{content: content}
}

func (m HelpModel) Init() tea.Cmd {
    return nil
}

func (m HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        if !m.ready {
            m.viewport = viewport.New(msg.Width, msg.Height)
            m.viewport.SetContent(m.content)
            m.ready = true
        } else {
            m.viewport.Width = msg.Width
            m.viewport.Height = msg.Height
        }

    case tea.KeyMsg:
        switch msg.String() {
        case "q", "esc", "ctrl+c":
            return m, tea.Quit
        }
    }

    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m HelpModel) View() string {
    if !m.ready {
        return ""
    }
    return lipgloss.JoinVertical(
        lipgloss.Left,
        m.viewport.View(),
        lipgloss.NewStyle().
            Foreground(lipgloss.Color("241")).
            Render("↑/↓ scroll - q quit"),
    )
}