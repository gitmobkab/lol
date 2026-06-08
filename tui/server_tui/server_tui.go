package server_tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	lolserver "github.com/gitmobkab/lol/server"
)

type serverTickMsg struct{}

func serverTick() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second)
		return serverTickMsg{}
	}
}

func waitForServerEvent(events <-chan lolserver.Event) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-events
		if !ok {
			return nil
		}
		return e
	}
}

type ServerModel struct {
	viewport  viewport.Model
	logs      []string
	events    <-chan lolserver.Event
	clients   int
	startTime time.Time
	width     int
	ready     bool
}

func NewServerModel(events <-chan lolserver.Event, initialLogs []string) ServerModel {
	return ServerModel{
		events:    events,
		startTime: time.Now(),
		logs:      initialLogs,
	}
}

func (m ServerModel) Run() error {
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func (m ServerModel) Init() tea.Cmd {
	return tea.Batch(
		waitForServerEvent(m.events),
		serverTick(),
	)
}

func (m ServerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if !m.ready {
			m.viewport = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-1),
			)
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - 1)
		}

	case lolserver.Event:
		ts := msg.Time.Format("15:04:05")
		var line string
		switch msg.Type {
		case lolserver.ClientConnected:
			m.clients++
			line = connectedLineStyle.Render(fmt.Sprintf("%s  + %s connected", ts, msg.Name))
		case lolserver.ClientDisconnected:
			if m.clients > 0 {
				m.clients--
			}
			line = disconnectedLineStyle.Render(fmt.Sprintf("%s  - %s disconnected", ts, msg.Name))
		}
		if line != "" {
			m.logs = append(m.logs, line)
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
			m.viewport.GotoBottom()
		}
		return m, waitForServerEvent(m.events)

	case serverTickMsg:
		return m, serverTick()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ServerModel) View() tea.View {
	var view tea.View
	view.AltScreen = true
	view.BackgroundColor = lipgloss.Black

	if !m.ready {
		view.SetContent("Starting...")
		return view
	}

	left := fmt.Sprintf(" %d clients connected", m.clients)
	right := fmt.Sprintf("uptime: %s ", time.Since(m.startTime).Round(time.Second))
	pad := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 0 {
		pad = 0
	}
	statusBar := serverStatusStyle.Width(m.width).Render(left + strings.Repeat(" ", pad) + right)

	view.SetContent(lipgloss.JoinVertical(lipgloss.Left,
		m.viewport.View(),
		statusBar,
	))
	return view
}
