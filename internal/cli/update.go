package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/spinner"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [input]",
	Short: "Updates flake inputs. Optionally updates a single [input].",
	Long:  `This command updates your flake's inputs by modifying the flake.lock file. Provide an optional input name to update only that specific dependency.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		spinner := spinner.NewSpinner("Updating flake inputs...")
		defer spinner.Stop()
		var inputName string
		if len(args) > 0 {
			inputName = args[0]
		}

		if _, err := api.Update(inputName); err != nil {
			fmt.Println("Error updating flake inputs:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
