package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/artifact"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/1psychoQAQ/verdict-agent/internal/storage"
	"github.com/google/uuid"
)

// mockLLMClient is a mock implementation of the LLM client for testing
type mockLLMClient struct {
	completeFunc     func(ctx context.Context, prompt string) (string, error)
	completeJSONFunc func(ctx context.Context, prompt string, result any) error
}

func (m *mockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, prompt)
	}
	return "", nil
}

func (m *mockLLMClient) CompleteJSON(ctx context.Context, prompt string, result any) error {
	if m.completeJSONFunc != nil {
		return m.completeJSONFunc(ctx, prompt, result)
	}
	return nil
}

// mockRepository is a mock implementation of the storage repository
type mockRepository struct {
	decisions map[uuid.UUID]*storage.Decision
	todos     map[uuid.UUID]*storage.Todo
	pingErr   error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		decisions: make(map[uuid.UUID]*storage.Decision),
		todos:     make(map[uuid.UUID]*storage.Todo),
	}
}

func (m *mockRepository) CreateDecision(ctx context.Context, d *storage.Decision) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	m.decisions[d.ID] = d
	return nil
}

func (m *mockRepository) GetDecision(ctx context.Context, id uuid.UUID) (*storage.Decision, error) {
	if d, ok := m.decisions[id]; ok {
		return d, nil
	}
	return nil, &notFoundError{message: "decision not found"}
}

func (m *mockRepository) CreateTodo(ctx context.Context, t *storage.Todo) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	m.todos[t.ID] = t
	return nil
}

func (m *mockRepository) GetTodo(ctx context.Context, id uuid.UUID) (*storage.Todo, error) {
	if t, ok := m.todos[id]; ok {
		return t, nil
	}
	return nil, &notFoundError{message: "todo not found"}
}

func (m *mockRepository) GetTodoByDecisionID(ctx context.Context, decisionID uuid.UUID) (*storage.Todo, error) {
	for _, t := range m.todos {
		if t.DecisionID == decisionID {
			return t, nil
		}
	}
	return nil, &notFoundError{message: "todo not found for decision"}
}

func (m *mockRepository) SaveArtifacts(ctx context.Context, d *storage.Decision, t *storage.Todo) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.DecisionID = d.ID
	m.decisions[d.ID] = d
	m.todos[t.ID] = t
	return nil
}

func (m *mockRepository) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *mockRepository) Close() {}

type notFoundError struct {
	message string
}

func (e *notFoundError) Error() string {
	return e.message
}

func TestHealthHandler(t *testing.T) {
	repo := newMockRepository()
	handlers := NewHandlers(nil, nil, repo)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handlers.HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", response["status"])
	}
}

func TestHealthHandler_Degraded(t *testing.T) {
	repo := newMockRepository()
	repo.pingErr = &notFoundError{message: "connection failed"}
	handlers := NewHandlers(nil, nil, repo)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handlers.HealthHandler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", response["status"])
	}
}

func TestVerdictHandler_EmptyInput(t *testing.T) {
	repo := newMockRepository()
	handlers := NewHandlers(nil, nil, repo)

	body := `{"input": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/verdict", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.VerdictHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Code != ErrCodeInputEmpty {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeInputEmpty, response.Code)
	}
}

func TestVerdictHandler_InputTooLong(t *testing.T) {
	repo := newMockRepository()
	handlers := NewHandlers(nil, nil, repo)

	// Create input that exceeds 10000 characters
	longInput := strings.Repeat("a", 10001)
	body := `{"input": "` + longInput + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/verdict", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.VerdictHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Code != ErrCodeInputTooLong {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeInputTooLong, response.Code)
	}
}

