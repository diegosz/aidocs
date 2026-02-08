// Package config handles parsing and validation of .aidocs.yaml configuration.
package config

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete aidocs configuration.
type Config struct {
	Content string        `yaml:"content"`
	Output  OutputConfig  `yaml:"output"`
	AI      AIConfig      `yaml:"ai"`
	Project ProjectConfig `yaml:"project"`

	// ConfigDir is the directory containing the config file (for relative path resolution)
	ConfigDir string `yaml:"-"`
}

// OutputConfig defines output file paths.
type OutputConfig struct {
	LLMsTxt  string `yaml:"llms_txt"`
	LLMsFull string `yaml:"llms_full"`
	Manifest string `yaml:"manifest"`
	Cache    string `yaml:"cache"`
}

// AIConfig defines AI-powered feature settings.
// Uses Claude Code CLI (claude -p) for AI features - no API key needed.
type AIConfig struct {
	Enabled                    bool `yaml:"enabled"`
	GenerateSummaries          bool `yaml:"generate_summaries"`
	GenerateMissingFrontmatter bool `yaml:"generate_missing_frontmatter"`
	GenerateDescriptions       bool `yaml:"generate_descriptions"`
	GenerateTags               bool `yaml:"generate_tags"`
}

// ProjectConfig defines project metadata.
type ProjectConfig struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version"`
	OptimizedFor []string `yaml:"optimized_for"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Content: "docs/SUMMARY.md",
		Output: OutputConfig{
			LLMsTxt:  "llms.txt",
			LLMsFull: "docs/llms-full.txt",
			Manifest: "docs/ai-optimization/manifest.json",
			Cache:    "docs/ai-optimization/.cache.json",
		},
		AI: AIConfig{
			Enabled:                    false,
			GenerateSummaries:          true,
			GenerateMissingFrontmatter: true,
			GenerateDescriptions:       true,
			GenerateTags:               true,
		},
		Project: ProjectConfig{
			Name:         "Project",
			Description:  "",
			Version:      "",
			OptimizedFor: []string{"Claude Code", "AI Agents"},
		},
	}
}

// Load reads and parses the configuration file.
// If the file doesn't exist, returns defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Use defaults if config doesn't exist
			cfg.ConfigDir, _ = os.Getwd()
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Store config directory for relative path resolution
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cfg.ConfigDir = filepath.Dir(absPath)

	// Resolve relative paths
	cfg.Content = cfg.resolvePath(cfg.Content)
	cfg.Output.LLMsTxt = cfg.resolvePath(cfg.Output.LLMsTxt)
	cfg.Output.LLMsFull = cfg.resolvePath(cfg.Output.LLMsFull)
	cfg.Output.Manifest = cfg.resolvePath(cfg.Output.Manifest)
	cfg.Output.Cache = cfg.resolvePath(cfg.Output.Cache)

	// Get version from git if not specified
	if cfg.Project.Version == "" {
		cfg.Project.Version = getGitVersion()
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// resolvePath resolves a path relative to the config directory.
func (c *Config) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.ConfigDir, path)
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Content == "" {
		return errors.New("content path is required")
	}

	// Check if SUMMARY.md exists
	if _, err := os.Stat(c.Content); err != nil {
		return err
	}

	// AI uses Claude Code CLI - no additional validation needed

	return nil
}

// getGitVersion attempts to get the version using git describe.
func getGitVersion() string {
	cmd := exec.Command("git", "describe", "--always", "--tags", "--match=v*")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
