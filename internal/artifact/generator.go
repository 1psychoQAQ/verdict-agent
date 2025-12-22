package artifact

import (
	"fmt"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/google/uuid"
)

// Generator generates decision and todo artifacts from pipeline results
type Generator struct{}

// Artifacts contains the generated decision.json and todo.md artifacts
type Artifacts struct {
	DecisionJSON []byte
	TodoMD       []byte
	ID           uuid.UUID
	CreatedAt    time.Time
}

// NewGenerator creates a new artifact generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates both decision.json and todo.md artifacts from pipeline result
// Generation is atomic - both artifacts are created or an error is returned
func (g *Generator) Generate(result *pipeline.PipelineResult) (*Artifacts, error) {
	if result == nil {
		return nil, fmt.Errorf("pipeline result cannot be nil")
	}
	if result.Verdict == nil {
		return nil, fmt.Errorf("verdict output cannot be nil")
	}
	if result.Execution == nil {
		return nil, fmt.Errorf("execution output cannot be nil")
	}

	// Generate UUID and timestamp for both artifacts
	id := uuid.New()
	createdAt := time.Now()

	// Generate decision.json
	decisionJSON, err := generateDecisionJSON(result.Input, result.Verdict, id, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate decision.json: %w", err)
	}

	// Generate todo.md
	todoMD, err := generateTodoMD(result.Verdict, result.Execution, id, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate todo.md: %w", err)
	}

	// Both artifacts generated successfully - atomic operation complete
	return &Artifacts{
		DecisionJSON: decisionJSON,
		TodoMD:       todoMD,
		ID:           id,
		CreatedAt:    createdAt,
	}, nil
}
