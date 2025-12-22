package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// mockExecutionLLMClient is a mock implementation of LLMClient for execution agent testing
type mockExecutionLLMClient struct {
	completeFunc     func(ctx context.Context, prompt string) (string, error)
	completeJSONFunc func(ctx context.Context, prompt string, result any) error
}

func (m *mockExecutionLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, prompt)
	}
	return "", nil
}

func (m *mockExecutionLLMClient) CompleteJSON(ctx context.Context, prompt string, result any) error {
	if m.completeJSONFunc != nil {
		return m.completeJSONFunc(ctx, prompt, result)
	}
	return nil
}

func TestNewExecutionAgent(t *testing.T) {
	client := &mockExecutionLLMClient{}
	agent := NewExecutionAgent(client)

	if agent == nil {
		t.Fatal("expected agent to be created")
	}

	if agent.client != client {
		t.Error("expected client to be set")
	}
}

func TestExecutionAgent_Process_Success(t *testing.T) {
	mockResponse := ExecutionOutput{
		MVPScope: []string{
			"Basic user authentication",
			"Simple dashboard view",
		},
		Phases: []Phase{
			{
				Name: "Foundation",
				Tasks: []string{
					"Set up database schema",
					"Implement user model",
					"Create auth endpoints",
				},
			},
			{
				Name: "UI Layer",
				Tasks: []string{
					"Build login page",
					"Build dashboard",
				},
			},
		},
		DoneCriteria: []string{
			"User can register and login",
			"Dashboard displays user data",
			"All tests pass",
		},
	}

	client := &mockExecutionLLMClient{
		completeJSONFunc: func(ctx context.Context, prompt string, result any) error {
			// Marshal and unmarshal to simulate real behavior
			data, _ := json.Marshal(mockResponse)
			return json.Unmarshal(data, result)
		},
	}

	agent := NewExecutionAgent(client)
	verdict := &VerdictOutput{
		Ruling:    "Implement basic user authentication system",
		Rationale: "User authentication is the foundation for all other features",
		Rejected:  []RejectedOption{},
		Ranking:   []int{1, 2, 3},
	}

	output, err := agent.Process(context.Background(), verdict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.MVPScope) != 2 {
		t.Errorf("expected 2 MVP scope items, got %d", len(output.MVPScope))
	}

	if len(output.Phases) != 2 {
		t.Errorf("expected 2 phases, got %d", len(output.Phases))
	}

	if len(output.DoneCriteria) != 3 {
		t.Errorf("expected 3 done criteria, got %d", len(output.DoneCriteria))
	}
}

func TestExecutionAgent_Process_NilVerdict(t *testing.T) {
	client := &mockExecutionLLMClient{}
	agent := NewExecutionAgent(client)

	_, err := agent.Process(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil verdict")
	}

	if err.Error() != "verdict cannot be nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecutionAgent_Process_LLMError(t *testing.T) {
	client := &mockExecutionLLMClient{
		completeJSONFunc: func(ctx context.Context, prompt string, result any) error {
			return errors.New("LLM API error")
		},
	}

	agent := NewExecutionAgent(client)
	verdict := &VerdictOutput{
		Ruling:    "Test ruling",
		Rationale: "Test rationale",
	}

	_, err := agent.Process(context.Background(), verdict)
	if err == nil {
		t.Fatal("expected error from LLM failure")
	}
}

func TestExecutionAgent_ValidateOutput_TooManyPhases(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope: []string{"Feature 1"},
		Phases: []Phase{
			{Name: "Phase 1", Tasks: []string{"Task 1"}},
			{Name: "Phase 2", Tasks: []string{"Task 1"}},
			{Name: "Phase 3", Tasks: []string{"Task 1"}},
			{Name: "Phase 4", Tasks: []string{"Task 1"}},
		},
		DoneCriteria: []string{"Criterion 1"},
	}

	err := agent.validateOutput(output)
	if err == nil {
		t.Fatal("expected error for too many phases")
	}
}

func TestExecutionAgent_ValidateOutput_TooManyTasks(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope: []string{"Feature 1"},
		Phases: []Phase{
			{
				Name: "Phase 1",
				Tasks: []string{
					"Task 1",
					"Task 2",
					"Task 3",
					"Task 4",
					"Task 5",
					"Task 6",
				},
			},
		},
		DoneCriteria: []string{"Criterion 1"},
	}

	err := agent.validateOutput(output)
	if err == nil {
		t.Fatal("expected error for too many tasks")
	}
}

func TestExecutionAgent_ValidateOutput_NoPhases(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope:     []string{"Feature 1"},
		Phases:       []Phase{},
		DoneCriteria: []string{"Criterion 1"},
	}

	err := agent.validateOutput(output)
	if err == nil {
		t.Fatal("expected error for no phases")
	}
}

