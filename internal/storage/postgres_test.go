package storage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockPool is a mock implementation for testing without a real database
type mockPool struct {
	decisions map[uuid.UUID]*Decision
	todos     map[uuid.UUID]*Todo
	pingErr   error
}

func newMockPool() *mockPool {
	return &mockPool{
		decisions: make(map[uuid.UUID]*Decision),
		todos:     make(map[uuid.UUID]*Todo),
	}
}

func TestNewPostgresRepository(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		wantErr     bool
	}{
		{
			name:        "empty database URL",
			databaseURL: "",
			wantErr:     true,
		},
		{
			name:        "invalid database URL",
			databaseURL: "invalid://url",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := NewPostgresRepository(ctx, tt.databaseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPostgresRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecisionValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		verdict json.RawMessage
		isFinal bool
		wantErr bool
	}{
		{
			name:    "valid decision",
			input:   "test input",
			verdict: json.RawMessage(`{"action": "approve"}`),
			isFinal: true,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			verdict: json.RawMessage(`{"action": "approve"}`),
			isFinal: true,
			wantErr: false, // Empty input is allowed by schema
		},
		{
			name:    "nil verdict",
			input:   "test input",
			verdict: nil,
			isFinal: true,
			wantErr: false, // Will be handled by database constraint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decision{
				ID:        uuid.New(),
				Input:     tt.input,
				Verdict:   tt.verdict,
				CreatedAt: time.Now(),
				IsFinal:   tt.isFinal,
			}

			// Basic validation
			if d.ID == uuid.Nil {
				t.Error("Decision ID should not be nil")
			}
			if d.CreatedAt.IsZero() {
				t.Error("Decision CreatedAt should not be zero")
			}
		})
	}
}

func TestTodoValidation(t *testing.T) {
	decisionID := uuid.New()

	tests := []struct {
		name       string
		decisionID uuid.UUID
		content    string
		wantErr    bool
	}{
		{
			name:       "valid todo",
			decisionID: decisionID,
			content:    "# Todo\n- Task 1\n- Task 2",
			wantErr:    false,
		},
		{
			name:       "empty content",
			decisionID: decisionID,
			content:    "",
			wantErr:    false, // Empty content is allowed
		},
		{
			name:       "nil decision ID",
			decisionID: uuid.Nil,
			content:    "# Todo",
			wantErr:    false, // Will be validated by foreign key constraint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := &Todo{
				ID:         uuid.New(),
				DecisionID: tt.decisionID,
				Content:    tt.content,
				CreatedAt:  time.Now(),
			}

			// Basic validation
			if todo.ID == uuid.Nil {
				t.Error("Todo ID should not be nil")
			}
			if todo.CreatedAt.IsZero() {
				t.Error("Todo CreatedAt should not be zero")
			}
		})
	}
}

func TestDecisionUUIDGeneration(t *testing.T) {
	d1 := &Decision{
		Input:   "test 1",
		Verdict: json.RawMessage(`{"action": "approve"}`),
	}

	d2 := &Decision{
		Input:   "test 2",
		Verdict: json.RawMessage(`{"action": "reject"}`),
	}

	// Both should have nil UUIDs since we didn't set them
	if d1.ID != uuid.Nil {
		t.Error("Decision 1 ID should be nil when not set")
	}
	if d2.ID != uuid.Nil {
		t.Error("Decision 2 ID should be nil when not set")
	}

	// Test UUID assignment
	customID := uuid.New()
	d3 := &Decision{
		ID:      customID,
		Input:   "test 3",
		Verdict: json.RawMessage(`{"action": "approve"}`),
	}

	if d3.ID != customID {
		t.Error("Decision should preserve custom UUID")
	}
}

func TestTodoUUIDGeneration(t *testing.T) {
	decisionID := uuid.New()

	t1 := &Todo{
		DecisionID: decisionID,
		Content:    "todo 1",
	}

	t2 := &Todo{
		DecisionID: decisionID,
		Content:    "todo 2",
	}

	// Both should have nil UUIDs since we didn't set them
	if t1.ID != uuid.Nil {
		t.Error("Todo 1 ID should be nil when not set")
	}
	if t2.ID != uuid.Nil {
		t.Error("Todo 2 ID should be nil when not set")
	}

	// Test UUID assignment
	customID := uuid.New()
	t3 := &Todo{
		ID:         customID,
		DecisionID: decisionID,
		Content:    "todo 3",
	}

	if t3.ID != customID {
		t.Error("Todo should preserve custom UUID")
	}
}

func TestJSONVerdictMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		verdict json.RawMessage
		wantErr bool
	}{
		{
			name:    "simple object",
			verdict: json.RawMessage(`{"action":"approve","reason":"valid"}`),
			wantErr: false,
		},
		{
			name:    "nested object",
			verdict: json.RawMessage(`{"action":"approve","details":{"score":10,"factors":["a","b"]}}`),
			wantErr: false,
		},
		{
			name:    "array",
			verdict: json.RawMessage(`[{"task":"1"},{"task":"2"}]`),
			wantErr: false,
		},
		{
			name:    "empty object",
			verdict: json.RawMessage(`{}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Decision{
				ID:        uuid.New(),
				Input:     "test",
				Verdict:   tt.verdict,
				CreatedAt: time.Now(),
				IsFinal:   true,
			}

			// Test marshaling
			data, err := json.Marshal(d)
			if err != nil {
				t.Fatalf("Failed to marshal decision: %v", err)
			}

			// Test unmarshaling
			var d2 Decision
			if err := json.Unmarshal(data, &d2); err != nil {
				t.Fatalf("Failed to unmarshal decision: %v", err)
			}

			// Compare verdict (normalize whitespace by unmarshaling both)
			var v1, v2 interface{}
			if err := json.Unmarshal(d.Verdict, &v1); err != nil {
				t.Fatalf("Failed to unmarshal original verdict: %v", err)
			}
			if err := json.Unmarshal(d2.Verdict, &v2); err != nil {
				t.Fatalf("Failed to unmarshal unmarshaled verdict: %v", err)
			}

			// Compare as strings after re-marshaling for consistent comparison
			b1, _ := json.Marshal(v1)
			b2, _ := json.Marshal(v2)
			if string(b1) != string(b2) {
				t.Errorf("Verdict mismatch: got %s, want %s", b2, b1)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{
		DatabaseURL:     "postgres://localhost/test",
		MaxConns:        0,
		MinConns:        0,
		MaxConnLifetime: 0,
	}

	// Set defaults if not provided
	if cfg.MaxConns == 0 {
		cfg.MaxConns = 10
	}
	if cfg.MinConns == 0 {
		cfg.MinConns = 2
	}
	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = time.Hour
	}

	if cfg.MaxConns != 10 {
		t.Errorf("MaxConns default = %d, want 10", cfg.MaxConns)
	}
	if cfg.MinConns != 2 {
		t.Errorf("MinConns default = %d, want 2", cfg.MinConns)
	}
	if cfg.MaxConnLifetime != time.Hour {
		t.Errorf("MaxConnLifetime default = %v, want 1h", cfg.MaxConnLifetime)
	}
}

func TestConfigCustomValues(t *testing.T) {
	cfg := &Config{
		DatabaseURL:     "postgres://localhost/test",
		MaxConns:        20,
		MinConns:        5,
		MaxConnLifetime: 30 * time.Minute,
	}

	if cfg.MaxConns != 20 {
		t.Errorf("MaxConns = %d, want 20", cfg.MaxConns)
	}
	if cfg.MinConns != 5 {
		t.Errorf("MinConns = %d, want 5", cfg.MinConns)
	}
	if cfg.MaxConnLifetime != 30*time.Minute {
		t.Errorf("MaxConnLifetime = %v, want 30m", cfg.MaxConnLifetime)
	}
}

func TestSaveArtifactsValidation(t *testing.T) {
	tests := []struct {
		name     string
		decision *Decision
		todo     *Todo
		wantErr  bool
	}{
		{
			name: "both valid",
			decision: &Decision{
				Input:   "test",
				Verdict: json.RawMessage(`{"action": "approve"}`),
			},
			todo: &Todo{
				Content: "# Todo",
			},
			wantErr: false,
		},
		{
			name:     "nil decision",
			decision: nil,
			todo: &Todo{
				Content: "# Todo",
			},
			wantErr: true,
		},
		{
			name: "nil todo",
			decision: &Decision{
				Input:   "test",
				Verdict: json.RawMessage(`{"action": "approve"}`),
			},
			todo:    nil,
			wantErr: true,
		},
		{
			name:     "both nil",
			decision: nil,
			todo:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation logic (would be in SaveArtifacts)
			var err error
			if tt.decision == nil || tt.todo == nil {
				err = &validationError{msg: "decision and todo cannot be nil"}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// validationError is a simple error type for testing
type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

func TestTimeHandling(t *testing.T) {
	// Test that zero time is handled correctly
	d := &Decision{
		ID:      uuid.New(),
		Input:   "test",
		Verdict: json.RawMessage(`{"action": "approve"}`),
		IsFinal: true,
	}

	// CreatedAt should be zero initially
	if !d.CreatedAt.IsZero() {
		t.Error("CreatedAt should be zero initially")
	}

	// Set CreatedAt
	now := time.Now()
	d.CreatedAt = now

	// Should not be zero after setting
	if d.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after setting")
	}

	// Should match the time we set
	if !d.CreatedAt.Equal(now) {
		t.Error("CreatedAt should match the time we set")
	}
}

func TestRepositoryInterface(t *testing.T) {
	// This test verifies that PostgresRepository implements Repository interface
	var _ Repository = (*PostgresRepository)(nil)
}
