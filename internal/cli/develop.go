package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var developCmd = &cobra.Command{
	Use:   "develop [shell]",
	Short: "Enters a persistent development shell from the flake.",
	Long:  `This command enters a persistent development shell from the flake. Lists available shells if none is specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.Develop(args); err != nil {
			fmt.Println("Error entering development shell:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(developCmd)
}
