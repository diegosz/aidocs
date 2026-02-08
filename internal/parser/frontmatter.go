package parser

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents the metadata at the top of a markdown file.
type Frontmatter struct {
	Title           string   `yaml:"title"`
	Description     string   `yaml:"description"`
	Category        string   `yaml:"category"`
	Tags            []string `yaml:"tags"`
	EstimatedTokens int      `yaml:"estimatedTokens,omitempty"`
	Summary         string   `yaml:"summary,omitempty"` // AI-generated summary
}

// h1Pattern matches H1 headers: # Title
var h1Pattern = regexp.MustCompile(`^#\s+(.+)$`)

// ExtractFrontmatter reads a markdown file and extracts its frontmatter and body.
// If no frontmatter exists, it extracts the H1 title as the title.
func ExtractFrontmatter(path string) (*Frontmatter, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	fm := &Frontmatter{}
	body := string(content)

	// Check for YAML frontmatter (between --- markers)
	if bytes.HasPrefix(content, []byte("---\n")) || bytes.HasPrefix(content, []byte("---\r\n")) {
		parts := bytes.SplitN(content[4:], []byte("\n---"), 2)
		if len(parts) == 2 {
			// Parse frontmatter
			if err := yaml.Unmarshal(parts[0], fm); err != nil {
				return nil, "", err
			}
			// Body is everything after the closing ---
			body = strings.TrimPrefix(string(parts[1]), "\n")
			body = strings.TrimPrefix(body, "\r\n")
		}
	}

	// If no title in frontmatter, extract from H1
	if fm.Title == "" {
		fm.Title = extractH1Title(body)
	}

	// Validate title matches H1 if both exist
	h1Title := extractH1Title(body)
	if fm.Title != "" && h1Title != "" && fm.Title != h1Title {
		// Use H1 as authoritative
		fm.Title = h1Title
	}

	return fm, body, nil
}

// extractH1Title finds the first H1 header in the content.
func extractH1Title(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if matches := h1Pattern.FindStringSubmatch(line); len(matches) == 2 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

// WriteFrontmatter writes/updates frontmatter in a markdown file.
func WriteFrontmatter(path string, fm *Frontmatter, body string, dryRun bool) error {
	if dryRun {
		return nil
	}

	// Generate YAML frontmatter
	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return err
	}

	// Combine frontmatter and body
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(body)

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// StripFrontmatter removes frontmatter from content, returning just the body.
// This is used for hash calculation.
func StripFrontmatter(content []byte) []byte {
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return content
	}

	// Find the closing ---
	rest := content[4:]
	idx := bytes.Index(rest, []byte("\n---"))
	if idx == -1 {
		return content
	}

	// Return everything after the closing ---
	body := rest[idx+4:]
	body = bytes.TrimPrefix(body, []byte("\n"))
	body = bytes.TrimPrefix(body, []byte("\r\n"))
	return body
}
