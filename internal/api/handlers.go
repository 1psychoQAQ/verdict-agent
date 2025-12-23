package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/artifact"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/1psychoQAQ/verdict-agent/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// VerdictRequest represents the request body for POST /api/verdict
type VerdictRequest struct {
	Input         string            `json:"input"`
	Clarification *ClarificationCtx `json:"clarification,omitempty"` // Optional clarification answers
	SkipClarify   bool              `json:"skip_clarify,omitempty"`  // Skip clarification check
}

// ClarificationCtx holds clarification answers from the user
type ClarificationCtx struct {
	Answers map[string]string `json:"answers"` // question_id -> answer
}

// VerdictResponse represents the response for POST /api/verdict
type VerdictResponse struct {
	// Status indicates the response type: "clarification_needed" or "verdict"
	Status       string          `json:"status"`
	DecisionID   string          `json:"decision_id,omitempty"`
	HistoryID    string          `json:"history_id,omitempty"` // User's history entry ID
	Decision     json.RawMessage `json:"decision,omitempty"`
	Todo         string          `json:"todo,omitempty"`         // Markdown content
	DoneCriteria []string        `json:"done_criteria,omitempty"` // Done criteria list for tracking
	// Clarification fields (when status is "clarification_needed")
	Questions []QuestionDTO `json:"questions,omitempty"`
	Reason    string        `json:"reason,omitempty"`
}

// QuestionDTO represents a clarifying question for the API response
type QuestionDTO struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Type     string   `json:"type"` // "text", "choice", "multiple_choice"
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
}

// DecisionResponse represents the response for GET /api/decisions/{id}
type DecisionResponse struct {
	ID        string          `json:"id"`
	Input     string          `json:"input"`
	Verdict   json.RawMessage `json:"verdict"`
	CreatedAt string          `json:"created_at"`
	IsFinal   bool            `json:"is_final"`
}

// TodoResponse represents the response for GET /api/todos/{id}
type TodoResponse struct {
	ID         string `json:"id"`
	DecisionID string `json:"decision_id"`
	Content    string `json:"content"` // Markdown content
	CreatedAt  string `json:"created_at"`
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Error codes
const (
	ErrCodeInputEmpty    = "INPUT_EMPTY"
	ErrCodeInputTooLong  = "INPUT_TOO_LONG"
	ErrCodeVerdictFailed = "VERDICT_FAILED"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeRateLimited   = "RATE_LIMITED"
	ErrCodeInvalidID     = "INVALID_ID"
	ErrCodeInternalError = "INTERNAL_ERROR"
)

// Handlers holds the dependencies for HTTP handlers
type Handlers struct {
	pipeline           *pipeline.Pipeline
	generator          *artifact.Generator
	repository         storage.Repository
	clarificationAgent *agent.ClarificationAgent
	memoryRepo         *storage.MemoryRepository // For history tracking
}

// NewHandlers creates a new Handlers instance
func NewHandlers(p *pipeline.Pipeline, g *artifact.Generator, r storage.Repository) *Handlers {
	return &Handlers{
		pipeline:   p,
		generator:  g,
		repository: r,
	}
}

// NewHandlersWithClarification creates Handlers with clarification support
func NewHandlersWithClarification(p *pipeline.Pipeline, g *artifact.Generator, r storage.Repository, ca *agent.ClarificationAgent) *Handlers {
	return &Handlers{
		pipeline:           p,
		generator:          g,
		repository:         r,
		clarificationAgent: ca,
	}
}

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, code, message, details string) {
	writeJSON(w, status, ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

// HealthHandler handles GET /health requests
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	httpStatus := http.StatusOK

	// Check database connection if repository is available
	if h.repository != nil {
		if err := h.repository.Ping(r.Context()); err != nil {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}
	}

	writeJSON(w, httpStatus, map[string]string{
		"status": status,
	})
}

// VerdictHandler handles POST /api/verdict requests
func (h *Handlers) VerdictHandler(w http.ResponseWriter, r *http.Request) {
	var req VerdictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInternalError, "Invalid JSON body", err.Error())
		return
	}

	// Validate input
	input := strings.TrimSpace(req.Input)
	if input == "" {
		writeError(w, http.StatusBadRequest, ErrCodeInputEmpty, "Input is required", "")
		return
	}
	if len(input) > 10000 {
		writeError(w, http.StatusBadRequest, ErrCodeInputTooLong, "Input exceeds 10000 characters", "")
		return
	}

	// Check if clarification is needed (if agent is available and not skipped)
	if h.clarificationAgent != nil && !req.SkipClarify && req.Clarification == nil {
		clarification, err := h.clarificationAgent.Analyze(r.Context(), input)
		if err != nil {
			// Log but continue without clarification
			// log.Printf("Clarification analysis failed: %v", err)
		} else if clarification != nil && clarification.NeedsClarification && len(clarification.Questions) > 0 {
			// Return clarification questions
			questions := make([]QuestionDTO, len(clarification.Questions))
			for i, q := range clarification.Questions {
				questions[i] = QuestionDTO{
					ID:       q.ID,
					Question: q.Question,
					Type:     q.Type,
					Options:  q.Options,
					Required: q.Required,
				}
			}
			writeJSON(w, http.StatusOK, VerdictResponse{
				Status:    "clarification_needed",
				Questions: questions,
				Reason:    clarification.Reason,
			})
			return
		}
	}

	// Build enriched input if clarification answers are provided
	enrichedInput := input
	if req.Clarification != nil && len(req.Clarification.Answers) > 0 {
		enrichedInput = h.buildEnrichedInput(input, req.Clarification.Answers)
	}

	// Execute pipeline
	result, err := h.pipeline.Execute(r.Context(), enrichedInput)
	if err != nil {
		switch {
		case errors.Is(err, pipeline.ErrInputEmpty):
			writeError(w, http.StatusBadRequest, ErrCodeInputEmpty, "Input is required", "")
		case errors.Is(err, pipeline.ErrInputTooLong):
			writeError(w, http.StatusBadRequest, ErrCodeInputTooLong, "Input exceeds 10000 characters", "")
		case errors.Is(err, pipeline.ErrTimeout):
			writeError(w, http.StatusGatewayTimeout, ErrCodeVerdictFailed, "Pipeline timeout", "")
		default:
			writeError(w, http.StatusInternalServerError, ErrCodeVerdictFailed, "Pipeline failed", err.Error())
		}
		return
	}

	// Generate artifacts
	artifacts, err := h.generator.Generate(result)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeVerdictFailed, "Failed to generate artifacts", err.Error())
		return
	}

	// Save to database
	decision := &storage.Decision{
		ID:        artifacts.ID,
		Input:     result.Input,
		Verdict:   artifacts.DecisionJSON,
		CreatedAt: artifacts.CreatedAt,
		IsFinal:   true,
	}
	todo := &storage.Todo{
		DecisionID: artifacts.ID,
		Content:    string(artifacts.TodoMD),
		CreatedAt:  artifacts.CreatedAt,
	}

	if err := h.repository.SaveArtifacts(r.Context(), decision, todo); err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeInternalError, "Failed to save artifacts", err.Error())
		return
	}

	// Save to user history if authenticated
	var historyID string
	if h.memoryRepo != nil {
		user := GetUserFromContext(r)
		if user != nil {
			// Extract done criteria from execution result
			doneCriteria := extractDoneCriteria(result.Execution)

			history := &storage.UserHistory{
				UserID:       user.ID,
				DecisionID:   artifacts.ID,
				Input:        result.Input,
				Verdict:      artifacts.DecisionJSON,
				Todo:         string(artifacts.TodoMD),
				DoneCriteria: doneCriteria,
				Score:        0, // Initial score is 0
			}
			if err := h.memoryRepo.CreateHistory(r.Context(), history); err == nil {
				historyID = history.ID.String()
			}
		}
	}

	// Return response
	resp := VerdictResponse{
		Status:     "verdict",
		DecisionID: artifacts.ID.String(),
		Decision:   artifacts.DecisionJSON,
		Todo:       string(artifacts.TodoMD),
	}
	// Add done criteria from execution result
	if result.Execution != nil && len(result.Execution.DoneCriteria) > 0 {
		resp.DoneCriteria = result.Execution.DoneCriteria
	}
	if historyID != "" {
		resp.HistoryID = historyID
	}
	writeJSON(w, http.StatusOK, resp)
}

