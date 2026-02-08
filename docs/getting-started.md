---
title: "Getting Started"
description: "Quick start guide for aidocs - generate LLM-friendly documentation"
category: "guide"
tags: ["quickstart", "installation", "setup"]
estimatedTokens: 400
---

# Getting Started

aidocs generates LLM-friendly documentation from your existing markdown files. It creates structured discovery files that help AI agents efficiently navigate your documentation.

## Installation

```bash
go install github.com/diegosz/aidocs/cmd/aidocs@latest
```

## Quick Start

1. Create a `docs/SUMMARY.md` file listing your documentation:

```markdown
# My Project

- [Introduction](intro.md)
- [API Reference](api.md)
- Examples
  - [Basic Usage](examples/basic.md)
```

2. Initialize aidocs configuration:

```bash
aidocs --init
```

3. Generate LLM optimization files:

```bash
aidocs
```

This creates:
- `llms.txt` - Root navigation for AI agents
- `docs/llms-full.txt` - Complete document index
- `docs/ai-optimization/manifest.json` - Structured metadata

## How It Works

aidocs uses `SUMMARY.md` as the single source of truth for your documentation structure. It extracts frontmatter metadata from each referenced file and generates optimized discovery files for LLM consumption.

The tool supports AI-powered features using Claude Code CLI to automatically generate:
- Document descriptions
- Relevant tags
- Concise summaries
