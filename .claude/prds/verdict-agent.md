---
name: verdict-agent
description: Meta-decision system with dual constrained agents that delivers verdicts and executable plans
status: completed
created: 2025-12-22T03:21:03Z
updated: 2025-12-22T05:41:20Z
---

# PRD: verdict-agent

## Executive Summary

Verdict-Agent is an AI-powered meta-decision system that transforms chaotic, unstructured ideas into decisive, executable outcomes. It uses two mutually-constrained agents—Agent A (Verdict) for value judgment and Agent B (Execution) for action planning—communicating exclusively via structured JSON protocols to deliver immutable decision artifacts and actionable todo lists.

The system solves judgment overload by delivering verdicts, not suggestions. It removes optionality by design.

## Problem Statement

### The Core Problem

Users suffer from three interconnected issues:
1. **Judgment Overload** - Too many options lead to decision paralysis
2. **Decision Instability** - Constant second-guessing and re-evaluation
3. **Execution Reversal** - Starting tasks then abandoning them due to doubt

### Why This Matters Now

AI assistants excel at generating options but fail at reducing them. Users don't need more ideas—they need fewer, with conviction. Existing tools act as consultants offering "you could also consider..." when users need a judge who rules definitively.

### Root Cause

The problem is NOT:
- Lack of information
- Need for more ideas
- Insufficient planning tools

The problem IS:
- No authority figure to make final calls
- No mechanism to freeze decisions
- No system that says "this, not that" definitively

## User Stories

### Primary Persona: The Overwhelmed Builder

**Profile:**
- Has multiple project ideas competing for attention
- Spends more time deciding what to work on than working
- Often starts projects, questions the decision, then abandons them
- Wants external validation but gets more options instead

**User Journey:**

1. **Input Phase**
   - User enters a fuzzy idea or set of competing options
   - Example: "I have 3 project ideas: a todo app, a habit tracker, or a journaling tool. I have 2 weeks."

2. **Verdict Phase**
   - Agent A compresses input into structured options
   - Agent A delivers singular ruling with explicit rejections
   - User receives: "Build the habit tracker. Reject: todo app (saturated market), journaling tool (scope too large for 2 weeks)."

3. **Execution Phase**
   - Agent B accepts verdict without dispute
   - Agent B outputs MVP scope, phases, and done criteria
   - User receives actionable todo.md with checkable items

4. **Artifact Phase**
   - decision.json frozen (immutable)
   - todo.md ready for execution
   - User begins work immediately

**Acceptance Criteria:**
- [ ] Single input produces exactly two artifacts
- [ ] Decision cannot be modified after creation
- [ ] Todo items are immediately actionable
- [ ] No "also consider" language anywhere in output

### Secondary Persona: The Technical Decision Maker

**Profile:**
- Needs to choose between technical approaches
- Paralyzed by "it depends" answers from AI tools
- Wants architecture decisions locked in before implementation

**User Journey:**
- Input: "Should I use REST or GraphQL for this internal API?"
- Verdict: "Use REST. Rejection rationale: GraphQL adds complexity without benefit for internal-only, single-client API."
- Execution: Clear implementation steps with done criteria

**Acceptance Criteria:**
- [ ] Technical tradeoffs are evaluated, not listed
- [ ] One approach is selected, others explicitly rejected
- [ ] Implementation plan is technology-specific

## Requirements

### Functional Requirements

#### FR1: Input Processing
- Accept free-form text input (Chinese or English)
- No structured format required from user
- Maximum input length: 10,000 characters

#### FR2: Agent A (Verdict Agent)
- Compress fuzzy input into structured options
- Evaluate each option against context and constraints
- Deliver singular ruling (not multiple valid options)
- Explicitly list rejected options with rationale
- Output ordered ranking when multiple items exist
- Communicate only via JSON protocol (no prose generation)

#### FR3: Agent B (Execution Agent)
- Accept Agent A's verdict without dispute
- Cannot question or modify the ruling
- Generate MVP scope definition
- Break execution into phases
- Define clear "done" criteria
- Output actionable checklist format

#### FR4: Artifact Generation
- Generate `decision.json` (immutable after creation)
- Generate `todo.md` (mutable, checkable)
- Support export to GitHub Issues
- Include timestamps and versioning

#### FR5: Web Interface
- Simple input form for ideas
- Display verdict results clearly
- Show both artifacts after processing
- Support Chinese and English UI

#### FR6: API Endpoints
- `POST /api/verdict` - Submit idea, receive verdict
- `GET /api/decisions/{id}` - Retrieve past decision
- `GET /api/todos/{id}` - Retrieve todo list

### Non-Functional Requirements

#### NFR1: Performance
- Pipeline completion: < 10 minutes (including LLM calls)
- API response for retrieval: < 500ms
- Support concurrent requests: 10 simultaneous users minimum

