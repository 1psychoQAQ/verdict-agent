package artifact

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/google/uuid"
)

const todoTemplate = `# Execution Plan: {{.Ruling}}

Generated: {{.Timestamp}}
Decision ID: {{.ID}}

## MVP Scope
{{range .MVPScope -}}
- {{.}}
{{end}}
## Phases
{{range .Phases}}
### Phase {{.Number}}: {{.Name}}
{{range .Tasks -}}
- [ ] {{.}}
{{end}}
{{end -}}
## Done Criteria
{{range .DoneCriteria -}}
- {{.}}
{{end}}`

// todoData represents the data structure for the todo template
type todoData struct {
	Ruling       string
	Timestamp    string
	ID           string
	MVPScope     []string
	Phases       []phaseData
	DoneCriteria []string
}

// phaseData represents a phase for template rendering
type phaseData struct {
	Number int
	Name   string
	Tasks  []string
}

// generateTodoMD creates the todo.md artifact
func generateTodoMD(verdict *agent.VerdictOutput, execution *agent.ExecutionOutput, id uuid.UUID, createdAt time.Time) ([]byte, error) {
	// Prepare phase data with numbering
	phases := make([]phaseData, len(execution.Phases))
	for i, phase := range execution.Phases {
		phases[i] = phaseData{
			Number: i + 1,
			Name:   phase.Name,
			Tasks:  phase.Tasks,
		}
	}

	data := todoData{
		Ruling:       verdict.Ruling,
		Timestamp:    createdAt.UTC().Format(time.RFC3339),
		ID:           id.String(),
		MVPScope:     execution.MVPScope,
		Phases:       phases,
		DoneCriteria: execution.DoneCriteria,
	}

	tmpl, err := template.New("todo").Parse(todoTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
