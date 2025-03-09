package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIClient_Chat(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Verify request header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type to be application/json, got %s", r.Header.Get("Content-Type"))
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("Expected Authorization header to start with Bearer, got %s", r.Header.Get("Authorization"))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var req OpenAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to parse request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify request parameters
		if req.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model to be gpt-3.5-turbo, got %s", req.Model)
		}

		if len(req.Messages) == 0 {
			t.Errorf("Expected at least one message")
			http.Error(w, "No messages provided", http.StatusBadRequest)
			return
		}

		// Return mock response
		resp := OpenAIResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1677858242,
			Model:   "gpt-3.5-turbo-0613",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "this is a test response",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}

		// If streaming request, return SSE format
		if req.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			// Mock streaming response
			chunk := struct {
				ID      string `json:"id"`
				Object  string `json:"object"`
				Created int64  `json:"created"`
				Model   string `json:"model"`
				Choices []struct {
					Index int `json:"index"`
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}{
				ID:      "chatcmpl-123",
				Object:  "chat.completion.chunk",
				Created: 1677858242,
				Model:   "gpt-3.5-turbo-0613",
				Choices: []struct {
					Index int `json:"index"`
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				}{
					{
						Index: 0,
						Delta: struct {
							Content string `json:"content"`
						}{
							Content: "this is a test response",
						},
						FinishReason: nil,
					},
				},
			}

			chunkData, _ := json.Marshal(chunk)
			w.Write([]byte("data: " + string(chunkData) + "\n\n"))
			w.Write([]byte("data: [DONE]\n\n"))
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client := NewOpenAIClient(ModelConfig{
		Name:   "test-openai",
		URL:    server.URL,
		APIKey: "test-api-key",
	})

	// Test normal chat
	t.Run("Normal chat", func(t *testing.T) {
		messages := []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		}

		opts := &ChatOptions{
			Temperature: 0.7,
			MaxTokens:   100,
			Stream:      false,
		}

		resp, err := client.Chat(context.Background(), messages, opts)
		if err != nil {
			t.Fatalf("Chat request failed: %v", err)
		}

		expected := "this is a test response"
		if resp != expected {
			t.Errorf("Expected response to be %s, got %s", expected, resp)
		}
	})

	// Test streaming chat
	t.Run("Streaming chat", func(t *testing.T) {
		messages := []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		}

		opts := &ChatOptions{
			Temperature: 0.7,
			MaxTokens:   100,
			Stream:      true,
		}

		resp, err := client.Chat(context.Background(), messages, opts)
		if err != nil {
			t.Fatalf("Streaming chat request failed: %v", err)
		}

		expected := "this is a test response"
		if resp != expected {
			t.Errorf("Expected response to be %s, got %s", expected, resp)
		}
	})
}

func TestOpenAIModel_Ask(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return mock response
		resp := OpenAIResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1677858242,
			Model:   "gpt-3.5-turbo-0613",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "this is a test response",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create model
	config := &ModelConfig{
		Name:   "test-openai",
		URL:    server.URL,
		APIKey: "test-api-key",
	}

	model := NewOpenAIModel(config)

	// Test Chat method
	resp, err := model.Chat(context.Background(), "test question")
	if err != nil {
		t.Fatalf("Chat method failed: %v", err)
	}

	expected := "this is a test response"
	if resp != expected {
		t.Errorf("Expected response to be %s, got %s", expected, resp)
	}
}

func TestOpenAIModel_AskWithFile(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var req OpenAIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to parse request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// 验证请求中包含文件内容
		if len(req.Messages) == 0 || !strings.Contains(req.Messages[0].Content, "test.go") {
			t.Errorf("Request should contain file name")
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Return mock response
		resp := OpenAIResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1677858242,
			Model:   "gpt-3.5-turbo-0613",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "this is a test response",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     30,
				CompletionTokens: 20,
				TotalTokens:      50,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create model
	config := &ModelConfig{
		Name:   "test-openai",
		URL:    server.URL,
		APIKey: "test-api-key",
	}

	model := NewOpenAIModel(config)

	// Test ChatWithFile method
	resp, err := model.ChatWithFile(context.Background(), "explain this code", "test.go", "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}")
	if err != nil {
		t.Fatalf("ChatWithFile method failed: %v", err)
	}

	expected := "this is a test response"
	if resp != expected {
		t.Errorf("Expected response to be %s, got %s", expected, resp)
	}
}

func TestHandleStreamResponse(t *testing.T) {
	// Create mock SSE format response
	sseResponse := `data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"这是"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"一个"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"流式"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"响应"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"测试"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677858242,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
`

	// Create a reader
	reader := strings.NewReader(sseResponse)

	// Create client
	client := NewOpenAIClient(ModelConfig{
		Name:   "test-openai",
		URL:    "https://api.openai.com/v1/chat/completions",
		APIKey: "test-api-key",
	})

	// Call handleStreamResponse
	result, err := client.handleStreamResponse(reader)
	if err != nil {
		t.Fatalf("Failed to handle stream response: %v", err)
	}

	// Verify result
	expected := "this is a test response"
	if result != expected {
		t.Errorf("Expected result to be %q, got %q", expected, result)
	}
}

// 测试错误处理
func TestHandleStreamResponseError(t *testing.T) {
	// Create a reader that will produce an error
	errorReader := &errorReader{err: fmt.Errorf("mock read error")}

	// Create client
	client := NewOpenAIClient(ModelConfig{
		Name:   "test-openai",
		URL:    "https://api.openai.com/v1/chat/completions",
		APIKey: "test-api-key",
	})

	// Call handleStreamResponse
	_, err := client.handleStreamResponse(errorReader)

	// Verify if an error is returned
	if err == nil {
		t.Error("Expected error, but got nil")
	}
}

// Reader for testing error cases
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
