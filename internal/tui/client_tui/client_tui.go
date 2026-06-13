package client_tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	km := DefaultKeyMap()

	text_area := textarea.New()
	text_area.Placeholder = "message…  enter to send  shift+enter for newline"
	text_area.ShowLineNumbers = false
	text_area.Prompt = ""
	text_area.SetHeight(1)
	text_area.DynamicHeight = true
	text_area.MinHeight = 1
	text_area.KeyMap.InsertNewline = km.NewLine
	text_area.Focus()

	return ClientModel{
		client:  c,
		ctx:     ctx,
		members: c.Members(),
		input:   text_area,
		histIdx: -1,
		keyMap:  km,
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

// buildMessages renders all message bubbles and returns the joined content
// alongside the content-line offset of each bubble's top border.
func (m ClientModel) buildMessages() (string, []int) {
	w := m.viewport.Width()
	parts := make([]string, len(m.messages))
	offsets := make([]int, len(m.messages))
	line := 0
	for i, msg := range m.messages {
		offsets[i] = line
		rendered := message_bubble.Render(msg, w)
		parts[i] = rendered
		line += lipgloss.Height(rendered)
	}
	return strings.Join(parts, "\n"), offsets
}

func (m ClientModel) renderMessages() string {
	content, _ := m.buildMessages()
	return content
}

func addMessage(m ClientModel, sender, body, ts string, own bool, kind message_bubble.MsgKind) ClientModel {
	m.messages = append(m.messages, message_bubble.Message{
		Sender: sender, Body: body, Ts: ts, Own: own, Kind: kind,
	})
	if m.ready {
		content, offsets := m.buildMessages()
		m.msgLines = offsets
		m.viewport.SetContent(content)
		m.viewport.GotoBottom()
	}
	return m
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func shortID(id uuid.UUID) string {
	return strings.ReplaceAll(id.String(), "-", "")[:6]
}

// memberTag returns the canonical "name#shortid" string used in message
// senders, system notices, and command completions.
func memberTag(name string, id uuid.UUID) string {
	return name + "#" + shortID(id)
}

// memberDisplay returns a styled version of memberTag where the #shortid
// suffix is rendered with shortIDStyle. Use this as the Sender field in
// message bubbles so the ID is visually de-emphasised but still present.
// It also handles the local client's own UUID.
func (m ClientModel) memberDisplay(id uuid.UUID) string {
	if id == m.client.Self {
		return m.client.Name + shortIDStyle.Render("#"+shortID(id))
	}
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
		chatHeight := max(msg.Height-m.input.Height()-m.overlayHeight()-3, 1)
		if !m.ready {
			m.viewport = viewport.New(
				viewport.WithWidth(chatWidth),
				viewport.WithHeight(chatHeight),
			)
			m.viewport.MouseWheelEnabled = true
			m.ready = true
		} else {
			m.viewport.SetWidth(chatWidth)
			m.viewport.SetHeight(chatHeight)
			content, offsets := m.buildMessages()
			m.msgLines = offsets
			m.viewport.SetContent(content)
		}
		if m.filePicker != nil {
			m.filePicker.SetHeight(min(10, max(msg.Height/3, 4)))
		}

	case lolclient.Event:
		ts := time.Now().Format("15:04")
		switch msg.Type {
		case protocol.BroadcastMessage:
			p := msg.Payload.(protocol.BroadcastPayload)
			m = addMessage(m, m.memberDisplay(p.From), ansi.Strip(p.Body), ts, false, message_bubble.KindBroadcast)
		case protocol.WhisperMessage:
			p := msg.Payload.(protocol.WhisperPayload)
			m = addMessage(m, "[DM] "+m.memberDisplay(p.From), ansi.Strip(p.Body), ts, false, message_bubble.KindDM)
		case protocol.JoinMessage:
			p := msg.Payload.(protocol.JoinPayload)
			p.Name = ansi.Strip(p.Name)
			m.members = append(m.members, protocol.Member{Name: p.Name, ID: p.Id})
			m = addMessage(m, "System", "→ "+memberTag(p.Name, p.Id)+" joined", ts, false, message_bubble.KindSystem)
		case protocol.LeaveMessage:
			p := msg.Payload.(protocol.LeavePayload)
			tag := p.From.String()[:8]
			for i, mem := range m.members {
				if mem.ID == p.From {
					tag = memberTag(mem.Name, mem.ID)
					m.members = append(m.members[:i], m.members[i+1:]...)
					break
				}
			}
			m = addMessage(m, "System", "← "+tag+" left", ts, false, message_bubble.KindSystem)
		case protocol.ErrorMessage:
			p := msg.Payload.(protocol.ErrorPayload)
			m = addMessage(m, "System", fmt.Sprintf("[error] %v", p.Body), ts, false, message_bubble.KindSystem)
		case protocol.PongMessage:
			m = addMessage(m, "System", "server responded: Pong", ts, false, message_bubble.KindSystem)
		case protocol.FileShareMessage:
			p := msg.Payload.(protocol.FileSharePayload)
			body := fmt.Sprintf("**%s** (%s) — type `/save %s` to download", p.Name, humanBytes(p.Size), p.Name)
			m = addMessage(m, m.memberDisplay(p.From), body, ts, false, message_bubble.KindFile)
		}
		return m, waitForClientEvent(m.client.Events)

	case connClosedMsg:
		return m, tea.Quit

	case tea.PasteMsg:
		if m.filePicker == nil {
			content := strings.TrimSpace(msg.Content)
			if !strings.ContainsAny(content, "\n\r") {
				// Strip surrounding quotes added by some terminals on drag-and-drop.
				if len(content) >= 2 && content[0] == '"' && content[len(content)-1] == '"' {
					content = content[1 : len(content)-1]
				}
				if info, err := os.Stat(content); err == nil && !info.IsDir() {
					m.input.SetValue("/upload " + content)
					m.input.MoveToEnd()
					return m, nil
				}
			}
		}

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft &&
			msg.Y < m.viewport.Height() &&
			msg.X < m.viewport.Width() {
			contentLine := m.viewport.YOffset() + msg.Y
			for i, startLine := range m.msgLines {
				if contentLine == startLine+1 { // +1: top border is line 0, header is line 1
					return m, tea.SetClipboard(m.messages[i].Body)
				}
			}
		}

	case tea.KeyPressMsg:
		if m.filePicker != nil {
			break // let the filepicker routing block below handle all keys
		}
		km := m.keyMap
		if m.scrollMode {
			switch {
			case key.Matches(msg, km.QuitScroll):
				return m, tea.Quit
			case key.Matches(msg, km.ToggleScroll):
				m.scrollMode = false
				m.input.Focus()
				return m, nil
			}
			m.viewport, viewCmd = m.viewport.Update(msg)
			return m, viewCmd
		}

		switch {
		case key.Matches(msg, km.ToggleScroll):
			// Arg completions take priority over scroll toggle.
			if suggs := argCompletions(m, m.input.Value()); len(suggs) > 0 {
				m.input.SetValue(applyCompletion(m.input.Value(), suggs[0]))
				m.input.MoveToEnd()
				return m, nil
			}
			// Single command match → complete the command name.
			if cmds := autocompleteSuggestions(m.input.Value()); len(cmds) == 1 {
				m.input.SetValue("/" + commandName(cmds[0].Usage) + " ")
				m.input.MoveToEnd()
				return m, nil
			}
			m.scrollMode = true
			m.input.Blur()
			return m, nil

		case key.Matches(msg, km.Send):
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
			m = addMessage(m, m.memberDisplay(m.client.Self), text, ts, true, message_bubble.KindBroadcast)

		case key.Matches(msg, km.HistoryPrev):
			if m.input.Line() == 0 {
				m.historyPrev()
				return m, nil
			}
			m.input, inputCmd = m.input.Update(msg)
			return m, inputCmd

		case key.Matches(msg, km.HistoryNext):
			if m.histIdx >= 0 && m.input.Line() == m.input.LineCount()-1 {
				m.historyNext()
				return m, nil
			}
			m.input, inputCmd = m.input.Update(msg)
			return m, inputCmd
		}
	}

	// Route remaining messages to the file picker when it is active.
	if m.filePicker != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			m.filePicker = nil
			m.input.Focus()
			return m, nil
		}
		newFP, fpCmd := m.filePicker.Update(msg)
		m.filePicker = &newFP
		if didSelect, path := newFP.DidSelectFile(msg); didSelect {
			m.filePicker = nil
			m.input.Focus()
			ts := time.Now().Format("15:04")
			if err := m.client.SendFile(m.ctx, path); err != nil {
				m = addMessage(m, "System", "upload failed: "+err.Error(), ts, false, message_bubble.KindSystem)
			} else {
				m = addMessage(m, "System", "sent "+filepath.Base(path), ts, false, message_bubble.KindSystem)
			}
		}
		return m, fpCmd
	}

	m.viewport, viewCmd = m.viewport.Update(msg)
	m.input, inputCmd = m.input.Update(msg)

	if m.ready {
		if newH := max(m.height-m.input.Height()-m.overlayHeight()-3, 1); newH != m.viewport.Height() {
			m.viewport.SetHeight(newH)
			content, offsets := m.buildMessages()
			m.msgLines = offsets
			m.viewport.SetContent(content)
			m.viewport.GotoBottom()
		}
	}

	return m, tea.Batch(viewCmd, inputCmd)
}

