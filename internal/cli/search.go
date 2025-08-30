package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"
	"pilo/internal/spinner"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Searches for packages in nixpkgs.",
	Long:  `This command searches for packages in nixpkgs.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		spinner := spinner.NewSpinner("Searching for packages...")
		spinner.Start()
		sortByPopularity, _ := cmd.Flags().GetBool("sort-by-popularity")
		freeOnly, _ := cmd.Flags().GetBool("free-only")
		out, err := api.Search(args, sortByPopularity, freeOnly)
		spinner.Stop()
		if err != nil {
			fmt.Println("Error searching for packages:", err)
			os.Exit(1)
		}
		for _, pkg := range out {
			fmt.Printf("%s - %s\n", pkg.Name, pkg.Description)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolP("sort-by-popularity", "p", false, "Sort results by popularity")
	searchCmd.Flags().BoolP("free-only", "f", false, "Show only free software")
}
