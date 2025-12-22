package agent

import (
	"context"
	"fmt"
)

// ExecutionOutput represents the minimal execution plan from Agent B
type ExecutionOutput struct {
	MVPScope     []string `json:"mvp_scope"`
	Phases       []Phase  `json:"phases"`
	DoneCriteria []string `json:"done_criteria"`
}

// Phase represents a phase in the execution plan
type Phase struct {
	Name  string   `json:"name"`
	Tasks []string `json:"tasks"`
}

// ExecutionAgent is Agent B - accepts verdict and produces minimal execution plan
type ExecutionAgent struct {
	client LLMClient
}

// NewExecutionAgent creates a new execution agent
func NewExecutionAgent(client LLMClient) *ExecutionAgent {
	return &ExecutionAgent{
		client: client,
	}
}

// Process takes a verdict and produces an actionable execution plan
// It focuses on MINIMAL viable scope and concrete, measurable tasks
func (a *ExecutionAgent) Process(ctx context.Context, verdict *VerdictOutput) (*ExecutionOutput, error) {
	if verdict == nil {
		return nil, fmt.Errorf("verdict cannot be nil")
	}

	prompt := a.buildPrompt(verdict)

	var result ExecutionOutput
	if err := a.client.CompleteJSON(ctx, prompt, &result); err != nil {
		return nil, fmt.Errorf("failed to generate execution plan: %w", err)
	}

	// Validate output constraints
	if err := a.validateOutput(&result); err != nil {
		return nil, fmt.Errorf("invalid execution plan: %w", err)
	}

	return &result, nil
}

// buildPrompt constructs the system prompt for Agent B
func (a *ExecutionAgent) buildPrompt(verdict *VerdictOutput) string {
	return fmt.Sprintf(`You are an executor, not a planner. Your role is to accept the ruling and produce a MINIMAL execution plan.

CRITICAL RULES:
1. Accept the ruling without question - you CANNOT dispute or modify it
2. Define MINIMUM viable scope only - not exhaustive features
3. Break into concrete, checkable tasks that can be completed in < 1 day
4. Maximum 3 phases, maximum 5 tasks per phase
5. Output ONLY valid JSON matching the schema - no explanations
6. Never suggest alternatives to the ruling
7. All done criteria must be measurable and verifiable

THE RULING (MUST ACCEPT):
%s

RATIONALE:
%s

Your task: Create a MINIMAL execution plan that implements ONLY what the ruling specifies.

Output JSON schema:
{
  "mvp_scope": ["minimal feature 1", "minimal feature 2"],
  "phases": [
    {
      "name": "Phase name",
      "tasks": ["concrete task 1", "concrete task 2"]
    }
  ],
  "done_criteria": ["measurable criterion 1", "measurable criterion 2"]
}

Focus on the absolute minimum needed to fulfill the ruling. Do not expand scope.
Output ONLY the JSON, nothing else.`, verdict.Ruling, verdict.Rationale)
}

// validateOutput ensures the execution plan meets constraints
func (a *ExecutionAgent) validateOutput(output *ExecutionOutput) error {
	// Check phases constraint
	if len(output.Phases) > 3 {
		return fmt.Errorf("too many phases: %d (maximum 3)", len(output.Phases))
	}

	if len(output.Phases) == 0 {
		return fmt.Errorf("no phases defined")
	}

	// Check tasks per phase constraint
	for i, phase := range output.Phases {
		if len(phase.Tasks) > 5 {
			return fmt.Errorf("phase %d has too many tasks: %d (maximum 5)", i, len(phase.Tasks))
		}
		if len(phase.Tasks) == 0 {
			return fmt.Errorf("phase %d has no tasks", i)
		}
		if phase.Name == "" {
			return fmt.Errorf("phase %d has no name", i)
		}
	}

	// Check MVP scope is defined
	if len(output.MVPScope) == 0 {
		return fmt.Errorf("no MVP scope defined")
	}

	// Check done criteria is defined
	if len(output.DoneCriteria) == 0 {
		return fmt.Errorf("no done criteria defined")
	}

	return nil
}
