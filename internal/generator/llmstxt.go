package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/diegosz/aidocs/internal/config"
)

// WriteLLMsFull generates the complete document index file.
func WriteLLMsFull(path string, docs []*Document, project config.ProjectConfig, dryRun bool) error {
	if dryRun {
		return nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("# %s - Complete Document Index", project.Name))
	lines = append(lines, "")
	if project.Version != "" {
		lines = append(lines, fmt.Sprintf("Version: %s", project.Version))
	}
	lines = append(lines, fmt.Sprintf("Total Documents: %d", len(docs)))
	lines = append(lines, "")

	// Group by category
	byCategory := make(map[string][]*Document)
	for _, doc := range docs {
		cat := doc.Frontmatter.Category
		if cat == "" {
			cat = doc.Category
		}
		if cat == "" {
			cat = "General"
		}
		byCategory[cat] = append(byCategory[cat], doc)
	}

	// Sort categories
	categories := make([]string, 0, len(byCategory))
	for cat := range byCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		catDocs := byCategory[cat]

		// Sort documents within category
		sort.Slice(catDocs, func(i, j int) bool {
			return catDocs[i].Path < catDocs[j].Path
		})

		titleCaser := cases.Title(language.English)
		lines = append(lines, fmt.Sprintf("## %s", titleCaser.String(cat)))
		lines = append(lines, "")

		for _, doc := range catDocs {
			lines = append(lines, fmt.Sprintf("### %s", doc.Frontmatter.Title))
			lines = append(lines, fmt.Sprintf("Path: %s", doc.Path))

			if doc.Frontmatter.Description != "" {
				lines = append(lines, fmt.Sprintf("Description: %s", doc.Frontmatter.Description))
			}

			if len(doc.Frontmatter.Tags) > 0 {
				lines = append(lines, fmt.Sprintf("Tags: %s", strings.Join(doc.Frontmatter.Tags, ", ")))
			}

			if doc.Frontmatter.EstimatedTokens > 0 {
				lines = append(lines, fmt.Sprintf("Tokens: %d", doc.Frontmatter.EstimatedTokens))
			}

			if doc.Frontmatter.Summary != "" {
				lines = append(lines, fmt.Sprintf("Summary: %s", doc.Frontmatter.Summary))
			}

			lines = append(lines, "")
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(content), 0o644)
}

// WriteLLMsTxt generates or updates the root llms.txt navigation file.
// If the file already exists, it is not overwritten unless force is true.
func WriteLLMsTxt(path string, docs []*Document, cfg *config.Config, dryRun bool) error {
	if dryRun {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		// File exists, don't overwrite
		return nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("# %s", cfg.Project.Name))
	lines = append(lines, "")

	if cfg.Project.Description != "" {
		lines = append(lines, fmt.Sprintf("> %s", cfg.Project.Description))
		lines = append(lines, "")
	}

	lines = append(lines, "## For AI Agents")
	lines = append(lines, "")

	// Add paths relative to llms.txt location
	llmsDir := filepath.Dir(path)

	manifestRel, _ := filepath.Rel(llmsDir, cfg.Output.Manifest)
	llmsFullRel, _ := filepath.Rel(llmsDir, cfg.Output.LLMsFull)

	lines = append(lines, fmt.Sprintf("- Pattern Manifest: /%s", manifestRel))
	lines = append(lines, fmt.Sprintf("- Full Index: /%s", llmsFullRel))
	lines = append(lines, "")

	lines = append(lines, "## Documentation")
	lines = append(lines, "")

	// Add document links
	for _, doc := range docs {
		relPath, _ := filepath.Rel(llmsDir, doc.Path)
		desc := doc.Frontmatter.Description
		if desc == "" && doc.Frontmatter.Summary != "" {
			// Truncate summary for description
			desc = doc.Frontmatter.Summary
			if len(desc) > 100 {
				desc = desc[:97] + "..."
			}
		}

		if desc != "" {
			lines = append(lines, fmt.Sprintf("- [%s](%s): %s", doc.Frontmatter.Title, relPath, desc))
		} else {
			lines = append(lines, fmt.Sprintf("- [%s](%s)", doc.Frontmatter.Title, relPath))
		}
	}

	lines = append(lines, "")

	// Add optimized for section
	if len(cfg.Project.OptimizedFor) > 0 {
		lines = append(lines, "---")
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Optimized for: %s", strings.Join(cfg.Project.OptimizedFor, ", ")))
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(content), 0o644)
}
