# aidocs - AI-Optimized Documentation Generator

## Executive Summary

A Go tool for generating LLM-friendly documentation with structured discovery files (manifest.json, tags.json, llms.txt). The tool uses SUMMARY.md as the source of truth for content structure, supports AI-powered summary generation via Claude Code CLI, and includes intelligent change detection.

**Tool Name:** `aidocs`
**Repository:** Separate repository (new project)
**Test Project:** `/home/diegos/_dev/goexogo/alibs/blind` (this repo serves as live test/example)

---

## Complete Architecture

```
CONFIG                          TOOL                       OUTPUT FILES

.aidocs.yaml ──────────────────→ aidocs ──┬───────────→ llms.txt (root)
  │                                │      │
  └─ content: docs/SUMMARY.md      │      ├───────────→ docs/llms-full.txt
                                   │      │
                                   ▼      └───────────→ docs/_ai/
                            Claude Code CLI                ├── manifest.json
                            (claude -p)                    ├── tags.json
                                                           ├── .cache.json
                                                           └── .gitignore (ignores .cache.json)

SOURCE FILES (referenced by SUMMARY.md)

docs/
├── SUMMARY.md ─────────────────→ Defines structure & content
├── records.md ─────────────────→ Must have frontmatter + H1 title
├── record_types.md                (aidocs adds missing frontmatter)
├── keys.md
├── record_keys.md
├── record_encoding.md
├── record_example.md
├── keys_example.md
├── dev_records.md
└── dev_services.md

EXISTING TOOLS (unchanged)

stitchmd ──────────────────────→ HANDBOOK.md
gomarkdoc ─────────────────────→ DOC.md (optional)
```

**Key principle:** SUMMARY.md is the single source of truth for what content to process.

---

## Configuration File: `.aidocs.yaml`

```yaml
# .aidocs.yaml - aidocs configuration
# All paths relative to this file's location

# Content source - defines structure and files to process
content: "docs/SUMMARY.md"

# Output settings
output:
  llms_txt: "llms.txt"                    # Root navigation (auto-generated if missing)
  llms_full: "docs/llms-full.txt"         # Complete index
  manifest: "docs/_ai/manifest.json"
  tags: "docs/_ai/tags.json"              # Aggregated tags index
  cache: "docs/_ai/.cache.json"

# AI features (uses Claude Code CLI - no API key needed)
ai:
  enabled: true                           # Enable AI-powered features
  generate_summaries: true                # Generate AI summaries for each doc
  generate_missing_frontmatter: true      # Add frontmatter to files missing it
  generate_descriptions: true             # Generate missing descriptions
  generate_tags: true                     # Generate missing tags

# Project metadata
project:
  name: "Blind Record System"
  description: "Secure record management with blind encryption"
  version: ""                             # Empty = use git tag
  optimized_for:
    - "Claude Code"
    - "BMAD BMM"
    - "AI Agents"
```

**Default behavior (no config file):**
- Looks for `docs/SUMMARY.md`
- Generates `llms.txt` at project root
- AI features disabled (fast mode)

---

## Reference Implementation & Dependencies

**Inspiration:** `$GODEV_KNOWLEDGE_FOLDER/tools/update-manifests/main.go`

**AI Integration:** Uses **Claude Code CLI** (`claude -p`)
- No API key required - uses existing Claude Code authentication
- Simple shell execution: `claude -p "prompt"`
- Parses JSON response for structured metadata
- No external LLM dependencies needed

```go
// Uses Claude Code CLI for AI features
func callClaudeCLI(prompt string) (*GeneratedMeta, error) {
    cmd := exec.Command("claude", "-p", prompt)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    return parseMetaJSON(string(output))
}
```

**Test/Example Project:** `/home/diegos/_dev/goexogo/alibs/blind`
- SUMMARY.md structure: `docs/SUMMARY.md`
- Existing docs: 9 markdown files
- Current frontmatter: None (will be generated)
- Expected output: manifest.json, llms.txt, llms-full.txt

---

## Key Design Decisions

### 1. SUMMARY.md as Source of Truth