func TestVerdictHandler_InvalidJSON(t *testing.T) {
	repo := newMockRepository()
	handlers := NewHandlers(nil, nil, repo)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/verdict", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.VerdictHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestVerdictHandler_Success(t *testing.T) {
	// Create mock LLM client that returns valid verdict and execution outputs
	llmClient := &mockLLMClient{
		completeJSONFunc: func(ctx context.Context, prompt string, result any) error {
			// Check what type is being requested
			switch v := result.(type) {
			case *agent.VerdictOutput:
				v.Ruling = "Use Go"
				v.Rationale = "Go is great for this project"
				v.Rejected = []agent.RejectedOption{
					{Option: "Python", Reason: "Not suitable for this use case"},
				}
			case *agent.ExecutionOutput:
				v.MVPScope = []string{"Basic implementation", "Core features"}
				v.Phases = []agent.Phase{
					{Name: "Phase 1", Tasks: []string{"Task 1", "Task 2"}},
				}
				v.DoneCriteria = []string{"All tests pass"}
			}
			return nil
		},
	}

	repo := newMockRepository()
	verdictAgent := agent.NewVerdictAgent(llmClient)
	executionAgent := agent.NewExecutionAgent(llmClient)
	p := pipeline.NewPipeline(verdictAgent, executionAgent, 10*time.Minute)
	generator := artifact.NewGenerator()
	handlers := NewHandlers(p, generator, repo)

	body := `{"input": "Should I use Go or Python for this project?"}`
	req := httptest.NewRequest(http.MethodPost, "/api/verdict", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.VerdictHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response VerdictResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.DecisionID == "" {
		t.Error("expected decision_id to be set")
	}

	if len(response.Decision) == 0 {
		t.Error("expected decision to be set")
	}

	if response.Todo == "" {
		t.Error("expected todo to be set")
	}
}

func TestGetDecisionHandler_Success(t *testing.T) {
	repo := newMockRepository()
	id := uuid.New()
	decision := &storage.Decision{
		ID:        id,
		Input:     "Test input",
		Verdict:   json.RawMessage(`{"ruling":"Test ruling"}`),
		CreatedAt: time.Now(),
		IsFinal:   true,
	}
	repo.decisions[id] = decision

	router := NewRouter(RouterConfig{Repository: repo})

	req := httptest.NewRequest(http.MethodGet, "/api/decisions/"+id.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response DecisionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != id.String() {
		t.Errorf("expected ID %s, got %s", id.String(), response.ID)
	}

	if response.Input != "Test input" {
		t.Errorf("expected input 'Test input', got '%s'", response.Input)
	}
}

func TestGetDecisionHandler_NotFound(t *testing.T) {
	repo := newMockRepository()
	router := NewRouter(RouterConfig{Repository: repo})

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/decisions/"+id.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Code != ErrCodeNotFound {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeNotFound, response.Code)
	}
}

func TestGetDecisionHandler_InvalidID(t *testing.T) {
	repo := newMockRepository()
	router := NewRouter(RouterConfig{Repository: repo})

	req := httptest.NewRequest(http.MethodGet, "/api/decisions/invalid-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Code != ErrCodeInvalidID {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeInvalidID, response.Code)
	}
}

func TestGetTodoHandler_Success(t *testing.T) {
	repo := newMockRepository()
	decisionID := uuid.New()
	todoID := uuid.New()
	todo := &storage.Todo{
		ID:         todoID,
		DecisionID: decisionID,
		Content:    "# Test Todo\n- [ ] Task 1",
		CreatedAt:  time.Now(),
	}
	repo.todos[todoID] = todo

	router := NewRouter(RouterConfig{Repository: repo})

	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+todoID.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response TodoResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != todoID.String() {
		t.Errorf("expected ID %s, got %s", todoID.String(), response.ID)
	}

	if response.DecisionID != decisionID.String() {
		t.Errorf("expected DecisionID %s, got %s", decisionID.String(), response.DecisionID)
	}
}

func TestGetTodoHandler_NotFound(t *testing.T) {
	repo := newMockRepository()
	router := NewRouter(RouterConfig{Repository: repo})

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/todos/"+id.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCORS(t *testing.T) {
	router := NewRouter(RouterConfig{
		CORSConfig: DefaultCORSConfig(),
	})

	// Test preflight request
	req := httptest.NewRequest(http.MethodOptions, "/api/verdict", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header to be '*', got '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestRateLimiting(t *testing.T) {
	repo := newMockRepository()

	// Add a decision to retrieve for testing rate limiting on GET endpoint
	id := uuid.New()
	decision := &storage.Decision{
		ID:        id,
		Input:     "Test input",
		Verdict:   json.RawMessage(`{"ruling":"Test ruling"}`),
		CreatedAt: time.Now(),
		IsFinal:   true,
	}
	repo.decisions[id] = decision

	router := NewRouter(RouterConfig{
		Repository: repo,
		RateLimit:  2, // Very low limit for testing
	})

	// Make requests until rate limited (using GET to avoid nil pipeline)
	var rateLimited bool
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/decisions/"+id.String(), nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			rateLimited = true
			break
		}
	}

	if !rateLimited {
		t.Error("expected to be rate limited after multiple requests")
	}
}