// buildEnrichedInput combines original input with clarification answers
func (h *Handlers) buildEnrichedInput(input string, answers map[string]string) string {
	var sb strings.Builder
	sb.WriteString(input)
	sb.WriteString("\n\n--- 用户补充信息 / User Clarifications ---\n")
	for id, answer := range answers {
		sb.WriteString("- ")
		sb.WriteString(id)
		sb.WriteString(": ")
		sb.WriteString(answer)
		sb.WriteString("\n")
	}
	return sb.String()
}

// GetDecisionHandler handles GET /api/decisions/{id} requests
func (h *Handlers) GetDecisionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid decision ID", "Must be a valid UUID")
		return
	}

	decision, err := h.repository.GetDecision(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Decision not found", "")
			return
		}
		writeError(w, http.StatusInternalServerError, ErrCodeInternalError, "Failed to retrieve decision", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, DecisionResponse{
		ID:        decision.ID.String(),
		Input:     decision.Input,
		Verdict:   decision.Verdict,
		CreatedAt: decision.CreatedAt.Format("2006-01-02T15:04:05Z"),
		IsFinal:   decision.IsFinal,
	})
}

// GetTodoHandler handles GET /api/todos/{id} requests
func (h *Handlers) GetTodoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid todo ID", "Must be a valid UUID")
		return
	}

	todo, err := h.repository.GetTodo(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Todo not found", "")
			return
		}
		writeError(w, http.StatusInternalServerError, ErrCodeInternalError, "Failed to retrieve todo", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, TodoResponse{
		ID:         todo.ID.String(),
		DecisionID: todo.DecisionID.String(),
		Content:    todo.Content,
		CreatedAt:  todo.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// extractDoneCriteria extracts done criteria from execution output
func extractDoneCriteria(execution *agent.ExecutionOutput) []storage.DoneCriterion {
	if execution == nil || len(execution.DoneCriteria) == 0 {
		return nil
	}

	criteria := make([]storage.DoneCriterion, len(execution.DoneCriteria))
	for i, text := range execution.DoneCriteria {
		criteria[i] = storage.DoneCriterion{
			Index:     i,
			Text:      text,
			Completed: false,
		}
	}
	return criteria
}
