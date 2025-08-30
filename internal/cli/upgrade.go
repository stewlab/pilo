package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/spinner"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades all packages in your user profile or NixOS system.",
	Long:  `This command upgrades all packages in your user profile or NixOS system.`,
	Run: func(cmd *cobra.Command, args []string) {
		spinner := spinner.NewSpinner("Upgrading packages...")
		defer spinner.Stop()
		if _, err := api.Upgrade(); err != nil {
			fmt.Println("Error upgrading:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
