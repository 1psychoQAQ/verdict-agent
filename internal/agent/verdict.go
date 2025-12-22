package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Error types
var (
	ErrInputTooLong = errors.New("input exceeds 10,000 characters")
	ErrEmptyInput   = errors.New("input cannot be empty")
	ErrEmptyRuling  = errors.New("verdict ruling is empty")
)

// VerdictAgent processes fuzzy user input and produces singular rulings
type VerdictAgent struct {
	client LLMClient
}

// NewVerdictAgent creates a new VerdictAgent with the given LLM client
func NewVerdictAgent(client LLMClient) *VerdictAgent {
	return &VerdictAgent{
		client: client,
	}
}

// Process takes user input and returns a decisive verdict with explicit rejections
func (a *VerdictAgent) Process(ctx context.Context, input string) (*VerdictOutput, error) {
	return a.ProcessWithContext(ctx, input, "")
}

// ProcessWithContext takes user input and optional search context for real-time information
func (a *VerdictAgent) ProcessWithContext(ctx context.Context, input string, searchContext string) (*VerdictOutput, error) {
	// Validate input
	if err := validateInput(input); err != nil {
		return nil, err
	}

	// Detect language and build prompt
	prompt := buildVerdictPromptWithContext(input, searchContext)

	// Call LLM
	var result VerdictOutput
	if err := a.client.CompleteJSON(ctx, prompt, &result); err != nil {
		return nil, fmt.Errorf("failed to get verdict: %w", err)
	}

	// Validate output
	if err := validateOutput(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// validateInput checks input constraints
func validateInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return ErrEmptyInput
	}
	if utf8.RuneCountInString(input) > 10000 {
		return ErrInputTooLong
	}
	return nil
}

// validateOutput ensures the verdict meets quality standards
func validateOutput(output *VerdictOutput) error {
	if strings.TrimSpace(output.Ruling) == "" {
		return ErrEmptyRuling
	}
	return nil
}

// detectLanguage returns "zh" for Chinese, "en" for English
func detectLanguage(input string) string {
	// Count Chinese characters and total characters
	chineseCount := 0
	totalChars := utf8.RuneCountInString(input)

	for _, r := range input {
		if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified Ideographs
			chineseCount++
		}
	}

	// If more than 20% of all characters are Chinese, treat as Chinese
	if totalChars > 0 && float64(chineseCount)/float64(totalChars) > 0.2 {
		return "zh"
	}
	return "en"
}

// buildVerdictPrompt constructs the system prompt based on input language (without search context)
func buildVerdictPrompt(input string) string {
	return buildVerdictPromptWithContext(input, "")
}

// buildVerdictPromptWithContext constructs the system prompt with optional search context
func buildVerdictPromptWithContext(input string, searchContext string) string {
	lang := detectLanguage(input)

	var systemPrompt string
	if lang == "zh" {
		systemPrompt = `你是一位法官，不是顾问。你的职责是做出单一、明确的裁决，而不是提供选项或建议。

核心原则：
1. 只给出一个裁决——绝不提供替代方案
2. 明确拒绝其他选项并说明理由
3. 绝不使用"你也可以"、"这取决于"、"另一个选择"等表述
4. 输出必须是有效的 JSON 格式
5. 如果提供了网络搜索结果，优先使用最新信息做出判断

输出格式（严格遵守）：
{
  "ruling": "你的唯一裁决",
  "rationale": "为什么这是正确的选择",
  "rejected": [
    {"option": "被拒绝的选项1", "reason": "拒绝的具体原因"},
    {"option": "被拒绝的选项2", "reason": "拒绝的具体原因"}
  ]
}

要求：
- ruling: 清晰、果断、可执行的单一决定
- rationale: 简洁有力的理由（2-3句话）
- rejected: 至少列出2个被拒绝的替代方案（如果适用）

严禁：
- 使用模糊语言
- 提供多个选项让用户选择
- 建议"根据情况而定"
- 在裁决中使用"可能"、"也许"等词

`
		if searchContext != "" {
			systemPrompt += "以下是与问题相关的最新网络搜索结果，请基于这些信息做出判断：\n\n" + searchContext + "\n\n"
		}
		systemPrompt += "现在，基于以下输入做出裁决：\n\n" + input
	} else {
		systemPrompt = `You are a judge, not a consultant. Your role is to deliver a SINGLE, DEFINITIVE ruling—not to offer options or suggestions.

Core Principles:
1. Deliver ONE ruling—no alternatives
2. Explicitly reject other options with reasons
3. Never use phrases like "you could also", "it depends", "another option would be"
4. Output ONLY valid JSON matching the schema
5. If web search results are provided, prioritize using the latest information

Output Format (strict adherence required):
{
  "ruling": "Your singular verdict",
  "rationale": "Why this is the correct choice",
  "rejected": [
    {"option": "Rejected option 1", "reason": "Specific reason for rejection"},
    {"option": "Rejected option 2", "reason": "Specific reason for rejection"}
  ]
}

Requirements:
- ruling: Clear, decisive, actionable single decision
- rationale: Concise, powerful reasoning (2-3 sentences)
- rejected: List at least 2 rejected alternatives (if applicable)

Prohibited:
- Hedging language
- Providing multiple options for user to choose from
- Suggesting "it depends on the situation"
- Using "maybe", "possibly", "could" in the ruling

`
		if searchContext != "" {
			systemPrompt += "The following are recent web search results relevant to the query. Use this information to make your judgment:\n\n" + searchContext + "\n\n"
		}
		systemPrompt += "Now, deliver your verdict based on the following input:\n\n" + input
	}

	return systemPrompt
}
