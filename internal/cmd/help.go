package cmd

import (
	"fmt"
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/gitmobkab/lol/internal/tui/markdown_viewer"
)

const HELP_FORMAT string = `
# %s - %s

> %s

%s
`

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
            if err != nil{
                return err
            }
        } else {
            target = cmd
        }

        content := fmt.Sprintf(HELP_FORMAT, 
            target.Name(), target.Use, target.Short, target.Long)

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
