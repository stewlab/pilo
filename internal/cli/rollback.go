package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/nix"
	"pilo/internal/spinner"

	"github.com/AlecAivazis/survey/v2"
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

		var password string
		if nix.GetNixMode() == nix.NixOS {
			prompt := &survey.Password{
				Message: "Please enter your password:",
			}
			survey.AskOne(prompt, &password)
		}

		if _, err := api.Rollback(password); err != nil {
			fmt.Println("Error rolling back:", err)
			os.Exit(1)
		}
		fmt.Println("Rollback complete!")
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
