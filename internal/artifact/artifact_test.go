package artifact

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
	"github.com/google/uuid"
)

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		result  *pipeline.PipelineResult
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pipeline result",
			result: &pipeline.PipelineResult{
				Input: "Should I build a mobile app or web app?",
				Verdict: &agent.VerdictOutput{
					Ruling:    "Build a mobile app first",
					Rationale: "Mobile-first approach reaches users faster",
					Rejected: []agent.RejectedOption{
						{Option: "web app", Reason: "requires more infrastructure"},
					},
					Ranking: []int{1, 2},
				},
				Execution: &agent.ExecutionOutput{
					MVPScope: []string{"User authentication", "Core features"},
					Phases: []agent.Phase{
						{
							Name:  "Foundation",
							Tasks: []string{"Setup project", "Configure build"},
						},
						{
							Name:  "Development",
							Tasks: []string{"Implement auth", "Add features"},
						},
					},
					DoneCriteria: []string{"App runs on iOS", "Users can login"},
				},
				Duration: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name:    "nil result",
			result:  nil,
			wantErr: true,
			errMsg:  "pipeline result cannot be nil",
		},
		{
			name: "nil verdict",
			result: &pipeline.PipelineResult{
				Input:     "test",
				Verdict:   nil,
				Execution: &agent.ExecutionOutput{},
			},
			wantErr: true,
			errMsg:  "verdict output cannot be nil",
		},
		{
			name: "nil execution",
			result: &pipeline.PipelineResult{
				Input:     "test",
				Verdict:   &agent.VerdictOutput{},
				Execution: nil,
			},
			wantErr: true,
			errMsg:  "execution output cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator()
			artifacts, err := g.Generate(tt.result)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Generate() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Generate() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Generate() unexpected error = %v", err)
				return
			}

			if artifacts == nil {
				t.Error("Generate() returned nil artifacts")
				return
			}

			// Verify artifacts are not empty
			if len(artifacts.DecisionJSON) == 0 {
				t.Error("DecisionJSON is empty")
			}
			if len(artifacts.TodoMD) == 0 {
				t.Error("TodoMD is empty")
			}

			// Verify ID and timestamp are set
			if artifacts.ID == uuid.Nil {
				t.Error("ID is nil UUID")
			}
			if artifacts.CreatedAt.IsZero() {
				t.Error("CreatedAt is zero time")
			}
		})
	}
}

