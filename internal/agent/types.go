package agent

// VerdictOutput represents the output from Agent A (Verdict Agent)
type VerdictOutput struct {
	Ruling    string           `json:"ruling"`
	Rationale string           `json:"rationale"`
	Rejected  []RejectedOption `json:"rejected"`
	Ranking   interface{}      `json:"ranking,omitempty"` // Can be []int or omitted
}

// RejectedOption represents an option that was rejected by the verdict
type RejectedOption struct {
	Option string `json:"option"`
	Reason string `json:"reason"`
}
