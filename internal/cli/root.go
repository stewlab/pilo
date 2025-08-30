package cli

import (
	"embed"
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/config"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pilo",
	Short: "A command-line tool for managing your Nix environment.",
	Long:  `A longer description that spans multiple lines and likely contains examples and usage of using your application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Check if the command is 'install' and if so, skip the auto-install logic
		if cmd.Name() == "install" {
			return
		}
		handleAutoInstall()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute(flakeFS embed.FS) {
	api.SetFlakeFS(flakeFS)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handleAutoInstall() {
	installPath := config.GetInstallPath()
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		fmt.Println("Pilo configuration not found. Installing the default configuration...")
		err := api.AutoInstallCLI(installPath)
		if err != nil {
			fmt.Printf("Failed to install Pilo configuration: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Println("Pilo has been installed successfully.")
		}
	}
}
