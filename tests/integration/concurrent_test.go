package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/artifact"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/1psychoQAQ/verdict-agent/tests/mocks"
)

func TestConcurrentPipelineExecution(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	numRequests := 10
	var wg sync.WaitGroup
	var successCount int32
	var errorCount int32

	ctx := context.Background()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			result, err := p.Execute(ctx, "Concurrent test input")
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				t.Logf("Request %d failed: %v", id, err)
				return
			}

			if result.Verdict != nil && result.Execution != nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&errorCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != int32(numRequests) {
		t.Errorf("Expected %d successful requests, got %d (errors: %d)",
			numRequests, successCount, errorCount)
	}
}

func TestConcurrentArtifactGeneration(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)
	gen := artifact.NewGenerator()

	numRequests := 5
	var wg sync.WaitGroup
	artifactIDs := make(map[string]bool)
	var mu sync.Mutex

	ctx := context.Background()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			result, err := p.Execute(ctx, "Concurrent artifact test")
			if err != nil {
				t.Logf("Request %d pipeline failed: %v", id, err)
				return
			}

			artifacts, err := gen.Generate(result)
			if err != nil {
				t.Logf("Request %d artifact generation failed: %v", id, err)
				return
			}

			// Store artifact ID to verify uniqueness
			mu.Lock()
			artifactIDs[artifacts.ID.String()] = true
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify all artifact IDs are unique
	if len(artifactIDs) != numRequests {
		t.Errorf("Expected %d unique artifact IDs, got %d", numRequests, len(artifactIDs))
	}
}

func TestConcurrentMixedOperations(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)
	gen := artifact.NewGenerator()

	var wg sync.WaitGroup
	var successCount int32
	var errorCount int32

	ctx := context.Background()

	// Mix of valid and invalid requests
	inputs := []string{
		"Valid input 1",
		"",           // Should fail - empty
		"Valid input 2",
		"Valid input 3",
		"",           // Should fail - empty
		"Valid input 4",
	}

	expectedSuccess := 4
	expectedErrors := 2

	for i, input := range inputs {
		wg.Add(1)
		go func(id int, inp string) {
			defer wg.Done()

			result, err := p.Execute(ctx, inp)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}

			_, err = gen.Generate(result)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}

			atomic.AddInt32(&successCount, 1)
		}(i, input)
	}

	wg.Wait()

	if int(successCount) != expectedSuccess {
		t.Errorf("Expected %d successful requests, got %d", expectedSuccess, successCount)
	}
	if int(errorCount) != expectedErrors {
		t.Errorf("Expected %d failed requests, got %d", expectedErrors, errorCount)
	}
}

func TestHighLoadConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load test in short mode")
	}

	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	numRequests := 50
	var wg sync.WaitGroup
	var successCount int32
	startTime := time.Now()

	ctx := context.Background()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			result, err := p.Execute(ctx, "High load test input")
			if err == nil && result.Verdict != nil && result.Execution != nil {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Processed %d requests in %v (%.2f req/sec)",
		successCount, duration, float64(successCount)/duration.Seconds())

	// All requests should succeed with mock LLM
	if successCount != int32(numRequests) {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}
}

func TestConcurrentWithCancellation(t *testing.T) {
	mockLLM := mocks.NewMockLLMClient()
	verdictAgent := agent.NewVerdictAgent(mockLLM)
	executionAgent := agent.NewExecutionAgent(mockLLM)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 5*time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Start several concurrent requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Request may succeed or fail due to cancellation
			p.Execute(ctx, "Test input")
		}(i)
	}

	// Cancel after a short delay
	time.Sleep(1 * time.Millisecond)
	cancel()

	wg.Wait()
	// Test passes if no panic occurred
}
