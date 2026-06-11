package client_tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	lolclient "github.com/gitmobkab/lol/internal/client"
	"github.com/gitmobkab/lol/internal/protocol"
	"github.com/google/uuid"
)

const (
	membersWidth    = 50
	bubbleWidthRatio = 55
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
	input := textinput.New()
	input.Placeholder = "message... (/dm <name> <msg> for DMs)"
	input.CharLimit = 500
	input.Focus()

	return ClientModel{
		client:  c,
		ctx:     ctx,
		members: c.Members(),
		input:   input,
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

func renderBubble(msg message, vpWidth int) string {
	maxWidth := max(vpWidth*bubbleWidthRatio/100, 20)
	maxWidth = min(maxWidth, vpWidth-2)
	innerWidth := max(maxWidth-2, 4)

	var headerStyle, borderStyle lipgloss.Style
	switch {
	case msg.kind == kindSystem:
		headerStyle = headerSystemStyle
		borderStyle = bubbleBorderStyle
	case msg.kind == kindDM:
		headerStyle = headerDMStyle
		borderStyle = bubbleBorderDMStyle
	case msg.own:
		headerStyle = headerOwnStyle
		borderStyle = bubbleBorderOwnStyle
	default:
		headerStyle = headerOtherStyle
		borderStyle = bubbleBorderStyle
	}

	header := headerStyle.Width(innerWidth).Render(msg.sender)
	body := bubbleBodyStyle.Width(innerWidth).Render(msg.body)
	timeRow := bubbleTimeStyle.Width(innerWidth).Align(lipgloss.Right).Render(msg.ts)
	inner := lipgloss.JoinVertical(lipgloss.Left, header, body, timeRow)
	bubble := borderStyle.Render(inner)

	bw := lipgloss.Width(bubble)
	switch {
	case msg.own:
		if pad := vpWidth - bw; pad > 0 {
			return indentLines(bubble, pad)
		}
	case msg.kind == kindSystem:
		if pad := (vpWidth - bw) / 2; pad > 0 {
			return indentLines(bubble, pad)
		}
	}
	return bubble
}

func indentLines(s string, n int) string {
	prefix := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

func (m ClientModel) renderMessages() string {
	w := m.viewport.Width()
	parts := make([]string, len(m.messages))
	for i, msg := range m.messages {
		parts[i] = renderBubble(msg, w)
	}
	return strings.Join(parts, "\n")
}

func addMessage(m ClientModel, sender, body, ts string, own bool, kind msgKind) ClientModel {
	m.messages = append(m.messages, message{sender: sender, body: body, ts: ts, own: own, kind: kind})
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

func (m ClientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var viewCmd, inputCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		chatWidth := msg.Width - membersWidth - 1
		chatHeight := msg.Height - 3
		m.input.SetWidth(msg.Width - 2)
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
			m = addMessage(m, m.memberDisplay(p.From), ansi.Strip(p.Body), ts, false, kindBroadcast)
		case protocol.WhisperMessage:
			p := msg.Payload.(protocol.WhisperPayload)
			m = addMessage(m, "[DM] "+m.memberName(p.From), ansi.Strip(p.Body), ts, false, kindDM)
		case protocol.JoinMessage:
			p := msg.Payload.(protocol.JoinPayload)
			p.Name = ansi.Strip(p.Name)
			m.members = append(m.members, protocol.Member{Name: p.Name, ID: p.Id})
			m = addMessage(m, "System", "→ "+p.Name+" joined", ts, false, kindSystem)
		case protocol.LeaveMessage:
			p := msg.Payload.(protocol.LeavePayload)
			name := m.memberName(p.From)
			for i, mem := range m.members {
				if mem.ID == p.From {
					m.members = append(m.members[:i], m.members[i+1:]...)
					break
				}
			}
			m = addMessage(m, "System", "← "+name+" left", ts, false, kindSystem)
		case protocol.ErrorMessage:
			p := msg.Payload.(protocol.ErrorPayload)
			m = addMessage(m, "System", fmt.Sprintf("[error] %v", p.Body), ts, false, kindSystem)
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
			ts := time.Now().Format("15:04")
			if strings.HasPrefix(text, "/dm ") {
				parts := strings.SplitN(text[4:], " ", 2)
				if len(parts) == 2 {
					rawTarget, body := parts[0], parts[1]
					targetName, targetPrefix, _ := strings.Cut(rawTarget, "#")

					var matches []protocol.Member
					for _, mem := range m.members {
						if !strings.EqualFold(mem.Name, targetName) {
							continue
						}
						if targetPrefix != "" && !strings.HasPrefix(shortID(mem.ID), strings.ToLower(targetPrefix)) {
							continue
						}
						matches = append(matches, mem)
					}

					switch len(matches) {
					case 0:
						m = addMessage(m, "System", fmt.Sprintf("user %q not found", rawTarget), ts, false, kindSystem)
					case 1:
						m.client.SendDM(m.ctx, matches[0].ID, body)
						m = addMessage(m, "DM → "+matches[0].Name, body, ts, true, kindDM)
					default:
						var ids []string
						for _, mem := range matches {
							ids = append(ids, mem.Name+"#"+shortID(mem.ID))
						}
						m = addMessage(m, "System", "ambiguous — be specific: "+strings.Join(ids, ", "), ts, false, kindSystem)
					}
				}
			} else {
				m.client.SendChat(m.ctx, text)
				m = addMessage(m, m.client.Name, text, ts, true, kindBroadcast)
			}
		}
	}

	m.viewport, viewCmd = m.viewport.Update(msg)
	m.input, inputCmd = m.input.Update(msg)
	return m, tea.Batch(viewCmd, inputCmd)
}

func (m ClientModel) renderMembers() string {
	lines := make([]string, 0, len(m.members)+1)
	lines = append(lines, "members")
	for _, mem := range m.members {
		lines = append(lines, "• "+mem.Name+" "+shortIDStyle.Render(shortID(mem.ID)))
	}
	return membersPanelStyle.
		Width(membersWidth).
		Height(m.height - 3).
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
	bottom := inputBarStyle.Width(m.width).Render(m.input.View())

	view.SetContent(lipgloss.JoinVertical(lipgloss.Left, top, bottom))
	return view
}
