package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/config"
	"pilo/internal/gui"
	"pilo/internal/nix"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuilds your NixOS or Home Manager configuration.",
	Long:  `This command rebuilds your NixOS or Home Manager configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		flakePath, _ := cmd.Flags().GetString("flake")
		if flakePath == "" {
			flakePath = config.GetFlakePath()
		}
		nixpkgsURL, _ := cmd.Flags().GetString("nixpkgs")
		homeManagerURL, _ := cmd.Flags().GetString("home-manager")

		var password string
		if nix.GetNixMode() == nix.NixOS {
			prompt := &survey.Password{
				Message: "Please enter your password:",
			}
			survey.AskOne(prompt, &password)
		}

		output, err := api.Rebuild(flakePath, password, nixpkgsURL, homeManagerURL)
		fmt.Println(output)
		if err != nil {
			fmt.Println("Error rebuilding:", err)
			os.Exit(1)
		} else {
			path := config.GetInstallPath()
			if err := api.GitAdd(path); err != nil {
				fmt.Println("Error adding changes:", err)
				os.Exit(1)
			}
			if err := api.GitCommit(path, "pilo: rebuild"); err != nil {
				fmt.Println("Error committing changes:", err)
				os.Exit(1)
			}
			gui.Refresh()
		}
	},
}

func init() {
	rebuildCmd.Flags().StringP("flake", "f", "", "Path to the flake to rebuild")
	rebuildCmd.Flags().String("nixpkgs", "", "URL of the nixpkgs flake to use")
	rebuildCmd.Flags().String("home-manager", "", "URL of the home-manager flake to use")
	rootCmd.AddCommand(rebuildCmd)
}
