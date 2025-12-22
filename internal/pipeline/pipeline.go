package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
)

// Error types
var (
	ErrInputEmpty      = errors.New("input is empty")
	ErrInputTooLong    = errors.New("input exceeds 10000 characters")
	ErrVerdictFailed   = errors.New("verdict agent failed")
	ErrExecutionFailed = errors.New("execution agent failed")
	ErrTimeout         = errors.New("pipeline timeout")
)

// Pipeline orchestrates the execution of Agent A (Verdict) → Agent B (Execution)
type Pipeline struct {
	verdictAgent   *agent.VerdictAgent
	executionAgent *agent.ExecutionAgent
	timeout        time.Duration
}

// PipelineResult contains the complete output of the pipeline execution
type PipelineResult struct {
	Input     string                `json:"input"`
	Verdict   *agent.VerdictOutput  `json:"verdict"`
	Execution *agent.ExecutionOutput `json:"execution"`
	Duration  time.Duration         `json:"duration"`
}

// NewPipeline creates a new pipeline with the given agents and timeout
func NewPipeline(verdictAgent *agent.VerdictAgent, executionAgent *agent.ExecutionAgent, timeout time.Duration) *Pipeline {
	if timeout <= 0 {
		timeout = 10 * time.Minute // Default timeout
	}

	return &Pipeline{
		verdictAgent:   verdictAgent,
		executionAgent: executionAgent,
		timeout:        timeout,
	}
}

// Execute runs the complete pipeline: validate input → Agent A → validate → Agent B → validate
func (p *Pipeline) Execute(ctx context.Context, input string) (*PipelineResult, error) {
	startTime := time.Now()

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	result := &PipelineResult{
		Input: input,
	}

	// Step 1: Validate input
	if err := p.validateInput(input); err != nil {
		return nil, err
	}

	// Step 2: Execute Agent A (Verdict)
	verdict, err := p.executeVerdictAgent(timeoutCtx, input)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrVerdictFailed, err)
	}
	result.Verdict = verdict

	// Step 3: Validate Agent A output
	if err := p.validateVerdictOutput(verdict); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVerdictFailed, err)
	}

	// Step 4: Execute Agent B (Execution)
	execution, err := p.executeExecutionAgent(timeoutCtx, verdict)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}
	result.Execution = execution

	// Step 5: Validate Agent B output
	if err := p.validateExecutionOutput(execution); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExecutionFailed, err)
	}

	// Calculate total duration
	result.Duration = time.Since(startTime)

	return result, nil
}

// validateInput checks input constraints
func (p *Pipeline) validateInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return ErrInputEmpty
	}
	if utf8.RuneCountInString(input) > 10000 {
		return ErrInputTooLong
	}
	return nil
}

// executeVerdictAgent calls Agent A with the user input
func (p *Pipeline) executeVerdictAgent(ctx context.Context, input string) (*agent.VerdictOutput, error) {
	return p.verdictAgent.Process(ctx, input)
}

// validateVerdictOutput ensures the verdict output meets quality standards
func (p *Pipeline) validateVerdictOutput(output *agent.VerdictOutput) error {
	if output == nil {
		return errors.New("verdict output is nil")
	}
	if strings.TrimSpace(output.Ruling) == "" {
		return errors.New("verdict ruling is empty")
	}
	if strings.TrimSpace(output.Rationale) == "" {
		return errors.New("verdict rationale is empty")
	}
	return nil
}

// executeExecutionAgent calls Agent B with the verdict
func (p *Pipeline) executeExecutionAgent(ctx context.Context, verdict *agent.VerdictOutput) (*agent.ExecutionOutput, error) {
	return p.executionAgent.Process(ctx, verdict)
}

// validateExecutionOutput ensures the execution output meets constraints
func (p *Pipeline) validateExecutionOutput(output *agent.ExecutionOutput) error {
	if output == nil {
		return errors.New("execution output is nil")
	}
	if len(output.Phases) == 0 {
		return errors.New("execution must have at least 1 phase with tasks")
	}

	// Verify at least one phase has tasks
	hasTasksInAnyPhase := false
	for _, phase := range output.Phases {
		if len(phase.Tasks) > 0 {
			hasTasksInAnyPhase = true
			break
		}
	}

	if !hasTasksInAnyPhase {
		return errors.New("execution must have at least 1 phase with tasks")
	}

	return nil
}
