// Package config handles parsing and validation of .aidocs.yaml configuration.
package config

import (
	"errors"
	"os"
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
	Tags     string `yaml:"tags"`
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
	OptimizedFor []string `yaml:"optimized_for"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Content: "docs/SUMMARY.md",
		Output: OutputConfig{
			LLMsTxt:  "llms.txt",
			LLMsFull: "docs/llms-full.txt",
			Manifest: "docs/_ai/manifest.json",
			Tags:     "docs/_ai/tags.json",
			Cache:    "docs/_ai/.cache.json",
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
	cfg.Output.Tags = cfg.resolvePath(cfg.Output.Tags)
	cfg.Output.Cache = cfg.resolvePath(cfg.Output.Cache)

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

// ValidateProjectName checks that the project name is non-empty.
// Called separately from Validate() since it runs after AI inference.
func (c *Config) ValidateProjectName() error {
	if strings.TrimSpace(c.Project.Name) == "" {
		return errors.New("project.name is required: set it in .aidocs.yaml or enable ai to infer it")
	}
	return nil
}

// SaveProjectFields updates the project.name and/or project.description fields
// in the config file, preserving comments and formatting.
// Pass empty string for fields that should not be updated.
func SaveProjectFields(path, name, description string) error {
	if name == "" && description == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return err
	}

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return errors.New("invalid YAML structure")
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return errors.New("expected mapping at root")
	}

	projectNode := findMapValue(root, "project")
	if projectNode == nil {
		return errors.New("project section not found in config")
	}
	if projectNode.Kind != yaml.MappingNode {
		return errors.New("project section is not a mapping")
	}

	if name != "" {
		setMapValue(projectNode, "name", name)
	}
	if description != "" {
		setMapValue(projectNode, "description", description)
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0o600)
}

// findMapValue finds the value node for a given key in a mapping node.
func findMapValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// setMapValue sets or adds a scalar value for a key in a mapping node.
func setMapValue(mapping *yaml.Node, key, value string) {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1].Value = value
			return
		}
	}
	// Key doesn't exist, add it
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value, Style: yaml.DoubleQuotedStyle},
	)
}
