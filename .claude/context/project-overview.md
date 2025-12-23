---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-23T05:46:55Z
version: 2.0
author: Claude Code PM System
---

# Project Overview

## System Summary

Verdict-Agent transforms messy, unstructured ideas into decisive, executable outcomes through a strictly constrained multi-agent pipeline with web search integration and interactive clarification.

## Feature List

### Core Features (MVP) - ✅ COMPLETE

| Feature | Status | Description |
|---------|--------|-------------|
| Input Processing | ✅ Complete | Accept fuzzy user ideas as text input |
| Agent A Pipeline | ✅ Complete | Compress and judge, deliver singular verdict |
| Agent B Pipeline | ✅ Complete | Accept verdict, output execution plan |
| Decision Artifact | ✅ Complete | Generate immutable `decision.json` |
| Todo Artifact | ✅ Complete | Generate actionable `todo.md` |
| Simple UI | ✅ Complete | Display input/output interface |
| Bilingual | ✅ Complete | Chinese and English support |

### Extended Features

| Feature | Status | Description |
|---------|--------|-------------|
| Clarification Agent | ✅ Complete | Interactive questions for context gathering |
| Web Search | ✅ Complete | Real-time info via Tavily/Google/DuckDuckGo |
| Multi-LLM | ✅ Complete | OpenAI, Anthropic, Gemini support |
| Progress UI | ✅ Complete | 4-step visual progress indicator |
| Auth UI | ✅ Complete | Login/Register button foundation |

### Agent Capabilities

**Clarification Agent (NEW)**
- Analyzes input ambiguity
- Generates targeted questions
- Supports skip option for direct verdict

**Agent A (Verdict Agent)**
- Structured option compression
- Value judgment with clear rationale
- Explicit rejection of alternatives
- Web search context integration

**Agent B (Execution Agent)**
- MVP boundary definition
- Phase breakdown with tasks
- Completion criteria
- Done state definition

## Current State

**Phase:** Production Ready
**Status:** All core features implemented and tested

### What Exists
- Complete Go backend with chi router
- Multi-provider LLM integration
- Web search integration
- Interactive clarification flow
- Bilingual web UI (Chinese/English)
- PostgreSQL and in-memory storage
- Docker configuration
- Integration tests

### What's Next
- Implement actual authentication flow
- Add decision history browsing
- Consider Tavily API for better Chinese search

## Integration Points

### Input
- Text input (user's idea/problem)
- REST API endpoint: `POST /api/verdict`
- Optional clarification answers

### Output
- `decision.json` - Structured decision artifact with UUID
- `todo.md` - Markdown task list with checkboxes
- API responses with decision ID

### External Services
- LLM providers: OpenAI, Anthropic, Google Gemini
- Web search: Tavily, Google Custom Search, DuckDuckGo

## Technical Highlights

- **Language:** Go (Golang) 1.21+
- **Architecture:** Multi-agent pipeline with JSON protocol
- **Framework:** chi router with middleware
- **Database:** PostgreSQL with pgx driver
- **Artifacts:** Immutable decision records + executable todo lists
- **UI:** Embedded static files with bilingual support
- **Search:** Optional real-time web context injection
