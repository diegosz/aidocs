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

// maxDescriptionLength is the maximum length for truncated descriptions.
const maxDescriptionLength = 100

// WriteLLMsFull generates the complete document index file.
func WriteLLMsFull(path string, docs []*Document, project config.ProjectConfig, dryRun bool) error {
	if dryRun {
		return nil
	}

	lines := []string{
		"# " + project.Name + " - Complete Document Index",
		"",
		fmt.Sprintf("Total Documents: %d", len(docs)),
		"",
	}

	// Group by category
	byCategory := make(map[string][]*Document)
	for _, doc := range docs {
		cat := doc.Frontmatter.Section
		if cat == "" {
			cat = doc.Section
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
		lines = append(lines,
			"## "+titleCaser.String(cat),
			"",
		)

		for _, doc := range catDocs {
			lines = append(lines,
				"### "+doc.Frontmatter.Title,
				"Path: "+doc.Path,
			)

			if doc.Frontmatter.Description != "" {
				lines = append(lines, "Description: "+doc.Frontmatter.Description)
			}

			if len(doc.Frontmatter.Tags) > 0 {
				lines = append(lines, "Tags: "+strings.Join(doc.Frontmatter.Tags, ", "))
			}

			if doc.Frontmatter.EstimatedTokens > 0 {
				lines = append(lines, fmt.Sprintf("Tokens: %d", doc.Frontmatter.EstimatedTokens))
			}

			if doc.Frontmatter.Summary != "" {
				lines = append(lines, "Summary: "+doc.Frontmatter.Summary)
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
	return os.WriteFile(path, []byte(content), 0o600)
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

	lines := []string{
		"# " + cfg.Project.Name,
		"",
	}

	if cfg.Project.Description != "" {
		lines = append(lines,
			"> "+cfg.Project.Description,
			"",
		)
	}

	lines = append(lines,
		"## For AI Agents",
		"",
	)

	// Add paths relative to llms.txt location
	llmsDir := filepath.Dir(path)

	manifestRel, _ := filepath.Rel(llmsDir, cfg.Output.Manifest)
	tagsRel, _ := filepath.Rel(llmsDir, cfg.Output.Tags)
	llmsFullRel, _ := filepath.Rel(llmsDir, cfg.Output.LLMsFull)

	lines = append(lines,
		"- Documents Manifest: /"+manifestRel,
		"- Tags Index: /"+tagsRel,
		"- Full Index: /"+llmsFullRel,
		"",
		"## Usage",
		"",
		"```bash",
		"# View manifest structure",
		"cat "+manifestRel+" | jq '.knowledgeBase'",
		"",
		"# List all documents",
		"cat "+manifestRel+" | jq '.documents[] | {id, name, section}'",
		"",
		"# Search by tag",
		"cat "+manifestRel+" | jq '.documents[] | select(.tags[]? | contains(\"encryption\"))'",
		"",
		"# Filter by section",
		"cat "+manifestRel+" | jq '.documents[] | select(.section == \"WIP\")'",
		"",
		"# List all tags with counts",
		"cat "+tagsRel+" | jq '.tags | to_entries | sort_by(-.value.count) | .[:10]'",
		"",
		"# Find documents by tag",
		"cat "+tagsRel+" | jq '.tags[\"encryption\"].documents'",
		"```",
		"",
		"## Documentation",
		"",
	)

	// Add document links
	for _, doc := range docs {
		relPath, _ := filepath.Rel(llmsDir, doc.Path)
		desc := doc.Frontmatter.Description
		if desc == "" && doc.Frontmatter.Summary != "" {
			// Truncate summary for description
			desc = doc.Frontmatter.Summary
			if len(desc) > maxDescriptionLength {
				desc = desc[:maxDescriptionLength-3] + "..."
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
		lines = append(lines,
			"---",
			"",
			"Optimized for: "+strings.Join(cfg.Project.OptimizedFor, ", "),
		)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(content), 0o600)
}
