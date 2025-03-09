package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pokitpeng/ai/pkg/models"
)

func main() {
	// Set API key
	os.Setenv("OPENAI_API_KEY", "sk-I99gFcEj0KF99J8Oan9uiSxnt2TW6k9oVvan7lm5OAI0jhK6")
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the OPENAI_API_KEY environment variable")
	}

	// Create model configuration
	config := &models.ModelConfig{
		Name:   "deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5",
		URL:    "http://106.75.1.116:30622",
		APIKey: apiKey,
	}

	// Create model
	model := models.NewOpenAIModel(config)

	// Create context
	ctx := context.Background()

	// Example: Streaming response
	question := "Please write a 500-word article about the history of artificial intelligence in Chinese, with paragraphs."
	fmt.Println("Question:", question)
	fmt.Println("\nStart streaming output (real-time display):")

	// Use streaming options
	startTime := time.Now()
	answer, err := model.Chat(ctx, question, models.WithStream(true))
	if err != nil {
		log.Fatalf("Failed to ask: %v", err)
	}
	duration := time.Since(startTime)

	fmt.Printf("\n\nCompleted! Time: %.2f seconds, Total characters: %d\n", duration.Seconds(), len([]rune(answer))) // Use rune to calculate Chinese characters
}
