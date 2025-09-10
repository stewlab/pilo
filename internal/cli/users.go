package cli

import (
	"fmt"
	"os"

	"pilo/internal/api"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  `Manage users in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := api.GetUsers()
		if err != nil {
			fmt.Println("Error getting users:", err)
			os.Exit(1)
		}
		for _, user := range users {
			fmt.Printf("Username: %s, Name: %s, Email: %s\n", user.Username, user.Name, user.Email)
		}
	},
}

var addUserCmd = &cobra.Command{
	Use:   "add [username] [name] [email]",
	Short: "Add a new user",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.AddUser(args[0], args[1], args[2]); err != nil {
			fmt.Println("Error adding user:", err)
			os.Exit(1)
		}
		fmt.Println("User added successfully.")
	},
}

var removeUserCmd = &cobra.Command{
	Use:   "remove [username]",
	Short: "Remove a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.RemoveUser(args[0]); err != nil {
			fmt.Println("Error removing user:", err)
			os.Exit(1)
		}
		fmt.Println("User removed successfully.")
	},
}

var updateUserCmd = &cobra.Command{
	Use:   "update [old_username] [new_username] [name] [email]",
	Short: "Update a user",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		if err := api.UpdateUser(args[0], args[1], args[2], args[3]); err != nil {
			fmt.Println("Error updating user:", err)
			os.Exit(1)
		}
		fmt.Println("User updated successfully.")
	},
}

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(listUsersCmd)
	usersCmd.AddCommand(addUserCmd)
	usersCmd.AddCommand(removeUserCmd)
	usersCmd.AddCommand(updateUserCmd)
}
