// Package parser provides parsing utilities for SUMMARY.md and markdown frontmatter.
package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SummaryEntry represents an entry in SUMMARY.md.
type SummaryEntry struct {
	Title    string         // Link text
	Path     string         // Relative path to .md file
	Category string         // Parent heading (e.g., "WIP")
	Children []SummaryEntry // Nested entries
}

// linkPattern matches markdown links: [Title](path.md)
var linkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+\.md)\)`)

// ParseSummary parses SUMMARY.md and extracts the document structure.
func ParseSummary(path string) ([]SummaryEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []SummaryEntry
	var currentCategory string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for category header (line that's not a link but contains text)
		// Categories are typically indented list items without links
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and the title
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check if this is a category (list item without link)
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			content := strings.TrimPrefix(strings.TrimPrefix(trimmed, "-"), "*")
			content = strings.TrimSpace(content)

			// If it doesn't contain a link, it's a category
			if !strings.Contains(content, "[") {
				currentCategory = content
				continue
			}
		}

		// Parse markdown link
		matches := linkPattern.FindStringSubmatch(line)
		if len(matches) == 3 {
			// Determine nesting level by leading whitespace/indentation
			indent := len(line) - len(strings.TrimLeft(line, " \t"))

			entry := SummaryEntry{
				Title: matches[1],
				Path:  matches[2],
			}

			// Apply category to indented entries (2+ spaces means nested)
			if indent >= 2 && currentCategory != "" {
				entry.Category = currentCategory
			} else if indent == 0 {
				// Non-indented entry resets the category
				currentCategory = ""
			}

			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// FindOrphans finds markdown files in the docs directory that are not referenced in SUMMARY.md.
func FindOrphans(summaryPath string, entries []SummaryEntry) ([]string, error) {
	docsDir := filepath.Dir(summaryPath)

	// Build a set of referenced files
	referenced := make(map[string]bool)
	for _, entry := range entries {
		absPath := filepath.Join(docsDir, entry.Path)
		referenced[absPath] = true
	}

	// Also exclude SUMMARY.md itself and common generated files
	referenced[summaryPath] = true

	var orphans []string

	// Walk the docs directory
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only check .md files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip common generated or special files
		base := filepath.Base(path)
		if base == "README.md" || base == "HANDBOOK.md" || base == "DOC.md" || base == "CHANGELOG.md" {
			return nil
		}

		// Check if file is referenced
		if !referenced[path] {
			relPath, _ := filepath.Rel(docsDir, path)
			orphans = append(orphans, relPath)
		}

		return nil
	})

	return orphans, err
}
