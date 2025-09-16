package cli

import (
	"fmt"
	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var devshellType string

func init() {
	devshellCmd.AddCommand(initDevshellCmd)
	devshellCmd.AddCommand(enterDevshellCmd)
	devshellCmd.AddCommand(runInDevshellCmd)
	rootCmd.AddCommand(devshellCmd)
}

var devshellCmd = &cobra.Command{
	Use:   "devshell",
	Short: "Manage and initialize development shells",
}

var initDevshellCmd = &cobra.Command{
	Use:   "init [template] [directory]",
	Short: "Initialize a new devshell from a template",
	Long:  "Initializes a new standalone devshell flake in the specified directory, based on a template.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		template := args[0]
		directory := args[1]
		if err := api.InitDevshellFromTemplate(template, directory); err != nil {
			fmt.Printf("Error initializing devshell: %v\n", err)
			return
		}
		fmt.Printf("Devshell '%s' initialized successfully in '%s'.\n", template, directory)
		fmt.Println("To enter the shell, run: cd", directory, "&& nix develop")
	},
}

var enterDevshellCmd = &cobra.Command{
	Use:   "enter [directory]",
	Short: "Enter a development shell in the specified directory",
	Long:  "Starts a development shell using the flake.nix in the target directory.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		directory := "."
		if len(args) > 0 {
			directory = args[0]
		}
		if err := api.EnterDevshell(directory); err != nil {
			fmt.Printf("Error entering devshell: %v\n", err)
			return
		}
	},
}

var runInDevshellCmd = &cobra.Command{
	Use:   "run [directory] [command]",
	Short: "Run a command in a development shell",
	Long:  "Runs a command within the context of the devshell defined in the specified directory.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		directory := "."
		command := ""
		if len(args) == 1 {
			command = args[0]
		} else {
			directory = args[0]
			command = args[1]
		}

		output, err := api.RunInDevshell(directory, command)
		if err != nil {
			fmt.Printf("Error running command in devshell: %v\n", err)
			return
		}
		fmt.Println(output)
	},
}
