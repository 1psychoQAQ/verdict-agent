# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Verdict-Agent is a meta-decision system with two constrained AI agents:
- **Agent A (Verdict Agent)**: Makes value judgments and delivers singular rulings (not suggestions)
- **Agent B (Execution Agent)**: Accepts Agent A's ruling and designs minimal execution plans

The system transforms fuzzy ideas into executable system assets (decision.json + todo.md).

## Key Constraints

- Agents communicate only via structured JSON protocols, not free-form text
- Agent A is a judge, not a consultant - no "also possible" options
- Agent B cannot dispute Agent A - only converts rulings to actions
- System produces immutable decision artifacts (like Git tags)

## Build & Test Commands

```bash
# Build
go build ./...

# Run tests
go test ./...

# Run single test
go test -run TestName ./path/to/package

# Run tests with coverage
go test -cover ./...

# Lint (if using golangci-lint)
golangci-lint run
```

## Project Management

This project uses CCPM (Claude Code Project Management). Key commands:
- `/pm:status` - View project status
- `/pm:prd-new <name>` - Create new PRD
- `/pm:epic-start <name>` - Start working on an epic
- `/pm:help` - Full command reference

## Architecture Notes

MVP output format:
- `decision.json` - Frozen decision artifact from Agent A
- `todo.md` - Executable tasks from Agent B

Backend: Go (Golang)
UI: Simple display interface (no multi-agent visualization)
Languages: Chinese and English support required
