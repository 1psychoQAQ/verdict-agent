---
created: 2025-12-22T03:16:21Z
last_updated: 2025-12-22T03:16:21Z
version: 1.0
author: Claude Code PM System
---

# Product Context

## Problem Statement

Users suffer from:
- **Judgment Overload** - Too many options, can't decide what's worth doing
- **Decision Instability** - Constantly second-guessing choices
- **Execution Reversal** - Starting tasks then abandoning them

The core pain is NOT "not knowing how to do something" but rather "not knowing if something is worth doing right now."

## Target Users

### Primary User
- Individuals with many ideas/options
- Those who struggle with prioritization
- People who over-research before acting
- Users who want definitive guidance, not more options

### User Needs
- **Clear verdicts** - Tell me what to do, not what I could do
- **Executable plans** - Break it down so I can start immediately
- **Persistent records** - Save decisions so I don't relitigate them

## Core Functionality

### What It Does
1. Takes fuzzy, unstructured user ideas as input
2. Compresses and evaluates them through Agent A
3. Delivers a singular verdict (not suggestions)
4. Generates executable plan through Agent B
5. Outputs persistent artifacts (decision.json + todo.md)

### What It Explicitly Does NOT Do
- Open-ended brainstorming
- Present multiple "also valid" options
- Engage in back-and-forth dialogue
- Expand scope or suggest additions
- Act as a general-purpose chatbot

## Use Cases

### Primary Use Case
**Input:** "I have 5 project ideas and limited time"
**Output:**
- Verdict on which to pursue now
- Why others are rejected
- Executable first steps

### Secondary Use Case
**Input:** "Should I refactor this or add the feature first?"
**Output:**
- Definitive ruling with rationale
- Minimal execution plan
- Clear done criteria

## Success Criteria

A successful interaction produces:
1. One frozen decision (can't be re-argued)
2. One actionable todo list (can be checked off)
3. Zero ambiguity about next steps

## Language Support

- Chinese (primary)
- English
- UI and outputs must support both languages
