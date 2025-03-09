package ai

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Start a new conversation session",
	Long:  `Start a fresh conversation session.`,
	Run: func(cmd *cobra.Command, args []string) {
		historyManager.New()
		fmt.Println("Started a new session.")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
