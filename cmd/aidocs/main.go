// Package main provides the aidocs CLI tool for generating LLM-friendly documentation.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/diegosz/aidocs/internal/ai"
	"github.com/diegosz/aidocs/internal/cache"
	"github.com/diegosz/aidocs/internal/config"
	"github.com/diegosz/aidocs/internal/generator"
	"github.com/diegosz/aidocs/internal/docparser"
)

var (
	configFile  = flag.String("config", ".aidocs.yaml", "Config file path")
	force       = flag.Bool("force", false, "Regenerate all, ignore cache")
	dryRun      = flag.Bool("dry-run", false, "Preview without writing")
	showOrphans = flag.Bool("show-orphans", false, "List files not in SUMMARY.md")
	initConfig  = flag.Bool("init", false, "Create default .aidocs.yaml")
	verbose     = flag.Bool("v", false, "Verbose output")
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	if *initConfig {
		if err := createDefaultConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created .aidocs.yaml")
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if *verbose {
		fmt.Printf("Using config: %s\n", *configFile)
		fmt.Printf("Content source: %s\n", cfg.Content)
	}

	// Parse SUMMARY.md - if missing, skip AI docs generation gracefully
	entries, err := docparser.ParseSummary(cfg.Content)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No %s found - skipping AI docs optimization\n", cfg.Content)
			return nil
		}
		return fmt.Errorf("parse summary: %w", err)
	}

	if *verbose {
		fmt.Printf("Found %d entries in SUMMARY.md\n", len(entries))
	}

	// Infer project metadata from SUMMARY.md if needed
	if cfg.AI.Enabled {
		needsName := strings.TrimSpace(cfg.Project.Name) == "" || cfg.Project.Name == "Project"
		needsDesc := strings.TrimSpace(cfg.Project.Description) == ""
		if needsName || needsDesc {
			summaryContent, readErr := os.ReadFile(cfg.Content)
			if readErr == nil {
				info, inferErr := ai.InferProjectInfo(string(summaryContent))
				if inferErr != nil {
					if *verbose {
						fmt.Printf("Warning: AI project inference failed: %v\n", inferErr)
					}
				} else {
					if needsName && strings.TrimSpace(info.Name) != "" {
						cfg.Project.Name = info.Name
					}
					if needsDesc && strings.TrimSpace(info.Description) != "" {
						cfg.Project.Description = info.Description
					}
				}
			} else if *verbose {
				fmt.Printf("Warning: could not read SUMMARY.md for project inference: %v\n", readErr)
			}
		}
	}

	// Validate project name after AI inference attempt
	if err := cfg.ValidateProjectName(); err != nil {
		return err
	}

	// Check for orphan files if requested
	if *showOrphans {
		orphans, err := docparser.FindOrphans(cfg.Content, entries)
		if err != nil {
			return fmt.Errorf("find orphans: %w", err)
		}
		if len(orphans) > 0 {
			fmt.Println("Orphan files (not in SUMMARY.md):")
			for _, o := range orphans {
				fmt.Printf("  - %s\n", o)
			}
		} else {
			fmt.Println("No orphan files found")
		}
		return nil
	}

	// Load cache
	var docCache *cache.Cache
	if !*force {
		docCache, err = cache.Load(cfg.Output.Cache)
		if err != nil && *verbose {
			fmt.Printf("Warning: could not load cache: %v\n", err)
		}
	}
	if docCache == nil {
		docCache = cache.New()
	}

	// Process each document
	docs := make([]*generator.Document, 0, len(entries))
	summaryDir := filepath.Dir(cfg.Content)

	for _, entry := range entries {
		docPath := filepath.Join(summaryDir, entry.Path)

		// Check if file needs processing
		needsProcess := *force
		if !needsProcess {
			changed, err := docCache.HasChanged(docPath)
			if err != nil {
				if *verbose {
					fmt.Printf("Warning: cache check failed for %s: %v\n", docPath, err)
				}
				needsProcess = true
			} else {
				needsProcess = changed
			}
		}

		// Extract or generate frontmatter
		fm, body, err := docparser.ExtractFrontmatter(docPath)
		if err != nil {
			fmt.Printf("Warning: skipping %s: %v\n", docPath, err)
			continue
		}

		// Use entry title if frontmatter title is missing
		if fm.Title == "" {
			fm.Title = entry.Title
		}

		// Use SUMMARY.md section if frontmatter section is missing
		if fm.Section == "" && entry.Section != "" {
			fm.Section = entry.Section
		}

		// Generate AI metadata if enabled and needed
		if cfg.AI.Enabled && needsProcess {
			// With --force: regenerate all enabled fields
			// Without --force: only generate missing fields
			needsDescription := cfg.AI.GenerateDescriptions && (*force || fm.Description == "")
			needsTags := cfg.AI.GenerateTags && (*force || len(fm.Tags) == 0)
			needsSummary := cfg.AI.GenerateSummaries && (*force || fm.Summary == "")

			if needsDescription || needsTags || needsSummary {
				meta, err := ai.GenerateMeta(body, fm.Title)
				if err != nil {
					if *verbose {
						fmt.Printf("Warning: AI generation failed for %s: %v\n", docPath, err)
					}
				} else {
					if needsDescription {
						fm.Description = meta.Description
					}
					if needsTags {
						fm.Tags = meta.Tags
					}
					if needsSummary {
						fm.Summary = meta.Summary
					}
				}
			}

			// Update frontmatter in file if configured
			if cfg.AI.GenerateMissingFrontmatter {
				if err := docparser.WriteFrontmatter(docPath, fm, body, *dryRun); err != nil {
					fmt.Printf("Warning: could not update frontmatter in %s: %v\n", docPath, err)
				} else if *verbose && !*dryRun {
					fmt.Printf("Updated frontmatter: %s\n", docPath)
				}
			}
		}

		// Update cache
		if err := docCache.Update(docPath, fm.Summary); err != nil && *verbose {
			fmt.Printf("Warning: cache update failed for %s: %v\n", docPath, err)
		}

		docs = append(docs, &generator.Document{
			Path:        docPath,
			Frontmatter: fm,
			Section:     entry.Section,
		})
	}

	if *dryRun {
		fmt.Println("Dry-run mode: no files will be written")
	}

	// Generate manifest.json
	manifest := generator.GenerateManifest(docs, cfg.Project)
	if err := generator.WriteManifest(cfg.Output.Manifest, manifest, *dryRun); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	fmt.Printf("Generated manifest.json (%d documents)\n", len(docs))

	// Generate llms-full.txt
	if err := generator.WriteLLMsFull(cfg.Output.LLMsFull, docs, cfg.Project, *dryRun); err != nil {
		return fmt.Errorf("write llms-full.txt: %w", err)
	}
	fmt.Println("Generated llms-full.txt")

	// Generate or update llms.txt at project root if missing
	if err := generator.WriteLLMsTxt(cfg.Output.LLMsTxt, docs, cfg, *dryRun); err != nil {
		return fmt.Errorf("write llms.txt: %w", err)
	}
	fmt.Println("Generated llms.txt")

	// Generate tags.json
	tags := generator.GenerateTags(docs)
	if err := generator.WriteTags(cfg.Output.Tags, tags, *dryRun); err != nil {
		return fmt.Errorf("write tags.json: %w", err)
	}
	fmt.Printf("Generated tags.json (%d unique tags)\n", tags.TotalTags)

	// Save cache
	if !*dryRun {
		if err := docCache.Save(cfg.Output.Cache); err != nil {
			fmt.Printf("Warning: could not save cache: %v\n", err)
		}
	}

	// Create .gitignore in output directory to ignore cache file
	if !*dryRun {
		if err := createOutputGitignore(cfg.Output.Cache); err != nil && *verbose {
			fmt.Printf("Warning: could not create .gitignore: %v\n", err)
		}
	}

	fmt.Println("Done!")
	return nil
}

