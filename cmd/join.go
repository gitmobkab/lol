package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join IP[:PORT]",
	Short: "Joins a room with the given IP address and port",
	Long: `The ip address need to be in the format IP:PORT.
Where *IP* is a valid IPv4 address, while **PORT** is an optional number between 1 and 65535.

Unless specified, the port will default to 8080.

# Examples:

- lol join
- lol join 192.168.1.100
- lol join 192.168.1.100:8080
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("join called")
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
}