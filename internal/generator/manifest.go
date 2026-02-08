// Package generator provides functions to generate manifest.json and llms.txt files.
package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/diegosz/aidocs/internal/config"
	"github.com/diegosz/aidocs/internal/parser"
)

// Document represents a processed documentation file.
type Document struct {
	Path        string
	Frontmatter *parser.Frontmatter
	Category    string
}

// Manifest represents the complete manifest.json structure.
type Manifest struct {
	KnowledgeBase map[string]any   `json:"knowledgeBase"`
	Patterns      []map[string]any `json:"patterns"`
	Categories    map[string]any   `json:"categories"`
	Metadata      map[string]any   `json:"metadata"`
}

// GenerateManifest creates a manifest from processed documents.
func GenerateManifest(docs []*Document, project config.ProjectConfig) *Manifest {
	m := &Manifest{
		KnowledgeBase: map[string]any{
			"name":         project.Name,
			"version":      project.Version,
			"description":  project.Description,
			"generatedBy":  "aidocs",
			"generatedAt":  time.Now().UTC().Format(time.RFC3339),
			"optimizedFor": project.OptimizedFor,
		},
		Patterns:   make([]map[string]any, 0, len(docs)),
		Categories: make(map[string]any),
		Metadata: map[string]any{
			"totalDocuments":      len(docs),
			"averageTokensPerDoc": 0,
		},
	}

	// Group patterns by category
	categoryDocs := make(map[string][]string)
	totalTokens := 0

	for _, doc := range docs {
		pattern := map[string]any{
			"id":   titleToID(doc.Frontmatter.Title),
			"name": doc.Frontmatter.Title,
			"paths": map[string]string{
				"doc": doc.Path,
			},
		}

		if doc.Frontmatter.Description != "" {
			pattern["description"] = doc.Frontmatter.Description
		}

		if doc.Frontmatter.Category != "" {
			pattern["category"] = doc.Frontmatter.Category
		} else if doc.Category != "" {
			pattern["category"] = doc.Category
		}

		if len(doc.Frontmatter.Tags) > 0 {
			pattern["tags"] = doc.Frontmatter.Tags
		}

		if doc.Frontmatter.EstimatedTokens > 0 {
			pattern["estimatedTokens"] = doc.Frontmatter.EstimatedTokens
			totalTokens += doc.Frontmatter.EstimatedTokens
		}

		if doc.Frontmatter.Summary != "" {
			pattern["summary"] = doc.Frontmatter.Summary
		}

		m.Patterns = append(m.Patterns, pattern)

		// Track categories
		cat := doc.Frontmatter.Category
		if cat == "" {
			cat = doc.Category
		}
		if cat == "" {
			cat = "general"
		}
		categoryDocs[cat] = append(categoryDocs[cat], titleToID(doc.Frontmatter.Title))
	}

	// Build categories map
	for cat, ids := range categoryDocs {
		sort.Strings(ids)
		m.Categories[cat] = map[string]any{
			"documents": ids,
			"count":     len(ids),
		}
	}

	// Calculate average tokens
	if len(docs) > 0 && totalTokens > 0 {
		m.Metadata["averageTokensPerDoc"] = totalTokens / len(docs)
	}

	return m
}

// WriteManifest writes the manifest to disk.
func WriteManifest(path string, m *Manifest, dryRun bool) error {
	if dryRun {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// titleToID converts a title to a URL-friendly ID.
func titleToID(title string) string {
	id := strings.ToLower(title)
	id = strings.ReplaceAll(id, " ", "-")

	// Remove special characters
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	// Clean up multiple dashes
	id = result.String()
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}
	id = strings.Trim(id, "-")

	return id
}
