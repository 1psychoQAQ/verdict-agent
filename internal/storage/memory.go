package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user account
type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserHistory represents a user's decision history entry
type UserHistory struct {
	ID              uuid.UUID              `json:"id"`
	UserID          uuid.UUID              `json:"user_id"`
	DecisionID      uuid.UUID              `json:"decision_id"`
	Input           string                 `json:"input"`
	Verdict         json.RawMessage        `json:"verdict"`
	Todo            string                 `json:"todo"`
	DoneCriteria    []DoneCriterion        `json:"done_criteria"`
	Score           float64                `json:"score"`
	UploadedContent []UploadedContent      `json:"uploaded_content,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// DoneCriterion represents a single done criterion with completion status
type DoneCriterion struct {
	Index     int    `json:"index"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

// UploadedContent represents user-uploaded proof/evidence
type UploadedContent struct {
	ID          uuid.UUID `json:"id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	URL         string    `json:"url"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// MemoryRepository implements Repository interface with in-memory storage
type MemoryRepository struct {
	mu        sync.RWMutex
	decisions map[uuid.UUID]*Decision
	todos     map[uuid.UUID]*Todo
	users     map[uuid.UUID]*User
	usernames map[string]uuid.UUID
	history   map[uuid.UUID]*UserHistory // keyed by history ID
	sessions  map[string]uuid.UUID       // token -> user ID
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		decisions: make(map[uuid.UUID]*Decision),
		todos:     make(map[uuid.UUID]*Todo),
		users:     make(map[uuid.UUID]*User),
		usernames: make(map[string]uuid.UUID),
		history:   make(map[uuid.UUID]*UserHistory),
		sessions:  make(map[string]uuid.UUID),
	}
}

// CreateDecision stores a new decision
func (r *MemoryRepository) CreateDecision(ctx context.Context, d *Decision) error {
	if d == nil {
		return fmt.Errorf("decision cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}

	r.decisions[d.ID] = d
	return nil
}

// GetDecision retrieves a decision by ID
func (r *MemoryRepository) GetDecision(ctx context.Context, id uuid.UUID) (*Decision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.decisions[id]
	if !ok {
		return nil, fmt.Errorf("decision not found")
	}
	return d, nil
}

// CreateTodo stores a new todo
func (r *MemoryRepository) CreateTodo(ctx context.Context, t *Todo) error {
	if t == nil {
		return fmt.Errorf("todo cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}

	r.todos[t.ID] = t
	return nil
}

// GetTodo retrieves a todo by ID
func (r *MemoryRepository) GetTodo(ctx context.Context, id uuid.UUID) (*Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.todos[id]
	if !ok {
		return nil, fmt.Errorf("todo not found")
	}
	return t, nil
}

// GetTodoByDecisionID retrieves a todo by decision ID
func (r *MemoryRepository) GetTodoByDecisionID(ctx context.Context, decisionID uuid.UUID) (*Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.todos {
		if t.DecisionID == decisionID {
			return t, nil
		}
	}
	return nil, fmt.Errorf("todo not found for decision")
}

// SaveArtifacts saves both decision and todo atomically
func (r *MemoryRepository) SaveArtifacts(ctx context.Context, d *Decision, t *Todo) error {
	if d == nil || t == nil {
		return fmt.Errorf("decision and todo cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	now := time.Now()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}

	t.DecisionID = d.ID
	r.decisions[d.ID] = d
	r.todos[t.ID] = t

	return nil
}

// Ping checks repository health
func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}

// Close cleans up resources
func (r *MemoryRepository) Close() {}

// === User Management ===

// CreateUser creates a new user account
func (r *MemoryRepository) CreateUser(ctx context.Context, username, password string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if username exists
	if _, exists := r.usernames[username]; exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	r.users[user.ID] = user
	r.usernames[username] = user.ID

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *MemoryRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, ok := r.usernames[username]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	return r.users[userID], nil
}

// GetUserByID retrieves a user by ID
func (r *MemoryRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// ValidatePassword checks if password matches user's hash
func (r *MemoryRepository) ValidatePassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// === Session Management ===

// CreateSession creates a new session token for user
func (r *MemoryRepository) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token := uuid.New().String()
	r.sessions[token] = userID
	return token, nil
}

// GetUserBySession retrieves user by session token
func (r *MemoryRepository) GetUserBySession(ctx context.Context, token string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, ok := r.sessions[token]
	if !ok {
		return nil, fmt.Errorf("invalid session")
	}

	user, ok := r.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// DeleteSession removes a session
func (r *MemoryRepository) DeleteSession(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, token)
	return nil
}

// === History Management ===

// CreateHistory creates a new history entry
func (r *MemoryRepository) CreateHistory(ctx context.Context, h *UserHistory) error {
	if h == nil {
		return fmt.Errorf("history cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	now := time.Now()
	if h.CreatedAt.IsZero() {
		h.CreatedAt = now
	}
	h.UpdatedAt = now

	r.history[h.ID] = h
	return nil
}

// GetHistory retrieves a history entry by ID
func (r *MemoryRepository) GetHistory(ctx context.Context, id uuid.UUID) (*UserHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, ok := r.history[id]
	if !ok {
		return nil, fmt.Errorf("history not found")
	}
	return h, nil
}

// GetHistoryByDecisionID retrieves history by decision ID
func (r *MemoryRepository) GetHistoryByDecisionID(ctx context.Context, decisionID uuid.UUID) (*UserHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, h := range r.history {
		if h.DecisionID == decisionID {
			return h, nil
		}
	}
	return nil, fmt.Errorf("history not found for decision")
}

// GetUserHistory retrieves all history entries for a user
func (r *MemoryRepository) GetUserHistory(ctx context.Context, userID uuid.UUID) ([]*UserHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*UserHistory
	for _, h := range r.history {
		if h.UserID == userID {
			result = append(result, h)
		}
	}
	return result, nil
}

// UpdateHistory updates a history entry
func (r *MemoryRepository) UpdateHistory(ctx context.Context, h *UserHistory) error {
	if h == nil {
		return fmt.Errorf("history cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.history[h.ID]; !ok {
		return fmt.Errorf("history not found")
	}

	h.UpdatedAt = time.Now()
	r.history[h.ID] = h
	return nil
}

// UpdateDoneCriteria updates the done criteria and recalculates score
func (r *MemoryRepository) UpdateDoneCriteria(ctx context.Context, historyID uuid.UUID, criteria []DoneCriterion) (*UserHistory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	h, ok := r.history[historyID]
	if !ok {
		return nil, fmt.Errorf("history not found")
	}

	h.DoneCriteria = criteria
	h.Score = calculateScore(criteria)
	h.UpdatedAt = time.Now()

	return h, nil
}

// calculateScore calculates completion percentage
func calculateScore(criteria []DoneCriterion) float64 {
	if len(criteria) == 0 {
		return 0
	}
	completed := 0
	for _, c := range criteria {
		if c.Completed {
			completed++
		}
	}
	return float64(completed) / float64(len(criteria)) * 100
}