func TestExecutionAgent_ValidateOutput_PhaseWithNoTasks(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope: []string{"Feature 1"},
		Phases: []Phase{
			{Name: "Phase 1", Tasks: []string{}},
		},
		DoneCriteria: []string{"Criterion 1"},
	}

	err := agent.validateOutput(output)
	if err == nil {
		t.Fatal("expected error for phase with no tasks")
	}
}

func TestExecutionAgent_ValidateOutput_PhaseWithNoName(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope: []string{"Feature 1"},
		Phases: []Phase{
			{Name: "", Tasks: []string{"Task 1"}},
		},
		DoneCriteria: []string{"Criterion 1"},
	}

	err := agent.validateOutput(output)
	if err == nil {
		t.Fatal("expected error for phase with no name")
	}
}

func TestExecutionAgent_ValidateOutput_NoMVPScope(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope: []string{},
		Phases: []Phase{
			{Name: "Phase 1", Tasks: []string{"Task 1"}},
		},
		DoneCriteria: []string{"Criterion 1"},
	}

	err := agent.validateOutput(output)
	if err == nil {
		t.Fatal("expected error for no MVP scope")
	}
}

func TestExecutionAgent_ValidateOutput_NoDoneCriteria(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope: []string{"Feature 1"},
		Phases: []Phase{
			{Name: "Phase 1", Tasks: []string{"Task 1"}},
		},
		DoneCriteria: []string{},
	}

	err := agent.validateOutput(output)
	if err == nil {
		t.Fatal("expected error for no done criteria")
	}
}

func TestExecutionAgent_ValidateOutput_ValidOutput(t *testing.T) {
	agent := &ExecutionAgent{}
	output := &ExecutionOutput{
		MVPScope: []string{"Feature 1", "Feature 2"},
		Phases: []Phase{
			{
				Name:  "Phase 1",
				Tasks: []string{"Task 1", "Task 2"},
			},
			{
				Name:  "Phase 2",
				Tasks: []string{"Task 3"},
			},
		},
		DoneCriteria: []string{"Criterion 1", "Criterion 2"},
	}

	err := agent.validateOutput(output)
	if err != nil {
		t.Errorf("unexpected error for valid output: %v", err)
	}
}

func TestExecutionAgent_BuildPrompt(t *testing.T) {
	agent := &ExecutionAgent{}
	verdict := &VerdictOutput{
		Ruling:    "Build a simple REST API",
		Rationale: "REST API is the most straightforward approach",
		Rejected: []RejectedOption{
			{Option: "GraphQL", Reason: "Too complex for MVP"},
		},
	}

	prompt := agent.buildPrompt(verdict)

	// Check that prompt contains key elements
	if !contains(prompt, "You are an executor, not a planner") {
		t.Error("prompt missing executor instruction")
	}

	if !contains(prompt, "Accept the ruling without question") {
		t.Error("prompt missing acceptance instruction")
	}

	if !contains(prompt, "MINIMUM viable scope") {
		t.Error("prompt missing minimal scope instruction")
	}

	if !contains(prompt, verdict.Ruling) {
		t.Error("prompt missing verdict ruling")
	}

	if !contains(prompt, verdict.Rationale) {
		t.Error("prompt missing verdict rationale")
	}

	if !contains(prompt, "Maximum 3 phases") {
		t.Error("prompt missing phase constraint")
	}

	if !contains(prompt, "maximum 5 tasks per phase") {
		t.Error("prompt missing task constraint")
	}

	if !contains(prompt, "Output ONLY valid JSON") {
		t.Error("prompt missing JSON instruction")
	}
}

func TestExecutionAgent_BilingualSupport(t *testing.T) {
	// Test that the agent handles non-English verdicts
	mockResponse := ExecutionOutput{
		MVPScope: []string{"基本用户认证"},
		Phases: []Phase{
			{
				Name:  "基础阶段",
				Tasks: []string{"创建数据库", "实现认证"},
			},
		},
		DoneCriteria: []string{"用户可以登录"},
	}

	client := &mockExecutionLLMClient{
		completeJSONFunc: func(ctx context.Context, prompt string, result any) error {
			data, _ := json.Marshal(mockResponse)
			return json.Unmarshal(data, result)
		},
	}

	agent := NewExecutionAgent(client)
	verdict := &VerdictOutput{
		Ruling:    "实现基本的用户认证系统",
		Rationale: "用户认证是所有其他功能的基础",
	}

	output, err := agent.Process(context.Background(), verdict)
	if err != nil {
		t.Fatalf("unexpected error with bilingual content: %v", err)
	}

	if len(output.MVPScope) == 0 {
		t.Error("expected MVP scope to be populated")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
