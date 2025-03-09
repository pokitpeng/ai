package models

import (
	"context"
)

// Model represents an AI model interface
type Model interface {
	// Name returns the model name
	Name() string

	// Chat sends a question to the model and returns the answer
	Chat(ctx context.Context, question string, options ...ChatOption) (string, error)

	// ChatWithFile sends a question with file content to the model
	ChatWithFile(ctx context.Context, question string, fileName string, fileContent string, options ...ChatOption) (string, error)
}

// ModelConfig stores model configuration
type ModelConfig struct {
	Name               string       `json:"name" yaml:"name"`
	URL                string       `json:"url" yaml:"url"`
	APIKey             string       `json:"api_key" yaml:"api_key"`
	DefaultEnabled     bool         `json:"default_enabled" yaml:"default_enabled"`
	DefaultChatOptions *ChatOptions `json:"default_chat_options" yaml:"default_chat_options"`
}

// ChatOption represents a chat option function
type ChatOption func(*ChatOptions)

// ChatOptions represents a collection of chat options
type ChatOptions struct {
	Temperature float64
	MaxTokens   int
	Stream      bool
	History     []Message
}

// WithTemperature sets the temperature parameter
func WithTemperature(temp float64) ChatOption {
	return func(o *ChatOptions) {
		o.Temperature = temp
	}
}

// WithMaxTokens sets the maximum number of tokens
func WithMaxTokens(tokens int) ChatOption {
	return func(o *ChatOptions) {
		o.MaxTokens = tokens
	}
}

// WithStream sets whether to use streaming output
func WithStream(stream bool) ChatOption {
	return func(o *ChatOptions) {
		o.Stream = stream
	}
}

// WithHistory adds conversation history to the chat options
func WithHistory(messages []Message) ChatOption {
	return func(o *ChatOptions) {
		o.History = messages
	}
}

// DefaultChatOptions returns default chat options
func DefaultChatOptions() *ChatOptions {
	return &ChatOptions{
		Temperature: 0.2,
		MaxTokens:   4096,
		Stream:      true,
	}
}
