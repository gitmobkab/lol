package client_tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"github.com/gitmobkab/lol/internal/protocol"
	"github.com/gitmobkab/lol/internal/tui/message_bubble"
	"github.com/gitmobkab/lol/internal/tui/theme"
)

type cmdHandler func(m ClientModel, args []string, tail string) (ClientModel, tea.Cmd)

// completeFn returns completion suggestions given already-confirmed args and the
// current partial token being typed. It is responsible for filtering by partial.
type completeFn func(m ClientModel, confirmedArgs []string, partial string) []string

type Command struct {
	Usage    string
	Help     string
	Args     int
	Tail     bool
	Run      cmdHandler
	Complete completeFn // optional; nil means no argument completions
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
		Complete: func(m ClientModel, confirmedArgs []string, partial string) []string {
			if len(confirmedArgs) > 0 {
				return nil // past the username; don't complete the message body
			}
			lower := strings.ToLower(partial)
			var out []string
			for _, mem := range m.members {
				if strings.HasPrefix(strings.ToLower(mem.Name), lower) {
					out = append(out, mem.Name)
				}
			}
			sort.Strings(out)
			return out
		},
	},
	"die": {
		Usage: "die",
		Help:  "quit the app",
		Run:   cmdDie,
	},
	"theme": {
		Usage: "theme <name>",
		Help:  "switch color theme (dark, dracula, nord)",
		Args:  1,
		Run:   cmdTheme,
		Complete: func(_ ClientModel, _ []string, partial string) []string {
			all := []string{"dark", "dracula", "nord"}
			var out []string
			for _, name := range all {
				if strings.HasPrefix(name, partial) {
					out = append(out, name)
				}
			}
			return out
		},
	},
	"upload": {
		Usage: "upload [path]",
		Help:  "send a file (no path opens a file picker, max 10 MB)",
		Args:  0,
		Tail:  true,
		Run:   cmdUpload,
		Complete: func(_ ClientModel, _ []string, partial string) []string {
			return listPathSuggestions(partial)
		},
	},
	"save": {
		Usage: "save <filename>",
		Help:  "save a received file to ~/Downloads/lol/",
		Args:  0,
		Tail:  true,
		Run:   cmdSave,
		Complete: func(m ClientModel, _ []string, partial string) []string {
			names := m.client.BufferedFileNames()
			var out []string
			for _, name := range names {
				if strings.HasPrefix(name, partial) {
					out = append(out, name)
				}
			}
			return out
		},
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

func cmdTheme(m ClientModel, args []string, _ string) (ClientModel, tea.Cmd) {
	ts := time.Now().Format("15:04")
	t, ok := theme.All[args[0]]
	if !ok {
		m = addMessage(m, "System", fmt.Sprintf("unknown theme %q — available: dark, dracula, nord", args[0]), ts, false, message_bubble.KindSystem)
		return m, nil
	}
	SetTheme(t)
	message_bubble.SetTheme(t)
	if m.ready {
		m.viewport.SetContent(m.renderMessages())
	}
	m = addMessage(m, "System", "theme set to "+t.Name, ts, false, message_bubble.KindSystem)
	return m, nil
}

func cmdUpload(m ClientModel, _ []string, tail string) (ClientModel, tea.Cmd) {
	ts := time.Now().Format("15:04")
	path := strings.TrimSpace(tail)
	if path == "" {
		fp := filepicker.New()
		fp.FileAllowed = true
		fp.DirAllowed = false
		fp.SetHeight(min(10, max(m.height/3, 4)))
		if home, err := os.UserHomeDir(); err == nil {
			fp.CurrentDirectory = home
		}
		m.filePicker = &fp
		return m, fp.Init()
	}
	if err := m.client.SendFile(m.ctx, path); err != nil {
		m = addMessage(m, "System", fmt.Sprintf("upload failed: %s", err), ts, false, message_bubble.KindSystem)
		return m, nil
	}
	m = addMessage(m, "System", fmt.Sprintf("sent %s", filepath.Base(path)), ts, false, message_bubble.KindSystem)
	return m, nil
}

// parseInputForCompletion extracts the command name, already-confirmed args, and
// the partial token currently being typed. inArgs is false when still completing
// the command name itself.
// For Tail-only commands (Args==0, Tail==true) the whole rest string is treated
// as a single partial so completions work on names that contain spaces.
func parseInputForCompletion(input string) (cmdName string, confirmedArgs []string, partial string, inArgs bool) {
	if !strings.HasPrefix(input, "/") {
		return
	}
	after := input[1:]
	spaceIdx := strings.IndexRune(after, ' ')
	if spaceIdx < 0 {
		return // still typing the command name
	}
	inArgs = true
	cmdName = after[:spaceIdx]
	rest := after[spaceIdx+1:]

	// Tail-only commands treat the entire rest as one argument.
	if cmd, ok := registry[cmdName]; ok && cmd.Tail && cmd.Args == 0 {
		confirmedArgs = nil
		partial = rest
		return
	}

	if strings.HasSuffix(rest, " ") || rest == "" {
		confirmedArgs = strings.Fields(rest)
		partial = ""
	} else {
		fields := strings.Fields(rest)
		confirmedArgs = fields[:len(fields)-1]
		partial = fields[len(fields)-1]
	}
	return
}

// argCompletions returns filtered argument suggestions for the current input.
func argCompletions(m ClientModel, input string) []string {
	cmdName, confirmedArgs, partial, inArgs := parseInputForCompletion(input)
	if !inArgs {
		return nil
	}
	cmd, ok := registry[cmdName]
	if !ok || cmd.Complete == nil {
		return nil
	}
	return cmd.Complete(m, confirmedArgs, partial)
}

// applyCompletion replaces the partial at the end of input with suggestion.
// For Tail-only commands it replaces everything after the command name so
// multi-word suggestions (filenames with spaces) are applied correctly.
func applyCompletion(input, suggestion string) string {
	if !strings.HasPrefix(input, "/") {
		return input
	}
	after := input[1:]
	spaceIdx := strings.IndexRune(after, ' ')
	if spaceIdx < 0 {
		return input
	}
	cmdName := after[:spaceIdx]
	if cmd, ok := registry[cmdName]; ok && cmd.Tail && cmd.Args == 0 {
		// Replace everything after "/cmdname " with the suggestion.
		return input[:1+spaceIdx+1] + suggestion
	}
	idx := strings.LastIndexByte(input, ' ')
	return input[:idx+1] + suggestion
}

// commandName returns the bare name from a Usage string (the first word).
func commandName(usage string) string {
	if i := strings.IndexByte(usage, ' '); i >= 0 {
		return usage[:i]
	}
	return usage
}

// listPathSuggestions returns filesystem entries whose name starts with the
// basename of partial, preserving the directory prefix.
func listPathSuggestions(partial string) []string {
	dir, prefix := filepath.Split(partial)
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		full := filepath.Join(dir, e.Name())
		if dir == "." {
			full = e.Name()
		}
		if e.IsDir() {
			full += string(filepath.Separator)
		}
		out = append(out, full)
	}
	return out
}

func cmdSave(m ClientModel, _ []string, tail string) (ClientModel, tea.Cmd) {
	ts := time.Now().Format("15:04")
	name := strings.TrimSpace(tail)
	if name == "" {
		m = addMessage(m, "System", "Usage: /save <filename>", ts, false, message_bubble.KindSystem)
		return m, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, "Downloads", "lol")
	if err := m.client.SaveFile(name, dir); err != nil {
		m = addMessage(m, "System", fmt.Sprintf("save failed: %s", err), ts, false, message_bubble.KindSystem)
		return m, nil
	}
	m = addMessage(m, "System", fmt.Sprintf("saved to %s", filepath.Join(dir, name)), ts, false, message_bubble.KindSystem)
	return m, nil
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
