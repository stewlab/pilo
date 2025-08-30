package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/spinner"

	"github.com/spf13/cobra"
)

var removePkgCmd = &cobra.Command{
	Use:   "remove [pkg...]",
	Short: "Removes packages from your user profile (non-NixOS).",
	Long:  `This command removes packages from your user profile (non-NixOS).`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		spinner := spinner.NewSpinner("Removing packages...")
		defer spinner.Stop()
		if err := api.Remove(args); err != nil {
			fmt.Println("Error removing packages:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(removePkgCmd)
}
