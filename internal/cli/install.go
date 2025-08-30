package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pilo/internal/api"
	"pilo/internal/config"
	"pilo/internal/spinner"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install pilo and configure your system.",
	Long:  `This command installs the pilo flake to your system and configures it as a Nix registry entry.`,
	Run: func(cmd *cobra.Command, args []string) {
		spinner := spinner.NewSpinner("Installing pilo...")
		defer spinner.Stop()

		var runErr error // Use a local error variable to track issues

		// Get flags
		path, _ := cmd.Flags().GetString("path")
		registry, _ := cmd.Flags().GetString("registry")
		remoteURL, _ := cmd.Flags().GetString("remote-url")
		sshKeyPath, _ := cmd.Flags().GetString("ssh-key-path")

		// Interactive prompts for missing flags
		if !cmd.Flags().Changed("path") {
			path = config.GetInstallPath()
			prompt := &survey.Input{
				Message: "Installation path:",
				Default: path,
			}
			if err := survey.AskOne(prompt, &path, survey.WithValidator(survey.Required)); err != nil {
				runErr = fmt.Errorf("error getting installation path: %w", err)
			}
		}

		if runErr == nil && !cmd.Flags().Changed("registry") {
			registry = config.GetRegistryName()
			prompt := &survey.Input{
				Message: "Nix registry name:",
				Default: registry,
			}
			if err := survey.AskOne(prompt, &registry, survey.WithValidator(survey.Required)); err != nil {
				runErr = fmt.Errorf("error getting registry name: %w", err)
			}
		}

		if runErr == nil && !cmd.Flags().Changed("remote-url") {
			remoteURL, _ = config.GetRemoteUrl()
			prompt := &survey.Input{
				Message: "Remote Git URL (optional):",
				Default: remoteURL,
			}
			if err := survey.AskOne(prompt, &remoteURL); err != nil {
				runErr = fmt.Errorf("error getting remote URL: %w", err)
			}
		}

		if runErr == nil && remoteURL != "" && !cmd.Flags().Changed("ssh-key-path") {
			sshKeyPath = config.GetSshKeyPath()
			prompt := &survey.Input{
				Message: "SSH Key Path:",
				Default: sshKeyPath,
			}
			if err := survey.AskOne(prompt, &sshKeyPath); err != nil {
				runErr = fmt.Errorf("error getting SSH key path: %w", err)
			}
		}

		// Persist settings
		config.SetRemoteGitUrl(remoteURL)
		config.SetSshKeyPath(sshKeyPath)

		// Expand path
		if runErr == nil && strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				runErr = fmt.Errorf("error getting home directory: %w", err)
			} else {
				path = filepath.Join(home, path[2:])
			}
		}

		// Add a confirmation prompt if the directory already exists
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			confirm := false
			prompt := &survey.Confirm{
				Message: "Installation path already exists. This will commit any existing changes and then reinstall. Are you sure you want to proceed?",
			}
			if err := survey.AskOne(prompt, &confirm); err != nil || !confirm {
				fmt.Println("Installation cancelled.")
				return
			}
		}

		if runErr == nil {
			if err := api.InstallPilo(path, registry, remoteURL); err != nil {
				runErr = fmt.Errorf("error installing pilo: %w", err)
			}
		}

		if runErr != nil {
			fmt.Println(runErr)
			os.Exit(1) // Exit after defer spinner.Stop() has been called
		} else {
			if err := api.GitAdd(path); err != nil {
				runErr = fmt.Errorf("error adding changes: %w", err)
			} else if err := api.GitCommit(path, "pilo: install"); err != nil {
				runErr = fmt.Errorf("error committing changes: %w", err)
			} else {
				fmt.Println("Pilo installed successfully!")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().String("path", "~/.config/pilo", "The path to install the configuration to.")
	installCmd.Flags().String("registry", "pilo", "The name for this flake in the Nix registry.")
	installCmd.Flags().String("remote-url", "", "The remote Git URL to install from (optional).")
	installCmd.Flags().String("ssh-key-path", "", "The path to the SSH key to use for remote Git operations (optional).")
}
