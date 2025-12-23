---
name: verdict-agent
status: completed
created: 2025-12-22T03:26:09Z
updated: 2025-12-23T03:09:03Z
progress: 100%
prd: .claude/prds/verdict-agent.md
github: https://github.com/1psychoQAQ/verdict-agent/issues/1
---

# Epic: verdict-agent

## Overview

Implement a dual-agent decision system in Go that accepts fuzzy user input and produces immutable decision artifacts (decision.json) and actionable execution plans (todo.md). The system uses structured JSON protocols between agents—no free-form dialogue—to enforce deterministic, non-negotiable verdicts.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| HTTP Framework | `chi` | Lightweight, stdlib-compatible, middleware support |
| LLM Client | Custom wrapper | Abstract OpenAI/Anthropic behind common interface |
| Database | PostgreSQL with `pgx` | Native Go driver, JSONB support, connection pooling |
| JSON Validation | `go-playground/validator` | Struct tag validation, widely used |
| Frontend | Embedded static files | Simple, no separate build step, `embed` package |
| Config | Environment variables | 12-factor app, no config files needed |

## Technical Approach

### Core Pipeline (Single Request Flow)

```
Input → Agent A (LLM) → Validate JSON → Agent B (LLM) → Generate Artifacts → Store → Response
```

Each step is synchronous within a single HTTP request. No queuing or async processing for MVP.

### Package Structure

```
cmd/server/main.go          # Entry point, wire dependencies
internal/
  agent/
    verdict.go              # Agent A implementation
    execution.go            # Agent B implementation
    llm.go                  # LLM client interface + OpenAI/Anthropic implementations
  pipeline/
    pipeline.go             # Orchestrates Agent A → Agent B → Artifacts
  artifact/
    decision.go             # decision.json generation
    todo.go                 # todo.md generation
  storage/
    postgres.go             # Database operations
  api/
    handlers.go             # HTTP handlers
    middleware.go           # Rate limiting, CORS, logging
web/
  static/                   # HTML, CSS, JS (minimal)
```

### Key Simplifications

1. **Single LLM interface** - Both agents use same client, different prompts
2. **No ORM** - Direct SQL with `pgx` for transparency
3. **Embedded frontend** - No npm/webpack, just static HTML with fetch API
4. **Atomic operations** - Single transaction for artifact + storage

## Implementation Strategy

### Development Order

1. **Foundation** - Go module, database schema, LLM client
2. **Agents** - Verdict and Execution agents with prompts
3. **Pipeline** - Orchestration and artifact generation
4. **API** - HTTP handlers and storage
5. **Frontend** - Simple form and display

### Risk Mitigation

| Risk | Mitigation |
|------|------------|
| LLM response variance | Strict JSON schema, retry with reformatted prompt |
| Slow pipeline (>10min) | Streaming response, timeout handling |
| Prompt injection | Input sanitization, structured prompts |

## Task Breakdown

| # | Task | Description |
|---|------|-------------|
| 1 | Project Setup | Go module, directory structure, PostgreSQL schema, config loading |
| 2 | LLM Client | Unified interface for OpenAI/Anthropic with retry logic |
| 3 | Agent A (Verdict) | Prompt engineering + JSON output for ruling/rejections |
| 4 | Agent B (Execution) | Prompt engineering + JSON output for MVP/phases/criteria |
| 5 | Pipeline Orchestrator | Chain agents, validate JSON, handle errors |
| 6 | Artifact Generator | Create decision.json and todo.md from pipeline output |
| 7 | Storage Layer | PostgreSQL CRUD for decisions and todos |
| 8 | HTTP API | POST /verdict, GET /decisions/{id}, GET /todos/{id} |
| 9 | Web Frontend | Input form, result display, bilingual support |
| 10 | Integration Testing | End-to-end tests with mock LLM responses |

## Dependencies

### External
- PostgreSQL 14+ instance
- OpenAI API key OR Anthropic API key

### Internal (Task Dependencies)
- Task 2 (LLM Client) blocks Tasks 3, 4
- Tasks 3, 4 block Task 5
- Task 5 blocks Task 6
- Tasks 6, 7 block Task 8
- Task 8 blocks Task 9

## Success Criteria (Technical)

| Criteria | Target |
|----------|--------|
| Pipeline latency | < 10 minutes end-to-end |
| API response (retrieval) | < 500ms |
| Test coverage | > 80% for pipeline package |
| Concurrent users | 10 simultaneous requests |
| Artifact atomicity | 100% (both or neither created) |

## Estimated Effort

| Phase | Tasks | Estimate |
|-------|-------|----------|
| Foundation | 1-2 | 1 day |
| Agents | 3-4 | 2 days |
| Pipeline | 5-6 | 1 day |
| API + Storage | 7-8 | 1 day |
| Frontend + Testing | 9-10 | 1 day |
| **Total** | 10 tasks | ~6 days |

## Notes

- Agent prompts are the most critical component—invest time in prompt engineering
- Keep frontend minimal (no React/Vue)—plain HTML + vanilla JS sufficient
- PostgreSQL JSONB allows flexible artifact storage without migrations for schema changes

## Tasks Created

| Issue | Task | Parallel | Depends On | Hours |
|-------|------|----------|------------|-------|
| [#2](https://github.com/1psychoQAQ/verdict-agent/issues/2) | Project Setup | false | - | 4-6 |
| [#3](https://github.com/1psychoQAQ/verdict-agent/issues/3) | LLM Client | false | #2 | 4-6 |
| [#4](https://github.com/1psychoQAQ/verdict-agent/issues/4) | Agent A (Verdict) | true | #3 | 6-8 |
| [#5](https://github.com/1psychoQAQ/verdict-agent/issues/5) | Agent B (Execution) | true | #3 | 6-8 |
| [#6](https://github.com/1psychoQAQ/verdict-agent/issues/6) | Pipeline Orchestrator | false | #4, #5 | 4-6 |
| [#7](https://github.com/1psychoQAQ/verdict-agent/issues/7) | Artifact Generator | false | #6 | 3-4 |
| [#8](https://github.com/1psychoQAQ/verdict-agent/issues/8) | Storage Layer | true | #2 | 4-6 |
| [#9](https://github.com/1psychoQAQ/verdict-agent/issues/9) | HTTP API | false | #7, #8 | 4-6 |
| [#10](https://github.com/1psychoQAQ/verdict-agent/issues/10) | Web Frontend | false | #9 | 4-6 |
| [#11](https://github.com/1psychoQAQ/verdict-agent/issues/11) | Integration Testing | false | #9, #10 | 4-6 |

**Summary:**
- Total tasks: 10
- Parallel tasks: 3 (#4, #5, #8)
- Sequential tasks: 7
- Estimated total effort: 44-62 hours (~6 days)
