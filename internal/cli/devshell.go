package cli

import (
	"fmt"
	"pilo/internal/api"
	"pilo/internal/config"

	"github.com/spf13/cobra"
)

var devshellType string

func init() {
	devshellCmd.AddCommand(addDevshellCmd)
	devshellCmd.AddCommand(removeDevshellCmd)

	enterDevshellCmd.Flags().StringP("flake", "f", "", "Path to the flake")
	devshellCmd.AddCommand(enterDevshellCmd)

	runInDevshellCmd.Flags().StringP("flake", "f", "", "Path to the flake")
	devshellCmd.AddCommand(runInDevshellCmd)

	rootCmd.AddCommand(devshellCmd)
}

var devshellCmd = &cobra.Command{
	Use:   "devshell",
	Short: "Manage development shells",
}

var addDevshellCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new development shell",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := api.AddDevshellWithContent(name, ""); err != nil {
			fmt.Printf("Error adding devshell: %v\n", err)
			return
		}
		fmt.Printf("Devshell '%s' added successfully.\n", name)
	},
}

var removeDevshellCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a development shell",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := api.RemoveDevshell(name); err != nil {
			fmt.Printf("Error removing devshell: %v\n", err)
			return
		}
		fmt.Printf("Devshell '%s' removed successfully.\n", name)
	},
}

var enterDevshellCmd = &cobra.Command{
	Use:   "enter [name]",
	Short: "Enter a development shell",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		flakePath, _ := cmd.Flags().GetString("flake")
		if flakePath == "" {
			flakePath = config.GetFlakePath()
		}
		if err := api.EnterDevshell(name, flakePath); err != nil {
			fmt.Printf("Error entering devshell: %v\n", err)
			return
		}
	},
}

var runInDevshellCmd = &cobra.Command{
	Use:   "run [name] [command]",
	Short: "Run a command in a development shell",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		command := args[1]
		flakePath, _ := cmd.Flags().GetString("flake")
		if flakePath == "" {
			flakePath = config.GetFlakePath()
		}
		output, err := api.RunInDevshell(name, command, flakePath)
		if err != nil {
			fmt.Printf("Error running command in devshell: %v\n", err)
			return
		}
		fmt.Println(output)
	},
}
