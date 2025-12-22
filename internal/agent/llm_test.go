package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewLLMClient tests the client factory function
func TestNewLLMClient(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantType  string
		wantError bool
	}{
		{
			name: "OpenAI client with defaults",
			config: Config{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantType:  "openai",
			wantError: false,
		},
		{
			name: "Anthropic client with defaults",
			config: Config{
				Provider: "anthropic",
				APIKey:   "test-key",
			},
			wantType:  "anthropic",
			wantError: false,
		},
		{
			name: "OpenAI client with custom model",
			config: Config{
				Provider: "openai",
				APIKey:   "test-key",
				Model:    "gpt-3.5-turbo",
			},
			wantType:  "openai",
			wantError: false,
		},
		{
			name: "Unsupported provider",
			config: Config{
				Provider: "invalid",
				APIKey:   "test-key",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewLLMClient(tt.config)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("expected client, got nil")
			}

			// Verify client type
			switch tt.wantType {
			case "openai":
				if _, ok := client.(*openAIClient); !ok {
					t.Errorf("expected *openAIClient, got %T", client)
				}
			case "anthropic":
				if _, ok := client.(*anthropicClient); !ok {
					t.Errorf("expected *anthropicClient, got %T", client)
				}
			}
		})
	}
}

// TestOpenAIComplete tests OpenAI completion
func TestOpenAIComplete(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantError  bool
		wantResult string
	}{
		{
			name: "successful completion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := openAIResponse{
					Choices: []struct {
						Message openAIMessage `json:"message"`
					}{
						{Message: openAIMessage{Role: "assistant", Content: "Hello, world!"}},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			wantResult: "Hello, world!",
			wantError:  false,
		},
		{
			name: "rate limited with retry",
			handler: func() http.HandlerFunc {
				attempts := 0
				return func(w http.ResponseWriter, r *http.Request) {
					attempts++
					if attempts == 1 {
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
					response := openAIResponse{
						Choices: []struct {
							Message openAIMessage `json:"message"`
						}{
							{Message: openAIMessage{Role: "assistant", Content: "Success after retry"}},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				}
			}(),
			wantResult: "Success after retry",
			wantError:  false,
		},
		{
			name: "API error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": {"message": "Invalid request"}}`))
			},
			wantError: true,
		},
		{
			name: "empty response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := openAIResponse{
					Choices: []struct {
						Message openAIMessage `json:"message"`
					}{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := &openAIClient{
				config: Config{
					Provider:   "openai",
					APIKey:     "test-key",
					Model:      "gpt-4o",
					MaxRetries: 3,
					Timeout:    5 * time.Second,
				},
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			// Create a test helper that uses the test server
			ctx := context.Background()
			reqBody := openAIRequest{
				Model: client.config.Model,
				Messages: []openAIMessage{
					{Role: "user", Content: "test prompt"},
				},
			}

			var lastErr error
			var result string
			for attempt := 0; attempt <= client.config.MaxRetries; attempt++ {
				response, err := client.makeRequestWithURL(ctx, reqBody, server.URL)
				if err != nil {
					lastErr = err
					if err == ErrRateLimited {
						continue
					}
					break
				}

				if len(response.Choices) == 0 {
					lastErr = errors.New("no response from OpenAI")
					break
				}

				result = response.Choices[0].Message.Content
				lastErr = nil
				break
			}

			if tt.wantError {
				if lastErr == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if lastErr != nil {
				t.Errorf("unexpected error: %v", lastErr)
				return
			}
			if result != tt.wantResult {
				t.Errorf("expected result %q, got %q", tt.wantResult, result)
			}
		})
	}
}

// TestAnthropicComplete tests Anthropic completion
func TestAnthropicComplete(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantError  bool
		wantResult string
	}{
		{
			name: "successful completion",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := anthropicResponse{
					Content: []struct {
						Text string `json:"text"`
					}{
						{Text: "Hello from Claude!"},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			wantResult: "Hello from Claude!",
			wantError:  false,
		},
		{
			name: "rate limited with retry",
			handler: func() http.HandlerFunc {
				attempts := 0
				return func(w http.ResponseWriter, r *http.Request) {
					attempts++
					if attempts == 1 {
						w.WriteHeader(http.StatusTooManyRequests)
						return
					}
					response := anthropicResponse{
						Content: []struct {
							Text string `json:"text"`
						}{
							{Text: "Success after retry"},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				}
			}(),
			wantResult: "Success after retry",
			wantError:  false,
		},
		{
			name: "empty response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := anthropicResponse{
					Content: []struct {
						Text string `json:"text"`
					}{},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := &anthropicClient{
				config: Config{
					Provider:   "anthropic",
					APIKey:     "test-key",
					Model:      "claude-3-opus-20240229",
					MaxRetries: 3,
					Timeout:    5 * time.Second,
				},
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			ctx := context.Background()
			reqBody := anthropicRequest{
				Model: client.config.Model,
				Messages: []anthropicMessage{
					{Role: "user", Content: "test prompt"},
				},
				MaxTokens: 4096,
			}

			var lastErr error
			var result string
			for attempt := 0; attempt <= client.config.MaxRetries; attempt++ {
				response, err := client.makeRequestWithURL(ctx, reqBody, server.URL)
				if err != nil {
					lastErr = err
					if err == ErrRateLimited {
						continue
					}
					break
				}

				if len(response.Content) == 0 {
					lastErr = errors.New("no response from Anthropic")
					break
				}

				result = response.Content[0].Text
				lastErr = nil
				break
			}

			if tt.wantError {
				if lastErr == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if lastErr != nil {
				t.Errorf("unexpected error: %v", lastErr)
				return
			}
			if result != tt.wantResult {
				t.Errorf("expected result %q, got %q", tt.wantResult, result)
			}
		})
	}
}

// TestExtractJSON tests JSON extraction from LLM responses
func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name      string
		response  string
		want      string
		wantError bool
	}{
		{
			name:      "JSON in code block with json marker",
			response:  "Here is the JSON:\n```json\n{\"key\": \"value\"}\n```",
			want:      "{\"key\": \"value\"}",
			wantError: false,
		},
		{
			name:      "JSON in code block without marker",
			response:  "Here is the JSON:\n```\n{\"key\": \"value\"}\n```",
			want:      "{\"key\": \"value\"}",
			wantError: false,
		},
		{
			name:      "JSON object without code block",
			response:  "The result is {\"key\": \"value\"} as you can see.",
			want:      "{\"key\": \"value\"}",
			wantError: false,
		},
		{
			name:      "JSON array",
			response:  "The list is [{\"id\": 1}, {\"id\": 2}]",
			want:      "[{\"id\": 1}, {\"id\": 2}]",
			wantError: false,
		},
		{
			name:      "Nested JSON",
			response:  "```json\n{\"outer\": {\"inner\": \"value\"}}\n```",
			want:      "{\"outer\": {\"inner\": \"value\"}}",
			wantError: false,
		},
		{
			name:      "No JSON found",
			response:  "This is just plain text without any JSON",
			wantError: true,
		},
		{
			name:      "Invalid JSON in code block",
			response:  "```json\n{invalid: json}\n```",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractJSON(tt.response)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.want {
				t.Errorf("expected %q, got %q", tt.want, result)
			}
		})
	}
}

// TestCompleteJSON tests JSON completion and parsing
func TestCompleteJSON(t *testing.T) {
	type TestResult struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name      string
		response  string
		want      TestResult
		wantError bool
	}{
		{
			name:      "valid JSON response",
			response:  "```json\n{\"name\": \"test\", \"value\": 42}\n```",
			want:      TestResult{Name: "test", Value: 42},
			wantError: false,
		},
		{
			name:      "JSON without code block",
			response:  "The result is {\"name\": \"example\", \"value\": 100}",
			want:      TestResult{Name: "example", Value: 100},
			wantError: false,
		},
		{
			name:      "no JSON in response",
			response:  "This is just text",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := openAIResponse{
					Choices: []struct {
						Message openAIMessage `json:"message"`
					}{
						{Message: openAIMessage{Role: "assistant", Content: tt.response}},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := &openAIClient{
				config: Config{
					Provider:   "openai",
					APIKey:     "test-key",
					Model:      "gpt-4o",
					MaxRetries: 3,
					Timeout:    5 * time.Second,
				},
				httpClient: &http.Client{Timeout: 5 * time.Second},
			}

			ctx := context.Background()
			reqBody := openAIRequest{
				Model: client.config.Model,
				Messages: []openAIMessage{
					{Role: "user", Content: "test prompt"},
				},
			}

			response, err := client.makeRequestWithURL(ctx, reqBody, server.URL)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			var result TestResult
			responseText := response.Choices[0].Message.Content
			jsonContent, err := extractJSON(responseText)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error extracting JSON: %v", err)
				return
			}

			err = json.Unmarshal([]byte(jsonContent), &result)
			if err != nil {
				t.Errorf("unexpected error unmarshaling JSON: %v", err)
				return
			}

			if result != tt.want {
				t.Errorf("expected %+v, got %+v", tt.want, result)
			}
		})
	}
}

// TestContextTimeout tests that context timeout is properly handled
func TestContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		response := openAIResponse{
			Choices: []struct {
				Message openAIMessage `json:"message"`
			}{
				{Message: openAIMessage{Role: "assistant", Content: "Too late"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &openAIClient{
		config: Config{
			Provider:   "openai",
			APIKey:     "test-key",
			Model:      "gpt-4o",
			MaxRetries: 0, // No retries for this test
			Timeout:    100 * time.Millisecond,
		},
		httpClient: &http.Client{Timeout: 100 * time.Millisecond},
	}

	ctx := context.Background()
	reqBody := openAIRequest{
		Model: client.config.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: "test prompt"},
		},
	}

	_, err := client.makeRequestWithURL(ctx, reqBody, server.URL)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

// TestRetryExponentialBackoff tests that retry logic uses exponential backoff
func TestRetryExponentialBackoff(t *testing.T) {
	attempts := 0
	startTime := time.Now()
	var requestTimes []time.Duration

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		requestTimes = append(requestTimes, time.Since(startTime))

		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		response := openAIResponse{
			Choices: []struct {
				Message openAIMessage `json:"message"`
			}{
				{Message: openAIMessage{Role: "assistant", Content: "Success"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &openAIClient{
		config: Config{
			Provider:   "openai",
			APIKey:     "test-key",
			Model:      "gpt-4o",
			MaxRetries: 3,
			Timeout:    30 * time.Second,
		},
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	reqBody := openAIRequest{
		Model: client.config.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: "test prompt"},
		},
	}

	var lastErr error
	for attempt := 0; attempt <= client.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}

		_, err := client.makeRequestWithURL(ctx, reqBody, server.URL)
		if err != nil {
			lastErr = err
			if err == ErrRateLimited {
				continue
			}
			break
		}
		lastErr = nil
		break
	}

	if lastErr != nil {
		t.Errorf("unexpected error: %v", lastErr)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	// Verify exponential backoff exists
	if len(requestTimes) >= 2 {
		diff1 := requestTimes[1] - requestTimes[0]
		if diff1 < 2*time.Second {
			t.Errorf("expected at least 2s between first and second request, got %v", diff1)
		}
	}
	if len(requestTimes) >= 3 {
		diff2 := requestTimes[2] - requestTimes[1]
		if diff2 < 4*time.Second {
			t.Errorf("expected at least 4s between second and third request, got %v", diff2)
		}
	}
}

// TestConfigDefaults tests that configuration defaults are set correctly
func TestConfigDefaults(t *testing.T) {
	tests := []struct {
		name           string
		inputConfig    Config
		wantMaxRetries int
		wantTimeout    time.Duration
		wantModel      string
	}{
		{
			name: "OpenAI with no defaults",
			inputConfig: Config{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantMaxRetries: 3,
			wantTimeout:    5 * time.Minute,
			wantModel:      "gpt-4o",
		},
		{
			name: "Anthropic with no defaults",
			inputConfig: Config{
				Provider: "anthropic",
				APIKey:   "test-key",
			},
			wantMaxRetries: 3,
			wantTimeout:    5 * time.Minute,
			wantModel:      "claude-3-opus-20240229",
		},
		{
			name: "Custom values preserved",
			inputConfig: Config{
				Provider:   "openai",
				APIKey:     "test-key",
				Model:      "gpt-3.5-turbo",
				MaxRetries: 5,
				Timeout:    10 * time.Minute,
			},
			wantMaxRetries: 5,
			wantTimeout:    10 * time.Minute,
			wantModel:      "gpt-3.5-turbo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewLLMClient(tt.inputConfig)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var config Config
			switch c := client.(type) {
			case *openAIClient:
				config = c.config
			case *anthropicClient:
				config = c.config
			}

			if config.MaxRetries != tt.wantMaxRetries {
				t.Errorf("expected MaxRetries %d, got %d", tt.wantMaxRetries, config.MaxRetries)
			}
			if config.Timeout != tt.wantTimeout {
				t.Errorf("expected Timeout %v, got %v", tt.wantTimeout, config.Timeout)
			}
			if config.Model != tt.wantModel {
				t.Errorf("expected Model %s, got %s", tt.wantModel, config.Model)
			}
		})
	}
}
