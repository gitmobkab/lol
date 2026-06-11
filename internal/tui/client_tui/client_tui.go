package client_tui

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	lolclient "github.com/gitmobkab/lol/internal/client"
	"github.com/gitmobkab/lol/internal/protocol"
	"github.com/gitmobkab/lol/internal/tui/message_bubble"
	"github.com/google/uuid"
)

const membersWidth = 50

// Regexes applied line-by-line to the textarea view for inline markdown highlighting.
// Processed in order: bold before italic so ** doesn't match as two * patterns.
var (
	reBold    = regexp.MustCompile(`\*\*[^*\n]+\*\*`)
	reItalic  = regexp.MustCompile(`_[^_\n]+_`)
	reCode    = regexp.MustCompile("`[^`\n]+`")
	reHeading = regexp.MustCompile(`^#{1,6} .+`)
)

func waitForClientEvent(events <-chan lolclient.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return connClosedMsg{}
		}
		return event
	}
}

func NewClientModel(c *lolclient.Client, ctx context.Context) ClientModel {
	text_area := textarea.New()
	text_area.Placeholder = "message…  enter to send  ctrl+enter for newline  /dm <name> <msg> for DMs"
	text_area.ShowLineNumbers = false
	text_area.Prompt = ""
	text_area.SetHeight(1)
	text_area.DynamicHeight = true
	text_area.MinHeight = 1
	text_area.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("ctrl+enter"), key.WithHelp("ctrl+enter", "insert newline"))
	text_area.Focus()

	return ClientModel{
		client:  c,
		ctx:     ctx,
		members: c.Members(),
		input:   text_area,
		histIdx: -1,
	}
}

func (m ClientModel) Run() error {
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func (m ClientModel) Init() tea.Cmd {
	return waitForClientEvent(m.client.Events)
}

// highlightLine applies inline markdown syntax coloring to a single rendered
// textarea line. It operates on already-ANSI-encoded text so patterns near the
// cursor may not match — that is acceptable; highlighting is best-effort.
func highlightLine(line string) string {
	line = reHeading.ReplaceAllStringFunc(line, func(s string) string { return mdHeadingStyle.Render(s) })
	line = reBold.ReplaceAllStringFunc(line, func(s string) string { return mdBoldStyle.Render(s) })
	line = reItalic.ReplaceAllStringFunc(line, func(s string) string { return mdItalicStyle.Render(s) })
	line = reCode.ReplaceAllStringFunc(line, func(s string) string { return mdCodeStyle.Render(s) })
	return line
}

func (m ClientModel) renderInput() string {
	raw := m.input.View()
	lines := strings.Split(raw, "\n")
	for i, l := range lines {
		lines[i] = highlightLine(l)
	}
	return strings.Join(lines, "\n")
}

func (m ClientModel) renderMessages() string {
	w := m.viewport.Width()
	parts := make([]string, len(m.messages))
	for i, msg := range m.messages {
		parts[i] = message_bubble.Render(msg, w)
	}
	return strings.Join(parts, "\n")
}

func addMessage(m ClientModel, sender, body, ts string, own bool, kind message_bubble.MsgKind) ClientModel {
	m.messages = append(m.messages, message_bubble.Message{
		Sender: sender, Body: body, Ts: ts, Own: own, Kind: kind,
	})
	if m.ready {
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}
	return m
}

func shortID(id uuid.UUID) string {
	return strings.ReplaceAll(id.String(), "-", "")[:6]
}

func (m ClientModel) memberName(id uuid.UUID) string {
	for _, mem := range m.members {
		if mem.ID == id {
			return mem.Name
		}
	}
	return id.String()[:8]
}

func (m ClientModel) memberDisplay(id uuid.UUID) string {
	for _, mem := range m.members {
		if mem.ID == id {
			return mem.Name + shortIDStyle.Render("#"+shortID(id))
		}
	}
	return id.String()[:8]
}

func (m *ClientModel) historyPrev() {
	if len(m.history) == 0 {
		return
	}
	if m.histIdx == -1 {
		m.draft = m.input.Value()
	}
	m.histIdx = min(m.histIdx+1, len(m.history)-1)
	m.input.SetValue(m.history[len(m.history)-1-m.histIdx])
	m.input.MoveToEnd()
}

func (m *ClientModel) historyNext() {
	if m.histIdx < 0 {
		return
	}
	m.histIdx--
	if m.histIdx == -1 {
		m.input.SetValue(m.draft)
	} else {
		m.input.SetValue(m.history[len(m.history)-1-m.histIdx])
	}
	m.input.MoveToEnd()
}

func (m ClientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var viewCmd, inputCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		chatWidth := msg.Width - membersWidth - 1
		m.input.SetWidth(msg.Width - 2)
		m.input.MaxHeight = max(msg.Height/3, 4)
		chatHeight := max(msg.Height-m.input.Height()-m.overlayHeight()-2, 1)
		if !m.ready {
			m.viewport = viewport.New(
				viewport.WithWidth(chatWidth),
				viewport.WithHeight(chatHeight),
			)
			m.ready = true
		} else {
			m.viewport.SetWidth(chatWidth)
			m.viewport.SetHeight(chatHeight)
			m.viewport.SetContent(m.renderMessages())
		}

	case lolclient.Event:
		ts := time.Now().Format("15:04")
		switch msg.Type {
		case protocol.BroadcastMessage:
			p := msg.Payload.(protocol.BroadcastPayload)
			m = addMessage(m, m.memberDisplay(p.From), ansi.Strip(p.Body), ts, false, message_bubble.KindBroadcast)
		case protocol.WhisperMessage:
			p := msg.Payload.(protocol.WhisperPayload)
			m = addMessage(m, "[DM] "+m.memberName(p.From), ansi.Strip(p.Body), ts, false, message_bubble.KindDM)
		case protocol.JoinMessage:
			p := msg.Payload.(protocol.JoinPayload)
			p.Name = ansi.Strip(p.Name)
			m.members = append(m.members, protocol.Member{Name: p.Name, ID: p.Id})
			m = addMessage(m, "System", "→ "+p.Name+" joined", ts, false, message_bubble.KindSystem)
		case protocol.LeaveMessage:
			p := msg.Payload.(protocol.LeavePayload)
			name := m.memberName(p.From)
			for i, mem := range m.members {
				if mem.ID == p.From {
					m.members = append(m.members[:i], m.members[i+1:]...)
					break
				}
			}
			m = addMessage(m, "System", "← "+name+" left", ts, false, message_bubble.KindSystem)
		case protocol.ErrorMessage:
			p := msg.Payload.(protocol.ErrorPayload)
			m = addMessage(m, "System", fmt.Sprintf("[error] %v", p.Body), ts, false, message_bubble.KindSystem)
		case protocol.PongMessage:
			m = addMessage(m, "System", "server responded: Pong", ts, false, message_bubble.KindSystem)
		}
		return m, waitForClientEvent(m.client.Events)

	case connClosedMsg:
		return m, tea.Quit

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			text := strings.TrimSpace(m.input.Value())
			m.input.Reset()
			if text == "" {
				break
			}
			m.history = append(m.history, text)
			m.histIdx = -1
			m.draft = ""
			if strings.HasPrefix(text, "/") {
				var cmd tea.Cmd
				m, cmd = dispatch(m, text[1:])
				return m, cmd
			}
			ts := time.Now().Format("15:04")
			m.client.SendChat(m.ctx, text)
			m = addMessage(m, m.client.Name, text, ts, true, message_bubble.KindBroadcast)

		case "up":
			if m.input.Line() == 0 {
				m.historyPrev()
				return m, nil
			}

		case "down":
			if m.histIdx >= 0 && m.input.Line() == m.input.LineCount()-1 {
				m.historyNext()
				return m, nil
			}
		}
	}

	m.viewport, viewCmd = m.viewport.Update(msg)
	m.input, inputCmd = m.input.Update(msg)

	if m.ready {
		if newH := max(m.height-m.input.Height()-m.overlayHeight()-2, 1); newH != m.viewport.Height() {
			m.viewport.SetHeight(newH)
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
	}

	return m, tea.Batch(viewCmd, inputCmd)
}

