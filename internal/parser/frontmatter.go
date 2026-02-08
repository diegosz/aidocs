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
	Section         string   `yaml:"section"`
	Tags            []string `yaml:"tags"`
	EstimatedTokens int      `yaml:"estimatedTokens,omitempty"`
	Summary         string   `yaml:"summary,omitempty"` // AI-generated summary
}

// h1Pattern matches H1 headers: # Title
var h1Pattern = regexp.MustCompile(`^#\s+(.+)$`)

// frontmatterDelimiter is the YAML frontmatter delimiter.
var frontmatterDelimiter = []byte("\n---")

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
		// Split into frontmatter and body at the closing ---
		rest := content[4:] // Skip opening "---\n"
		fmContent, bodyContent, found := bytes.Cut(rest, frontmatterDelimiter)
		if found {
			// Parse frontmatter
			if err := yaml.Unmarshal(fmContent, fm); err != nil {
				return nil, "", err
			}
			// Body is everything after the closing ---, trimmed of leading whitespace
			body = strings.TrimLeft(string(bodyContent), "\n\r\t ")
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
		// h1Pattern captures: [0]=full match, [1]=title
		if matches := h1Pattern.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

// WriteFrontmatter writes/updates frontmatter in a markdown file.
// Ensures exactly one empty line between frontmatter and H1 title.
func WriteFrontmatter(path string, fm *Frontmatter, body string, dryRun bool) error {
	if dryRun {
		return nil
	}

	// Generate YAML frontmatter
	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		return err
	}

	// Trim leading whitespace from body to ensure consistent spacing
	body = strings.TrimLeft(body, "\n\r\t ")

	// Combine frontmatter and body with exactly one empty line
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n\n")
	buf.WriteString(body)

	return os.WriteFile(path, buf.Bytes(), 0o600)
}

// StripFrontmatter removes frontmatter from content, returning just the body.
// This is used for hash calculation.
func StripFrontmatter(content []byte) []byte {
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return content
	}

	// Find the closing ---
	rest := content[4:]
	_, body, found := bytes.Cut(rest, frontmatterDelimiter)
	if !found {
		return content
	}

	// Return everything after the closing ---, trimmed of leading whitespace
	body = bytes.TrimLeft(body, "\n\r\t ")
	return body
}
