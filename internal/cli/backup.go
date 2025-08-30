package cli

import (
	"fmt"
	"pilo/internal/api"
	"pilo/internal/config"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the current Pilo configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		path := config.GetInstallPath()
		err := api.GitBackup(path)
		if err != nil {
			fmt.Printf("Failed to create backup: %v\n", err)
		} else {
			fmt.Println("Backup created successfully!")
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
