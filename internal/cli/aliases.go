package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var aliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "Manage aliases",
	Long:  `Manage aliases in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var addAliasCmd = &cobra.Command{
	Use:   "add [name] [command]",
	Short: "Add a new alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.AddAlias(args[0], args[1]); err != nil {
			fmt.Println("Error adding alias:", err)
			os.Exit(1)
		}
		fmt.Println("Alias added successfully.")
	},
}

var removeAliasCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.RemoveAlias(args[0]); err != nil {
			fmt.Println("Error removing alias:", err)
			os.Exit(1)
		}
		fmt.Println("Alias removed successfully.")
	},
}

var updateAliasCmd = &cobra.Command{
	Use:   "update [old_name] [new_name] [command]",
	Short: "Update an alias",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.UpdateAlias(args[0], args[1], args[2]); err != nil {
			fmt.Println("Error updating alias:", err)
			os.Exit(1)
		}
		fmt.Println("Alias updated successfully.")
	},
}

var duplicateAliasCmd = &cobra.Command{
	Use:   "duplicate [name] [command]",
	Short: "Duplicate an alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.DuplicateAlias(args[0], args[1]); err != nil {
			fmt.Println("Error duplicating alias:", err)
			os.Exit(1)
		}
		fmt.Println("Alias duplicated successfully.")
	},
}

func init() {
	rootCmd.AddCommand(aliasesCmd)
	aliasesCmd.AddCommand(addAliasCmd)
	aliasesCmd.AddCommand(removeAliasCmd)
	aliasesCmd.AddCommand(updateAliasCmd)
	aliasesCmd.AddCommand(duplicateAliasCmd)
}
