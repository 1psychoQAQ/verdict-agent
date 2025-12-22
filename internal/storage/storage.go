package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Decision represents a stored decision with its verdict
type Decision struct {
	ID        uuid.UUID       `json:"id"`
	Input     string          `json:"input"`
	Verdict   json.RawMessage `json:"verdict"` // JSONB
	CreatedAt time.Time       `json:"created_at"`
	IsFinal   bool            `json:"is_final"`
}

// Todo represents a stored todo item linked to a decision
type Todo struct {
	ID         uuid.UUID `json:"id"`
	DecisionID uuid.UUID `json:"decision_id"`
	Content    string    `json:"content"` // Markdown content
	CreatedAt  time.Time `json:"created_at"`
}

// Repository defines the interface for data persistence operations
type Repository interface {
	// Decisions
	CreateDecision(ctx context.Context, d *Decision) error
	GetDecision(ctx context.Context, id uuid.UUID) (*Decision, error)

	// Todos
	CreateTodo(ctx context.Context, t *Todo) error
	GetTodo(ctx context.Context, id uuid.UUID) (*Todo, error)
	GetTodoByDecisionID(ctx context.Context, decisionID uuid.UUID) (*Todo, error)

	// Atomic operations
	SaveArtifacts(ctx context.Context, d *Decision, t *Todo) error

	// Health
	Ping(ctx context.Context) error

	// Cleanup
	Close()
}
