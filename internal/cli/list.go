package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [packages|generations]",
	Short: "Lists installed packages or system generations (NixOS).",
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
		default:
			fmt.Println("Invalid argument. Use 'packages' or 'generations'.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
