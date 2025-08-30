package cli

import (
	"bufio"
	"fmt"
	"os"
	"pilo/internal/api"
	"pilo/internal/config"
	"pilo/internal/gui"
	"strings"

	"github.com/spf13/cobra"
)

var branch string

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the Pilo configuration from a remote Git repository.",
	Run: func(cmd *cobra.Command, args []string) {
		path := config.GetInstallPath()
		remoteURL, _ := config.GetRemoteUrl()
		if remoteURL == "" {
			fmt.Println("Remote URL is not set. Please set it in the preferences.")
			return
		}
		branch, _ := cmd.Flags().GetString("branch")

		err := api.GitRestore(path, remoteURL, branch, nil, "")
		if err != nil {
			if err == api.ErrDirtyRepository {
				fmt.Println("Your local repository has uncommitted changes.")
				fmt.Print("Would you like to back them up, commit them, discard them, or cancel? (backup/commit/discard/cancel): ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				var strategy api.GitRestoreStrategy
				var commitMessage string

				switch input {
				case "backup":
					strategy = api.GitRestoreBackup
				case "commit":
					strategy = api.GitRestoreCommit
					fmt.Print("Enter a commit message: ")
					commitMessage, _ = reader.ReadString('\n')
					commitMessage = strings.TrimSpace(commitMessage)
				case "discard":
					strategy = api.GitRestoreDiscard
				default:
					fmt.Println("Restore canceled.")
					return
				}

				err = api.GitRestore(path, remoteURL, branch, &strategy, commitMessage)
				if err != nil {
					fmt.Printf("Failed to restore configuration: %v\n", err)
				} else {
					fmt.Println("Configuration restored successfully!")
					gui.Refresh()
				}
			} else {
				fmt.Printf("Failed to restore configuration: %v\n", err)
			}
		} else {
			fmt.Println("Configuration restored successfully!")
			gui.Refresh()
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVarP(&branch, "branch", "b", "", "Specify a branch to restore from")
}