func (m ClientModel) overlayHeight() int {
	if m.filePicker != nil {
		return m.filePicker.Height() + 2 // +2 for border
	}
	input := m.input.Value()
	if args := argCompletions(m, input); len(args) > 0 {
		return min(len(args), 5) + 2
	}
	if cmds := autocompleteSuggestions(input); len(cmds) > 0 {
		return len(cmds) + 2
	}
	return 0
}

func (m ClientModel) renderAutocomplete() string {
	input := m.input.Value()

	// Argument completions (shown when in the args phase of a command).
	if args := argCompletions(m, input); len(args) > 0 {
		limit := min(len(args), 5)
		lines := make([]string, limit)
		for i := range limit {
			lines[i] = "  " + overlayUsageStyle.Render(args[i])
		}
		return overlayStyle.Width(m.width).Render(strings.Join(lines, "\n"))
	}

	// Command name completions (shown while typing /command).
	suggestions := autocompleteSuggestions(input)
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
		Height(m.height - m.input.Height() - m.overlayHeight() - 3).
		Render(strings.Join(lines, "\n"))
}

func (m ClientModel) renderHelpBar() string {
	var hint string
	switch {
	case m.filePicker != nil:
		hint = "↑/↓: navigate · enter: select · esc: cancel"
	case m.scrollMode:
		hint = "tab: input mode · ↑/↓: scroll · esc/q/ctrl+c: quit"
	default:
		hint = "tab: complete/scroll · enter: send · shift+enter: newline · ↑/↓: history · ctrl+c: copy · ctrl+v: paste"
	}
	return helpBarStyle.Width(m.width).Render(hint)
}

func (m ClientModel) View() tea.View {
	var view tea.View
	view.AltScreen = true
	view.BackgroundColor = currentBackground
	view.MouseMode = tea.MouseModeCellMotion

	if !m.ready {
		view.SetContent("Connecting...")
		return view
	}

	top := lipgloss.JoinHorizontal(lipgloss.Top,
		m.viewport.View(),
		m.renderMembers(),
	)
	help := m.renderHelpBar()

	if m.filePicker != nil {
		bottom := inputBarStyle.Width(m.width).Render(m.filePicker.View())
		view.SetContent(lipgloss.JoinVertical(lipgloss.Left, top, bottom, help))
		return view
	}

	bar := inputBarStyle
	if m.scrollMode {
		bar = inputBarBlurredStyle
	}
	bottom := bar.Width(m.width).Render(m.renderInput())

	overlay := m.renderAutocomplete()
	if overlay != "" {
		view.SetContent(lipgloss.JoinVertical(lipgloss.Left, top, overlay, bottom, help))
	} else {
		view.SetContent(lipgloss.JoinVertical(lipgloss.Left, top, bottom, help))
	}
	return view
}
