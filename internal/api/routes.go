package api

import (
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/artifact"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/1psychoQAQ/verdict-agent/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterConfig holds configuration for the API router
type RouterConfig struct {
	Pipeline    *pipeline.Pipeline
	Generator   *artifact.Generator
	Repository  storage.Repository
	RateLimit   int           // Requests per minute per IP (default: 10)
	Timeout     time.Duration // Request timeout (default: 10 minutes)
	CORSConfig  CORSConfig
}

// DefaultRouterConfig returns a default router configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		RateLimit:  10,
		Timeout:    10 * time.Minute,
		CORSConfig: DefaultCORSConfig(),
	}
}

// NewRouter creates a new chi router with all API routes configured
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Apply defaults
	if cfg.RateLimit == 0 {
		cfg.RateLimit = 10
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Minute
	}

	// Create handlers
	handlers := NewHandlers(cfg.Pipeline, cfg.Generator, cfg.Repository)

	// Create rate limiter (10 requests per minute per IP)
	rateLimiter := NewRateLimiter(cfg.RateLimit, time.Minute)

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.Timeout))
	r.Use(CORS(cfg.CORSConfig))

	// Health check (not rate limited)
	r.Get("/health", handlers.HealthHandler)

	// API routes with rate limiting
	r.Route("/api", func(r chi.Router) {
		// Apply rate limiting to API routes
		r.Use(RateLimitMiddleware(rateLimiter))

		// POST /api/verdict - Submit idea, receive verdict + todo
		r.Post("/verdict", handlers.VerdictHandler)

		// GET /api/decisions/{id} - Retrieve decision by ID
		r.Get("/decisions/{id}", handlers.GetDecisionHandler)

		// GET /api/todos/{id} - Retrieve todo by ID
		r.Get("/todos/{id}", handlers.GetTodoHandler)
	})

	return r
}
