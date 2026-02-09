# aidocs

AI-Optimized Documentation Generator - A Go tool for generating LLM-friendly documentation with structured discovery files.

## Features

- **SUMMARY.md as Source of Truth**: Uses your existing SUMMARY.md to define document structure
- **Manifest Generation**: Creates `manifest.json` with documents, sections, and metadata
- **llms.txt Generation**: Creates navigation files optimized for LLM discovery
- **AI-Powered Summaries**: Optional AI-generated descriptions and tags via Claude Code CLI
- **Change Detection**: SHA256-based caching to only regenerate changed files
- **Orphan Detection**: Find documentation files not referenced in SUMMARY.md

## Installation

```bash
go install github.com/diegosz/aidocs/cmd/aidocs@latest
```

## Quick Start

```bash
# Show version
aidocs --version

# Create default configuration
aidocs --init

# Generate documentation files
aidocs

# Preview without writing
aidocs --dry-run

# Force regeneration
aidocs --force

# Check for orphan files
aidocs --show-orphans
```

## Configuration

Create `.aidocs.yaml` in your project root:

```yaml
# Content source - defines structure and files to process
content: "docs/SUMMARY.md"

# Output settings
output:
  llms_txt: "llms.txt"
  llms_full: "docs/llms-full.txt"
  manifest: "docs/_ai/manifest.json"
  tags: "docs/_ai/tags.json"
  cache: "docs/_ai/.cache.json"

# AI features (uses Claude Code CLI - no API key needed)
ai:
  enabled: false
  generate_summaries: true
  generate_missing_frontmatter: true
  generate_descriptions: true
  generate_tags: true

# Project metadata
project:
  name: "My Project"
  description: "Project description"
  optimized_for:
    - "Claude Code"
    - "AI Agents"
```

## AI Features

When `ai.enabled: true`, aidocs uses **Claude Code CLI** (`claude -p`) for AI-powered features:

- No API key required - uses your existing Claude Code authentication
- Generates summaries, descriptions, and tags for documents
- Can auto-generate missing frontmatter

Requirements:
- Claude Code CLI installed and authenticated (`claude` command available in PATH)

## Output Files

### manifest.json

Complete index of all documents with metadata:

```json
{
  "knowledgeBase": {
    "name": "Project Name",
    "generatedBy": "aidocs",
    "optimizedFor": ["Claude Code", "AI Agents"]
  },
  "documents": [...],
  "sections": {...},
  "metadata": {
    "totalDocuments": 10,
    "averageTokensPerDoc": 500
  }
}
```

### llms.txt

Root navigation file for AI agents:

```markdown
# Project Name

> Project description

## For AI Agents
- Documents Manifest: /docs/_ai/manifest.json
- Tags Index: /docs/_ai/tags.json
- Full Index: /docs/llms-full.txt

## Usage
# (jq examples for querying manifest and tags)

## Documentation
- [Getting Started](docs/start.md): Quick start guide
```

### llms-full.txt

Complete document index organized by category.

## Frontmatter

Documents can include YAML frontmatter:

```markdown
---
title: "Document Title"
description: "One-line description"
section: "reference"
tags: ["topic1", "topic2"]
estimatedTokens: 500
---

# Document Title

Content...
```

When AI features are enabled, missing frontmatter can be auto-generated.

## LLM Usage Flow

1. AI agent discovers project via `llms.txt` (~500 tokens)
2. Agent loads `manifest.json` for full index (~2K tokens)
3. Agent queries documents by tags/sections
4. Agent fetches only relevant docs
5. Result: ~4K tokens instead of loading all docs

## Development

```bash
# Run tests
make test

# Build
make build

# Install locally
make install
```

## License

MIT
