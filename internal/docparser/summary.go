// Package docparser provides parsing utilities for SUMMARY.md and markdown frontmatter.
package docparser

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
	Section  string         // Parent heading (e.g., "WIP")
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
	var currentSection string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for section header (line that's not a link but contains text)
		// Sections are typically indented list items without links
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and the title
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check if this is a section (list item without link)
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			content := strings.TrimPrefix(strings.TrimPrefix(trimmed, "-"), "*")
			content = strings.TrimSpace(content)

			// If it doesn't contain a link, it's a section
			if !strings.Contains(content, "[") {
				currentSection = content
				continue
			}
		}

		// Parse markdown link
		// linkPattern captures: [0]=full match, [1]=title, [2]=path
		matches := linkPattern.FindStringSubmatch(line)
		if len(matches) > 2 { //nolint:mnd // regex capture groups
			// Determine nesting level by leading whitespace/indentation
			indent := len(line) - len(strings.TrimLeft(line, " \t"))

			entry := SummaryEntry{
				Title: matches[1],
				Path:  matches[2],
			}

			// Apply section to indented entries (2+ spaces means nested)
			if indent >= 2 && currentSection != "" {
				entry.Section = currentSection
			} else if indent == 0 {
				// Non-indented entry resets the section
				currentSection = ""
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
