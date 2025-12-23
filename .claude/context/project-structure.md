---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-23T05:46:55Z
version: 2.0
author: Claude Code PM System
---

# Project Structure

## Current Directory Layout

```
verdict-agent/
├── .claude/              # Claude Code PM system
│   ├── agents/           # Agent definitions
│   ├── commands/         # PM slash commands
│   ├── context/          # Project context (this directory)
│   ├── epics/            # Epic tracking
│   ├── hooks/            # Git hooks
│   ├── prds/             # Product requirement documents
│   ├── rules/            # Development rules
│   └── scripts/          # PM automation scripts
├── cmd/
│   └── server/
│       └── main.go       # Application entry point
├── internal/
│   ├── agent/            # LLM agent implementations
│   │   ├── llm.go        # Multi-provider LLM client
│   │   ├── verdict.go    # Agent A - Verdict/Ruling
│   │   ├── execution.go  # Agent B - Execution planner
│   │   ├── clarification.go # Clarification agent
│   │   └── types.go      # Shared types
│   ├── api/              # HTTP API layer
│   │   ├── handlers.go   # Request handlers
│   │   ├── auth_handlers.go # Authentication handlers
│   │   ├── middleware.go # CORS, rate limiting, auth
│   │   └── routes.go     # Route definitions
│   ├── artifact/         # Artifact generation
│   │   ├── decision.go   # decision.json generator
│   │   ├── todo.go       # todo.md generator
│   │   └── generator.go  # Unified generator
│   ├── config/
│   │   └── config.go     # Environment configuration
│   ├── pipeline/
│   │   └── pipeline.go   # Agent orchestration
│   ├── search/
│   │   └── search.go     # Web search integration
│   └── storage/
│       ├── storage.go    # Storage interface
│       ├── postgres.go   # PostgreSQL implementation
│       └── memory.go     # In-memory implementation
├── migrations/
│   └── 001_initial.sql   # Database schema
├── tests/
│   ├── integration/      # Integration tests
│   └── mocks/            # Mock implementations
├── web/
│   ├── embed.go          # Static file embedding
│   └── static/           # Frontend assets
│       ├── index.html    # Main HTML
│       ├── app.js        # Application logic
│       ├── styles.css    # Styling
│       └── i18n.js       # Internationalization
├── .env.example          # Environment template
├── docker-compose.yml    # Docker configuration
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
├── CLAUDE.md             # Claude Code guidance
└── README.md             # Project documentation
```

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/server/` | HTTP server entry point |
| `internal/agent/` | LLM agents (verdict, execution, clarification) |
| `internal/api/` | HTTP handlers, middleware, routes |
| `internal/artifact/` | Decision and Todo artifact generation |
| `internal/pipeline/` | Agent orchestration pipeline |
| `internal/search/` | Web search providers (Tavily, Google, DuckDuckGo) |
| `internal/storage/` | Data persistence (PostgreSQL, memory) |
| `web/static/` | Frontend UI with bilingual support |
| `tests/` | Integration tests and mocks |
| `migrations/` | Database migrations |

## File Naming Conventions

- Go files: `lowercase_snake_case.go`
- Test files: `*_test.go` in same directory as code
- SQL migrations: `NNN_description.sql`
- Static assets: lowercase with appropriate extensions

## Module Organization

The project follows Go's standard layout with `internal/` for private code:
- **Agent layer**: LLM interactions, prompt engineering
- **Pipeline layer**: Agent orchestration, flow control
- **API layer**: HTTP handlers, authentication, middleware
- **Storage layer**: PostgreSQL and in-memory backends
- **Artifact layer**: Output generation (decision.json, todo.md)
- **Search layer**: Real-time web search integration
