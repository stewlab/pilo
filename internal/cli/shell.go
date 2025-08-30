package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [pkg...]",
	Short: "Enters a temporary shell with the specified packages.",
	Long:  `This command enters a temporary shell with the specified packages.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.Shell(args); err != nil {
			fmt.Println("Error entering temporary shell:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
