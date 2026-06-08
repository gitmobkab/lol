package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gitmobkab/lol/server"
	server_tui "github.com/gitmobkab/lol/tui/server_tui"
	"github.com/gitmobkab/lol/tui/choice_picker"
	"github.com/spf13/cobra"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the server.",
	Long: `This command will start the server and listen for **incoming requests**.

You can specify additional flags to configure the server behavior.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := pickAddr(port)
		if err != nil {
			return err
		}

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		s := server.New(logger)

		errCh := make(chan error, 1)
		go func() { errCh <- s.Start(addr) }()

		select {
		case err := <-errCh:
			return err
		case <-time.After(100 * time.Millisecond):
		}

		initialLogs := []string{"join with: lol join " + addr}
		return server_tui.NewServerModel(s.Events, initialLogs).Run()
	},
}

// --- IP picker ---

func pickAddr(port string) (string, error) {
	ips, err := lanIPs()
	if err != nil {
		return "", err
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no valid network interface found")
	}
	if len(ips) == 1 {
		return ips[0] + ":" + port, nil
	}

	result, err := tea.NewProgram(choice_picker.NewPickerModel(ips)).Run()
	if err != nil {
		return "", err
	}
	chosen := result.(choice_picker.PickerModel).Chosen
	if chosen == "" {
		return "", fmt.Errorf("no interface selected")
	}
	return chosen + ":" + port, nil
}

func lanIPs() ([]string, error) {
	ifaces, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, iface := range ifaces {
		ipnet, ok := iface.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.To4() == nil {
			continue
		}
		ips = append(ips, ipnet.IP.String())
	}
	return ips, nil
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
}
