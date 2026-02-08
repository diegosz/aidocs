---
title: "CLI Reference"
description: "Command-line interface reference for aidocs"
category: "reference"
tags: ["cli", "commands", "flags"]
estimatedTokens: 350
---

# CLI Reference

## Usage

```bash
aidocs [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Path to config file (default: `.aidocs.yaml`) |
| `--init` | Create default `.aidocs.yaml` configuration |
| `--force` | Regenerate all files, ignore cache |
| `--dry-run` | Preview changes without writing files |
| `--show-orphans` | List markdown files not in SUMMARY.md |
| `-v, --verbose` | Enable verbose output |

## Examples

### Initialize a new project

```bash
aidocs --init
```

Creates a default `.aidocs.yaml` in the current directory.

### Generate documentation

```bash
aidocs
```

Processes SUMMARY.md and generates output files.

### Preview changes

```bash
aidocs --dry-run -v
```

Shows what would be generated without writing files.

### Force regeneration

```bash
aidocs --force
```

Ignores cache and regenerates all files.

### Find orphan files

```bash
aidocs --show-orphans
```

Lists markdown files in the docs directory that are not referenced in SUMMARY.md.

### Custom config location

```bash
aidocs --config path/to/config.yaml
```

Uses a config file from a different location.