// createOutputGitignore creates a .gitignore file in the output directory
// to ignore the cache file.
func createOutputGitignore(cachePath string) error {
	dir := filepath.Dir(cachePath)
	gitignorePath := filepath.Join(dir, ".gitignore")

	// Only create if it doesn't exist
	if _, err := os.Stat(gitignorePath); err == nil {
		return nil
	}

	cacheFile := filepath.Base(cachePath)
	content := fmt.Sprintf("# aidocs cache (auto-generated)\n%s\n", cacheFile)

	return os.WriteFile(gitignorePath, []byte(content), 0o600)
}

func createDefaultConfig() error {
	defaultConfig := `# .aidocs.yaml - aidocs configuration
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
  enabled: false                          # Enable AI-powered features
  generate_summaries: true                # Generate AI summaries for each doc
  generate_missing_frontmatter: true      # Add frontmatter to files missing it
  generate_descriptions: true             # Generate missing descriptions
  generate_tags: true                     # Generate missing tags

# Project metadata
project:
  name: "My Project"
  description: "Project description"
  optimized_for:
    - "Claude Code"
    - "AI Agents"
`
	return os.WriteFile(".aidocs.yaml", []byte(defaultConfig), 0o600)
}

func printVersion() {
	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
	fmt.Printf("aidocs %s\n", version)
}
