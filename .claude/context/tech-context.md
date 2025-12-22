---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T03:16:21Z
version: 1.0
author: Claude Code PM System
---

# Technical Context

## Technology Stack

### Core Language
- **Go (Golang)** - Primary backend language
- Version: TBD (recommend Go 1.21+)

### Planned Dependencies

| Category | Technology | Purpose |
|----------|------------|---------|
| HTTP Framework | `net/http` or `chi`/`gin` | REST API endpoints |
| JSON | `encoding/json` | Protocol serialization |
| LLM Integration | OpenAI/Anthropic SDK | Agent AI backend |
| Configuration | `viper` or `envconfig` | Config management |
| Logging | `slog` (stdlib) | Structured logging |

### Development Tools

| Tool | Purpose |
|------|---------|
| `go test` | Unit and integration testing |
| `golangci-lint` | Code linting |
| `make` | Build automation |
| `gh` CLI | GitHub operations |

## Build Commands

```bash
# Initialize module
go mod init github.com/1psychoQAQ/verdict-agent

# Build
go build ./...

# Test
go test ./...

# Lint
golangci-lint run

# Run server
go run cmd/server/main.go
```

## Environment Configuration

```bash
# Required
OPENAI_API_KEY=     # or ANTHROPIC_API_KEY for Claude
PORT=8080           # HTTP server port

# Optional
LOG_LEVEL=info
CONFIG_PATH=./configs/config.yaml
```

## External Integrations

### AI Provider
- Primary: OpenAI API or Anthropic API for agent reasoning
- Agents communicate via structured JSON, not raw LLM responses

### Output Formats
- `decision.json` - Immutable decision artifact
- `todo.md` - Markdown task list

## Development Environment

- **IDE:** GoLand (JetBrains) - indicated by `.idea/` directory
- **OS:** macOS (Darwin) - indicated by project paths
- **Version Control:** Git with GitHub origin

## Notes

- No existing Go dependencies yet (no `go.mod`)
- Project uses CCPM for project management
- Bilingual support required (Chinese/English)
