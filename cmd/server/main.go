package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/api"
	"github.com/1psychoQAQ/verdict-agent/internal/artifact"
	"github.com/1psychoQAQ/verdict-agent/internal/config"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/1psychoQAQ/verdict-agent/internal/search"
	"github.com/1psychoQAQ/verdict-agent/internal/storage"
	"github.com/1psychoQAQ/verdict-agent/web"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Configuration validation failed: %v", err)
		log.Printf("Starting server with limited functionality (health check only)")
		startHealthOnlyServer(8080)
		return
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	repo, err := storage.NewPostgresRepository(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	// Initialize LLM client
	var llmClient agent.LLMClient
	switch cfg.LLMProvider {
	case "openai":
		llmClient, err = agent.NewLLMClient(agent.Config{
			Provider: "openai",
			APIKey:   cfg.OpenAIAPIKey,
		})
	case "anthropic":
		llmClient, err = agent.NewLLMClient(agent.Config{
			Provider: "anthropic",
			APIKey:   cfg.AnthropicAPIKey,
		})
	case "gemini":
		llmClient, err = agent.NewLLMClient(agent.Config{
			Provider: "gemini",
			APIKey:   cfg.GeminiAPIKey,
		})
	}
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	// Initialize agents
	verdictAgent := agent.NewVerdictAgent(llmClient)
	executionAgent := agent.NewExecutionAgent(llmClient)
	clarificationAgent := agent.NewClarificationAgent(llmClient)

	// Initialize search client (optional)
	var searchClient search.Client
	if cfg.SearchEnabled && cfg.SearchProvider != "" {
		var searchAPIKey string
		switch cfg.SearchProvider {
		case "tavily":
			searchAPIKey = cfg.TavilyAPIKey
		case "google":
			searchAPIKey = cfg.GoogleSearchKey
		}

		searchClient, err = search.NewClient(search.Config{
			Provider: cfg.SearchProvider,
			APIKey:   searchAPIKey,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize search client: %v (continuing without search)", err)
			searchClient = nil
		} else {
			log.Printf("Search enabled with provider: %s", cfg.SearchProvider)
		}
	}

	// Initialize pipeline with search
	p := pipeline.NewPipelineWithSearch(verdictAgent, executionAgent, searchClient, 10*time.Minute)

	// Initialize artifact generator
	generator := artifact.NewGenerator()

	// Create router with configuration
	routerCfg := api.RouterConfig{
		Pipeline:           p,
		Generator:          generator,
		Repository:         repo,
		ClarificationAgent: clarificationAgent,
		RateLimit:          10,
		Timeout:            10 * time.Minute,
		CORSConfig:         api.DefaultCORSConfig(),
		StaticFS:           web.StaticFS(),
		IndexHandler:       web.IndexHandler(),
	}
	log.Println("Clarification mode enabled")
	router := api.NewRouter(routerCfg)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 10*time.Minute + 15*time.Second, // Pipeline timeout + buffer
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// startHealthOnlyServer starts a minimal server with just the health check endpoint and frontend
func startHealthOnlyServer(port int) {
	router := api.NewRouter(api.RouterConfig{
		RateLimit:    10,
		Timeout:      10 * time.Minute,
		StaticFS:     web.StaticFS(),
		IndexHandler: web.IndexHandler(),
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting health-only server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}
