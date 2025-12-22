---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T03:16:21Z
version: 1.0
author: Claude Code PM System
---

# System Patterns

## Core Architecture

### Dual-Agent Pipeline

```
User Input (fuzzy idea)
        │
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

1. **Structured Protocol Only**
   - Agents communicate via JSON, never free-form text
   - Protocol defines exact fields and constraints
   - No "conversation" between agents

2. **Unidirectional Authority**
   - Agent A has absolute ruling authority
   - Agent B cannot dispute or negotiate
   - Clear hierarchy prevents decision loops

3. **Immutable Artifacts**
   - Decision artifacts are frozen once created
   - Like Git tags - can't be modified after creation
   - Creates audit trail of decisions

## Agent Patterns

### Agent A (Verdict Agent)

**Role:** Judge, not consultant

**Behavior:**
- Compresses fuzzy input into structured options
- Delivers singular ruling (not multiple options)
- Explicitly rejects alternatives
- Outputs: ruling + ranking + rejections

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
Agent A Output: {
    ruling: Decision,
    rejected: Decision[],
    ranking: number[]
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

## Artifact Patterns

### Decision Artifact (decision.json)
- Immutable once created
- Contains ruling rationale
- Timestamped
- Can be referenced but not modified

### Execution Artifact (todo.md)
- Decomposable tasks
- Checkable items
- Exportable to GitHub Issues
- Progressive state tracking

## Template-Based Output

Agents output structured data, not documents:
- JSON + templates = rendered documents
- Documents are reproducible from state
- No prose generation by agents
