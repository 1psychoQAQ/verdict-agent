---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T03:16:21Z
version: 1.0
author: Claude Code PM System
---

# Project Brief

## One-Line Summary

An AI decision engine that compresses chaotic thinking into executable system state through two constrained, cooperating agents.

## What It Is

**Verdict-Agent** is a meta-decision system that uses two mutually-constrained AI agents to:
1. Make definitive rulings on what's worth doing (Agent A)
2. Convert those rulings into minimal executable plans (Agent B)

## Why It Exists

### The Real Problem
Users don't lack ideas or information - they suffer from:
- Judgment overload (too many options)
- Decision instability (constant re-evaluation)
- Execution reversal (starting then stopping)

### The Solution
A system that delivers verdicts, not suggestions. It removes optionality by design.

## Goals

1. **Eliminate decision paralysis** - One clear ruling, not multiple options
2. **Create executable outcomes** - Output is actionable todo list, not prose
3. **Persist decisions** - Artifacts are frozen, preventing re-litigation
4. **Maintain simplicity** - Minimal UI, no multi-agent visualization

## Non-Goals

- Being a general-purpose AI chatbot
- Generating ideas or brainstorming
- Providing balanced "pros and cons"
- Expanding scope or suggesting additions
- Creating extensive documentation

## Success Metrics

| Metric | Target |
|--------|--------|
| Input → Output pipeline | Complete in single interaction |
| Decision artifacts | 100% immutable after creation |
| Execution artifacts | Exportable to GitHub Issues |
| User re-litigation | 0% (decisions are final) |

## Scope

### MVP Scope
- Single input → two file outputs
- `decision.json` (verdict from Agent A)
- `todo.md` (execution plan from Agent B)
- Simple display UI
- Chinese/English support

### Out of Scope (MVP)
- Multi-agent visualization
- Decision history browsing
- Collaborative features
- Mobile app