#### NFR2: Reliability
- Artifact creation is atomic (both files or neither)
- Failed LLM calls trigger graceful retry (3 attempts)
- Partial failures do not corrupt state

#### NFR3: Security
- API keys stored securely (environment variables)
- No user input passed directly to shell commands
- Rate limiting on public endpoints

#### NFR4: Scalability
- PostgreSQL for decision/todo storage and history
- Stateless application servers
- Horizontal scaling capability

#### NFR5: Internationalization
- UI supports Chinese and English
- Agent outputs respect input language
- Error messages localized

### Technical Requirements

#### TR1: LLM Integration
- Support OpenAI API (GPT-4 family)
- Support Anthropic API (Claude family)
- Provider configurable via environment variable
- Graceful fallback between providers

#### TR2: Protocol
- Agent-to-agent communication via JSON only
- No free-form text between agents
- Schema-validated at each handoff

#### TR3: Storage
- PostgreSQL for persistent storage
- Store decision artifacts with full history
- Query capability for past decisions

## Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| Pipeline completion rate | > 95% | Successful artifact generation / total requests |
| User re-litigation rate | < 5% | Users attempting to modify frozen decisions |
| Artifact completeness | 100% | Both files generated per request |
| Decision uniqueness | 100% | Single ruling per request (no "also valid" options) |
| API availability | > 99% | Uptime during business hours |

## Constraints & Assumptions

### Technical Constraints
- Go 1.21+ required for stdlib improvements
- PostgreSQL 14+ for JSONB support
- LLM API availability dependent on external provider

### Design Constraints
- Agents MUST NOT engage in dialogue
- Decisions MUST be immutable after creation
- Agent B CANNOT dispute Agent A
- No multi-agent visualization in UI

### Assumptions
- Users have valid LLM API credentials
- Single-user decisions (not team consensus)
- English and Chinese language support sufficient for MVP

## Out of Scope

The following are explicitly NOT part of this PRD:

1. **Multi-agent visualization** - No UI showing agent "conversations"
2. **Decision history browsing** - Beyond simple retrieval by ID
3. **Collaborative decisions** - No multi-user input aggregation
4. **Mobile application** - Web-only for MVP
5. **Offline mode** - Requires active LLM API connection
6. **Custom agent training** - Uses base LLM capabilities only
7. **Integration with external PM tools** - Beyond GitHub Issues export
8. **Decision modification** - Artifacts are immutable by design
9. **Brainstorming features** - System delivers verdicts, not ideas

## Dependencies

### External Dependencies
- OpenAI API access (primary or fallback LLM)
- Anthropic API access (primary or fallback LLM)
- PostgreSQL database instance

### Internal Dependencies
- JSON protocol schema must be defined before agent implementation
- Artifact templates must be finalized before generation logic
- API design must precede frontend development

## Technical Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                      Web UI                             │
│                (Input Form + Display)                   │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                    HTTP API Layer                       │
│              (Go net/http or chi/gin)                   │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                  Verdict Pipeline                       │
│  ┌─────────────┐    JSON     ┌─────────────┐           │
│  │  Agent A    │ ──────────► │  Agent B    │           │
│  │  (Verdict)  │  Protocol   │ (Execution) │           │
│  └─────────────┘             └─────────────┘           │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                 Artifact Generator                      │
│         (decision.json + todo.md creation)              │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                    PostgreSQL                           │
│            (Decisions, Todos, History)                  │
└─────────────────────────────────────────────────────────┘
```

### Data Flow

1. User submits idea via Web UI or API
2. Pipeline orchestrator invokes Agent A with structured prompt
3. Agent A returns verdict JSON (ruling + rejections + ranking)
4. Pipeline validates Agent A output against schema
5. Pipeline invokes Agent B with Agent A's verdict
6. Agent B returns execution JSON (MVP + phases + criteria)
7. Artifact generator creates decision.json and todo.md
8. Artifacts stored in PostgreSQL
9. Response returned to user with both artifacts

## Appendix

### decision.json Schema

```json
{
  "id": "uuid",
  "created_at": "ISO-8601",
  "input": "original user input",
  "verdict": {
    "ruling": "the chosen option",
    "rationale": "why this was chosen",
    "rejected": [
      {"option": "name", "reason": "why rejected"}
    ],
    "ranking": [1, 2, 3]
  },
  "is_final": true
}
```

### todo.md Template

```markdown
# Execution Plan: {ruling}

Generated: {timestamp}
Decision ID: {uuid}

## MVP Scope
- {scope item 1}
- {scope item 2}

## Phases

### Phase 1: {name}
- [ ] {task 1}
- [ ] {task 2}

### Phase 2: {name}
- [ ] {task 1}

## Done Criteria
- {criterion 1}
- {criterion 2}
```