func TestGenerateDecisionJSON(t *testing.T) {
	input := "Should I use Go or Python?"
	verdict := &agent.VerdictOutput{
		Ruling:    "Use Go",
		Rationale: "Better performance and concurrency",
		Rejected: []agent.RejectedOption{
			{Option: "Python", Reason: "slower execution"},
		},
		Ranking: []int{1, 2},
	}
	id := uuid.New()
	createdAt := time.Date(2025, 12, 22, 3, 28, 32, 0, time.UTC)

	jsonBytes, err := generateDecisionJSON(input, verdict, id, createdAt)
	if err != nil {
		t.Fatalf("generateDecisionJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var decision Decision
	if err := json.Unmarshal(jsonBytes, &decision); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify structure
	if decision.ID != id.String() {
		t.Errorf("ID = %v, want %v", decision.ID, id.String())
	}
	if decision.CreatedAt != "2025-12-22T03:28:32Z" {
		t.Errorf("CreatedAt = %v, want 2025-12-22T03:28:32Z", decision.CreatedAt)
	}
	if decision.Input != input {
		t.Errorf("Input = %v, want %v", decision.Input, input)
	}
	if decision.Verdict.Ruling != "Use Go" {
		t.Errorf("Ruling = %v, want Use Go", decision.Verdict.Ruling)
	}
	if decision.Verdict.Rationale != "Better performance and concurrency" {
		t.Errorf("Rationale = %v, want Better performance and concurrency", decision.Verdict.Rationale)
	}
	if len(decision.Verdict.Rejected) != 1 {
		t.Errorf("Rejected count = %v, want 1", len(decision.Verdict.Rejected))
	}
	if len(decision.Verdict.Ranking) != 2 {
		t.Errorf("Ranking count = %v, want 2", len(decision.Verdict.Ranking))
	}
	if !decision.IsFinal {
		t.Error("IsFinal = false, want true")
	}

	// Verify JSON is properly indented
	if !strings.Contains(string(jsonBytes), "\n") {
		t.Error("JSON is not indented")
	}
}

func TestGenerateDecisionJSON_EmptyRejected(t *testing.T) {
	verdict := &agent.VerdictOutput{
		Ruling:    "Proceed with plan A",
		Rationale: "Clear best choice",
		Rejected:  []agent.RejectedOption{},
		Ranking:   nil,
	}
	id := uuid.New()
	createdAt := time.Now()

	jsonBytes, err := generateDecisionJSON("test input", verdict, id, createdAt)
	if err != nil {
		t.Fatalf("generateDecisionJSON() error = %v", err)
	}

	var decision Decision
	if err := json.Unmarshal(jsonBytes, &decision); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify rejected is empty array, not null
	if decision.Verdict.Rejected == nil {
		t.Error("Rejected is nil, want empty array")
	}
	if len(decision.Verdict.Rejected) != 0 {
		t.Errorf("Rejected count = %v, want 0", len(decision.Verdict.Rejected))
	}
}

func TestGenerateTodoMD(t *testing.T) {
	verdict := &agent.VerdictOutput{
		Ruling:    "Build API service",
		Rationale: "Backend needed first",
	}
	execution := &agent.ExecutionOutput{
		MVPScope: []string{"REST API", "Database schema"},
		Phases: []agent.Phase{
			{
				Name:  "Setup",
				Tasks: []string{"Initialize project", "Configure dependencies"},
			},
			{
				Name:  "Implementation",
				Tasks: []string{"Create models", "Build endpoints"},
			},
		},
		DoneCriteria: []string{"API responds to requests", "Data persists"},
	}
	id := uuid.New()
	createdAt := time.Date(2025, 12, 22, 3, 28, 32, 0, time.UTC)

	mdBytes, err := generateTodoMD(verdict, execution, id, createdAt)
	if err != nil {
		t.Fatalf("generateTodoMD() error = %v", err)
	}

	md := string(mdBytes)

	// Verify header
	if !strings.Contains(md, "# Execution Plan: Build API service") {
		t.Error("Missing or incorrect header")
	}

	// Verify metadata
	if !strings.Contains(md, "Generated: 2025-12-22T03:28:32Z") {
		t.Error("Missing or incorrect timestamp")
	}
	if !strings.Contains(md, "Decision ID: "+id.String()) {
		t.Error("Missing or incorrect ID")
	}

	// Verify MVP Scope
	if !strings.Contains(md, "## MVP Scope") {
		t.Error("Missing MVP Scope section")
	}
	if !strings.Contains(md, "- REST API") {
		t.Error("Missing MVP scope item")
	}
	if !strings.Contains(md, "- Database schema") {
		t.Error("Missing MVP scope item")
	}

	// Verify Phases
	if !strings.Contains(md, "## Phases") {
		t.Error("Missing Phases section")
	}
	if !strings.Contains(md, "### Phase 1: Setup") {
		t.Error("Missing Phase 1")
	}
	if !strings.Contains(md, "### Phase 2: Implementation") {
		t.Error("Missing Phase 2")
	}

	// Verify tasks with checkboxes
	if !strings.Contains(md, "- [ ] Initialize project") {
		t.Error("Missing task checkbox")
	}
	if !strings.Contains(md, "- [ ] Configure dependencies") {
		t.Error("Missing task checkbox")
	}
	if !strings.Contains(md, "- [ ] Create models") {
		t.Error("Missing task checkbox")
	}
	if !strings.Contains(md, "- [ ] Build endpoints") {
		t.Error("Missing task checkbox")
	}

	// Verify Done Criteria
	if !strings.Contains(md, "## Done Criteria") {
		t.Error("Missing Done Criteria section")
	}
	if !strings.Contains(md, "- API responds to requests") {
		t.Error("Missing done criterion")
	}
	if !strings.Contains(md, "- Data persists") {
		t.Error("Missing done criterion")
	}
}

func TestGenerateTodoMD_MultiplePhases(t *testing.T) {
	verdict := &agent.VerdictOutput{
		Ruling:    "Three phase project",
		Rationale: "Structured approach",
	}
	execution := &agent.ExecutionOutput{
		MVPScope: []string{"Core feature"},
		Phases: []agent.Phase{
			{Name: "Phase A", Tasks: []string{"Task A1"}},
			{Name: "Phase B", Tasks: []string{"Task B1"}},
			{Name: "Phase C", Tasks: []string{"Task C1"}},
		},
		DoneCriteria: []string{"Complete"},
	}
	id := uuid.New()
	createdAt := time.Now()

	mdBytes, err := generateTodoMD(verdict, execution, id, createdAt)
	if err != nil {
		t.Fatalf("generateTodoMD() error = %v", err)
	}

	md := string(mdBytes)

	// Verify all phases are numbered correctly
	if !strings.Contains(md, "### Phase 1: Phase A") {
		t.Error("Missing Phase 1")
	}
	if !strings.Contains(md, "### Phase 2: Phase B") {
		t.Error("Missing Phase 2")
	}
	if !strings.Contains(md, "### Phase 3: Phase C") {
		t.Error("Missing Phase 3")
	}
}

func TestArtifacts_UUIDAndTimestamp(t *testing.T) {
	g := NewGenerator()
	result := &pipeline.PipelineResult{
		Input: "test",
		Verdict: &agent.VerdictOutput{
			Ruling:    "test ruling",
			Rationale: "test rationale",
		},
		Execution: &agent.ExecutionOutput{
			MVPScope: []string{"scope"},
			Phases: []agent.Phase{
				{Name: "phase", Tasks: []string{"task"}},
			},
			DoneCriteria: []string{"done"},
		},
	}

	artifacts1, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond) // Ensure time difference

	artifacts2, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Each generation should have unique UUID
	if artifacts1.ID == artifacts2.ID {
		t.Error("Generated artifacts have same UUID")
	}

	// Timestamps should be different
	if artifacts1.CreatedAt.Equal(artifacts2.CreatedAt) {
		t.Error("Generated artifacts have same timestamp")
	}

	// Verify UUID format in both JSON and markdown
	var decision1 Decision
	if err := json.Unmarshal(artifacts1.DecisionJSON, &decision1); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if _, err := uuid.Parse(decision1.ID); err != nil {
		t.Errorf("Invalid UUID in JSON: %v", err)
	}

	md1 := string(artifacts1.TodoMD)
	if !strings.Contains(md1, artifacts1.ID.String()) {
		t.Error("UUID not found in markdown")
	}
}

func TestDecisionJSON_ISO8601Timestamp(t *testing.T) {
	result := &pipeline.PipelineResult{
		Input: "test",
		Verdict: &agent.VerdictOutput{
			Ruling:    "test",
			Rationale: "test",
		},
		Execution: &agent.ExecutionOutput{
			MVPScope:     []string{"test"},
			Phases:       []agent.Phase{{Name: "test", Tasks: []string{"test"}}},
			DoneCriteria: []string{"test"},
		},
	}

	g := NewGenerator()
	artifacts, err := g.Generate(result)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var decision Decision
	if err := json.Unmarshal(artifacts.DecisionJSON, &decision); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify ISO 8601 format
	_, err = time.Parse(time.RFC3339, decision.CreatedAt)
	if err != nil {
		t.Errorf("Timestamp not in ISO 8601 format: %v", err)
	}

	// Verify it ends with Z (UTC)
	if !strings.HasSuffix(decision.CreatedAt, "Z") {
		t.Error("Timestamp should end with Z for UTC")
	}
}
