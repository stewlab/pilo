package cli

import (
	"pilo/internal/gui"

	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launches the Fyne GUI for pilo.",
	Long:  `This command launches the Fyne GUI for pilo.`,
	Run: func(cmd *cobra.Command, args []string) {
		gui.Run() // This should now correctly call the public Run function
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
