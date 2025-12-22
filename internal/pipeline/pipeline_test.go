package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
)

// Mock LLM client for testing
type mockLLMClient struct {
	verdictResponse   *agent.VerdictOutput
	executionResponse *agent.ExecutionOutput
	verdictError      error
	executionError    error
	delay             time.Duration
	callCount         int
}

func (m *mockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	// Not used in pipeline tests, but required by interface
	return "", nil
}

func (m *mockLLMClient) CompleteJSON(ctx context.Context, prompt string, result interface{}) error {
	m.callCount++

	// Simulate delay if configured
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Check if this is a verdict call or execution call based on the result type
	switch v := result.(type) {
	case *agent.VerdictOutput:
		if m.verdictError != nil {
			return m.verdictError
		}
		if m.verdictResponse != nil {
			*v = *m.verdictResponse
		}
	case *agent.ExecutionOutput:
		if m.executionError != nil {
			return m.executionError
		}
		if m.executionResponse != nil {
			*v = *m.executionResponse
		}
	}

	return nil
}

func TestNewPipeline(t *testing.T) {
	client := &mockLLMClient{}
	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)

	t.Run("with valid timeout", func(t *testing.T) {
		timeout := 5 * time.Minute
		p := NewPipeline(verdictAgent, executionAgent, timeout)
		if p == nil {
			t.Fatal("expected non-nil pipeline")
		}
		if p.timeout != timeout {
			t.Errorf("expected timeout %v, got %v", timeout, p.timeout)
		}
	})

	t.Run("with zero timeout uses default", func(t *testing.T) {
		p := NewPipeline(verdictAgent, executionAgent, 0)
		if p.timeout != 10*time.Minute {
			t.Errorf("expected default timeout 10m, got %v", p.timeout)
		}
	})

	t.Run("with negative timeout uses default", func(t *testing.T) {
		p := NewPipeline(verdictAgent, executionAgent, -1*time.Second)
		if p.timeout != 10*time.Minute {
			t.Errorf("expected default timeout 10m, got %v", p.timeout)
		}
	})
}

func TestPipeline_Execute_Success(t *testing.T) {
	client := &mockLLMClient{
		verdictResponse: &agent.VerdictOutput{
			Ruling:    "Build a REST API",
			Rationale: "REST APIs are simple and widely supported",
			Rejected: []agent.RejectedOption{
				{Option: "GraphQL", Reason: "Too complex for MVP"},
			},
		},
		executionResponse: &agent.ExecutionOutput{
			MVPScope: []string{"User authentication", "Basic CRUD"},
			Phases: []agent.Phase{
				{
					Name:  "Phase 1: Setup",
					Tasks: []string{"Create project", "Setup database"},
				},
			},
			DoneCriteria: []string{"API responds to requests", "Database stores data"},
		},
	}

	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

	result, err := p.Execute(context.Background(), "Should I build a REST API or GraphQL API?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Verdict == nil {
		t.Fatal("expected non-nil verdict")
	}
	if result.Verdict.Ruling != "Build a REST API" {
		t.Errorf("unexpected ruling: %s", result.Verdict.Ruling)
	}

	if result.Execution == nil {
		t.Fatal("expected non-nil execution")
	}
	if len(result.Execution.Phases) == 0 {
		t.Error("expected at least one phase")
	}

	if result.Duration == 0 {
		t.Error("expected non-zero duration")
	}
}

