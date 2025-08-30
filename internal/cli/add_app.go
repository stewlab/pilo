package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var addAppCmd = &cobra.Command{
	Use:   "add-app [pname] [version] [url] [sha256]",
	Short: "Adds a new Flake App.",
	Long:  `This command adds a new Flake App to your flake/packages directory.`,
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		app := api.App{
			Pname:   args[0],
			Version: args[1],
			URL:     args[2],
			Sha256:  args[3],
		}
		if err := api.AddApp(app); err != nil {
			fmt.Println("Error adding app:", err)
			os.Exit(1)
		}
		fmt.Printf("App '%s' added successfully.\n", app.Pname)
	},
}

func init() {
	rootCmd.AddCommand(addAppCmd)
}
