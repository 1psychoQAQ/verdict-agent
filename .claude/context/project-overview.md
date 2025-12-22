---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T03:16:21Z
version: 1.0
author: Claude Code PM System
---

# Project Overview

## System Summary

Verdict-Agent transforms messy, unstructured ideas into decisive, executable outcomes through a strictly constrained dual-agent pipeline.

## Feature List

### Core Features (MVP)

| Feature | Status | Description |
|---------|--------|-------------|
| Input Processing | Planned | Accept fuzzy user ideas as text input |
| Agent A Pipeline | Planned | Compress and judge, deliver singular verdict |
| Agent B Pipeline | Planned | Accept verdict, output execution plan |
| Decision Artifact | Planned | Generate immutable `decision.json` |
| Todo Artifact | Planned | Generate actionable `todo.md` |
| Simple UI | Planned | Display input/output interface |
| Bilingual | Planned | Chinese and English support |

### Agent Capabilities

**Agent A (Verdict Agent)**
- Structured option compression
- Value judgment with clear rationale
- Explicit rejection of alternatives
- Ordered ranking output

**Agent B (Execution Agent)**
- MVP boundary definition
- Phase breakdown
- Completion criteria
- Done state definition

## Current State

**Phase:** Pre-development
**Status:** Project initialization complete, architecture planning

### What Exists
- Project repository on GitHub
- CCPM project management system
- Core requirements documentation (`mind.md`)
- Claude Code guidance (`CLAUDE.md`)

### What's Next
- Go module initialization
- JSON protocol schema design
- Agent implementation
- Basic HTTP API
- Simple display UI

## Integration Points

### Input
- Text input (user's idea/problem)
- REST API endpoint

### Output
- `decision.json` - Structured decision artifact
- `todo.md` - Markdown task list
- GitHub Issues export (optional)

### External Services
- LLM provider (OpenAI/Anthropic) for agent reasoning

## Technical Highlights

- **Language:** Go (Golang)
- **Architecture:** Dual-agent pipeline with JSON protocol
- **Artifacts:** Immutable decision records + mutable todo lists
- **UI:** Simple web interface (no agent visualization)
