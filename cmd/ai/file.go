package ai

import (
	"github.com/spf13/cobra"
)

// fileCmd represents the file subcommand
var fileCmd = &cobra.Command{
	Use:   "file <file_path> <question>",
	Short: "Ask AI questions based on file content",
	Long:  `Use file content as context to ask AI model related questions.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		question := args[1]

		askWithFile(filePath, question)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)
}
