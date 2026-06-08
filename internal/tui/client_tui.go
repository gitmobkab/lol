package tui

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

const membersWidth = 22

var (
	membersPanelStyle = lipgloss.NewStyle().
				BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder()).
				PaddingLeft(1)
	inputBarStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder())
	dmStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#a855f7"))
	systemStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
	shortIDStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
)

type connClosedMsg struct{}

func waitForClientEvent(events <-chan lolclient.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return connClosedMsg{}
		}
		return event
	}
}

type ClientModel struct {
	viewport viewport.Model
	input    textinput.Model
	messages []string
	members  []protocol.Member
	client   *lolclient.Client
	ctx      context.Context
	width    int
	height   int
	ready    bool
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

func addMessage(m ClientModel, line string) ClientModel {
	m.messages = append(m.messages, line)
	if m.ready {
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
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
		}

	case lolclient.Event:
		ts := time.Now().Format("15:04")
		switch msg.Type {
		case protocol.BroadcastMessage:
			p := msg.Payload.(protocol.BroadcastPayload)
			m = addMessage(m, fmt.Sprintf("%s  %s: %s", ts, m.memberName(p.From), ansi.Strip(p.Body)))
		case protocol.WhisperMessage:
			p := msg.Payload.(protocol.WhisperPayload)
			m = addMessage(m, dmStyle.Render(fmt.Sprintf("%s  [DM] %s: %s", ts, m.memberName(p.From), ansi.Strip(p.Body))))
		case protocol.JoinMessage:
			p := msg.Payload.(protocol.JoinPayload)
			p.Name = ansi.Strip(p.Name)
			m.members = append(m.members, protocol.Member{Name: p.Name, ID: p.Id})
			m = addMessage(m, systemStyle.Render(fmt.Sprintf("%s  → %s joined", ts, p.Name)))
		case protocol.LeaveMessage:
			p := msg.Payload.(protocol.LeavePayload)
			name := m.memberName(p.From)
			for i, mem := range m.members {
				if mem.ID == p.From {
					m.members = append(m.members[:i], m.members[i+1:]...)
					break
				}
			}
			m = addMessage(m, systemStyle.Render(fmt.Sprintf("%s  ← %s left", ts, name)))
		case protocol.ErrorMessage:
			p := msg.Payload.(protocol.ErrorPayload)
			m = addMessage(m, systemStyle.Render(fmt.Sprintf("%s  [error] %v", ts, p.Body)))
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

					ts := time.Now().Format("15:04")
					switch len(matches) {
					case 0:
						m = addMessage(m, systemStyle.Render(fmt.Sprintf("%s  user %q not found", ts, rawTarget)))
					case 1:
						m.client.SendDM(m.ctx, matches[0].ID, body)
						m = addMessage(m, dmStyle.Render(fmt.Sprintf("%s  [DM → %s]: %s", ts, matches[0].Name, body)))
					default:
						var ids []string
						for _, mem := range matches {
							ids = append(ids, mem.Name+"#"+shortID(mem.ID))
						}
						m = addMessage(m, systemStyle.Render(fmt.Sprintf("%s  ambiguous — be specific: %s", ts, strings.Join(ids, ", "))))
					}
				}
			} else {
				m.client.SendChat(m.ctx, text)
				ts := time.Now().Format("15:04")
				m = addMessage(m, fmt.Sprintf("%s  %s: %s", ts, m.client.Name, text))
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
