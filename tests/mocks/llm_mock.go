package mocks

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
)

// MockLLMClient is a mock implementation of the LLM client for testing
type MockLLMClient struct {
	VerdictResponse   *agent.VerdictOutput
	ExecutionResponse *agent.ExecutionOutput
	ShouldFail        bool
	FailAfter         int32
	callCount         int32
	ErrorMessage      string
}

// NewMockLLMClient creates a new mock LLM client with default responses
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		VerdictResponse: &agent.VerdictOutput{
			Ruling:    "Use Go",
			Rationale: "Go is well-suited for this project due to its performance and simplicity",
			Rejected: []agent.RejectedOption{
				{Option: "Python", Reason: "Not ideal for this type of system"},
				{Option: "Node.js", Reason: "Less suitable for production workloads"},
			},
			Ranking: []int{1, 2, 3},
		},
		ExecutionResponse: &agent.ExecutionOutput{
			MVPScope: []string{
				"Implement core functionality",
				"Add basic error handling",
			},
			Phases: []agent.Phase{
				{
					Name:  "Phase 1: Setup",
					Tasks: []string{"Initialize project", "Setup dependencies"},
				},
				{
					Name:  "Phase 2: Implementation",
					Tasks: []string{"Implement core logic", "Add tests"},
				},
			},
			DoneCriteria: []string{
				"All tests pass",
				"Code coverage > 80%",
			},
		},
	}
}

// Complete returns a mock completion response
func (m *MockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	count := atomic.AddInt32(&m.callCount, 1)

	if m.ShouldFail {
		errMsg := m.ErrorMessage
		if errMsg == "" {
			errMsg = "mock error"
		}
		return "", errors.New(errMsg)
	}

	if m.FailAfter > 0 && count > m.FailAfter {
		errMsg := m.ErrorMessage
		if errMsg == "" {
			errMsg = "mock error after limit"
		}
		return "", errors.New(errMsg)
	}

	// Return appropriate response based on prompt content
	if strings.Contains(prompt, "verdict") || strings.Contains(prompt, "ruling") ||
		strings.Contains(prompt, "You are a decisive") {
		data, _ := json.Marshal(m.VerdictResponse)
		return "```json\n" + string(data) + "\n```", nil
	}

	if strings.Contains(prompt, "execution") || strings.Contains(prompt, "phase") ||
		strings.Contains(prompt, "You are an executor") {
		data, _ := json.Marshal(m.ExecutionResponse)
		return "```json\n" + string(data) + "\n```", nil
	}

	// Default: return verdict response
	data, _ := json.Marshal(m.VerdictResponse)
	return "```json\n" + string(data) + "\n```", nil
}

// CompleteJSON returns a mock completion response parsed as JSON
func (m *MockLLMClient) CompleteJSON(ctx context.Context, prompt string, result any) error {
	count := atomic.AddInt32(&m.callCount, 1)

	if m.ShouldFail {
		errMsg := m.ErrorMessage
		if errMsg == "" {
			errMsg = "mock error"
		}
		return errors.New(errMsg)
	}

	if m.FailAfter > 0 && count > m.FailAfter {
		errMsg := m.ErrorMessage
		if errMsg == "" {
			errMsg = "mock error after limit"
		}
		return errors.New(errMsg)
	}

	// Based on the result type, populate with mock data
	switch v := result.(type) {
	case *agent.VerdictOutput:
		*v = *m.VerdictResponse
	case *agent.ExecutionOutput:
		*v = *m.ExecutionResponse
	}

	return nil
}

// Reset resets the call count
func (m *MockLLMClient) Reset() {
	atomic.StoreInt32(&m.callCount, 0)
}

// CallCount returns the current call count
func (m *MockLLMClient) CallCount() int32 {
	return atomic.LoadInt32(&m.callCount)
}
