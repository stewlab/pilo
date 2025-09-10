package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [packages|generations|users|aliases]",
	Short: "Lists installed packages, system generations (NixOS), users, or aliases.",
	Long:  `This command lists installed packages for the user and system (NixOS) or system generations.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var itemToList string
		if len(args) == 0 {
			itemToList = "packages"
		} else {
			itemToList = args[0]
		}

		switch itemToList {
		case "packages":
			out, err := api.List()
			if err != nil {
				fmt.Println("Error listing packages:", err)
				os.Exit(1)
			}
			fmt.Println(out)
		case "generations":
			out, err := api.ListGenerations()
			if err != nil {
				fmt.Println("Error listing generations:", err)
				os.Exit(1)
			}
			fmt.Println(string(out))
		case "users":
			users, err := api.GetUsers()
			if err != nil {
				fmt.Println("Error getting users:", err)
				os.Exit(1)
			}
			for _, user := range users {
				fmt.Printf("Username: %s, Name: %s, Email: %s\n", user.Username, user.Name, user.Email)
			}
		case "aliases":
			aliases, err := api.GetAliases()
			if err != nil {
				fmt.Println("Error getting aliases:", err)
				os.Exit(1)
			}
			for name, command := range aliases {
				fmt.Printf("%s: %s\n", name, command)
			}
		default:
			fmt.Println("Invalid argument. Use 'packages', 'generations', 'users', or 'aliases'.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
