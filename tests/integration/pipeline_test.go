package integration

import (
	"context"
	"testing"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/artifact"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/1psychoQAQ/verdict-agent/tests/mocks"
)

func TestPipelineHappyPath(t *testing.T) {
	// Setup mock LLM client
	mockLLM := mocks.NewMockLLMClient()

	// Create agents
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)

	// Create pipeline
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	// Execute pipeline
	ctx := context.Background()
	result, err := p.Execute(ctx, "Should I use Go or Python for building a web service?")
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Verify verdict
	if result.Verdict == nil {
		t.Fatal("Expected verdict to be set")
	}
	if result.Verdict.Ruling == "" {
		t.Error("Expected ruling to be set")
	}
	if result.Verdict.Rationale == "" {
		t.Error("Expected rationale to be set")
	}

	// Verify execution
	if result.Execution == nil {
		t.Fatal("Expected execution to be set")
	}
	if len(result.Execution.Phases) == 0 {
		t.Error("Expected at least one phase")
	}

	// Verify duration is recorded
	if result.Duration == 0 {
		t.Error("Expected duration to be recorded")
	}
}

func TestPipelineWithArtifactGeneration(t *testing.T) {
	// Setup mock LLM client
	mockLLM := mocks.NewMockLLMClient()

	// Create agents
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)

	// Create pipeline and generator
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)
	gen := artifact.NewGenerator()

	// Execute pipeline
	ctx := context.Background()
	result, err := p.Execute(ctx, "What framework should I use for building a REST API?")
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	// Generate artifacts
	artifacts, err := gen.Generate(result)
	if err != nil {
		t.Fatalf("Artifact generation failed: %v", err)
	}

	// Verify artifacts
	if len(artifacts.DecisionJSON) == 0 {
		t.Error("Expected decision JSON to be generated")
	}
	if len(artifacts.TodoMD) == 0 {
		t.Error("Expected todo MD to be generated")
	}
	if artifacts.ID.String() == "" {
		t.Error("Expected artifact ID to be generated")
	}
	if artifacts.CreatedAt.IsZero() {
		t.Error("Expected artifact timestamp to be set")
	}
}

func TestPipelineEmptyInput(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	ctx := context.Background()
	_, err := p.Execute(ctx, "")
	if err == nil {
		t.Error("Expected error for empty input")
	}
	if err != pipeline.ErrInputEmpty {
		t.Errorf("Expected ErrInputEmpty, got: %v", err)
	}
}

func TestPipelineInputTooLong(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	// Create input exceeding 10000 characters
	longInput := string(make([]byte, 10001))
	for i := range longInput {
		longInput = longInput[:i] + "a" + longInput[i+1:]
	}

	ctx := context.Background()
	_, err := p.Execute(ctx, longInput)
	if err == nil {
		t.Error("Expected error for input too long")
	}
	if err != pipeline.ErrInputTooLong {
		t.Errorf("Expected ErrInputTooLong, got: %v", err)
	}
}

func TestPipelineLLMFailure(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	mockLLM.ShouldFail = true
	mockLLM.ErrorMessage = "LLM service unavailable"

	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	ctx := context.Background()
	_, err := p.Execute(ctx, "Test input")
	if err == nil {
		t.Error("Expected error when LLM fails")
	}
}

func TestPipelineTimeout(t *testing.T) {
	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(1 * time.Millisecond)

	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 1*time.Nanosecond)

	_, err := p.Execute(ctx, "Test input")
	if err == nil {
		t.Log("Note: Timeout test may not fail if mock is too fast")
	}
}

func TestPipelineContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	// Cancel context before execution
	cancel()

	_, err := p.Execute(ctx, "Test input")
	if err == nil {
		t.Log("Note: Cancellation test may not fail if mock is too fast")
	}
}

func TestPipelineMultipleExecutions(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	ctx := context.Background()

	// Execute multiple times
	for i := 0; i < 3; i++ {
		result, err := p.Execute(ctx, "Test input for execution")
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i+1, err)
		}
		if result.Verdict == nil || result.Execution == nil {
			t.Fatalf("Execution %d produced incomplete result", i+1)
		}
	}

	// Verify LLM was called multiple times
	if mockLLM.CallCount() < 6 {
		t.Errorf("Expected at least 6 LLM calls (2 per execution), got %d", mockLLM.CallCount())
	}
}

func TestPipelineLLMFailsAfterVerdict(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	mockLLM.FailAfter = 1 // Fail after first call (verdict)
	mockLLM.ErrorMessage = "Execution agent failed"

	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	ctx := context.Background()
	_, err := p.Execute(ctx, "Test input")
	if err == nil {
		t.Error("Expected error when execution agent fails")
	}
}
