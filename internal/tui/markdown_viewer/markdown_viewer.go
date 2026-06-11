package markdown_viewer

import (
    "charm.land/bubbles/v2/viewport"
    "charm.land/bubbles/v2/help"
    "charm.land/bubbles/v2/key"
    tea "charm.land/bubbletea/v2"
    "charm.land/glamour/v2"
    "charm.land/lipgloss/v2"

)



func NewMarkdownViewer(content string) MarkdownViewerModel {
    rendered_content, _ := glamour.Render(content, "dark")
    return MarkdownViewerModel{
        content: rendered_content,
        keys: keys,
        help: help.New(),
    }
}

func (m MarkdownViewerModel) Init() tea.Cmd {
    return nil
}

func (m MarkdownViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        width := msg.Width - 20
        height := msg.Height - 10
        if !m.ready {
            m.help.SetWidth(msg.Width)
            m.viewport = viewport.New(
                viewport.WithWidth(width),
                viewport.WithHeight(height),
            )
            m.viewport.SetContent(m.content)
            m.ready = true
        } else {
            m.viewport.SetWidth(width)
            m.viewport.SetHeight(height)
        }

    case tea.KeyPressMsg:
        switch {
        
        case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        }
    }

    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m MarkdownViewerModel) View() tea.View {
    var view tea.View
    view.AltScreen = true
    view.MouseMode = tea.MouseModeAllMotion
    view.BackgroundColor = lipgloss.Black
    if !m.ready {
        loading := CenteredStyle.
            Width(m.viewport.Width()).
            Height(m.viewport.Height()).
            Render("Loading...")
        view.SetContent(loading)
        return view
    }

    Viewport := m.viewport.View()
    HelpView := m.help.View(m.keys)

    rendered := lipgloss.JoinVertical(
        lipgloss.Left,
        Viewport,
        HelpView,
    )

    view.SetContent(rendered)

    return view
}