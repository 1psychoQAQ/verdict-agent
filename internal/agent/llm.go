package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Error types
var (
	ErrRateLimited = errors.New("rate limited")
	ErrTimeout     = errors.New("request timeout")
	ErrInvalidJSON = errors.New("invalid JSON in response")
)

// LLMClient defines the interface for interacting with LLM providers
type LLMClient interface {
	Complete(ctx context.Context, prompt string) (string, error)
	CompleteJSON(ctx context.Context, prompt string, result any) error
}

// Config holds the configuration for LLM clients
type Config struct {
	Provider   string        // "openai" or "anthropic"
	APIKey     string
	Model      string        // "gpt-4" or "claude-3-opus-20240229"
	MaxRetries int
	Timeout    time.Duration
}

// NewLLMClient creates a new LLM client based on the configuration
func NewLLMClient(cfg Config) (LLMClient, error) {
	// Set defaults
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}

	// Set default models based on provider
	if cfg.Model == "" {
		switch cfg.Provider {
		case "openai":
			cfg.Model = "gpt-4"
		case "anthropic":
			cfg.Model = "claude-3-opus-20240229"
		default:
			return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
		}
	}

	switch cfg.Provider {
	case "openai":
		return &openAIClient{
			config:     cfg,
			httpClient: &http.Client{Timeout: cfg.Timeout},
		}, nil
	case "anthropic":
		return &anthropicClient{
			config:     cfg,
			httpClient: &http.Client{Timeout: cfg.Timeout},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// openAIClient implements LLMClient for OpenAI
type openAIClient struct {
	config     Config
	httpClient *http.Client
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *openAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	reqBody := openAIRequest{
		Model: c.config.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt seconds
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		response, err := c.makeRequest(ctx, reqBody)
		if err != nil {
			lastErr = err
			// Retry on rate limit or timeout
			if errors.Is(err, ErrRateLimited) || errors.Is(err, ErrTimeout) {
				continue
			}
			return "", err
		}

		if len(response.Choices) == 0 {
			return "", errors.New("no response from OpenAI")
		}

		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *openAIClient) CompleteJSON(ctx context.Context, prompt string, result any) error {
	response, err := c.Complete(ctx, prompt)
	if err != nil {
		return err
	}

	jsonContent, err := extractJSON(response)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(jsonContent), result); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return nil
}

func (c *openAIClient) makeRequest(ctx context.Context, reqBody openAIRequest) (*openAIResponse, error) {
	return c.makeRequestWithURL(ctx, reqBody, "https://api.openai.com/v1/chat/completions")
}

func (c *openAIClient) makeRequestWithURL(ctx context.Context, reqBody openAIRequest, url string) (*openAIResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response openAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	return &response, nil
}

// anthropicClient implements LLMClient for Anthropic
type anthropicClient struct {
	config     Config
	httpClient *http.Client
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *anthropicClient) Complete(ctx context.Context, prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model: c.config.Model,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 4096,
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 2^attempt seconds
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		response, err := c.makeRequest(ctx, reqBody)
		if err != nil {
			lastErr = err
			// Retry on rate limit or timeout
			if errors.Is(err, ErrRateLimited) || errors.Is(err, ErrTimeout) {
				continue
			}
			return "", err
		}

		if len(response.Content) == 0 {
			return "", errors.New("no response from Anthropic")
		}

		return response.Content[0].Text, nil
	}

	return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *anthropicClient) CompleteJSON(ctx context.Context, prompt string, result any) error {
	response, err := c.Complete(ctx, prompt)
	if err != nil {
		return err
	}

	jsonContent, err := extractJSON(response)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(jsonContent), result); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return nil
}

func (c *anthropicClient) makeRequest(ctx context.Context, reqBody anthropicRequest) (*anthropicResponse, error) {
	return c.makeRequestWithURL(ctx, reqBody, "https://api.anthropic.com/v1/messages")
}

func (c *anthropicClient) makeRequestWithURL(ctx context.Context, reqBody anthropicRequest, url string) (*anthropicResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response anthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("Anthropic API error: %s", response.Error.Message)
	}

	return &response, nil
}

// extractJSON extracts JSON content from LLM response (between ```json and ```)
func extractJSON(response string) (string, error) {
	// Try to find JSON block marked with ```json
	jsonBlockRegex := regexp.MustCompile("(?s)```json\\s*\\n(.*?)\\n```")
	matches := jsonBlockRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		content := strings.TrimSpace(matches[1])
		// Verify it's valid JSON
		if json.Valid([]byte(content)) {
			return content, nil
		}
		// If not valid, continue to other patterns
	}

	// Try to find any code block
	codeBlockRegex := regexp.MustCompile("(?s)```\\s*\\n(.*?)\\n```")
	matches = codeBlockRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		content := strings.TrimSpace(matches[1])
		// Verify it's valid JSON
		if json.Valid([]byte(content)) {
			return content, nil
		}
	}

	// Try to find JSON object in the response
	jsonObjectRegex := regexp.MustCompile("(?s)\\{.*\\}")
	match := jsonObjectRegex.FindString(response)
	if match != "" && json.Valid([]byte(match)) {
		return match, nil
	}

	// Try to find JSON array in the response
	jsonArrayRegex := regexp.MustCompile("(?s)\\[.*\\]")
	match = jsonArrayRegex.FindString(response)
	if match != "" && json.Valid([]byte(match)) {
		return match, nil
	}

	return "", fmt.Errorf("%w: no valid JSON found in response", ErrInvalidJSON)
}
