/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/gitmobkab/lol/tui"
	"github.com/spf13/cobra"
)

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "opens the help screen for a command",
	Long: `# Opens the help screen for a command.

If no command is provided, it shows the help for itself.`,
	Run: func(cmd *cobra.Command, args []string) {
        target := cmd
        if len(args) > 0 {
            target, _, _ = rootCmd.Find(args)
        }

        rendered, err := glamour.Render(target.Long, "dark")
        if err != nil {
            fmt.Println(target.Long)
            return
        }

        p := tea.NewProgram(
            tui.NewHelpModel(rendered),
            tea.WithAltScreen(),
        )
        p.Run()
    },
}

func init() {
	rootCmd.SetHelpCommand(helpCmd)
}
