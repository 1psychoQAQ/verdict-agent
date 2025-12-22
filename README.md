# Verdict Agent

A multi-agent pipeline system for generating and executing verdicts using LLMs.

## Project Structure

```
verdict-agent/
├── cmd/server/          # Server entry point
├── internal/            # Internal packages
│   ├── agent/          # Agent logic
│   ├── api/            # HTTP API handlers
│   ├── artifact/       # Artifact management
│   ├── config/         # Configuration loading
│   ├── pipeline/       # Pipeline processing
│   └── storage/        # Data persistence
├── migrations/          # Database migrations
├── web/static/         # Static assets
└── docker-compose.yml  # Local development environment
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose (for PostgreSQL)

### Setup

1. Start PostgreSQL:
```bash
docker-compose up -d
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env with your API keys
```

3. Run the server:
```bash
go run cmd/server/main.go
```

4. Test health check:
```bash
curl http://localhost:8080/health
# Response: {"status":"ok"}
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| DATABASE_URL | Yes | - | PostgreSQL connection string |
| LLM_PROVIDER | Yes | openai | LLM provider: 'openai' or 'anthropic' |
| OPENAI_API_KEY | Conditional | - | Required if LLM_PROVIDER=openai |
| ANTHROPIC_API_KEY | Conditional | - | Required if LLM_PROVIDER=anthropic |
| PORT | No | 8080 | Server port |

## Database Schema

The database includes:

- `decisions` - Stores verdicts and their inputs
- `todos` - Stores action items linked to decisions

See `migrations/001_initial.sql` for the full schema.

## Development

### Build
```bash
go build ./...
```

### Test
```bash
go test ./...
```

### Run with custom port
```bash
PORT=8081 go run cmd/server/main.go
```

## API Endpoints

### Health Check
```
GET /health
Response: {"status":"ok"}
```

## License

TBD