- Only files referenced in SUMMARY.md are processed
- Structure in SUMMARY.md defines categories/hierarchy
- Files NOT in SUMMARY.md are ignored (with optional warning flag)

```markdown
# Project Title

- [Blind Records](records.md)       ← processed
- [Record Types](record_types.md)   ← processed
- WIP
  - [Dev Notes](dev_records.md)     ← processed, category: "WIP"
```

### 2. Frontmatter + H1 Title Requirement

Files MUST have both frontmatter AND matching H1 title:

```markdown
---
title: "Blind Records"
description: "Core blind record concepts"
category: reference
tags: [records, encryption]
estimatedTokens: 500
---

# Blind Records

Content starts here...
```

**If missing:** `aidocs` can auto-generate frontmatter using LLM (when `ai.generate_missing_frontmatter: true`).

### 3. SHA256 Change Detection

Hash is calculated on **content + H1 title, EXCLUDING frontmatter**:

```go
func contentHash(filePath string) string {
    content := readFile(filePath)
    // Remove frontmatter (between --- markers)
    body := stripFrontmatter(content)
    // Hash the body (includes H1 title)
    return sha256(body)
}
```

**Why exclude frontmatter?**
- Frontmatter may be updated by aidocs (descriptions, tags)
- Actual content changes should trigger regeneration
- H1 title is part of content identity

### 4. Orphan File Detection

