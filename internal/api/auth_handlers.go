package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/1psychoQAQ/verdict-agent/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AuthHandlers holds dependencies for auth-related handlers
type AuthHandlers struct {
	repo *storage.MemoryRepository
}

// NewAuthHandlers creates new AuthHandlers
func NewAuthHandlers(repo *storage.MemoryRepository) *AuthHandlers {
	return &AuthHandlers{repo: repo}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// UserResponse represents user info response
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

// HistoryResponse represents a history entry in API response
type HistoryResponse struct {
	ID              string                       `json:"id"`
	DecisionID      string                       `json:"decision_id"`
	Input           string                       `json:"input"`
	Verdict         json.RawMessage              `json:"verdict"`
	Todo            string                       `json:"todo"`
	DoneCriteria    []storage.DoneCriterion      `json:"done_criteria"`
	Score           float64                      `json:"score"`
	UploadedContent []storage.UploadedContent    `json:"uploaded_content,omitempty"`
	CreatedAt       string                       `json:"created_at"`
	UpdatedAt       string                       `json:"updated_at"`
}

// UpdateDoneCriteriaRequest represents request to update done criteria
type UpdateDoneCriteriaRequest struct {
	DoneCriteria []storage.DoneCriterion `json:"done_criteria"`
}

// RegisterHandler handles POST /api/auth/register
func (h *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", err.Error())
		return
	}

	// Validate
	if strings.TrimSpace(req.Username) == "" {
		writeError(w, http.StatusBadRequest, "USERNAME_REQUIRED", "Username is required", "")
		return
	}
	if len(req.Password) < 4 {
		writeError(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "Password must be at least 4 characters", "")
		return
	}

	// Create user
	user, err := h.repo.CreateUser(r.Context(), req.Username, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeError(w, http.StatusConflict, "USERNAME_EXISTS", "Username already exists", "")
			return
		}
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create user", err.Error())
		return
	}

	// Create session
	token, err := h.repo.CreateSession(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_FAILED", "Failed to create session", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
	})
}

// LoginHandler handles POST /api/auth/login
func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", err.Error())
		return
	}

	// Get user
	user, err := h.repo.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password", "")
		return
	}

	// Validate password
	if !h.repo.ValidatePassword(user, req.Password) {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password", "")
		return
	}

	// Create session
	token, err := h.repo.CreateSession(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_FAILED", "Failed to create session", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
	})
}

// LogoutHandler handles POST /api/auth/logout
func (h *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token != "" {
		h.repo.DeleteSession(r.Context(), token)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

// MeHandler handles GET /api/auth/me
func (h *AuthHandlers) MeHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", "")
		return
	}

	writeJSON(w, http.StatusOK, UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// GetHistoryHandler handles GET /api/history
func (h *AuthHandlers) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", "")
		return
	}

	histories, err := h.repo.GetUserHistory(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch history", err.Error())
		return
	}

	response := make([]HistoryResponse, len(histories))
	for i, h := range histories {
		response[i] = historyToResponse(h)
	}

	writeJSON(w, http.StatusOK, response)
}

// GetHistoryByIDHandler handles GET /api/history/{id}
func (h *AuthHandlers) GetHistoryByIDHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", "")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid history ID", "")
		return
	}

	history, err := h.repo.GetHistory(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "History not found", "")
		return
	}

	// Verify ownership
	if history.UserID != user.ID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Access denied", "")
		return
	}

	writeJSON(w, http.StatusOK, historyToResponse(history))
}

// UpdateDoneCriteriaHandler handles PUT /api/history/{id}/criteria
func (h *AuthHandlers) UpdateDoneCriteriaHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", "")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid history ID", "")
		return
	}

	// Get existing history
	history, err := h.repo.GetHistory(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "History not found", "")
		return
	}

	// Verify ownership
	if history.UserID != user.ID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Access denied", "")
		return
	}

	var req UpdateDoneCriteriaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", err.Error())
		return
	}

	// Update criteria
	updated, err := h.repo.UpdateDoneCriteria(r.Context(), id, req.DoneCriteria)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update criteria", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, historyToResponse(updated))
}

// historyToResponse converts UserHistory to HistoryResponse
func historyToResponse(h *storage.UserHistory) HistoryResponse {
	return HistoryResponse{
		ID:              h.ID.String(),
		DecisionID:      h.DecisionID.String(),
		Input:           h.Input,
		Verdict:         h.Verdict,
		Todo:            h.Todo,
		DoneCriteria:    h.DoneCriteria,
		Score:           h.Score,
		UploadedContent: h.UploadedContent,
		CreatedAt:       h.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       h.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

