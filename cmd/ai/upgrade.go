package ai

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade subcommand
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade AI terminal tool",
	Long:  `Check for updates and upgrade the AI terminal tool to the latest version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Upgrading AI terminal tool...")

		// Execute go install command
		upgradeCmd := exec.Command("go", "install", "github.com/pokitpeng/ai@latest")
		upgradeCmd.Stdout = os.Stdout
		upgradeCmd.Stderr = os.Stderr

		if err := upgradeCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Upgrade failed: %v\n", err)
			return
		}

		fmt.Println("AI terminal tool has been upgraded to the latest version!")
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
