package client_tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/gitmobkab/lol/internal/protocol"
	"github.com/gitmobkab/lol/internal/tui/message_bubble"
)

type cmdHandler func(m ClientModel, args []string, tail string) (ClientModel, tea.Cmd)

type Command struct {
	Usage string
	Help  string
	Args  int
	Tail  bool
	Run   cmdHandler
}

var registry = map[string]Command{
	"ping": {
		Usage: "ping",
		Help:  "send a ping request to the server",
		Run:   cmdPing,
	},
	"dm": {
		Usage: "dm <user> <message>",
		Help:  "send a direct message",
		Args:  1,
		Tail:  true,
		Run:   cmdDM,
	},
	"die": {
		Usage: "die",
		Help:  "quit the app",
		Run:   cmdDie,
	},
}

func dispatch(m ClientModel, input string) (ClientModel, tea.Cmd) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return m, nil
	}
	name, rest := tokens[0], tokens[1:]

	cmd, ok := registry[name]
	if !ok {
		ts := time.Now().Format("15:04")
		m = addMessage(m, "System", fmt.Sprintf("Unknown command: /%s", name), ts, false, message_bubble.KindSystem)
		return m, nil
	}

	if cmd.Args > 0 && len(rest) < cmd.Args {
		ts := time.Now().Format("15:04")
		m = addMessage(m, "System", "Usage: /"+cmd.Usage, ts, false, message_bubble.KindSystem)
		return m, nil
	}

	args := rest[:cmd.Args]
	tail := ""
	if cmd.Tail {
		tail = strings.Join(rest[cmd.Args:], " ")
	}

	return cmd.Run(m, args, tail)
}

func cmdPing(m ClientModel, _ []string, _ string) (ClientModel, tea.Cmd) {
	m.client.SendPing(m.ctx)
	return m, nil
}

func cmdDM(m ClientModel, args []string, tail string) (ClientModel, tea.Cmd) {
	ts := time.Now().Format("15:04")
	rawTarget := args[0]
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
		m = addMessage(m, "System", fmt.Sprintf("user %q not found", rawTarget), ts, false, message_bubble.KindSystem)
	case 1:
		m.client.SendDM(m.ctx, matches[0].ID, tail)
		m = addMessage(m, "DM → "+matches[0].Name, tail, ts, true, message_bubble.KindDM)
	default:
		ids := make([]string, len(matches))
		for i, mem := range matches {
			ids[i] = mem.Name + "#" + shortID(mem.ID)
		}
		m = addMessage(m, "System", "ambiguous — be specific: "+strings.Join(ids, ", "), ts, false, message_bubble.KindSystem)
	}

	return m, nil
}

func cmdDie(m ClientModel, _ []string, _ string) (ClientModel, tea.Cmd) {
	return m, tea.Quit
}

// autocompleteSuggestions returns commands whose name starts with the typed
// fragment after '/'. Returns nil once the user types a space (args phase).
func autocompleteSuggestions(input string) []Command {
	if !strings.HasPrefix(input, "/") {
		return nil
	}
	fragment := input[1:]
	if strings.ContainsRune(fragment, ' ') {
		return nil
	}
	var matches []Command
	for name, cmd := range registry {
		if strings.HasPrefix(name, fragment) {
			matches = append(matches, cmd)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Usage < matches[j].Usage
	})
	return matches
}
