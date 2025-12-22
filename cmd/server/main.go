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
	"github.com/1psychoQAQ/verdict-agent/internal/storage"
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
	}
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	// Initialize agents
	verdictAgent := agent.NewVerdictAgent(llmClient)
	executionAgent := agent.NewExecutionAgent(llmClient)

	// Initialize pipeline
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 10*time.Minute)

	// Initialize artifact generator
	generator := artifact.NewGenerator()

	// Create router with configuration
	routerCfg := api.RouterConfig{
		Pipeline:   p,
		Generator:  generator,
		Repository: repo,
		RateLimit:  10,
		Timeout:    10 * time.Minute,
		CORSConfig: api.DefaultCORSConfig(),
	}
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

// startHealthOnlyServer starts a minimal server with just the health check endpoint
func startHealthOnlyServer(port int) {
	router := api.NewRouter(api.RouterConfig{
		RateLimit: 10,
		Timeout:   10 * time.Minute,
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting health-only server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
		os.Exit(1)
	}
}
