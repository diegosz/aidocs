---
title: "Frontmatter Specification"
description: "YAML frontmatter format for aidocs documentation files"
section: "reference"
tags: ["frontmatter", "yaml", "metadata", "specification"]
estimatedTokens: 400
---

# Frontmatter Specification

aidocs extracts metadata from YAML frontmatter at the top of markdown files.

## Format

```yaml
---
title: "Document Title"
description: "One-line description (max 100 chars)"
section: "reference"
tags: ["tag1", "tag2", "tag3"]
estimatedTokens: 500
summary: "2-3 sentence summary of the document content."
---

# Document Title

Content starts here...
```

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | No | Document title (falls back to H1) |
| `description` | string | No | One-line description |
| `section` | string | No | Document section |
| `tags` | string[] | No | Topic tags |
| `estimatedTokens` | int | No | Estimated token count |
| `summary` | string | No | AI-generated summary |

## Title Resolution

1. If frontmatter has `title`, use it
2. If document has H1 header, prefer H1 (authoritative)
3. Fall back to SUMMARY.md link text

## Sections

Sections can come from:
1. Frontmatter `section` field
2. Parent heading in SUMMARY.md (for nested entries)

## Without Frontmatter

Files without frontmatter are still processed:
- Title extracted from H1 header
- Section from SUMMARY.md structure
- Other fields remain empty

Enable `ai.generate_missing_frontmatter` to auto-add frontmatter.

## Change Detection

Content hashes are calculated **excluding frontmatter**. This means:
- Updating frontmatter doesn't trigger regeneration
- Only body content changes require reprocessing
