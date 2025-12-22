package agent

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ClarificationAgent analyzes input to determine if clarifying questions are needed
type ClarificationAgent struct {
	client LLMClient
}

// ClarificationOutput represents the result of clarification analysis
type ClarificationOutput struct {
	NeedsClarification bool       `json:"needs_clarification"`
	Questions          []Question `json:"questions,omitempty"`
	Reason             string     `json:"reason,omitempty"`
}

// Question represents a clarifying question to ask the user
type Question struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Type     string   `json:"type"` // "text", "choice", "multiple_choice"
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
}

// ClarificationContext holds the user's answers to clarifying questions
type ClarificationContext struct {
	OriginalInput string            `json:"original_input"`
	Answers       map[string]string `json:"answers"` // question_id -> answer
}

// NewClarificationAgent creates a new ClarificationAgent with the given LLM client
func NewClarificationAgent(client LLMClient) *ClarificationAgent {
	return &ClarificationAgent{
		client: client,
	}
}

// Analyze checks if the input needs clarification before making a decision
func (a *ClarificationAgent) Analyze(ctx context.Context, input string) (*ClarificationOutput, error) {
	// Validate input
	if strings.TrimSpace(input) == "" {
		return nil, ErrEmptyInput
	}
	if utf8.RuneCountInString(input) > 10000 {
		return nil, ErrInputTooLong
	}

	prompt := buildClarificationPrompt(input)

	var result ClarificationOutput
	if err := a.client.CompleteJSON(ctx, prompt, &result); err != nil {
		return nil, fmt.Errorf("failed to analyze for clarification: %w", err)
	}

	// Assign IDs to questions if not set
	for i := range result.Questions {
		if result.Questions[i].ID == "" {
			result.Questions[i].ID = fmt.Sprintf("q%d", i+1)
		}
		if result.Questions[i].Type == "" {
			result.Questions[i].Type = "text"
		}
	}

	return &result, nil
}

// BuildEnrichedInput combines original input with clarification answers
func (a *ClarificationAgent) BuildEnrichedInput(clarification *ClarificationContext) string {
	if clarification == nil || len(clarification.Answers) == 0 {
		return clarification.OriginalInput
	}

	var sb strings.Builder
	sb.WriteString(clarification.OriginalInput)
	sb.WriteString("\n\n--- 补充信息 / Additional Context ---\n")

	for questionID, answer := range clarification.Answers {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", questionID, answer))
	}

	return sb.String()
}

// buildClarificationPrompt constructs the prompt for clarification analysis
func buildClarificationPrompt(input string) string {
	lang := detectLanguage(input)

	var systemPrompt string
	if lang == "zh" {
		systemPrompt = `你是一位信息分析专家。分析用户输入，判断是否需要更多上下文信息才能做出准确决策。

判断标准：
1. 输入是否涉及具体的个人情况（账号、订阅、设备等）？
2. 输入是否涉及实时变化的信息（政策、价格、流程等）？
3. 输入是否有多种可能的解读？
4. 是否缺少关键的约束条件（预算、时间、技术水平等）？

如果需要澄清，生成2-4个简洁、关键的问题。

输出格式（严格遵守JSON）：
{
  "needs_clarification": true/false,
  "reason": "为什么需要/不需要澄清",
  "questions": [
    {
      "id": "q1",
      "question": "问题内容",
      "type": "text/choice/multiple_choice",
      "options": ["选项1", "选项2"],
      "required": true/false
    }
  ]
}

如果不需要澄清，questions 数组为空。

分析以下输入：

` + input
	} else {
		systemPrompt = `You are an information analyst. Analyze the user input to determine if more context is needed for an accurate decision.

Criteria for clarification:
1. Does the input involve specific personal situations (accounts, subscriptions, devices)?
2. Does it involve real-time changing information (policies, prices, procedures)?
3. Are there multiple possible interpretations?
4. Are key constraints missing (budget, time, skill level)?

If clarification is needed, generate 2-4 concise, critical questions.

Output Format (strict JSON):
{
  "needs_clarification": true/false,
  "reason": "Why clarification is/isn't needed",
  "questions": [
    {
      "id": "q1",
      "question": "Question content",
      "type": "text/choice/multiple_choice",
      "options": ["Option 1", "Option 2"],
      "required": true/false
    }
  ]
}

If no clarification needed, questions array should be empty.

Analyze the following input:

` + input
	}

	return systemPrompt
}
