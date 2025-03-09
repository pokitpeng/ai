package models

import (
	"context"
	"fmt"
	"strings"
)

// Factory function for creating model instances
func CreateModel(config *ModelConfig) (Model, error) {
	// Determine model type based on name or URL characteristics
	modelType := determineModelType(config.Name, config.URL)

	switch modelType {
	case "openai":
		return NewOpenAIModel(config), nil
	case "anthropic":
		return NewAnthropicModel(config), nil
	default:
		// Use generic model by default
		return NewOpenAIModel(config), nil
	}
}

// Determine model type based on name and URL
func determineModelType(name, url string) string {
	name = strings.ToLower(name)

	// Determine model type based on name
	if strings.Contains(name, "openai") || strings.Contains(name, "gpt") {
		return "openai"
	}
	if strings.Contains(name, "anthropic") || strings.Contains(name, "claude") {
		return "anthropic"
	}

	// Determine model type based on URL
	if strings.Contains(url, "openai.com") {
		return "openai"
	}

	// Default to openai model
	return "openai"
}

// Base model implementation
type baseModel struct {
	config *ModelConfig
}

func (m *baseModel) Name() string {
	return m.config.Name
}

// OpenAIModel implementation
type OpenAIModel struct {
	baseModel
}

func NewOpenAIModel(config *ModelConfig) *OpenAIModel {
	return &OpenAIModel{
		baseModel: baseModel{config: config},
	}
}

// OpenAIModel's Chat and ChatWithFile methods are implemented in openai.go

// AnthropicModel implementation
type AnthropicModel struct {
	baseModel
}

func NewAnthropicModel(config *ModelConfig) *AnthropicModel {
	return &AnthropicModel{
		baseModel: baseModel{config: config},
	}
}

func (m *AnthropicModel) Chat(ctx context.Context, question string, options ...ChatOption) (string, error) {
	// Apply default options from model config if available
	var opts *ChatOptions
	if m.config.DefaultChatOptions != nil {
		// Create a copy of default options
		defaultOpts := *m.config.DefaultChatOptions
		opts = &defaultOpts
	} else {
		// Use global defaults
		opts = DefaultChatOptions()
	}

	// Apply user-provided options
	for _, option := range options {
		option(opts)
	}

	// Implement actual Anthropic API call here
	return fmt.Sprintf("[Anthropic] Response to: %s", question), nil
}

func (m *AnthropicModel) ChatWithFile(ctx context.Context, question string, fileName string, fileContent string, options ...ChatOption) (string, error) {
	// Apply default options from model config if available
	var opts *ChatOptions
	if m.config.DefaultChatOptions != nil {
		// Create a copy of default options
		defaultOpts := *m.config.DefaultChatOptions
		opts = &defaultOpts
	} else {
		// Use global defaults
		opts = DefaultChatOptions()
	}

	// Apply user-provided options
	for _, option := range options {
		option(opts)
	}

	// Implement actual Anthropic API call here
	return fmt.Sprintf("[Anthropic] Response to file %s question: %s", fileName, question), nil
}
