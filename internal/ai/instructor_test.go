// Package ai provides AI-powered content generation using Claude Code CLI.
package ai

import (
	"os/exec"
	"testing"

	"github.com/diegosz/aidocs/internal/config"
)

func TestParseMetaJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDesc    string
		wantTags    []string
		wantSummary string
		wantErr     bool
	}{
		{
			name: "clean JSON",
			input: `{
				"description": "Core record concepts",
				"tags": ["records", "encryption", "storage"],
				"summary": "This document covers blind records."
			}`,
			wantDesc:    "Core record concepts",
			wantTags:    []string{"records", "encryption", "storage"},
			wantSummary: "This document covers blind records.",
		},
		{
			name: "JSON with markdown code fence",
			input: "```json\n{\"description\": \"Test desc\", \"tags\": [\"tag1\"], \"summary\": \"Test summary\"}\n```",
			wantDesc:    "Test desc",
			wantTags:    []string{"tag1"},
			wantSummary: "Test summary",
		},
		{
			name: "JSON with surrounding text",
			input: "Here is the metadata:\n{\"description\": \"Desc\", \"tags\": [], \"summary\": \"Sum\"}\nHope this helps!",
			wantDesc:    "Desc",
			wantTags:    []string{},
			wantSummary: "Sum",
		},
		{
			name:    "invalid JSON",
			input:   "not json at all",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := parseMetaJSON(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if meta.Description != tt.wantDesc {
				t.Errorf("description = %q, want %q", meta.Description, tt.wantDesc)
			}
			if meta.Summary != tt.wantSummary {
				t.Errorf("summary = %q, want %q", meta.Summary, tt.wantSummary)
			}
			if len(meta.Tags) != len(tt.wantTags) {
				t.Errorf("tags count = %d, want %d", len(meta.Tags), len(tt.wantTags))
			}
			for i, tag := range meta.Tags {
				if i < len(tt.wantTags) && tag != tt.wantTags[i] {
					t.Errorf("tag[%d] = %q, want %q", i, tag, tt.wantTags[i])
				}
			}
		})
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "clean JSON",
			input: `{"key": "value"}`,
			want:  `{"key": "value"}`,
		},
		{
			name:  "with code fence",
			input: "```json\n{\"key\": \"value\"}\n```",
			want:  `{"key": "value"}`,
		},
		{
			name:  "with surrounding text",
			input: "Here is the JSON: {\"key\": \"value\"} That's it!",
			want:  `{"key": "value"}`,
		},
		{
			name:  "nested objects",
			input: `{"outer": {"inner": "value"}}`,
			want:  `{"outer": {"inner": "value"}}`,
		},
		{
			name:  "no JSON",
			input: "no json here",
			want:  "no json here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.want {
				t.Errorf("extractJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateContent(t *testing.T) {
	short := "short content"
	if got := truncateContent(short, 100); got != short {
		t.Errorf("truncateContent should not modify short content")
	}

	long := "a]" // 100 chars
	for i := 0; i < 50; i++ {
		long = long + "ab"
	}
	truncated := truncateContent(long, 50)
	if len(truncated) <= 50 {
		t.Error("truncated content should include truncation notice")
	}
	if truncated[len(truncated)-4:] != "...]" {
		t.Error("truncated content should end with truncation marker")
	}
}

// TestClaudeCLIIntegration tests actual Claude CLI calls.
// Skip if claude is not installed.
func TestClaudeCLIIntegration(t *testing.T) {
	// Check if claude CLI is available
	_, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude CLI not installed, skipping integration test")
	}

	cfg := config.AIConfig{
		Enabled:           true,
		GenerateSummaries: true,
		GenerateTags:      true,
	}

	content := `# Blind Records

Blind records are encrypted data structures that allow secure storage
without revealing content to the storage provider. Each record has a
unique identifier and is encrypted with a key derived from the user's
master key.

## Features

- End-to-end encryption
- Zero-knowledge storage
- Deterministic key derivation
`

	meta, err := GenerateMeta(cfg, content, "Blind Records")
	if err != nil {
		t.Fatalf("GenerateMeta failed: %v", err)
	}

	// Verify we got something back
	if meta.Description == "" {
		t.Error("expected non-empty description")
	}
	if len(meta.Tags) == 0 {
		t.Error("expected at least one tag")
	}
	if meta.Summary == "" {
		t.Error("expected non-empty summary")
	}

	t.Logf("AI Generated metadata:")
	t.Logf("  Description: %s", meta.Description)
	t.Logf("  Tags: %v", meta.Tags)
	t.Logf("  Summary: %s", meta.Summary)
}
