package api

import (
	"net/http"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/artifact"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/1psychoQAQ/verdict-agent/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterConfig holds configuration for the API router
type RouterConfig struct {
	Pipeline           *pipeline.Pipeline
	Generator          *artifact.Generator
	Repository         storage.Repository
	MemoryRepository   *storage.MemoryRepository // In-memory repo for auth/history
	ClarificationAgent *agent.ClarificationAgent // Optional: enables clarification flow
	RateLimit          int                       // Requests per minute per IP (default: 10)
	Timeout            time.Duration             // Request timeout (default: 10 minutes)
	CORSConfig         CORSConfig
	StaticFS           http.FileSystem  // Filesystem for serving static files
	IndexHandler       http.HandlerFunc // Handler for serving index.html
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

	// Create handlers (with or without clarification support)
	var handlers *Handlers
	if cfg.ClarificationAgent != nil {
		handlers = NewHandlersWithClarification(cfg.Pipeline, cfg.Generator, cfg.Repository, cfg.ClarificationAgent)
	} else {
		handlers = NewHandlers(cfg.Pipeline, cfg.Generator, cfg.Repository)
	}

	// Set memory repository for history tracking
	if cfg.MemoryRepository != nil {
		handlers.memoryRepo = cfg.MemoryRepository
	}

	// Create auth handlers if memory repo is available
	var authHandlers *AuthHandlers
	if cfg.MemoryRepository != nil {
		authHandlers = NewAuthHandlers(cfg.MemoryRepository)
	}

	// Create rate limiter (10 requests per minute per IP)
	rateLimiter := NewRateLimiter(cfg.RateLimit, time.Minute)

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.Timeout))
	r.Use(CORS(cfg.CORSConfig))

	// Add auth middleware if memory repo is available
	if cfg.MemoryRepository != nil {
		r.Use(AuthMiddleware(cfg.MemoryRepository))
	}

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

		// Auth routes (if auth handlers available)
		if authHandlers != nil {
			r.Route("/auth", func(r chi.Router) {
				r.Post("/register", authHandlers.RegisterHandler)
				r.Post("/login", authHandlers.LoginHandler)
				r.Post("/logout", authHandlers.LogoutHandler)
				r.Get("/me", authHandlers.MeHandler)
			})

			// History routes (require authentication)
			r.Route("/history", func(r chi.Router) {
				r.Use(RequireAuth)
				r.Get("/", authHandlers.GetHistoryHandler)
				r.Get("/{id}", authHandlers.GetHistoryByIDHandler)
				r.Put("/{id}/criteria", authHandlers.UpdateDoneCriteriaHandler)
			})
		}
	})

	// Static files and frontend (if configured)
	if cfg.StaticFS != nil {
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(cfg.StaticFS)))
	}
	if cfg.IndexHandler != nil {
		r.Get("/", cfg.IndexHandler)
	}

	return r
}
