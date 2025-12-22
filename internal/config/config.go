package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL      string
	OpenAIAPIKey     string
	AnthropicAPIKey  string
	GeminiAPIKey     string
	LLMProvider      string
	Port             int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),
		LLMProvider:      getEnv("LLM_PROVIDER", "openai"),
		Port:             getEnvAsInt("PORT", 8080),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Validate LLM provider
	if cfg.LLMProvider != "openai" && cfg.LLMProvider != "anthropic" && cfg.LLMProvider != "gemini" {
		return nil, fmt.Errorf("LLM_PROVIDER must be 'openai', 'anthropic', or 'gemini'")
	}

	// Validate API key based on provider
	if cfg.LLMProvider == "openai" && cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required when LLM_PROVIDER is 'openai'")
	}
	if cfg.LLMProvider == "anthropic" && cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when LLM_PROVIDER is 'anthropic'")
	}
	if cfg.LLMProvider == "gemini" && cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when LLM_PROVIDER is 'gemini'")
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
