package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/pokitpeng/ai/pkg/models"
)

func main() {
	// Get API key from environment variable
	os.Setenv("OPENAI_API_KEY", "sk-LyjXB90w2ENzg3h9DgEacd6zzbmeHQmKns2QA8ZPVhHeZq74")
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the OPENAI_API_KEY environment variable")
	}

	// Create model configuration
	config := &models.ModelConfig{
		Name:   "Qwen/Qwen2.5-3B-Instruct",
		URL:    "http://106.75.1.116:32635",
		APIKey: apiKey,
	}

	// Create model
	model := models.NewOpenAIModel(config)

	// Create context
	ctx := context.Background()

	// Example 1: Basic question
	question := "Write a simple HTTP server in Go"
	fmt.Println("Question:", question)
	fmt.Println("Requesting answer...")

	answer, err := model.Chat(ctx, question)
	if err != nil {
		log.Fatalf("Question failed: %v", err)
	}

	fmt.Println("\nAnswer:")
	fmt.Println(answer)

	// Example 2: Question with file
	fileQuestion := "Explain the functionality of this code and suggest improvements"
	fileName := "example.go"
	fileContent := `package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	
	go func() {
		time.Sleep(2 * time.Second)
		ch <- "Hello, World!"
	}()
	
	msg := <-ch
	fmt.Println(msg)
}`

	fmt.Println("\n\nFile question:", fileQuestion)
	fmt.Println("File:", fileName)
	fmt.Println("File content:")
	fmt.Println(fileContent)
	fmt.Println("\nRequesting answer...")

	fileAnswer, err := model.ChatWithFile(ctx, fileQuestion, fileName, fileContent)
	if err != nil {
		log.Fatalf("File question failed: %v", err)
	}

	fmt.Println("\nAnswer:")
	fmt.Println(fileAnswer)
}
