---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T03:16:21Z
version: 1.0
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
├── .git/                 # Git repository
├── .gitignore            # Go-specific ignores
├── CLAUDE.md             # Claude Code guidance
└── mind.md               # Core requirements document
```

## Planned Structure (Post-Initialization)

```
verdict-agent/
├── cmd/                  # Application entrypoints
│   └── server/           # Main server binary
│       └── main.go
├── internal/             # Private application code
│   ├── agent/            # Agent implementations
│   │   ├── verdict/      # Agent A - Verdict/Ruling agent
│   │   └── execution/    # Agent B - Execution planner
│   ├── protocol/         # JSON protocol definitions
│   ├── artifact/         # Decision/Todo artifact handling
│   └── api/              # HTTP API handlers
├── pkg/                  # Public library code (if any)
├── web/                  # UI assets (simple display interface)
├── configs/              # Configuration files
├── docs/                 # Additional documentation
├── go.mod                # Go module definition
├── go.sum                # Dependency checksums
└── Makefile              # Build automation
```

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/` | Application entry points (main packages) |
| `internal/agent/` | Agent A and B implementations |
| `internal/protocol/` | JSON protocol structs and validation |
| `internal/artifact/` | Decision and Todo artifact generation |
| `web/` | Simple UI for displaying results |

## File Naming Conventions

- Go files: `lowercase_snake_case.go`
- Test files: `*_test.go` in same directory as code
- JSON schemas: `*.schema.json`
- Configuration: `config.yaml` or `config.json`

## Module Organization

The project will follow Go's standard layout with `internal/` for private code and clear separation between:
- Agent logic (verdict, execution)
- Protocol definitions (JSON schemas)
- Artifact generation (decision.json, todo.md)
- API layer (HTTP handlers)
