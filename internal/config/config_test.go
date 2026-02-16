package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSaveProjectFields(t *testing.T) {
	const input = `# .aidocs.yaml
content: "docs/SUMMARY.md"

# AI features
ai:
  enabled: true

# Project metadata
project:
  name: ""
  description: ""
  optimized_for:
    - "Claude Code"
`

	t.Run("saves both name and description", func(t *testing.T) {
		path := writeTempConfig(t, input)

		err := SaveProjectFields(path, "MyProject", "A cool project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cfg := loadAndVerify(t, path)
		if cfg.Project.Name != "MyProject" {
			t.Errorf("expected name MyProject, got %q", cfg.Project.Name)
		}
		if cfg.Project.Description != "A cool project" {
			t.Errorf("expected description 'A cool project', got %q", cfg.Project.Description)
		}
	})

	t.Run("saves only name when description is empty", func(t *testing.T) {
		path := writeTempConfig(t, input)

		err := SaveProjectFields(path, "MyProject", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cfg := loadAndVerify(t, path)
		if cfg.Project.Name != "MyProject" {
			t.Errorf("expected name MyProject, got %q", cfg.Project.Name)
		}
		if cfg.Project.Description != "" {
			t.Errorf("expected description to remain empty, got %q", cfg.Project.Description)
		}
	})

	t.Run("saves only description when name is empty", func(t *testing.T) {
		path := writeTempConfig(t, input)

		err := SaveProjectFields(path, "", "A cool project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cfg := loadAndVerify(t, path)
		if cfg.Project.Description != "A cool project" {
			t.Errorf("expected description 'A cool project', got %q", cfg.Project.Description)
		}
	})

	t.Run("no-op when both empty", func(t *testing.T) {
		path := writeTempConfig(t, input)

		err := SaveProjectFields(path, "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, _ := os.ReadFile(path)
		if string(data) != input {
			t.Errorf("file should not have been modified")
		}
	})

	t.Run("preserves comments", func(t *testing.T) {
		path := writeTempConfig(t, input)

		err := SaveProjectFields(path, "MyProject", "A cool project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, _ := os.ReadFile(path)
		content := string(data)

		if !strings.Contains(content, "# AI features") {
			t.Errorf("expected comments to be preserved, got:\n%s", content)
		}
		if !strings.Contains(content, "# Project metadata") {
			t.Errorf("expected comments to be preserved, got:\n%s", content)
		}
	})

	t.Run("preserves other fields", func(t *testing.T) {
		path := writeTempConfig(t, input)

		err := SaveProjectFields(path, "MyProject", "A cool project")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, _ := os.ReadFile(path)
		content := string(data)

		if !strings.Contains(content, "enabled: true") {
			t.Errorf("expected ai.enabled to be preserved, got:\n%s", content)
		}
		if !strings.Contains(content, "Claude Code") {
			t.Errorf("expected optimized_for to be preserved, got:\n%s", content)
		}
	})

	t.Run("error on missing file", func(t *testing.T) {
		err := SaveProjectFields("/nonexistent/path.yaml", "Name", "Desc")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".aidocs.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}

// loadAndVerify re-parses the saved config to verify field values.
func loadAndVerify(t *testing.T, path string) *Config {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse saved config: %v", err)
	}
	return &cfg
}
