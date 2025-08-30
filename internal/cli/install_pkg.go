package cli

import (
	"fmt"
	"os"
	"strings"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var installPkgCmd = &cobra.Command{
	Use:   "install-pkg [pkg...]",
	Short: "Installs packages to your user profile (non-NixOS) or provides a temporary shell (NixOS).",
	Long:  `This command installs packages to your user profile (non-NixOS) or provides a temporary shell (NixOS).`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, pkg := range args {
			if strings.HasPrefix(pkg, "github:") {
				if err := api.AddGitPackage(pkg); err != nil {
					fmt.Printf("Error adding git package %s: %v\n", pkg, err)
					os.Exit(1)
				}
				fmt.Printf("Added git package %s. Rebuilding system...", pkg)
				if _, err := api.Rebuild("", "", "", ""); err != nil {
					fmt.Printf("Error rebuilding system: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("System rebuilt successfully.")
			} else {
				if err := api.Install([]string{pkg}); err != nil {
					fmt.Println("Error installing packages:", err)
					os.Exit(1)
				}
				fmt.Printf("Installed package %s. Rebuilding system...", pkg)
				if _, err := api.Rebuild("", "", "", ""); err != nil {
					fmt.Printf("Error rebuilding system: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("System rebuilt successfully.")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(installPkgCmd)
}
