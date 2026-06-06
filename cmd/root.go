/*
Copyright © 2026 gitmobkab <richmondaurel77@gmail.com>


*/
package cmd

import (
	"os"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lol",
	Short: "An app to chat on a LAN",
	Long: `## Flow
	
the main flow to use lol is:

1 - start up a lol server with **lol serve**

2 - join a server with **lol join IP[:PORT]**
`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}