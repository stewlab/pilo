package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/spinner"

	"github.com/spf13/cobra"
)

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "Runs the garbage collector to free up disk space.",
	Long:  `This command runs the garbage collector to free up disk space.`,
	Run: func(cmd *cobra.Command, args []string) {
		spinner := spinner.NewSpinner("Running garbage collector...")
		spinner.Start()
		defer spinner.Stop()
		if _, err := api.GC(); err != nil {
			fmt.Println("Error running garbage collector:", err)
			os.Exit(1)
		}
		fmt.Println("Garbage collection complete!")
	},
}

func init() {
	rootCmd.AddCommand(gcCmd)
}
