package cli

import (
	"fmt"
	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clears the Pilo logs.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.Reset(); err != nil {
			fmt.Printf("Error clearing logs: %v\n", err)
			return
		}
		fmt.Println("Logs have been cleared.")
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
