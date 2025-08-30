package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/spinner"

	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rolls back to the previous generation.",
	Long:  `This command rolls back to the previous generation.`,
	Run: func(cmd *cobra.Command, args []string) {
		spinner := spinner.NewSpinner("Rolling back...")
		spinner.Start()
		defer spinner.Stop()
		if _, err := api.Rollback(""); err != nil {
			fmt.Println("Error rolling back:", err)
			os.Exit(1)
		}
		fmt.Println("Rollback complete!")
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
