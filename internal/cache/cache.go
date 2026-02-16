// Package cache provides SHA256-based change detection for documentation files.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/diegosz/aidocs/internal/docparser"
)

// Cache stores hashes and metadata for processed files.
type Cache struct {
	Version     string                `json:"version"`
	GeneratedAt time.Time             `json:"generatedAt"`
	Files       map[string]*FileCache `json:"files"`
}

// FileCache stores cache data for a single file.
type FileCache struct {
	FileHash    string    `json:"fileHash"`          // Hash of entire file (quick check)
	ContentHash string    `json:"contentHash"`       // Hash excluding frontmatter (content change detection)
	Summary     string    `json:"summary,omitempty"` // Cached AI summary
	GeneratedAt time.Time `json:"generatedAt"`
}

// New creates a new empty cache.
func New() *Cache {
	return &Cache{
		Version:     "1",
		GeneratedAt: time.Now(),
		Files:       make(map[string]*FileCache),
	}
}

// Load reads cache from disk.
func Load(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	if cache.Files == nil {
		cache.Files = make(map[string]*FileCache)
	}

	return &cache, nil
}

// Save writes cache to disk.
func (c *Cache) Save(path string) error {
	c.GeneratedAt = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// HasChanged checks if a file's content has changed since last processing.
// Returns true if the file has changed or is not in cache.
// Uses two-level hashing: first checks fileHash (quick), then contentHash if needed.
func (c *Cache) HasChanged(path string) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return true, err
	}

	// Check if file is in cache
	cached, ok := c.Files[absPath]
	if !ok {
		return true, nil
	}

	// Read file once
	content, err := os.ReadFile(path)
	if err != nil {
		return true, err
	}

	// Quick check: compare full file hash first
	fileHash := hashBytes(content)
	if cached.FileHash == fileHash {
		// File unchanged - no need to parse frontmatter
		return false, nil
	}

	// File changed - check if actual content (excluding frontmatter) changed
	body := docparser.StripFrontmatter(content)
	contentHash := hashBytes(body)

	return cached.ContentHash != contentHash, nil
}

// Update updates the cache entry for a file.
func (c *Cache) Update(path, summary string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fileHash := hashBytes(content)
	body := docparser.StripFrontmatter(content)
	contentHash := hashBytes(body)

	c.Files[absPath] = &FileCache{
		FileHash:    fileHash,
		ContentHash: contentHash,
		Summary:     summary,
		GeneratedAt: time.Now(),
	}

	return nil
}

// GetSummary returns the cached summary for a file, if available.
func (c *Cache) GetSummary(path string) (string, bool) {
	absPath, _ := filepath.Abs(path)
	if cached, ok := c.Files[absPath]; ok && cached.Summary != "" {
		return cached.Summary, true
	}
	return "", false
}

// ContentHash calculates SHA256 hash of file content, excluding frontmatter.
func ContentHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Strip frontmatter - we only hash the body content
	body := docparser.StripFrontmatter(content)
	return hashBytes(body), nil
}

// hashBytes computes SHA256 hash of bytes.
func hashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