func (m ClientModel) overlayHeight() int {
	s := autocompleteSuggestions(m.input.Value())
	if len(s) == 0 {
		return 0
	}
	return len(s) + 2 // content lines + top/bottom border
}

func (m ClientModel) renderAutocomplete() string {
	suggestions := autocompleteSuggestions(m.input.Value())
	if len(suggestions) == 0 {
		return ""
	}
	lines := make([]string, len(suggestions))
	for i, cmd := range suggestions {
		lines[i] = "  " + overlayUsageStyle.Render("/"+cmd.Usage) + "  " + cmd.Help
	}
	return overlayStyle.Width(m.width).Render(strings.Join(lines, "\n"))
}

func (m ClientModel) renderMembers() string {
	lines := make([]string, 0, len(m.members)+1)
	lines = append(lines, "members")
	for _, mem := range m.members {
		lines = append(lines, "• "+mem.Name+" "+shortIDStyle.Render(shortID(mem.ID)))
	}
	return membersPanelStyle.
		Width(membersWidth).
		Height(m.height - m.input.Height() - m.overlayHeight() - 2).
		Render(strings.Join(lines, "\n"))
}

func (m ClientModel) View() tea.View {
	var view tea.View
	view.AltScreen = true
	view.BackgroundColor = lipgloss.Black

	if !m.ready {
		view.SetContent("Connecting...")
		return view
	}

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		m.viewport.View(),
		m.renderMembers(),
	)
	bottom := inputBarStyle.Width(m.width).Render(m.renderInput())

	overlay := m.renderAutocomplete()
	if overlay != "" {
		view.SetContent(lipgloss.JoinVertical(lipgloss.Left, top, overlay, bottom))
	} else {
		view.SetContent(lipgloss.JoinVertical(lipgloss.Left, top, bottom))
	}
	return view
}
