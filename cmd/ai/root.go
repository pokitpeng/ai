package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pokitpeng/ai/pkg/history"
	"github.com/pokitpeng/ai/pkg/models"
	"github.com/pokitpeng/ai/pkg/util"
	"github.com/spf13/cobra"
)

var (
	modelManager   *models.ModelManager
	historyManager *history.Manager
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI Terminal Tool - Use AI large language models in the command line",
	Long: `AI Terminal Tool is a command line tool that allows programmers to interact directly with AI large language models in the terminal.
Supports multiple models, file input, and querying multiple models simultaneously.

Examples:
  ai "How to implement quicksort algorithm?"
  ai file main.go "Explain what this code does"
  ai model list
  ai multi openai,anthropic "What is functional programming?"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Direct question mode
		question := args[0]

		// Get default model
		model, err := modelManager.GetDefaultModel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Create context
		ctx := context.Background()

		// Get history if needed
		var chatOptions []models.ChatOption
		noHistory, _ := cmd.Flags().GetBool("no-history")

		if !noHistory && !historyManager.IsEmpty() {
			// Convert history to model messages
			modelMessages := convertToModelMessages(historyManager.GetMessages())
			chatOptions = append(chatOptions, models.WithHistory(modelMessages))
		}

		// Send question with options
		response, err := model.Chat(ctx, question, chatOptions...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Add to history
		historyManager.AddUserMessage(question)
		historyManager.AddAssistantMessage(response)

		// Print response, remove this line to disable response printing
		// fmt.Println(response)
	},
}

// Initialization function
func init() {
	// Create and initialize model manager
	modelManager = models.NewModelManager()
	if err := modelManager.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize model manager: %v\n", err)
	}

	// Create and initialize history manager
	homeDir, _ := os.UserHomeDir()
	historyPath := filepath.Join(homeDir, ".ai", "history")
	var err error
	historyManager, err = history.NewManager(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize history manager: %v\n", err)
	}

	// Add flags
	rootCmd.PersistentFlags().Bool("no-history", false, "Don't use conversation history")
}

// Execute executes the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// askWithFile asks a question based on file content
func askWithFile(filePath, question string) {
	// Get file content
	content, language, err := util.GetFileInfo(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		return
	}

	// Get default model
	model, err := modelManager.GetDefaultModel()
	if err != nil {
		fmt.Println("No default model set. Please add a model first:")
		fmt.Println("  ai model add <model> <url> <apikey>")
		return
	}

	// Create context
	ctx := context.Background()

	// Execute question
	resp, err := model.ChatWithFile(ctx, question, filePath, content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Question failed: %v\n", err)
		return
	}

	// Print file info and response
	fmt.Printf("File: %s (%s)\n", filePath, language)
	fmt.Printf("Question: %s\n\n", question)
	fmt.Println(resp)
}

// askMultiModels asks multiple models simultaneously
func askMultiModels(modelNames []string, question string) {
	var wg sync.WaitGroup
	responsesCh := make(chan struct {
		modelName string
		response  string
		err       error
	}, len(modelNames))

	// Create context
	ctx := context.Background()

	// Ask all models in parallel
	for _, name := range modelNames {
		wg.Add(1)
		go func(modelName string) {
			defer wg.Done()

			model, err := modelManager.GetModel(modelName)
			if err != nil {
				responsesCh <- struct {
					modelName string
					response  string
					err       error
				}{modelName, "", err}
				return
			}

			// Disable streaming output
			resp, err := model.Chat(ctx, question, models.WithStream(false))

			responsesCh <- struct {
				modelName string
				response  string
				err       error
			}{modelName, resp, err}
		}(name)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(responsesCh)
	}()

	// Collect and display results
	fmt.Printf("Question: %s\n\n", question)

	for resp := range responsesCh {
		fmt.Printf("===== Model: %s =====\n", resp.modelName)
		if resp.err != nil {
			fmt.Printf("Error: %v\n", resp.err)
		} else {
			fmt.Println(resp.response)
		}
		fmt.Println()
	}
}

// convertToModelMessages converts history messages to model messages
func convertToModelMessages(historyMessages []history.Message) []models.Message {
	modelMessages := make([]models.Message, len(historyMessages))
	for i, msg := range historyMessages {
		modelMessages[i] = models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return modelMessages
}
