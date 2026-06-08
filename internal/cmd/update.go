package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [version]",
	Short: "Updates the application to the latest version.",
	Long: `This command if called without any arguments, it will update the application to the latest stable version.
	
If a version is provided, it will update to that specific version. 
The version should be in the semantic versioning format (e.g., 1.2.3).

# Examples:

- lol update
- lol update 0.9.0
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
