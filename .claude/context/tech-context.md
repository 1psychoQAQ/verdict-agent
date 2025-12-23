---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-23T05:46:55Z
version: 2.1
author: Claude Code PM System
---

# Technical Context

## Technology Stack

### Core Language
- **Go (Golang)** - Primary backend language
- Version: Go 1.21+

### Current Dependencies

| Category | Technology | Purpose |
|----------|------------|---------|
| HTTP Framework | `chi` | REST API routing with middleware |
| JSON | `encoding/json` | Protocol serialization |
| LLM Integration | Custom clients | OpenAI, Anthropic, Gemini support |
| Configuration | `godotenv` | Environment variable loading |
| Database | `pgx` | PostgreSQL driver |
| UUID | `google/uuid` | Decision ID generation |
| File Embedding | `embed` | Static file serving |

### LLM Providers

| Provider | Model | Package |
|----------|-------|---------|
| OpenAI | gpt-4o | Custom HTTP client |
| Anthropic | claude-sonnet-4-20250514 | Custom HTTP client |
| Google | gemini-2.5-flash | Custom HTTP client |

### Web Search Providers

| Provider | API | Status |
|----------|-----|--------|
| Tavily | REST API | Supported |
| Google | Custom Search | Supported |
| DuckDuckGo | Instant Answer | Supported (limited) |

### Development Tools

| Tool | Purpose |
|------|---------|
| `go test` | Unit and integration testing |
| `golangci-lint` | Code linting |
| `make` | Build automation |
| `gh` CLI | GitHub operations |
| Docker | Containerization |

## Build Commands

```bash
# Build
go build -o verdict-server ./cmd/server

# Test
go test ./...

# Lint
golangci-lint run

# Run server
./verdict-server
# or
go run cmd/server/main.go
```

## Environment Configuration

```bash
# Required - LLM Provider (choose one)
OPENAI_API_KEY=         # For OpenAI
ANTHROPIC_API_KEY=      # For Anthropic
GEMINI_API_KEY=         # For Google Gemini
LLM_PROVIDER=gemini     # openai, anthropic, or gemini

# Required - Database
DATABASE_URL=postgres://...

# Optional - Server
PORT=9999               # HTTP server port (default: 8080)

# Optional - Web Search
SEARCH_ENABLED=true
SEARCH_PROVIDER=duckduckgo  # tavily, google, or duckduckgo
TAVILY_API_KEY=         # For Tavily search
GOOGLE_SEARCH_KEY=      # For Google Custom Search
```

## External Integrations

### AI Provider
- Multi-provider support: OpenAI, Anthropic, Gemini
- Agents communicate via structured JSON
- Search context injected into prompts

### Web Search
- Optional real-time information retrieval
- Helps reduce model hallucination for current events
- DuckDuckGo works without API key

### Output Formats
- `decision.json` - Immutable decision artifact
- `todo.md` - Markdown task list with checkboxes

## Development Environment

- **IDE:** GoLand (JetBrains)
- **OS:** macOS (Darwin)
- **Version Control:** Git with GitHub origin
- **Database:** PostgreSQL

## Notes

- Server embeds static files at compile time
- Rebuild required after static file changes
- godotenv auto-loads `.env` file
- Bilingual support (Chinese/English) in frontend
- In-memory storage available for development/testing
- Auth UI present, backend implementation ready for extension
