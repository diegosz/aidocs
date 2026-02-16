// Package ai provides AI-powered content generation using Claude Code CLI.
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// maxContentLength is the maximum content length to send to the LLM.
const maxContentLength = 4000

// GeneratedMeta contains AI-generated metadata for a document.
type GeneratedMeta struct {
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Summary     string   `json:"summary"`
}

// GenerateMeta generates metadata for document content using Claude Code CLI.
func GenerateMeta(content, title string) (*GeneratedMeta, error) {
	prompt := fmt.Sprintf(`Analyze this documentation and provide metadata in JSON format.

Title: %s

Content:
%s

Respond with ONLY a JSON object (no markdown, no explanation, no code fences) with these fields:
- "description": one-line description (max 100 chars)
- "tags": array of 3-5 relevant topic tags (lowercase, hyphenated)
- "summary": 2-3 sentence summary of key concepts`, title, truncateContent(content, maxContentLength))

	// Use Claude Code CLI
	return callClaudeCLI(context.Background(), prompt)
}

// callClaudeRaw calls the claude CLI and returns the raw output string.
func callClaudeRaw(ctx context.Context, prompt string) (string, error) {
	// Check if claude CLI is available
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return "", errors.New("claude CLI not found in PATH - install Claude Code or use 'ai.enabled: false'")
	}

	// Run claude with --print flag for non-interactive mode
	cmd := exec.CommandContext(ctx, claudePath, "-p", prompt)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("claude CLI error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("claude CLI failed: %w", err)
	}

	return string(output), nil
}

// callClaudeCLI uses the claude command-line tool to generate responses.
func callClaudeCLI(ctx context.Context, prompt string) (*GeneratedMeta, error) {
	output, err := callClaudeRaw(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return parseMetaJSON(output)
}

func parseMetaJSON(text string) (*GeneratedMeta, error) {
	// Try to find JSON in the response
	text = strings.TrimSpace(text)
	text = extractJSON(text)

	var meta GeneratedMeta
	if err := json.Unmarshal([]byte(text), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w\nResponse was: %s", err, text)
	}

	return &meta, nil
}

// extractJSON attempts to find a JSON object in the text.
func extractJSON(text string) string {
	// Remove markdown code fences if present
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	// Find first { and last }
	start := -1
	end := -1
	depth := 0

	for i, c := range text {
		if c == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 && start != -1 {
				end = i + 1
				break
			}
		}
	}

	if start != -1 && end != -1 {
		return text[start:end]
	}

	return text
}

// GeneratedProjectInfo contains AI-inferred project metadata.
type GeneratedProjectInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InferProjectInfo uses the Claude CLI to infer project name and description
// from SUMMARY.md content.
func InferProjectInfo(summaryContent string) (*GeneratedProjectInfo, error) {
	prompt := fmt.Sprintf(`Analyze this documentation table of contents (SUMMARY.md) and infer the project name and a short description.

Content:
%s

Respond with ONLY a JSON object (no markdown, no explanation, no code fences) with these fields:
- "name": the project name (concise, typically 1-3 words)
- "description": one-line project description (max 120 chars)`, truncateContent(summaryContent, maxContentLength))

	output, err := callClaudeRaw(context.Background(), prompt)
	if err != nil {
		return nil, err
	}

	return parseProjectInfoJSON(output)
}

func parseProjectInfoJSON(text string) (*GeneratedProjectInfo, error) {
	text = strings.TrimSpace(text)
	text = extractJSON(text)

	var info GeneratedProjectInfo
	if err := json.Unmarshal([]byte(text), &info); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w\nResponse was: %s", err, text)
	}

	return &info, nil
}

// truncateContent limits content size to avoid overwhelming the LLM.
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "\n\n[Content truncated...]"
}
