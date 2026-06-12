package cmd

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gitmobkab/lol/internal/tui/markdown_viewer"
)

const HELP_FORMAT string = `
# %s - %s

> %s

%s
%s%s`

func buildSubcommandsSection(cmd *cobra.Command) string {
	var lines []string
	for _, c := range cmd.Commands() {
		if !c.Hidden {
			lines = append(lines, fmt.Sprintf("| `%s` | %s |", c.Name(), c.Short))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Available Commands\n\n")
	sb.WriteString("| Command | Description |\n")
	sb.WriteString("|---------|-------------|\n")
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
	return sb.String()
}

func buildFlagsSection(cmd *cobra.Command) string {
	var lines []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		label := fmt.Sprintf("`--%s`", f.Name)
		if f.Shorthand != "" {
			label += fmt.Sprintf(", `-%s`", f.Shorthand)
		}
		lines = append(lines, fmt.Sprintf("| %s | `%s` | %s |", label, f.DefValue, f.Usage))
	})
	if len(lines) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("# Flags\n\n")
	sb.WriteString("| Flag | Default | Description |\n")
	sb.WriteString("|------|---------|-------------|\n")
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
	return sb.String()
}

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "opens the help screen for a command",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("Too many arguments, expected at most one argument")
		}
		return nil
	},
	Long: `If no command is provided, it shows the help for itself.

# Examples:

- lol help

- lol help serve

- lol help help <- **same as lol help**`,

	RunE: func(cmd *cobra.Command, args []string) error {
		var target *cobra.Command
		var err error = nil

		if len(args) == 1 {
			target, _, err = rootCmd.Find(args)
			if err != nil {
				return err
			}
		} else {
			target = cmd
		}

		content := fmt.Sprintf(HELP_FORMAT,
			target.Name(), target.Use, target.Short, target.Long,
			buildSubcommandsSection(target),
			buildFlagsSection(target))

		p := tea.NewProgram(
			markdown_viewer.NewMarkdownViewer(content),
		)

		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.SetHelpCommand(helpCmd)
}
