package ai

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// multiCmd represents the multi subcommand
var multiCmd = &cobra.Command{
	Use:   "multi <model1,model2,...> <question>",
	Short: "Ask multiple models simultaneously",
	Long:  `Ask the same question to multiple models simultaneously and compare their answers.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse model list
		modelList := strings.Split(args[0], ",")
		if len(modelList) < 1 {
			fmt.Println("Please specify at least one model")
			return
		}

		// Question
		question := args[1]

		// Execute multi-model questioning
		askMultiModels(modelList, question)
	},
}

func init() {
	rootCmd.AddCommand(multiCmd)
}
