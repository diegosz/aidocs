# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [v0.0.5] - 2026-02-16

### Added

- Persist AI-inferred `project.name` and `project.description` back to `.aidocs.yaml` so subsequent runs skip the AI call.

## [v0.0.4] - 2026-02-16

### Changed

- Updated README to clarify project name and description inference from SUMMARY.md.

### Removed

- Removed `version` field from project configuration.

## [v0.0.3] - 2026-02-16

### Added

- AI inference of `project.name` and `project.description` from SUMMARY.md when left empty in config.

### Changed

- Removed unused version code.

## [v0.0.2] - 2026-02-09

### Changed

- Renamed testdata scenario folder.

## [v0.0.1] - 2026-02-09

### Added

- Initial implementation of aidocs tool.
- SUMMARY.md parsing as source of truth for document structure.
- `manifest.json` generation with documents, sections, and metadata.
- `llms.txt` and `llms-full.txt` generation for LLM discovery.
- `tags.json` generation for tag-based document discovery.
- AI-powered summaries, descriptions, and tags via Claude Code CLI.
- SHA256-based caching with `fileHash` for change detection.
- Orphan file detection for files not referenced in SUMMARY.md.
- `--init` flag to create default `.aidocs.yaml`.
- `--dry-run`, `--force`, and `--version` flags.
- YAML frontmatter extraction and auto-generation.
- Usage section with `jq` examples in `llms.txt`.
- MIT license.
