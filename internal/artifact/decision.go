package artifact

import (
	"encoding/json"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/google/uuid"
)

// Decision represents the complete decision artifact
type Decision struct {
	ID        string          `json:"id"`
	CreatedAt string          `json:"created_at"`
	Input     string          `json:"input"`
	Verdict   DecisionVerdict `json:"verdict"`
	IsFinal   bool            `json:"is_final"`
}

// DecisionVerdict represents the verdict portion of the decision
type DecisionVerdict struct {
	Ruling    string           `json:"ruling"`
	Rationale string           `json:"rationale"`
	Rejected  []RejectedOption `json:"rejected"`
	Ranking   interface{}      `json:"ranking,omitempty"` // Can be []int or omitted
}

// RejectedOption represents a rejected option in the decision
type RejectedOption struct {
	Option string `json:"option"`
	Reason string `json:"reason"`
}

// generateDecisionJSON creates the decision.json artifact
func generateDecisionJSON(input string, verdict *agent.VerdictOutput, id uuid.UUID, createdAt time.Time) ([]byte, error) {
	decision := Decision{
		ID:        id.String(),
		CreatedAt: createdAt.UTC().Format(time.RFC3339),
		Input:     input,
		Verdict: DecisionVerdict{
			Ruling:    verdict.Ruling,
			Rationale: verdict.Rationale,
			Rejected:  convertRejectedOptions(verdict.Rejected),
			Ranking:   verdict.Ranking,
		},
		IsFinal: true,
	}

	// Marshal with indentation for readability
	return json.MarshalIndent(decision, "", "  ")
}

// convertRejectedOptions converts agent rejected options to decision format
func convertRejectedOptions(rejected []agent.RejectedOption) []RejectedOption {
	if len(rejected) == 0 {
		return []RejectedOption{}
	}

	result := make([]RejectedOption, len(rejected))
	for i, r := range rejected {
		result[i] = RejectedOption{
			Option: r.Option,
			Reason: r.Reason,
		}
	}
	return result
}
