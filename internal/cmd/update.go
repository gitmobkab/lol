package cmd

import (
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"

	"github.com/gitmobkab/lol/internal/data"
	"github.com/gitmobkab/lol/internal/tui/markdown_viewer"
)

var noChangelog bool

var updateCmd = &cobra.Command{
	Use:   "update [version]",
	Short: "Updates the application to the latest version.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments: expected at most one version")
		}
		return nil
	},
	Long: `Updates the lol binary in-place.

Without arguments, installs the latest stable release.
With a version argument, installs that specific release.
The version should be in semantic versioning format (e.g., 1.2.3).

# Examples:

- lol update
- lol update 0.1.0
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		updater := selfupdate.DefaultUpdater()
		slug := selfupdate.ParseSlug("gitmobkab/lol")

		var (
			release *selfupdate.Release
			found   bool
			err     error
		)

		specificVersion := len(args) == 1
		if specificVersion {
			release, found, err = updater.DetectVersion(ctx, slug, normalizeVersion(args[0]))
		} else {
			release, found, err = updater.DetectLatest(ctx, slug)
		}
		if err != nil {
			return fmt.Errorf("could not check for releases: %w", err)
		}
		if !found {
			return fmt.Errorf("no release found for this platform")
		}

		// Skip version check for dev builds so local installs always update.
		// When a specific version was requested, always install it (allow downgrade).
		if !specificVersion && data.Version != "dev" && !release.GreaterThan(data.Version) {
			fmt.Printf("Already on the latest version (%s), nothing to do.\n", data.Version)
			return nil
		}

		if !noChangelog {
			notes := release.ReleaseNotes
			if notes == "" {
				notes = "No Changelog Notes"
			}
			fullNotes := fmt.Sprintf("\nWhat's new in %s:\n\n%s\n", release.Version(), notes)
			markdownModel := markdown_viewer.NewMarkdownViewer(fullNotes)
			if _, err := tea.NewProgram(markdownModel).Run(); err != nil {
				return err
			}
		}

		fmt.Printf("Perform update %s → %s? [y/N]: ", data.Version, release.Version())
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Update cancelled.")
			return nil
		}

		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not resolve executable path: %w", err)
		}

		fmt.Printf("Updating %s → %s...\n", data.Version, release.Version())
		if err := updater.UpdateTo(ctx, release, exePath); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Println("Done. Run `lol --version` to confirm.")
		return nil
	},
}

func normalizeVersion(v string) string {
	if len(v) == 0 || v[0] != 'v' {
		return "v" + v
	}
	return v
}

func init() {
	updateCmd.Flags().BoolVarP(&noChangelog, "no-changelog", "n", false, "skip showing the changelog before updating")
	rootCmd.AddCommand(updateCmd)
}
