package client_tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInputForCompletion(t *testing.T) {
	cases := []struct {
		input            string
		wantCmd          string
		wantConfirmedLen int
		wantConfirmedNil bool
		wantPartial      string
		wantInArgs       bool
	}{
		// Not a command — no slash prefix
		{"hello world", "", 0, true, "", false},
		// Still typing the command name (no space yet)
		{"/up", "", 0, true, "", false},
		{"/upload", "", 0, true, "", false},
		// Tail-only commands: entire rest is the partial
		{"/upload ", "upload", 0, true, "", true},
		{"/upload /home/ri", "upload", 0, true, "/home/ri", true},
		{"/upload /home/user/my file", "upload", 0, true, "/home/user/my file", true},
		{"/save HOW TO RUN GAME!!.txt", "save", 0, true, "HOW TO RUN GAME!!.txt", true},
		// Normal commands: split on spaces
		{"/dm ", "dm", 0, false, "", true},
		{"/dm alice", "dm", 0, false, "alice", true},
		{"/dm alice ", "dm", 1, false, "", true},
		{"/dm alice hello", "dm", 1, false, "hello", true},
		{"/theme d", "theme", 0, false, "d", true},
		{"/theme ", "theme", 0, false, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			cmd, confirmed, partial, inArgs := parseInputForCompletion(tc.input)
			if cmd != tc.wantCmd {
				t.Errorf("cmd: got %q, want %q", cmd, tc.wantCmd)
			}
			if inArgs != tc.wantInArgs {
				t.Errorf("inArgs: got %v, want %v", inArgs, tc.wantInArgs)
			}
			if partial != tc.wantPartial {
				t.Errorf("partial: got %q, want %q", partial, tc.wantPartial)
			}
			if tc.wantConfirmedNil && confirmed != nil {
				t.Errorf("confirmedArgs: got %v, want nil", confirmed)
			}
			if !tc.wantConfirmedNil && len(confirmed) != tc.wantConfirmedLen {
				t.Errorf("confirmedArgs len: got %d, want %d (%v)", len(confirmed), tc.wantConfirmedLen, confirmed)
			}
		})
	}
}

func TestApplyCompletion(t *testing.T) {
	cases := []struct {
		input, suggestion, want string
	}{
		// Tail-only: replaces everything after "/cmd "
		{"/upload /home/ri", "/home/richm/", "/upload /home/richm/"},
		{"/upload ", "/home/richm/", "/upload /home/richm/"},
		{"/save HOW", "HOW TO RUN GAME!!.txt", "/save HOW TO RUN GAME!!.txt"},
		// Normal: replaces the last space-delimited token
		{"/dm al", "alice#abc123", "/dm alice#abc123"},
		{"/theme d", "dark", "/theme dark"},
		{"/theme ", "dark", "/theme dark"},
		// No slash — returned unchanged
		{"hello", "suggestion", "hello"},
	}
	for _, tc := range cases {
		t.Run(tc.input+"→"+tc.suggestion, func(t *testing.T) {
			got := applyCompletion(tc.input, tc.suggestion)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCommandName(t *testing.T) {
	cases := []struct {
		usage, want string
	}{
		{"ping", "ping"},
		{"die", "die"},
		{"upload [path]", "upload"},
		{"dm <user> <message>", "dm"},
		{"theme <name>", "theme"},
		{"save <filename>", "save"},
	}
	for _, tc := range cases {
		got := commandName(tc.usage)
		if got != tc.want {
			t.Errorf("commandName(%q) = %q, want %q", tc.usage, got, tc.want)
		}
	}
}

func TestListPathSuggestions(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "beta.txt"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(dir, "gamma"), 0755)

	t.Run("prefix match", func(t *testing.T) {
		got := listPathSuggestions(filepath.Join(dir, "al"))
		if len(got) != 1 {
			t.Fatalf("got %v, want 1 result", got)
		}
		if got[0] != filepath.Join(dir, "alpha.txt") {
			t.Errorf("got %q, want %q", got[0], filepath.Join(dir, "alpha.txt"))
		}
	})

	t.Run("list all with trailing separator", func(t *testing.T) {
		got := listPathSuggestions(dir + string(filepath.Separator))
		if len(got) != 3 {
			t.Fatalf("got %d results, want 3: %v", len(got), got)
		}
	})

	t.Run("directory entry has trailing separator", func(t *testing.T) {
		got := listPathSuggestions(dir + string(filepath.Separator))
		found := false
		for _, s := range got {
			if strings.HasSuffix(s, "gamma"+string(filepath.Separator)) {
				found = true
			}
		}
		if !found {
			t.Errorf("directory entry missing trailing separator: %v", got)
		}
	})

	t.Run("no match returns nil", func(t *testing.T) {
		got := listPathSuggestions(filepath.Join(dir, "zzz"))
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("invalid parent returns nil", func(t *testing.T) {
		got := listPathSuggestions(filepath.Join(dir, "nonexistent", "foo"))
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestArgCompletionsTheme(t *testing.T) {
	m := ClientModel{}

	cases := []struct {
		input string
		want  []string
	}{
		{"/theme ", []string{"dark", "dracula", "nord"}},
		{"/theme d", []string{"dark", "dracula"}},
		{"/theme no", []string{"nord"}},
		{"/theme dr", []string{"dracula"}},
		{"/theme xyz", nil},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := argCompletions(m, tc.input)
			if len(got) != len(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
				return
			}
			for i, w := range tc.want {
				if got[i] != w {
					t.Errorf("[%d] got %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

func TestArgCompletionsUnknownCommand(t *testing.T) {
	m := ClientModel{}
	got := argCompletions(m, "/notacommand foo")
	if got != nil {
		t.Errorf("expected nil for unknown command, got %v", got)
	}
}

func TestArgCompletionsNotInArgs(t *testing.T) {
	m := ClientModel{}
	got := argCompletions(m, "/theme")
	if got != nil {
		t.Errorf("expected nil when not yet in args, got %v", got)
	}
}
