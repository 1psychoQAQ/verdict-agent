package storage_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/1psychoQAQ/verdict-agent/internal/storage"
	"github.com/google/uuid"
)

// This example demonstrates basic usage of the PostgreSQL repository
func Example_basicUsage() {
	_ = context.Background()

	// This would normally come from environment variable
	// For this example, we show the connection string format
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	// Create repository (would fail without real database, so we skip in tests)
	_ = databaseURL

	// Example decision data
	decision := &storage.Decision{
		Input:   "Should we implement feature X?",
		Verdict: json.RawMessage(`{"action":"approve","confidence":0.85}`),
		IsFinal: true,
	}

	// Example todo data
	todo := &storage.Todo{
		Content: "# Implementation Tasks\n- Setup database\n- Create API endpoints\n- Add tests",
	}

	fmt.Printf("Decision input: %s\n", decision.Input)
	fmt.Printf("Todo content length: %d\n", len(todo.Content))
	// Output:
	// Decision input: Should we implement feature X?
	// Todo content length: 74
}

// This example demonstrates atomic save of decision and todo
func Example_atomicSave() {
	// Example showing the SaveArtifacts pattern
	decision := &storage.Decision{
		Input:   "Feature request analysis",
		Verdict: json.RawMessage(`{"approved":true,"priority":"high"}`),
		IsFinal: true,
	}

	_ = &storage.Todo{
		Content: "# Tasks\n- Design API\n- Implement handlers\n- Write tests",
	}

	// In real usage, this would save both in a transaction
	// SaveArtifacts ensures decision and todo are linked via decision_id
	fmt.Printf("Decision will be saved with input: %s\n", decision.Input)
	fmt.Printf("Todo will be linked to decision\n")
	// Output:
	// Decision will be saved with input: Feature request analysis
	// Todo will be linked to decision
}

// This example demonstrates repository configuration
func Example_configuration() {
	cfg := &storage.Config{
		DatabaseURL:     "postgres://localhost/mydb",
		MaxConns:        10, // Maximum number of connections
		MinConns:        2,  // Minimum number of connections
		MaxConnLifetime: 3600000000000, // 1 hour in nanoseconds
	}

	fmt.Printf("Max connections: %d\n", cfg.MaxConns)
	fmt.Printf("Min connections: %d\n", cfg.MinConns)
	// Output:
	// Max connections: 10
	// Min connections: 2
}

// This example shows how to use the repository in a real application
func ExampleNewPostgresRepository() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/verdict_db?sslmode=disable"

	// Create repository
	repo, err := storage.NewPostgresRepository(ctx, databaseURL)
	if err != nil {
		log.Printf("Failed to create repository: %v", err)
		return
	}
	defer repo.Close()

	// Check connection health
	if err := repo.Ping(ctx); err != nil {
		log.Printf("Database unhealthy: %v", err)
		return
	}

	// Create a decision
	decision := &storage.Decision{
		ID:      uuid.New(),
		Input:   "Evaluate proposal",
		Verdict: json.RawMessage(`{"decision":"approved"}`),
		IsFinal: true,
	}

	if err := repo.CreateDecision(ctx, decision); err != nil {
		log.Printf("Failed to create decision: %v", err)
		return
	}

	// Create associated todo
	todo := &storage.Todo{
		ID:         uuid.New(),
		DecisionID: decision.ID,
		Content:    "# Next Steps\n- Review code\n- Deploy to staging",
	}

	if err := repo.CreateTodo(ctx, todo); err != nil {
		log.Printf("Failed to create todo: %v", err)
		return
	}

	log.Printf("Successfully saved decision %s and todo %s", decision.ID, todo.ID)
}

// This example shows atomic save of both decision and todo
func ExamplePostgresRepository_SaveArtifacts() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/verdict_db?sslmode=disable"

	repo, err := storage.NewPostgresRepository(ctx, databaseURL)
	if err != nil {
		log.Printf("Failed to create repository: %v", err)
		return
	}
	defer repo.Close()

	// Prepare decision and todo
	decision := &storage.Decision{
		Input:   "Project feasibility analysis",
		Verdict: json.RawMessage(`{"feasible":true,"risk":"low"}`),
		IsFinal: true,
	}

	todo := &storage.Todo{
		Content: "# Project Tasks\n- Create project structure\n- Setup CI/CD\n- Write documentation",
	}

	// Save both atomically - if either fails, both are rolled back
	if err := repo.SaveArtifacts(ctx, decision, todo); err != nil {
		log.Printf("Failed to save artifacts: %v", err)
		return
	}

	log.Printf("Successfully saved decision %s with todo %s", decision.ID, todo.ID)
}
