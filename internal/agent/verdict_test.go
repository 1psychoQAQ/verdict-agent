package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// mockVerdictLLMClient is a mock implementation for verdict agent testing
type mockVerdictLLMClient struct {
	jsonResponse  *VerdictOutput
	shouldError   bool
	errorToReturn error
}

func (m *mockVerdictLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	if m.shouldError {
		return "", m.errorToReturn
	}
	data, _ := json.Marshal(m.jsonResponse)
	return string(data), nil
}

func (m *mockVerdictLLMClient) CompleteJSON(ctx context.Context, prompt string, result any) error {
	if m.shouldError {
		return m.errorToReturn
	}
	data, _ := json.Marshal(m.jsonResponse)
	return json.Unmarshal(data, result)
}

func TestNewVerdictAgent(t *testing.T) {
	client := &mockVerdictLLMClient{}
	agent := NewVerdictAgent(client)

	if agent == nil {
		t.Fatal("NewVerdictAgent returned nil")
	}
	if agent.client != client {
		t.Error("VerdictAgent client not set correctly")
	}
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid input",
			input:   "Should I use Go or Python?",
			wantErr: nil,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: ErrEmptyInput,
		},
		{
			name:    "whitespace only",
			input:   "   \n\t  ",
			wantErr: ErrEmptyInput,
		},
		{
			name:    "input too long",
			input:   strings.Repeat("a", 10001),
			wantErr: ErrInputTooLong,
		},
		{
			name:    "input at max length",
			input:   strings.Repeat("a", 10000),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  *VerdictOutput
		wantErr error
	}{
		{
			name: "valid output",
			output: &VerdictOutput{
				Ruling:    "Use Go for this project",
				Rationale: "Better performance",
				Rejected:  []RejectedOption{{Option: "Python", Reason: "Slower"}},
			},
			wantErr: nil,
		},
		{
			name: "empty ruling",
			output: &VerdictOutput{
				Ruling:    "",
				Rationale: "Some reason",
			},
			wantErr: ErrEmptyRuling,
		},
		{
			name: "whitespace ruling",
			output: &VerdictOutput{
				Ruling:    "   \n  ",
				Rationale: "Some reason",
			},
			wantErr: ErrEmptyRuling,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutput(tt.output)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validateOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "english text",
			input:    "Should I use Go or Python?",
			expected: "en",
		},
		{
			name:     "chinese text",
			input:    "我应该用Go还是Python？",
			expected: "zh",
		},
		{
			name:     "mixed with more chinese",
			input:    "我想知道是用Go还是Python好，which is better?",
			expected: "zh",
		},
		{
			name:     "mixed with more english",
			input:    "Should I use Go还是Python?",
			expected: "en",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLanguage(tt.input)
			if result != tt.expected {
				t.Errorf("detectLanguage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildVerdictPrompt(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedLang    string
		expectedContain []string
	}{
		{
			name:         "english input",
			input:        "Should I use microservices?",
			expectedLang: "en",
			expectedContain: []string{
				"You are a judge",
				"Deliver ONE ruling",
				"Output Format",
				"Should I use microservices?",
			},
		},
		{
			name:         "chinese input",
			input:        "我应该使用微服务架构吗？",
			expectedLang: "zh",
			expectedContain: []string{
				"你是一位法官",
				"只给出一个裁决",
				"输出格式",
				"我应该使用微服务架构吗？",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildVerdictPrompt(tt.input)

			for _, expected := range tt.expectedContain {
				if !strings.Contains(prompt, expected) {
					t.Errorf("buildVerdictPrompt() missing expected content: %s", expected)
				}
			}
		})
	}
}

func TestVerdictAgent_Process(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		mockResponse *VerdictOutput
		mockError    error
		wantErr      bool
		checkOutput  func(*testing.T, *VerdictOutput)
	}{
		{
			name:  "successful verdict - technical decision",
			input: "Should we use REST or GraphQL for our API?",
			mockResponse: &VerdictOutput{
				Ruling:    "Use GraphQL for your API",
				Rationale: "GraphQL provides better flexibility for evolving requirements and reduces over-fetching",
				Rejected: []RejectedOption{
					{Option: "REST", Reason: "Less flexible for complex data requirements"},
					{Option: "gRPC", Reason: "Unnecessary complexity for web clients"},
				},
				Ranking: []int{1, 2, 3},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, output *VerdictOutput) {
				if output.Ruling != "Use GraphQL for your API" {
					t.Errorf("unexpected ruling: %s", output.Ruling)
				}
				if len(output.Rejected) != 2 {
					t.Errorf("expected 2 rejected options, got %d", len(output.Rejected))
				}
			},
		},
		{
			name:  "successful verdict - chinese input",
			input: "我们应该用MongoDB还是PostgreSQL？",
			mockResponse: &VerdictOutput{
				Ruling:    "使用PostgreSQL",
				Rationale: "PostgreSQL提供更好的数据一致性和ACID保证，适合业务系统",
				Rejected: []RejectedOption{
					{Option: "MongoDB", Reason: "缺乏强一致性保证"},
					{Option: "MySQL", Reason: "PostgreSQL功能更强大"},
				},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, output *VerdictOutput) {
				if !strings.Contains(output.Ruling, "PostgreSQL") {
					t.Errorf("unexpected ruling: %s", output.Ruling)
				}
			},
		},
		{
			name:  "empty input error",
			input: "",
			mockResponse: &VerdictOutput{
				Ruling: "Should not reach here",
			},
			wantErr: true,
		},
		{
			name:  "input too long error",
			input: strings.Repeat("a", 10001),
			mockResponse: &VerdictOutput{
				Ruling: "Should not reach here",
			},
			wantErr: true,
		},
		{
			name:      "LLM error",
			input:     "Valid input",
			mockError: errors.New("API error"),
			wantErr:   true,
		},
		{
			name:  "empty ruling in response",
			input: "Valid input",
			mockResponse: &VerdictOutput{
				Ruling:    "",
				Rationale: "Some rationale",
			},
			wantErr: true,
		},
		{
			name:  "project choice verdict",
			input: "Should we build our own CMS or use WordPress?",
			mockResponse: &VerdictOutput{
				Ruling:    "Build a custom CMS",
				Rationale: "Your specific requirements justify custom development for better long-term maintainability",
				Rejected: []RejectedOption{
					{Option: "WordPress", Reason: "Plugin dependency creates security and maintenance burden"},
					{Option: "Contentful", Reason: "Third-party dependency limits control"},
				},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, output *VerdictOutput) {
				if !strings.Contains(output.Ruling, "custom") {
					t.Errorf("unexpected ruling: %s", output.Ruling)
				}
				if len(output.Rejected) < 2 {
					t.Errorf("expected at least 2 rejected options")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockVerdictLLMClient{
				jsonResponse:  tt.mockResponse,
				shouldError:   tt.mockError != nil,
				errorToReturn: tt.mockError,
			}

			agent := NewVerdictAgent(mock)
			ctx := context.Background()

			output, err := agent.Process(ctx, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Process() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Process() unexpected error: %v", err)
				return
			}

			if output == nil {
				t.Fatal("Process() returned nil output")
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

func TestVerdictOutput_JSONMarshaling(t *testing.T) {
	output := &VerdictOutput{
		Ruling:    "Use microservices architecture",
		Rationale: "Better scalability for your use case",
		Rejected: []RejectedOption{
			{Option: "Monolith", Reason: "Scaling limitations"},
			{Option: "Serverless", Reason: "Higher complexity"},
		},
		Ranking: []int{1, 2, 3},
	}

	// Test marshaling
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Test unmarshaling
	var decoded VerdictOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Verify fields
	if decoded.Ruling != output.Ruling {
		t.Errorf("ruling mismatch: got %s, want %s", decoded.Ruling, output.Ruling)
	}
	if decoded.Rationale != output.Rationale {
		t.Errorf("rationale mismatch: got %s, want %s", decoded.Rationale, output.Rationale)
	}
	if len(decoded.Rejected) != len(output.Rejected) {
		t.Errorf("rejected count mismatch: got %d, want %d", len(decoded.Rejected), len(output.Rejected))
	}
	if len(decoded.Ranking) != len(output.Ranking) {
		t.Errorf("ranking count mismatch: got %d, want %d", len(decoded.Ranking), len(output.Ranking))
	}
}

func TestVerdictAgent_PromptContainsNoHedging(t *testing.T) {
	inputs := []string{
		"Should I use Docker or Kubernetes?",
		"我应该选择哪个框架？",
	}

	for _, input := range inputs {
		prompt := buildVerdictPrompt(input)
		promptLower := strings.ToLower(prompt)

		// Check that prompt explicitly prohibits hedging language
		if !strings.Contains(promptLower, "never") && !strings.Contains(promptLower, "prohibited") && !strings.Contains(promptLower, "严禁") {
			t.Error("Prompt should explicitly prohibit hedging language")
		}

		// Check for key decisive language requirements
		hasDecisiveRequirement := strings.Contains(promptLower, "one ruling") ||
			strings.Contains(promptLower, "single") ||
			strings.Contains(promptLower, "一个裁决")

		if !hasDecisiveRequirement {
			t.Error("Prompt should require singular, decisive ruling")
		}

		// Check for rejection requirement
		hasRejectionRequirement := strings.Contains(promptLower, "reject") ||
			strings.Contains(promptLower, "拒绝")

		if !hasRejectionRequirement {
			t.Error("Prompt should require explicit rejections")
		}
	}
}
