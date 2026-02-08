---
title: "Configuration"
description: "Complete configuration reference for .aidocs.yaml"
category: "reference"
tags: ["config", "yaml", "settings"]
estimatedTokens: 500
---

# Configuration

aidocs is configured via `.aidocs.yaml` in your project root.

## Configuration File

```yaml
# .aidocs.yaml - aidocs configuration

# Content source - defines structure and files to process
content: "docs/SUMMARY.md"

# Output settings
output:
  llms_txt: "llms.txt"                    # Root navigation file
  llms_full: "docs/llms-full.txt"         # Complete document index
  manifest: "docs/ai-optimization/manifest.json"
  cache: "docs/ai-optimization/.cache.json"

# AI features (uses Claude Code CLI)
ai:
  enabled: false                          # Enable AI-powered features
  generate_summaries: true                # Generate AI summaries
  generate_missing_frontmatter: true      # Add frontmatter to files
  generate_descriptions: true             # Generate descriptions
  generate_tags: true                     # Generate tags

# Project metadata
project:
  name: "My Project"
  description: "Project description"
  version: ""                             # Empty = use git tag
  optimized_for:
    - "Claude Code"
    - "AI Agents"
```

## Configuration Options

### content

Path to your SUMMARY.md file. This file defines the documentation structure.

### output

- `llms_txt`: Path for the root llms.txt file (default: `llms.txt`)
- `llms_full`: Path for the complete index (default: `docs/llms-full.txt`)
- `manifest`: Path for manifest.json (default: `docs/ai-optimization/manifest.json`)
- `cache`: Path for cache file (default: `docs/ai-optimization/.cache.json`)

### ai

AI features require Claude Code CLI to be installed and authenticated.

- `enabled`: Master switch for AI features
- `generate_summaries`: Create 2-3 sentence summaries
- `generate_missing_frontmatter`: Add frontmatter to files without it
- `generate_descriptions`: Generate one-line descriptions
- `generate_tags`: Generate relevant topic tags

### project

Metadata included in generated files:

- `name`: Project name
- `description`: Short description
- `version`: Version string (empty uses git tag)
- `optimized_for`: List of target AI systems
