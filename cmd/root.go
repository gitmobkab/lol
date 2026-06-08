/*
Copyright © 2026 gitmobkab <richmondaurel77@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/gitmobkab/lol/data"
	"github.com/spf13/cobra"
)

var versionFull bool

var rootCmd = &cobra.Command{
	Version: data.GetVersion(),
	Use:     "lol",
	Short:   "An app to chat on a LAN",
	Long: `## Flow

1 - One person on the LAN starts the server:

    lol serve

2 - That person finds their LAN IP (Windows: ipconfig, Linux/Mac: ip addr or ifconfig)
    and shares it with everyone else.

3 - Everyone else joins using that IP:

    lol join 192.168.1.5
    lol join 192.168.1.5 --name Alice
    lol join 192.168.1.5:9090          (if the server was started with --addr :9090)
`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFull {
			fmt.Println(data.GetDetailedVersion())
			return
		}
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&versionFull, "version-full", false, "Print detailed version info and exit")
}
