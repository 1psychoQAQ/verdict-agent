---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T03:16:21Z
version: 1.0
author: Claude Code PM System
---

# Project Style Guide

## Go Code Style

### General Principles
- Follow standard Go conventions (Effective Go, Go Code Review Comments)
- Use `gofmt` for formatting (automatic)
- Prefer simplicity over cleverness

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Packages | lowercase, single word | `verdict`, `protocol` |
| Exported functions | PascalCase | `ProcessVerdict()` |
| Unexported functions | camelCase | `parseInput()` |
| Interfaces | -er suffix when appropriate | `Decider`, `Executor` |
| Struct types | PascalCase | `VerdictResult` |
| Constants | PascalCase or SCREAMING_SNAKE | `MaxRetries`, `DEFAULT_PORT` |

### File Organization

```go
// Package declaration
package verdict

// Imports (grouped: stdlib, external, internal)
import (
    "context"
    "encoding/json"

    "github.com/pkg/errors"

    "github.com/1psychoQAQ/verdict-agent/internal/protocol"
)

// Constants and package-level vars
const (
    DefaultTimeout = 30 * time.Second
)

// Types
type Verdict struct { ... }

// Functions (public first, then private)
func NewVerdict() *Verdict { ... }

func (v *Verdict) Process() error { ... }

func parseResponse(data []byte) (*Response, error) { ... }
```

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to process verdict: %w", err)
}

// Define package-level errors for common cases
var (
    ErrInvalidInput = errors.New("invalid input")
    ErrVerdictFailed = errors.New("verdict generation failed")
)
```

## JSON Protocol Style

### Field Naming
- Use snake_case for JSON fields
- Use camelCase for Go struct fields with json tags

```go
type DecisionArtifact struct {
    Ruling     string    `json:"ruling"`
    Rejected   []string  `json:"rejected"`
    CreatedAt  time.Time `json:"created_at"`
    IsFinal    bool      `json:"is_final"`
}
```

### Schema Structure
- All protocols must be versioned
- Include `type` field for polymorphic structures
- Timestamps use ISO 8601 format

## Comment Style

### Package Comments
```go
// Package verdict implements Agent A, the value judgment agent
// that delivers singular rulings on user ideas.
package verdict
```

### Function Comments
```go
// Process takes a user's fuzzy idea and returns a definitive verdict.
// It compresses options, evaluates worth, and delivers a single ruling.
// Returns ErrInvalidInput if the input cannot be processed.
func (v *Verdict) Process(input string) (*Result, error) {
```

### Inline Comments
- Explain "why", not "what"
- Place above the code, not at end of line

## Testing Style

### Test Naming
```go
func TestVerdict_Process_ValidInput(t *testing.T) { ... }
func TestVerdict_Process_EmptyInput_ReturnsError(t *testing.T) { ... }
```

### Table-Driven Tests
```go
tests := []struct {
    name    string
    input   string
    want    *Result
    wantErr bool
}{
    {"valid input", "build feature X", &Result{...}, false},
    {"empty input", "", nil, true},
}
```

## Git Commit Style

- Imperative mood: "Add verdict processing" not "Added..."
- Reference issues: "Issue #123: Add verdict processing"
- Keep subject line under 72 characters

## Documentation Style

- Keep docs minimal and actionable
- Prefer examples over explanations
- Update docs with code changes
- No excessive prose