func TestPipeline_Execute_InputValidation(t *testing.T) {
	client := &mockLLMClient{}
	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "empty input",
			input:       "",
			expectedErr: ErrInputEmpty,
		},
		{
			name:        "whitespace only input",
			input:       "   \n\t  ",
			expectedErr: ErrInputEmpty,
		},
		{
			name:        "input too long",
			input:       strings.Repeat("a", 10001),
			expectedErr: ErrInputTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Execute(context.Background(), tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestPipeline_Execute_VerdictValidation(t *testing.T) {
	tests := []struct {
		name            string
		verdictResponse *agent.VerdictOutput
		expectError     bool
		errorContains   string
	}{
		{
			name:            "nil verdict",
			verdictResponse: nil,
			expectError:     true,
			errorContains:   "verdict agent failed",
		},
		{
			name: "empty ruling",
			verdictResponse: &agent.VerdictOutput{
				Ruling:    "",
				Rationale: "Some rationale",
			},
			expectError:   true,
			errorContains: "ruling is empty",
		},
		{
			name: "empty rationale",
			verdictResponse: &agent.VerdictOutput{
				Ruling:    "Some ruling",
				Rationale: "",
			},
			expectError:   true,
			errorContains: "rationale is empty",
		},
		{
			name: "whitespace ruling",
			verdictResponse: &agent.VerdictOutput{
				Ruling:    "   ",
				Rationale: "Some rationale",
			},
			expectError:   true,
			errorContains: "ruling is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockLLMClient{
				verdictResponse: tt.verdictResponse,
			}
			verdictAgent := agent.NewVerdictAgent(client)
			executionAgent := agent.NewExecutionAgent(client)
			p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

			_, err := p.Execute(context.Background(), "valid input")
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPipeline_Execute_ExecutionValidation(t *testing.T) {
	validVerdict := &agent.VerdictOutput{
		Ruling:    "Build a REST API",
		Rationale: "REST APIs are simple",
	}

	tests := []struct {
		name              string
		executionResponse *agent.ExecutionOutput
		expectError       bool
		errorContains     string
	}{
		{
			name:              "nil execution",
			executionResponse: nil,
			expectError:       true,
			errorContains:     "execution agent failed",
		},
		{
			name: "no phases",
			executionResponse: &agent.ExecutionOutput{
				MVPScope:     []string{"Feature 1"},
				Phases:       []agent.Phase{},
				DoneCriteria: []string{"Done"},
			},
			expectError:   true,
			errorContains: "no phases",
		},
		{
			name: "phases with no tasks",
			executionResponse: &agent.ExecutionOutput{
				MVPScope: []string{"Feature 1"},
				Phases: []agent.Phase{
					{Name: "Phase 1", Tasks: []string{}},
				},
				DoneCriteria: []string{"Done"},
			},
			expectError:   true,
			errorContains: "no tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockLLMClient{
				verdictResponse:   validVerdict,
				executionResponse: tt.executionResponse,
			}
			verdictAgent := agent.NewVerdictAgent(client)
			executionAgent := agent.NewExecutionAgent(client)
			p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

			_, err := p.Execute(context.Background(), "valid input")
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPipeline_Execute_Timeout(t *testing.T) {
	client := &mockLLMClient{
		delay: 2 * time.Second, // Simulate slow LLM
		verdictResponse: &agent.VerdictOutput{
			Ruling:    "Build a REST API",
			Rationale: "REST APIs are simple",
		},
	}

	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 500*time.Millisecond)

	_, err := p.Execute(context.Background(), "Should I build a REST API?")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestPipeline_Execute_ContextCancellation(t *testing.T) {
	client := &mockLLMClient{
		delay: 2 * time.Second,
		verdictResponse: &agent.VerdictOutput{
			Ruling:    "Build a REST API",
			Rationale: "REST APIs are simple",
		},
	}

	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := p.Execute(ctx, "Should I build a REST API?")
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestPipeline_Execute_VerdictAgentError(t *testing.T) {
	client := &mockLLMClient{
		verdictError: errors.New("LLM service unavailable"),
	}

	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

	_, err := p.Execute(context.Background(), "valid input")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrVerdictFailed) {
		t.Errorf("expected ErrVerdictFailed, got %v", err)
	}
}

func TestPipeline_Execute_ExecutionAgentError(t *testing.T) {
	client := &mockLLMClient{
		verdictResponse: &agent.VerdictOutput{
			Ruling:    "Build a REST API",
			Rationale: "REST APIs are simple",
		},
		executionError: errors.New("LLM service unavailable"),
	}

	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

	_, err := p.Execute(context.Background(), "valid input")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrExecutionFailed) {
		t.Errorf("expected ErrExecutionFailed, got %v", err)
	}
}

func TestPipeline_Execute_DurationTracking(t *testing.T) {
	client := &mockLLMClient{
		delay: 100 * time.Millisecond,
		verdictResponse: &agent.VerdictOutput{
			Ruling:    "Build a REST API",
			Rationale: "REST APIs are simple",
		},
		executionResponse: &agent.ExecutionOutput{
			MVPScope: []string{"Feature 1"},
			Phases: []agent.Phase{
				{Name: "Phase 1", Tasks: []string{"Task 1"}},
			},
			DoneCriteria: []string{"Done"},
		},
	}

	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

	result, err := p.Execute(context.Background(), "valid input")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should track duration (at least 200ms due to 2 calls with 100ms delay each)
	if result.Duration < 200*time.Millisecond {
		t.Errorf("expected duration >= 200ms, got %v", result.Duration)
	}
}

func TestPipeline_Execute_InputMaxLength(t *testing.T) {
	client := &mockLLMClient{
		verdictResponse: &agent.VerdictOutput{
			Ruling:    "Build a REST API",
			Rationale: "REST APIs are simple",
		},
		executionResponse: &agent.ExecutionOutput{
			MVPScope: []string{"Feature 1"},
			Phases: []agent.Phase{
				{Name: "Phase 1", Tasks: []string{"Task 1"}},
			},
			DoneCriteria: []string{"Done"},
		},
	}

	verdictAgent := agent.NewVerdictAgent(client)
	executionAgent := agent.NewExecutionAgent(client)
	p := NewPipeline(verdictAgent, executionAgent, 1*time.Minute)

	// Test exactly 10000 characters (should succeed)
	input := strings.Repeat("a", 10000)
	_, err := p.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error for 10000 chars: %v", err)
	}

	// Test 10001 characters (should fail)
	input = strings.Repeat("a", 10001)
	_, err = p.Execute(context.Background(), input)
	if !errors.Is(err, ErrInputTooLong) {
		t.Errorf("expected ErrInputTooLong for 10001 chars, got %v", err)
	}
}
