package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements the Repository interface using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// Config holds the database connection configuration
type Config struct {
	DatabaseURL     string
	MaxConns        int32         // default 10
	MinConns        int32         // default 2
	MaxConnLifetime time.Duration // default 1 hour
}

// NewPostgresRepository creates a new PostgreSQL repository with connection pooling
func NewPostgresRepository(ctx context.Context, databaseURL string) (*PostgresRepository, error) {
	config := &Config{
		DatabaseURL:     databaseURL,
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
	}
	return NewPostgresRepositoryWithConfig(ctx, config)
}

// NewPostgresRepositoryWithConfig creates a new PostgreSQL repository with custom configuration
func NewPostgresRepositoryWithConfig(ctx context.Context, cfg *Config) (*PostgresRepository, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set connection pool limits
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

// CreateDecision inserts a new decision into the database
func (r *PostgresRepository) CreateDecision(ctx context.Context, d *Decision) error {
	if d == nil {
		return fmt.Errorf("decision cannot be nil")
	}

	query := `
		INSERT INTO decisions (id, input, verdict, created_at, is_final)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Generate UUID if not provided
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	// Set created_at if not provided
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query, d.ID, d.Input, d.Verdict, d.CreatedAt, d.IsFinal)
	if err != nil {
		return fmt.Errorf("failed to create decision: %w", err)
	}

	return nil
}

// GetDecision retrieves a decision by its ID
func (r *PostgresRepository) GetDecision(ctx context.Context, id uuid.UUID) (*Decision, error) {
	query := `
		SELECT id, input, verdict, created_at, is_final
		FROM decisions
		WHERE id = $1
	`

	var d Decision
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID,
		&d.Input,
		&d.Verdict,
		&d.CreatedAt,
		&d.IsFinal,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("decision not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	return &d, nil
}

// CreateTodo inserts a new todo into the database
func (r *PostgresRepository) CreateTodo(ctx context.Context, t *Todo) error {
	if t == nil {
		return fmt.Errorf("todo cannot be nil")
	}

	query := `
		INSERT INTO todos (id, decision_id, content, created_at)
		VALUES ($1, $2, $3, $4)
	`

	// Generate UUID if not provided
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	// Set created_at if not provided
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}

	_, err := r.pool.Exec(ctx, query, t.ID, t.DecisionID, t.Content, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	return nil
}

// GetTodo retrieves a todo by its ID
func (r *PostgresRepository) GetTodo(ctx context.Context, id uuid.UUID) (*Todo, error) {
	query := `
		SELECT id, decision_id, content, created_at
		FROM todos
		WHERE id = $1
	`

	var t Todo
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.DecisionID,
		&t.Content,
		&t.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("todo not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return &t, nil
}

// GetTodoByDecisionID retrieves a todo by its associated decision ID
func (r *PostgresRepository) GetTodoByDecisionID(ctx context.Context, decisionID uuid.UUID) (*Todo, error) {
	query := `
		SELECT id, decision_id, content, created_at
		FROM todos
		WHERE decision_id = $1
	`

	var t Todo
	err := r.pool.QueryRow(ctx, query, decisionID).Scan(
		&t.ID,
		&t.DecisionID,
		&t.Content,
		&t.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("todo not found for decision: %w", err)
		}
		return nil, fmt.Errorf("failed to get todo by decision ID: %w", err)
	}

	return &t, nil
}

// SaveArtifacts atomically saves both a decision and its associated todo
func (r *PostgresRepository) SaveArtifacts(ctx context.Context, d *Decision, t *Todo) error {
	if d == nil || t == nil {
		return fmt.Errorf("decision and todo cannot be nil")
	}

	// Begin transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Generate UUIDs if not provided
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	// Set created_at if not provided
	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}

	// Ensure todo references the decision
	t.DecisionID = d.ID

	// Insert decision
	decisionQuery := `
		INSERT INTO decisions (id, input, verdict, created_at, is_final)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, decisionQuery, d.ID, d.Input, d.Verdict, d.CreatedAt, d.IsFinal)
	if err != nil {
		return fmt.Errorf("failed to insert decision in transaction: %w", err)
	}

	// Insert todo
	todoQuery := `
		INSERT INTO todos (id, decision_id, content, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, todoQuery, t.ID, t.DecisionID, t.Content, t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert todo in transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Ping checks if the database connection is healthy
func (r *PostgresRepository) Ping(ctx context.Context) error {
	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// Close closes the database connection pool
func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}
