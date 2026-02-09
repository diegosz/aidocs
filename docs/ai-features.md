---
title: "AI Features"
description: "AI-powered documentation enhancement using Claude Code CLI"
section: "guide"
tags: ["ai", "claude", "summaries", "automation"]
estimatedTokens: 450
---

# AI Features

aidocs integrates with Claude Code CLI to automatically enhance your documentation with AI-generated metadata.

## Requirements

- Claude Code CLI installed (`claude` command available)
- Active Claude Code authentication

## Enabling AI Features

Set `ai.enabled: true` in your `.aidocs.yaml`:

```yaml
ai:
  enabled: true
  generate_summaries: true
  generate_descriptions: true
  generate_tags: true
  generate_missing_frontmatter: true
```

## What Gets Generated

### Summaries

2-3 sentence summaries capturing the key concepts of each document. Stored in the cache and included in manifest.json.

### Descriptions

One-line descriptions (max 100 chars) suitable for document listings and navigation.

### Tags

3-5 relevant topic tags in lowercase, hyphenated format (e.g., `api-reference`, `getting-started`).

### Frontmatter

When `generate_missing_frontmatter` is enabled, aidocs will add YAML frontmatter to files that lack it.

## Caching

AI-generated content is cached in `.cache.json` to avoid redundant API calls. The cache uses SHA256 hashes of document content (excluding frontmatter) to detect changes.

Only modified documents trigger AI regeneration.

## Running Without AI

When `ai.enabled: false` (the default), aidocs runs in fast mode:
- Extracts existing frontmatter metadata
- Falls back to H1 titles when no frontmatter exists
- No external API calls
- Suitable for CI/CD pipelines
