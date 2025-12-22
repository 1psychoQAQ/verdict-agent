---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T09:42:24Z
version: 2.0
author: Claude Code PM System
---

# System Patterns

## Core Architecture

### Multi-Agent Pipeline

```
User Input (fuzzy idea)
        │
        ▼
┌───────────────────┐
│ Clarification     │ ◄── Optional: Gathers context
│ Agent             │     if input is ambiguous
└─────────┬─────────┘
          │ Questions or Skip
          ▼
┌───────────────────┐
│ Web Search        │ ◄── Optional: Fetches real-time
│ Integration       │     information to reduce hallucination
└─────────┬─────────┘
          │ Search Context
          ▼
┌───────────────────┐
│   Agent A         │
│   (Verdict)       │
│   - Compresses    │
│   - Judges        │
│   - Rules         │
└─────────┬─────────┘
          │ JSON Protocol
          ▼
┌───────────────────┐
│   Agent B         │
│   (Execution)     │
│   - Accepts       │
│   - Plans         │
│   - Outputs       │
└─────────┬─────────┘
          │
          ▼
    Artifacts
    ├── decision.json
    └── todo.md
```

### Key Design Decisions

1. **Interactive Clarification**
   - System can ask clarifying questions before deciding
   - User can skip clarification for direct verdict
   - Questions are typed (text, choice, multiple_choice)

2. **Web Search Integration**
   - Optional real-time information retrieval
   - Reduces hallucination for current events
   - Multiple providers (Tavily, Google, DuckDuckGo)

3. **Structured Protocol Only**
   - Agents communicate via JSON, never free-form text
   - Protocol defines exact fields and constraints
   - No "conversation" between agents

4. **Unidirectional Authority**
   - Agent A has absolute ruling authority
   - Agent B cannot dispute or negotiate
   - Clear hierarchy prevents decision loops

5. **Immutable Artifacts**
   - Decision artifacts are frozen once created
   - Like Git tags - can't be modified after creation
   - Creates audit trail of decisions

## Agent Patterns

### Clarification Agent (NEW)

**Role:** Context Gatherer

**Behavior:**
- Analyzes if input needs clarification
- Generates targeted questions
- Returns structured question format

**Output:**
```json
{
  "needs_clarification": true,
  "questions": [
    {"id": "q1", "question": "...", "type": "choice", "options": [...], "required": true}
  ],
  "reason": "Why clarification is needed"
}
```

### Agent A (Verdict Agent)

**Role:** Judge, not consultant

**Behavior:**
- Compresses fuzzy input into structured options
- Incorporates search context for real-time info
- Delivers singular ruling (not multiple options)
- Explicitly rejects alternatives
- Outputs: ruling + rationale + rejections

**Anti-patterns to avoid:**
- "You could also consider..."
- "Another option might be..."
- Presenting multiple valid choices

### Agent B (Execution Agent)

**Role:** Executor, not planner

**Behavior:**
- Accepts Agent A's ruling without question
- Designs minimal execution path
- Outputs: MVP scope + phases + done criteria

**Anti-patterns to avoid:**
- Questioning the ruling
- Expanding scope
- Adding "nice to have" features

## Data Flow

```
Input: string (user's fuzzy idea)
    │
    ▼
Clarification Check: {
    needs_clarification: boolean,
    questions?: Question[],
    reason?: string
}
    │
    ▼ (if clarification provided or skipped)
Web Search: {
    results: SearchResult[],
    formatted_context: string
}
    │
    ▼
Agent A Output: {
    ruling: Decision,
    rationale: string,
    rejected: Decision[]
}
    │
    ▼
Agent B Output: {
    mvp_scope: Scope,
    phases: Phase[],
    done_criteria: Criteria[]
}
    │
    ▼
Artifacts: decision.json + todo.md
```

## API Patterns

### Verdict Request Flow

```
POST /api/verdict
├── Input only → Check clarification
├── Input + skip_clarify → Direct verdict
└── Input + clarification.answers → Verdict with context
```

### Response Types

```json
// Clarification needed
{"status": "clarification_needed", "questions": [...], "reason": "..."}

// Verdict complete
{"status": "verdict", "decision_id": "...", "decision": {...}, "todo": "..."}
```

## Frontend Patterns

### Progress Display
- 4-step progress indicator
- Animated active state with pulse
- Checkmark for completed steps
- Bilingual labels (EN/ZH)

### Clarification UI
- Dynamic question rendering
- Supports text, choice, multiple_choice types
- Skip option for direct verdict
- Form validation for required fields

## Artifact Patterns

### Decision Artifact (decision.json)
- Immutable once created
- Contains ruling rationale
- Timestamped with UUID
- Can be referenced but not modified

### Execution Artifact (todo.md)
- Decomposable tasks with checkboxes
- Phase-based organization
- MVP scope clearly defined
- Done criteria for validation
