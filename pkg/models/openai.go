package models

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAI API response structure
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// OpenAI API request structure
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Choice represents a choice returned by the API
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents the usage of the API call
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIClient implements the OpenAI API client
type OpenAIClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
	model      string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(modelConfig ModelConfig) *OpenAIClient {
	// Create HTTP client, set timeout
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	return &OpenAIClient{
		apiKey:     modelConfig.APIKey,
		apiURL:     modelConfig.URL,
		httpClient: httpClient,
		model:      modelConfig.Name,
	}
}

// SetModel sets the model to use
func (c *OpenAIClient) SetModel(model string) {
	c.model = model
}

// Chat sends a chat request
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (string, error) {
	// Prepare request
	req := OpenAIRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stream:      opts.Stream,
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}

	// Create HTTP request
	apiURL := c.apiURL
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "v1/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set request headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	// Handle stream response
	if opts.Stream {
		return c.handleStreamResponse(resp.Body)
	}

	// Handle normal response
	return c.handleNormalResponse(resp.Body)
}

// handleNormalResponse handles normal responses
func (c *OpenAIClient) handleNormalResponse(respBody io.Reader) (string, error) {
	var apiResp OpenAIResponse

	// Parse response
	if err := json.NewDecoder(respBody).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if there are choices
	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("API returned empty response")
	}

	// Return the content of the first choice
	return apiResp.Choices[0].Message.Content, nil
}

// handleStreamResponse handles stream responses
func (c *OpenAIClient) handleStreamResponse(respBody io.Reader) (string, error) {
	// Use bufio.Scanner to read line by line in SSE format
	scanner := bufio.NewScanner(respBody)
	var fullContent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if it's a data line
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract data part
		data := strings.TrimPrefix(line, "data: ")

		// Check if it's the end marker
		if data == "[DONE]" {
			break
		}

		// Parse JSON data
		var chunk struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Role    string `json:"role,omitempty"`
					Content string `json:"content,omitempty"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// Parse error, skip this line
			continue
		}

		// Extract content and add to result
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				fullContent.WriteString(content)
				// Print content in real time
				fmt.Print(content)
			}
		}
	}

	// Check if there was an error during scanning
	if err := scanner.Err(); err != nil {
		return fullContent.String(), fmt.Errorf("error scanning stream response: %w", err)
	}

	// Output newline, making subsequent output more pretty
	fmt.Println()

	return fullContent.String(), nil
}

// Enhance OpenAIModel implementation
func (m *OpenAIModel) Chat(ctx context.Context, question string, options ...ChatOption) (string, error) {
	// Apply options
	opts := DefaultChatOptions()
	for _, option := range options {
		option(opts)
	}

	// Create messages array
	messages := []Message{}

	// Add history messages if provided
	if len(opts.History) > 0 {
		messages = append(messages, opts.History...)
	}

	// Add current question
	messages = append(messages, Message{
		Role:    "user",
		Content: question,
	})

	// Send to API
	client := NewOpenAIClient(ModelConfig{
		Name:   m.config.Name,
		URL:    m.config.URL,
		APIKey: m.config.APIKey,
	})
	return client.Chat(ctx, messages, opts)
}

// Enhance OpenAIModel's file question implementation
func (m *OpenAIModel) ChatWithFile(ctx context.Context, question string, fileName string, fileContent string, options ...ChatOption) (string, error) {
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

	// Create client
	client := NewOpenAIClient(ModelConfig{
		Name:   m.config.Name,
		URL:    m.config.URL,
		APIKey: m.config.APIKey,
	})

	// Build prompt with file content
	prompt := fmt.Sprintf("file name: %s\n\nfile content:\n%s\n\nquestion: %s", fileName, fileContent, question)

	// Create messages
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Send request
	return client.Chat(ctx, messages, opts)
}