Flag `--show-orphans` lists docs/*.md files NOT in SUMMARY.md:

```bash
$ aidocs --show-orphans
⚠ Orphan files (not in SUMMARY.md):
  - docs/old_draft.md
  - docs/temp_notes.md
```

### 5. llms.txt Auto-Generation

If `llms.txt` doesn't exist at project root, aidocs generates a default:

```markdown
# {project.name}

> {project.description}

## For AI Agents
- Pattern Manifest: /docs/_ai/manifest.json
- Tags Index: /docs/_ai/tags.json
- Full Index: /docs/llms-full.txt

## Documentation
{links from SUMMARY.md}
```

---

## Implementation Plan

### Phase 1: Create `aidocs` Repository

1. **New repository structure:**
   ```
   aidocs/
   ├── cmd/aidocs/main.go         # CLI entry point
   ├── internal/
   │   ├── config/config.go       # Parse .aidocs.yaml
   │   ├── parser/summary.go      # Parse SUMMARY.md structure
   │   ├── parser/frontmatter.go  # Extract/generate frontmatter
   │   ├── generator/manifest.go  # Generate manifest.json
   │   ├── generator/tags.go      # Generate tags.json
   │   ├── generator/llmstxt.go   # Generate llms.txt, llms-full.txt
   │   ├── cache/cache.go         # SHA256 change detection
   │   └── ai/instructor.go       # Claude Code CLI integration
   ├── docs/                      # Self-documentation (dogfooding)
   │   ├── SUMMARY.md
   │   └── *.md
   ├── .aidocs.yaml               # Own config (dogfooding)
   ├── go.mod
   └── README.md
   ```

2. **CLI flags:**
   ```go
   var (
       configFile  = flag.String("config", ".aidocs.yaml", "Config file path")
       force       = flag.Bool("force", false, "Regenerate all, ignore cache")
       dryRun      = flag.Bool("dry-run", false, "Preview without writing")
       showOrphans = flag.Bool("show-orphans", false, "List files not in SUMMARY.md")
       init        = flag.Bool("init", false, "Create default .aidocs.yaml")
   )
   ```

### Phase 2: SUMMARY.md Parser

3. **Parse SUMMARY.md to extract structure:**
   ```go
   type SummaryEntry struct {
       Title    string         // Link text
       Path     string         // Relative path to .md file
       Category string         // Parent heading (e.g., "WIP")
       Children []SummaryEntry // Nested entries
   }

   func ParseSummary(path string) ([]SummaryEntry, error)
   ```

4. **Resolve paths relative to SUMMARY.md location:**
   ```go
   // If SUMMARY.md is at docs/SUMMARY.md
   // and entry is [Records](records.md)
   // resolved path is docs/records.md
   ```

### Phase 3: Frontmatter Management

5. **Frontmatter structure:**
   ```go
   type Frontmatter struct {
       Title           string   `yaml:"title"`
       Description     string   `yaml:"description"`
       Category        string   `yaml:"category"`
       Tags            []string `yaml:"tags"`
       EstimatedTokens int      `yaml:"estimatedTokens"`
   }
   ```

6. **Validation rules:**
   - Title in frontmatter MUST match H1 title in content
   - If mismatch, warn user or auto-fix based on config

7. **Auto-generation with Claude Code CLI:**
   ```go
   type GeneratedMeta struct {
       Description string   `json:"description"`
       Tags        []string `json:"tags"`
       Summary     string   `json:"summary"`
   }

   func GenerateMeta(cfg config.AIConfig, content, title string) (*GeneratedMeta, error) {
       prompt := fmt.Sprintf(`Analyze this documentation and provide metadata in JSON format.
   Title: %s
   Content: %s
   Respond with ONLY a JSON object with: description, tags, summary`, title, content)

       // Use Claude Code CLI - no API key needed
       cmd := exec.Command("claude", "-p", prompt)
       output, err := cmd.Output()
       if err != nil {
           return nil, err
       }
       return parseMetaJSON(string(output))
   }
   ```

### Phase 4: Change Detection (Excluding Frontmatter)

8. **Hash calculation:**
   ```go
   func ContentHash(filePath string) (string, error) {
       content, _ := os.ReadFile(filePath)

       // Remove frontmatter (between --- markers)
       body := stripFrontmatter(string(content))

       // Hash includes H1 title but not frontmatter
       hash := sha256.Sum256([]byte(body))
       return hex.EncodeToString(hash[:]), nil
   }
   ```

9. **Cache structure:**
   ```json
   {
     "version": "1",
     "generatedAt": "2024-01-15T10:30:00Z",
     "files": {
       "docs/records.md": {
         "contentHash": "a1b2c3d4...",
         "frontmatterHash": "e5f6g7h8...",
         "summary": "Cached AI summary...",
         "generatedAt": "2024-01-15T10:30:00Z"
       }
     }
   }
   ```

### Phase 5: Output Generation

10. **Generate `llms.txt`** (at project root, if missing):
    ```markdown
    # Blind Record System

    > Secure record management with blind encryption

    ## For AI Agents
    - Pattern Manifest: /docs/_ai/manifest.json
- Tags Index: /docs/_ai/tags.json
    - Full Index: /docs/llms-full.txt

    ## Documentation
    - [Blind Records](docs/records.md): Core record concepts
    - [Keys](docs/keys.md): Key hierarchy and management
    ```

11. **Generate `manifest.json`:**
    ```json
    {
      "knowledgeBase": {
        "name": "Blind Record System",
        "version": "v0.4.96",
        "generatedBy": "aidocs",
        "generatedAt": "2024-01-15T10:30:00Z",
        "optimizedFor": ["Claude Code", "BMAD BMM", "AI Agents"]
      },
      "patterns": [...],
      "categories": {...},
      "metadata": {
        "totalDocuments": 9,
        "averageTokensPerDoc": 600,
        "contentFile": "docs/SUMMARY.md"
      }
    }
    ```

12. **Generate `tags.json`:**
    ```json
    {
      "tags": {
        "encryption": {
          "count": 3,
          "documents": ["blind-records", "blind-keys", "example-keys"]
        },
        "example": {
          "count": 2,
          "documents": ["example-record", "example-keys"]
        }
      },
      "totalTags": 15,
      "topTags": ["encryption", "records", "keys", "example", "storage"]
    }
    ```
    - `tags`: Map of tag name to documents and count
    - `totalTags`: Total unique tags across all documents
    - `topTags`: Top 10 most common tags (sorted by count, then alphabetically)

### Phase 6: Integration Tests with Blind Project

12. **Test fixtures structure (source folder pattern):**
    ```
    aidocs/
    └── testdata/
        └── blind/
            ├── .aidocs.yaml          # Test config
            ├── source/               # Source files (copied fresh for each test)
            │   ├── .aidocs.yaml
            │   └── docs/
            │       ├── SUMMARY.md
            │       ├── records.md          # WITHOUT frontmatter
            │       ├── record_types.md     # WITHOUT frontmatter
            │       ├── keys.md             # WITHOUT frontmatter
            │       ├── record_keys.md      # WITHOUT frontmatter
            │       ├── dev_records.md      # WITHOUT frontmatter
            │       ├── dev_services.md     # WITHOUT frontmatter
            │       ├── record_encoding.md  # WITH frontmatter
            │       ├── record_example.md   # WITH frontmatter
            │       └── keys_example.md     # WITH frontmatter
            └── expected/             # Expected outputs for assertions
                ├── llms.txt
                ├── llms-full.txt
                └── manifest.json
    ```

    **Key design:** Source files have a MIX of files WITH and WITHOUT frontmatter to test both scenarios in every test run.

13. **Integration test structure:**
    ```go
    // internal/integration_test.go

    // copyDir copies source to temp directory for isolated testing
    func copyDir(src, dst string) error

    // runAidocs simulates running the aidocs tool
    func runAidocs(t *testing.T, configPath string) (string, error)

    // TestGreenfieldGeneration - tests fresh generation with mixed frontmatter
    func TestGreenfieldGeneration(t *testing.T) {
        // Copies source/ to temp dir
        // Runs aidocs
        // Verifies outputs exist and are valid
        // Checks: 3 docs WITH frontmatter, 6 WITHOUT
    }

    // TestBrownfieldGeneration - tests idempotent re-generation
    func TestBrownfieldGeneration(t *testing.T) {
        // Copies source/ to temp dir
        // Runs aidocs TWICE
        // Verifies outputs are identical (idempotent)
    }

    // TestCachePreservesState - tests cache tracks file changes
    func TestCachePreservesState(t *testing.T) {
        // Runs aidocs, modifies file, runs again
        // Verifies cache detects the change
    }

    // TestExpectedOutputComparison - compares with expected fixtures
    func TestExpectedOutputComparison(t *testing.T) {
        // Runs aidocs
        // Compares manifest.json with expected/manifest.json
    }

    // TestFrontmatterPreservedInOutput - verifies frontmatter data in manifest
    func TestFrontmatterPreservedInOutput(t *testing.T) {
        // Verifies files WITH frontmatter have metadata in manifest
        // (description, tags, category, estimatedTokens)
    }
    ```

14. **Test scenarios:**
    | Test | Description |
    |------|-------------|
    | `TestGreenfieldGeneration` | Fresh generation with mixed frontmatter files |
    | `TestBrownfieldGeneration` | Idempotent re-generation (run twice, same output) |
    | `TestCachePreservesState` | Cache detects file modifications |
    | `TestExpectedOutputComparison` | Compare with expected fixtures |
    | `TestFrontmatterPreservedInOutput` | Frontmatter metadata appears in manifest |
    | `TestChangeDetection` | SHA256 hash detects content changes |
    | `TestChangeDetectionExcludesFrontmatter` | Frontmatter changes don't trigger regeneration |
    | `TestOrphanDetection` | Files not in SUMMARY.md detected |
    | `TestFrontmatterExtraction` | Parse frontmatter correctly |
    | `TestSummaryParsing` | Parse SUMMARY.md structure |
    | `TestConfigDefaults` | Default config values |
    | `TestParseMetaJSON` | Parse JSON from various Claude response formats |
    | `TestExtractJSON` | Extract JSON from markdown, surrounding text |
    | `TestTruncateContent` | Content truncation for large documents |
    | `TestClaudeCLIIntegration` | Real Claude CLI call (skipped if not installed) |
    | `TestFrontmatterSpacingIdempotent` | Running --force twice produces identical output |
    | `TestFrontmatterSpacingWithExistingNewlines` | Leading whitespace variations handled correctly |

15. **Run tests:**
    ```bash
    go test ./... -v
    go test -run TestGreenfieldGeneration -v
    go test -run TestBrownfield -v
    ```

### Phase 7: Self-Documentation (Dogfooding)

14. **aidocs documents itself:**
    ```
    aidocs/
    ├── .aidocs.yaml              # Own config
    ├── docs/
    │   ├── SUMMARY.md            # Tool documentation structure
    │   ├── getting-started.md
    │   ├── configuration.md
    │   ├── ai-features.md
    │   └── examples.md
    ├── llms.txt                  # Generated
    └── docs/_ai/
        ├── manifest.json         # Generated
        ├── tags.json             # Generated
        ├── .cache.json           # Cache (auto git-ignored)
        └── .gitignore            # Auto-generated to ignore .cache.json
    ```

---

## Repository Structure (aidocs)

```
aidocs/                              # NEW REPOSITORY
├── cmd/aidocs/
│   └── main.go                      # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go                # Parse .aidocs.yaml
│   ├── parser/
│   │   ├── summary.go               # Parse SUMMARY.md structure
│   │   └── frontmatter.go           # Extract/validate/generate frontmatter
│   ├── generator/
│   │   ├── manifest.go              # Generate manifest.json
│   │   ├── tags.go                  # Generate tags.json
│   │   └── llmstxt.go               # Generate llms.txt, llms-full.txt
│   ├── cache/
│   │   └── cache.go                 # SHA256 change detection
│   ├── ai/
│   │   └── instructor.go            # Claude Code CLI integration
│   └── integration_test.go          # Integration tests
├── testdata/
│   └── blind/                       # Test fixtures
│       ├── .aidocs.yaml
│       ├── source/                  # Source files (copied fresh for each test)
│       │   ├── .aidocs.yaml
│       │   └── docs/
│       │       ├── SUMMARY.md
│       │       ├── *.md             # Mix: 6 WITHOUT frontmatter, 3 WITH
│       │       └── ...
│       └── expected/                # Expected outputs for assertions
│           ├── llms.txt
│           ├── llms-full.txt
│           ├── manifest.json
│           └── tags.json
├── docs/                            # Self-documentation (dogfooding)
│   ├── SUMMARY.md
│   ├── getting-started.md
│   ├── configuration.md
│   ├── ai-features.md
│   └── examples.md
├── .aidocs.yaml                     # Own config (dogfooding)
├── llms.txt                         # Generated by self
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

**Dependencies:**
```go
require (
    golang.org/x/text v0.x.x   // Title casing
    gopkg.in/yaml.v3 v3.x.x    // YAML parsing
)
// AI features use Claude Code CLI (claude -p) - no LLM library needed
```

---

## Test Fixtures Source (blind repo)

**Source Location:** `/home/diegos/_dev/goexogo/alibs/blind`

**Files to copy to `testdata/blind/`:**
| Source File | Purpose |
|-------------|---------|
| `docs/SUMMARY.md` | Content structure definition |
| `docs/records.md` | Sample doc - core concepts |
| `docs/record_types.md` | Sample doc - types |
| `docs/keys.md` | Sample doc - key hierarchy |
| `docs/record_keys.md` | Sample doc - record keys |
| `docs/record_encoding.md` | Sample doc - encoding |
| `docs/record_example.md` | Sample doc - examples |
| `docs/keys_example.md` | Sample doc - key examples |
| `docs/dev_records.md` | Sample doc - WIP category |
| `docs/dev_services.md` | Sample doc - WIP category |

**Create expected outputs manually for assertions:**
```
testdata/blind/expected/
├── llms.txt              # Expected root navigation
├── llms-full.txt         # Expected full index
└── manifest.json         # Expected manifest structure
```

**After aidocs is ready, integrate into blind repo:**
```
blind/
├── llms.txt                         # Generated
├── .aidocs.yaml                     # Config
└── docs/
    ├── llms-full.txt                # Generated
    ├── *.md                         # Updated with frontmatter
    └── _ai/
        ├── manifest.json            # Generated
        ├── tags.json                # Generated
        ├── .cache.json              # Cache (auto git-ignored)
        └── .gitignore               # Auto-generated
```

---

## Summary

| Aspect | Decision |
|--------|----------|
| Tool name | `aidocs` |
| Repository | Separate new repository |
| Content source | SUMMARY.md (single source of truth) |
| Config file | `.aidocs.yaml` |
| AI backend | Claude Code CLI (`claude -p`) - no API key needed |
| Summaries | AI-generated (optional, cached) |
| Change detection | SHA256 on content excluding frontmatter |
| Frontmatter | Auto-generated if missing (via Claude CLI) |
| Orphan detection | `--show-orphans` flag |
| Output format | manifest.json + tags.json + llms.txt + llms-full.txt |
| Self-documentation | Dogfooding (aidocs documents itself) |
| Testing | Integration tests with fixtures from blind repo |

---

## Usage in Target Projects

**Installation:**
```bash
go install github.com/diegosz/aidocs@latest
```

**Quick start:**
```bash
cd /path/to/project

# Create default config
aidocs --init

# Generate without AI (fast)
aidocs

# Generate with AI summaries
aidocs  # (when ai.enabled: true in config)

# Check for orphan files
aidocs --show-orphans

# Force regenerate all
aidocs --force
```

**Makefile integration (blind repo example):**
```makefile
.PHONY: genaidocs

# Generate LLM optimization files
genaidocs:
	@aidocs

# Include in main docs target
gendocs: gendocsbook genaidocs
```

---

## LLM Usage Flow

```
1. Claude Code discovers project
   └──→ Reads llms.txt (~500 tokens)
        "This is Blind Record System with manifest available"

2. Agent needs info about "encryption keys"
   └──→ Loads manifest.json (~2K tokens)
        └──→ Queries: patterns where tags include "keys"
             └──→ Finds: keys.md, record_keys.md, keys_example.md

3. Agent fetches only relevant docs
   └──→ Reads keys.md (~800 tokens)
   └──→ Reads record_keys.md (~600 tokens)
        Total: ~3.9K tokens instead of ~15K for all docs

4. Agent answers with precise context ✓
```

---

## Implementation Status

### Completed ✓

- [x] CLI with all planned flags (`--init`, `--force`, `--dry-run`, `--show-orphans`, `-v`)
- [x] Config parsing (`.aidocs.yaml`)
- [x] SUMMARY.md parser with category detection
- [x] Frontmatter extraction (reads existing, extracts H1 as fallback)
- [x] Frontmatter writing with idempotent spacing (exactly 1 empty line before H1)
- [x] manifest.json generation
- [x] tags.json generation (aggregated tags with counts and top tags)
- [x] llms.txt generation with Tags Index reference and Usage section (jq examples)
- [x] JSON output uses "section" instead of "category" (reflects SUMMARY.md structure)
- [x] llms-full.txt generation
- [x] SHA256 cache for change detection (excludes frontmatter)
- [x] Auto-generated `.gitignore` in output directory (ignores `.cache.json`)
- [x] Orphan file detection
- [x] AI integration via Claude Code CLI (`claude -p`) - no API key needed
- [x] AI unit tests (JSON parsing, extraction, truncation)
- [x] AI integration test (real Claude CLI call, skipped if not installed)
- [x] Frontmatter spacing tests (idempotent, various leading whitespace)
- [x] Integration tests passing
- [x] Test fixtures from blind repo
- [x] Self-documentation (dogfooding): 6 docs in guide/reference categories
- [x] Default output folder: `docs/_ai/` (shorter path)

### Next Steps

1. **Create aidocs repository** with structure above ✓
2. **Copy test fixtures** from blind repo to `testdata/blind/` ✓
3. **Implement core features** (SUMMARY.md parser, frontmatter, manifest generation) ✓
4. **Write integration tests** using blind fixtures ✓
5. **Add AI features** via Claude Code CLI ✓
6. **Document aidocs using aidocs** (dogfooding) ✓
7. **Publish aidocs** and integrate into blind repo via Makefile - TODO
