package cmd

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/gitmobkab/lol/internal/client"
	client_tui "github.com/gitmobkab/lol/internal/tui/client_tui"
	"github.com/spf13/cobra"
)

var name string

var joinCmd = &cobra.Command{
	Use:   "join IP[:PORT]",
	Short: "Joins a room with the given IP address and port",
	Long: `The ip address need to be in the format IP:PORT.
Where *IP* is a valid IPv4 address, while **PORT** is an optional number between 1 and 65535.

Unless specified, the port will default to 8080.

# Examples:

- lol join 192.168.1.100
- lol join 192.168.1.100 --name Alice
- lol join 192.168.1.100:8080 --name Alice
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := args[0]
		if !strings.Contains(addr, ":") {
			addr += ":8080"
		}

		logger := slog.Default()
		ctx := context.Background()

		c, err := client.Connect(ctx, "ws://"+addr, name, logger)
		if err != nil {
			return err
		}
		defer c.Close()

		go c.ReadLoop(ctx)

		return client_tui.NewClientModel(c, ctx).Run()
	},
}

func init() {
	hostname, _ := os.Hostname()

	rootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringVarP(&name, "name", "n", hostname, "Your display name")
}
