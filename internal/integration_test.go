package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/diegosz/aidocs/internal/cache"
	"github.com/diegosz/aidocs/internal/config"
	"github.com/diegosz/aidocs/internal/generator"
	"github.com/diegosz/aidocs/internal/parser"
)

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// runAidocs simulates running aidocs on a directory.
func runAidocs(t *testing.T, configPath string) ([]*generator.Document, *config.Config) {
	t.Helper()

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	entries, err := parser.ParseSummary(cfg.Content)
	if err != nil {
		t.Fatalf("Failed to parse SUMMARY.md: %v", err)
	}

	docs := make([]*generator.Document, 0, len(entries))
	summaryDir := filepath.Dir(cfg.Content)

	for _, entry := range entries {
		docPath := filepath.Join(summaryDir, entry.Path)

		fm, _, err := parser.ExtractFrontmatter(docPath)
		if err != nil {
			t.Errorf("Failed to extract frontmatter from %s: %v", docPath, err)
			continue
		}

		if fm.Title == "" {
			fm.Title = entry.Title
		}
		if fm.Category == "" && entry.Category != "" {
			fm.Category = entry.Category
		}

		docs = append(docs, &generator.Document{
			Path:        docPath,
			Frontmatter: fm,
			Category:    entry.Category,
		})
	}

	// Generate outputs
	manifest := generator.GenerateManifest(docs, cfg.Project)
	if err := generator.WriteManifest(cfg.Output.Manifest, manifest, false); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	if err := generator.WriteLLMsFull(cfg.Output.LLMsFull, docs, cfg.Project, false); err != nil {
		t.Fatalf("Failed to write llms-full.txt: %v", err)
	}

	if err := generator.WriteLLMsTxt(cfg.Output.LLMsTxt, docs, cfg, false); err != nil {
		t.Fatalf("Failed to write llms.txt: %v", err)
	}

	// Generate tags.json
	tags := generator.GenerateTags(docs)
	if err := generator.WriteTags(cfg.Output.Tags, tags, false); err != nil {
		t.Fatalf("Failed to write tags.json: %v", err)
	}

	// Save cache
	docCache := cache.New()
	for _, doc := range docs {
		if err := docCache.Update(doc.Path, ""); err != nil {
			t.Errorf("Failed to update cache for %s: %v", doc.Path, err)
		}
	}
	if err := docCache.Save(cfg.Output.Cache); err != nil {
		t.Errorf("Failed to save cache: %v", err)
	}

	return docs, cfg
}

// TestGreenfieldGeneration tests generation from scratch with mix of frontmatter/no-frontmatter files.
func TestGreenfieldGeneration(t *testing.T) {
	// Copy source to temp directory
	sourceDir := filepath.Join("..", "testdata", "blind", "source")
	tempDir := t.TempDir()

	if err := copyDir(sourceDir, tempDir); err != nil {
		t.Fatalf("Failed to copy source: %v", err)
	}

	configPath := filepath.Join(tempDir, ".aidocs.yaml")

	// Run aidocs
	docs, cfg := runAidocs(t, configPath)

	// Verify document count
	if len(docs) != 9 {
		t.Errorf("Expected 9 documents, got %d", len(docs))
	}

	// Verify mix of frontmatter states
	withFrontmatter := 0
	withoutFrontmatter := 0
	for _, doc := range docs {
		if doc.Frontmatter.Description != "" {
			withFrontmatter++
		} else {
			withoutFrontmatter++
		}
	}

	if withFrontmatter == 0 {
		t.Error("Expected some documents with frontmatter")
	}
	if withoutFrontmatter == 0 {
		t.Error("Expected some documents without frontmatter")
	}
	t.Logf("Documents with frontmatter: %d, without: %d", withFrontmatter, withoutFrontmatter)

	// Verify outputs exist
	if _, err := os.Stat(cfg.Output.Manifest); err != nil {
		t.Errorf("Manifest not created: %v", err)
	}
	if _, err := os.Stat(cfg.Output.LLMsFull); err != nil {
		t.Errorf("llms-full.txt not created: %v", err)
	}
	if _, err := os.Stat(cfg.Output.LLMsTxt); err != nil {
		t.Errorf("llms.txt not created: %v", err)
	}
	if _, err := os.Stat(cfg.Output.Cache); err != nil {
		t.Errorf("Cache not created: %v", err)
	}
	if _, err := os.Stat(cfg.Output.Tags); err != nil {
		t.Errorf("tags.json not created: %v", err)
	}

	// Verify tags.json structure
	tagsData, _ := os.ReadFile(cfg.Output.Tags)
	var tagsFile map[string]any
	if err := json.Unmarshal(tagsData, &tagsFile); err != nil {
		t.Fatalf("Failed to parse tags.json: %v", err)
	}
	if _, ok := tagsFile["tags"]; !ok {
		t.Error("tags.json missing 'tags' field")
	}
	if _, ok := tagsFile["totalTags"]; !ok {
		t.Error("tags.json missing 'totalTags' field")
	}
	if _, ok := tagsFile["topTags"]; !ok {
		t.Error("tags.json missing 'topTags' field")
	}

	// Verify manifest structure
	manifestData, _ := os.ReadFile(cfg.Output.Manifest)
	var manifest map[string]any
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	// Check sections include those from frontmatter and SUMMARY.md
	sections, ok := manifest["sections"].(map[string]any)
	if !ok {
		t.Fatal("Expected sections to be a map")
	}
	if _, ok := sections["examples"]; !ok {
		t.Error("Expected 'examples' section from frontmatter")
	}
	if _, ok := sections["reference"]; !ok {
		t.Error("Expected 'reference' section from frontmatter")
	}
	if _, ok := sections["WIP"]; !ok {
		t.Error("Expected 'WIP' section from SUMMARY.md")
	}

	// Verify documents have frontmatter data where applicable
	documents, ok := manifest["documents"].([]any)
	if !ok {
		t.Fatal("Expected documents to be an array")
	}
	foundWithTags := false
	foundWithoutTags := false
	for _, p := range documents {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		if _, ok := pm["tags"]; ok {
			foundWithTags = true
		} else {
			foundWithoutTags = true
		}
	}
	if !foundWithTags {
		t.Error("Expected some documents with tags from frontmatter")
	}
	if !foundWithoutTags {
		t.Error("Expected some documents without tags")
	}
}

// TestBrownfieldGeneration tests re-generation on already processed files (idempotent).
func TestBrownfieldGeneration(t *testing.T) {
	// Copy source to temp directory
	sourceDir := filepath.Join("..", "testdata", "blind", "source")
	tempDir := t.TempDir()

	if err := copyDir(sourceDir, tempDir); err != nil {
		t.Fatalf("Failed to copy source: %v", err)
	}

	configPath := filepath.Join(tempDir, ".aidocs.yaml")

	// First run
	docs1, cfg := runAidocs(t, configPath)

	// Read first manifest
	manifest1Data, _ := os.ReadFile(cfg.Output.Manifest)
	llmsTxt1, _ := os.ReadFile(cfg.Output.LLMsTxt)
	llmsFull1, _ := os.ReadFile(cfg.Output.LLMsFull)

	// Second run (brownfield)
	docs2, _ := runAidocs(t, configPath)

	// Read second manifest
	manifest2Data, _ := os.ReadFile(cfg.Output.Manifest)
	llmsTxt2, _ := os.ReadFile(cfg.Output.LLMsTxt)
	llmsFull2, _ := os.ReadFile(cfg.Output.LLMsFull)

	// Verify same document count
	if len(docs1) != len(docs2) {
		t.Errorf("Document count changed: %d -> %d", len(docs1), len(docs2))
	}

	// Parse manifests and compare (ignoring timestamps)
	var m1, m2 map[string]any
	if err := json.Unmarshal(manifest1Data, &m1); err != nil {
		t.Fatalf("Failed to parse manifest1: %v", err)
	}
	if err := json.Unmarshal(manifest2Data, &m2); err != nil {
		t.Fatalf("Failed to parse manifest2: %v", err)
	}

	// Compare documents
	p1, _ := m1["documents"].([]any)
	p2, _ := m2["documents"].([]any)
	if len(p1) != len(p2) {
		t.Errorf("Document count changed: %d -> %d", len(p1), len(p2))
	}

	// Compare sections
	c1, _ := m1["sections"].(map[string]any)
	c2, _ := m2["sections"].(map[string]any)
	if len(c1) != len(c2) {
		t.Errorf("Section count changed: %d -> %d", len(c1), len(c2))
	}

	// llms.txt should be identical (not regenerated if exists)
	if !bytes.Equal(llmsTxt1, llmsTxt2) {
		t.Error("llms.txt changed on second run - should be preserved")
	}

	// llms-full.txt content should be equivalent
	if len(llmsFull1) != len(llmsFull2) {
		t.Logf("llms-full.txt size: %d -> %d", len(llmsFull1), len(llmsFull2))
	}
}

// TestCachePreservesState tests that cache correctly tracks file changes.
func TestCachePreservesState(t *testing.T) {
	// Copy source to temp directory
	sourceDir := filepath.Join("..", "testdata", "blind", "source")
	tempDir := t.TempDir()

	if err := copyDir(sourceDir, tempDir); err != nil {
		t.Fatalf("Failed to copy source: %v", err)
	}

	configPath := filepath.Join(tempDir, ".aidocs.yaml")
	cfg, _ := config.Load(configPath)

	// First run to populate cache
	runAidocs(t, configPath)

	// Load cache
	docCache, err := cache.Load(cfg.Output.Cache)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	// Check all source files are in cache
	entries, _ := parser.ParseSummary(cfg.Content)
	summaryDir := filepath.Dir(cfg.Content)

	for _, entry := range entries {
		docPath := filepath.Join(summaryDir, entry.Path)
		changed, err := docCache.HasChanged(docPath)
		if err != nil {
			t.Errorf("HasChanged failed for %s: %v", entry.Path, err)
		}
		if changed {
			t.Errorf("File %s marked as changed but was just cached", entry.Path)
		}
	}

	// Modify one file
	recordsPath := filepath.Join(summaryDir, "records.md")
	content, _ := os.ReadFile(recordsPath)
	newContent := string(content) + "\n\n## New Section\n\nAdded content."
	if err := os.WriteFile(recordsPath, []byte(newContent), 0o600); err != nil {
		t.Fatalf("Failed to write modified file: %v", err)
	}

	// Check only modified file is marked as changed
	for _, entry := range entries {
		docPath := filepath.Join(summaryDir, entry.Path)
		changed, _ := docCache.HasChanged(docPath)

		if entry.Path == "records.md" {
			if !changed {
				t.Error("Modified file records.md should be marked as changed")
			}
		} else {
			if changed {
				t.Errorf("Unmodified file %s should not be marked as changed", entry.Path)
			}
		}
	}
}

// TestExpectedOutputComparison compares generated output with expected fixtures.
func TestExpectedOutputComparison(t *testing.T) {
	// Copy source to temp directory
	sourceDir := filepath.Join("..", "testdata", "blind", "source")
	expectedDir := filepath.Join("..", "testdata", "blind", "expected")
	tempDir := t.TempDir()

	if err := copyDir(sourceDir, tempDir); err != nil {
		t.Fatalf("Failed to copy source: %v", err)
	}

	configPath := filepath.Join(tempDir, ".aidocs.yaml")

	// Run aidocs
	_, cfg := runAidocs(t, configPath)

	// Compare manifest structure with expected
	generatedManifest, _ := os.ReadFile(cfg.Output.Manifest)
	expectedManifest, err := os.ReadFile(filepath.Join(expectedDir, "manifest.json"))
	if err != nil {
		t.Fatalf("Failed to read expected manifest: %v", err)
	}

	var genM, expM map[string]any
	if err := json.Unmarshal(generatedManifest, &genM); err != nil {
		t.Fatalf("Failed to parse generated manifest: %v", err)
	}
	if err := json.Unmarshal(expectedManifest, &expM); err != nil {
		t.Fatalf("Failed to parse expected manifest: %v", err)
	}

	// Compare document IDs
	genDocs, _ := genM["documents"].([]any)
	expDocs, _ := expM["documents"].([]any)

	if len(genDocs) != len(expDocs) {
		t.Errorf("Document count mismatch: generated %d, expected %d", len(genDocs), len(expDocs))
	}

	expIDs := make(map[string]bool)
	for _, p := range expDocs {
		if pm, ok := p.(map[string]any); ok {
			if id, ok := pm["id"].(string); ok {
				expIDs[id] = true
			}
		}
	}

	for _, p := range genDocs {
		if pm, ok := p.(map[string]any); ok {
			if id, ok := pm["id"].(string); ok {
				if !expIDs[id] {
					t.Errorf("Unexpected document ID: %s", id)
				}
			}
		}
	}

	// Compare sections
	genSecs, _ := genM["sections"].(map[string]any)
	expSecs, _ := expM["sections"].(map[string]any)

	for secName := range expSecs {
		if _, ok := genSecs[secName]; !ok {
			t.Errorf("Missing expected section: %s", secName)
		}
	}
}

func TestChangeDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.md")
	content1 := "# Test\n\nThis is test content."
	if err := os.WriteFile(testFile, []byte(content1), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create cache
	c := cache.New()

	// First check should indicate change (not in cache)
	changed, err := c.HasChanged(testFile)
	if err != nil {
		t.Fatalf("HasChanged failed: %v", err)
	}
	if !changed {
		t.Error("Expected file to be marked as changed (not in cache)")
	}

	// Update cache
	if err := c.Update(testFile, "test summary"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Second check should indicate no change
	changed, err = c.HasChanged(testFile)
	if err != nil {
		t.Fatalf("HasChanged failed: %v", err)
	}
	if changed {
		t.Error("Expected file to be marked as unchanged")
	}

	// Modify content
	content2 := "# Test\n\nThis is different content."
	if err := os.WriteFile(testFile, []byte(content2), 0o600); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Should detect change
	changed, err = c.HasChanged(testFile)
	if err != nil {
		t.Fatalf("HasChanged failed: %v", err)
	}
	if !changed {
		t.Error("Expected file to be marked as changed after modification")
	}

	// Save and reload cache
	cachePath := filepath.Join(tempDir, "cache.json")
	if err := c.Save(cachePath); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}

	c2, err := cache.Load(cachePath)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	// Verify summary was preserved
	summary, ok := c2.GetSummary(testFile)
	if !ok || summary != "test summary" {
		t.Errorf("Expected summary 'test summary', got '%s'", summary)
	}
}

func TestChangeDetectionExcludesFrontmatter(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file with frontmatter
	testFile := filepath.Join(tempDir, "test.md")
	content1 := `---
title: "Test"
description: "Original description"
---

# Test

This is test content.`
	if err := os.WriteFile(testFile, []byte(content1), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Get initial hash
	hash1, err := cache.ContentHash(testFile)
	if err != nil {
		t.Fatalf("ContentHash failed: %v", err)
	}

	// Modify only frontmatter
	content2 := `---
title: "Test"
description: "Updated description"
tags: ["new", "tags"]
---

# Test

This is test content.`
	if err := os.WriteFile(testFile, []byte(content2), 0o600); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Hash should be the same (frontmatter excluded)
	hash2, err := cache.ContentHash(testFile)
	if err != nil {
		t.Fatalf("ContentHash failed: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Hash changed when only frontmatter was modified - frontmatter should be excluded")
	}

	// Modify body content
	content3 := `---
title: "Test"
description: "Updated description"
tags: ["new", "tags"]
---

# Test

This is DIFFERENT content.`
	if err := os.WriteFile(testFile, []byte(content3), 0o600); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Hash should be different now
	hash3, err := cache.ContentHash(testFile)
	if err != nil {
		t.Fatalf("ContentHash failed: %v", err)
	}

	if hash1 == hash3 {
		t.Error("Hash should change when body content is modified")
	}
}

func TestOrphanDetection(t *testing.T) {
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// Create SUMMARY.md with one reference
	summaryContent := `# Test Project

- [Intro](intro.md)
`
	summaryPath := filepath.Join(docsDir, "SUMMARY.md")
	if err := os.WriteFile(summaryPath, []byte(summaryContent), 0o600); err != nil {
		t.Fatalf("Failed to write SUMMARY.md: %v", err)
	}

	// Create referenced file
	if err := os.WriteFile(filepath.Join(docsDir, "intro.md"), []byte("# Intro"), 0o600); err != nil {
		t.Fatalf("Failed to write intro.md: %v", err)
	}

	// Create orphan file
	if err := os.WriteFile(filepath.Join(docsDir, "orphan.md"), []byte("# Orphan"), 0o600); err != nil {
		t.Fatalf("Failed to write orphan.md: %v", err)
	}

	entries, _ := parser.ParseSummary(summaryPath)
	orphans, err := parser.FindOrphans(summaryPath, entries)
	if err != nil {
		t.Fatalf("FindOrphans failed: %v", err)
	}

	if !slices.Contains(orphans, "orphan.md") {
		t.Errorf("Expected to find orphan.md, got: %v", orphans)
	}
}

func TestFrontmatterExtraction(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		content        string
		expectedTitle  string
		expectedDesc   string
		expectedTags   int
		expectedHasErr bool
	}{
		{
			name: "no_frontmatter",
			content: `# My Document

This is content.`,
			expectedTitle: "My Document",
			expectedDesc:  "",
			expectedTags:  0,
		},
		{
			name: "with_frontmatter",
			content: `---
title: "Explicit Title"
description: "A description"
tags: ["one", "two"]
---

# My Document

This is content.`,
			expectedTitle: "My Document", // H1 takes precedence
			expectedDesc:  "A description",
			expectedTags:  2,
		},
		{
			name: "frontmatter_only",
			content: `---
title: "Only Frontmatter"
description: "No H1 header"
---

Some content without H1.`,
			expectedTitle: "Only Frontmatter",
			expectedDesc:  "No H1 header",
			expectedTags:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, tc.name+".md")
			if err := os.WriteFile(testFile, []byte(tc.content), 0o600); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			fm, body, err := parser.ExtractFrontmatter(testFile)
			if tc.expectedHasErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if fm.Title != tc.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tc.expectedTitle, fm.Title)
			}

			if fm.Description != tc.expectedDesc {
				t.Errorf("Expected description '%s', got '%s'", tc.expectedDesc, fm.Description)
			}

			if len(fm.Tags) != tc.expectedTags {
				t.Errorf("Expected %d tags, got %d", tc.expectedTags, len(fm.Tags))
			}

			if body == "" {
				t.Error("Expected non-empty body")
			}
		})
	}
}

func TestFrontmatterSpacingIdempotent(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file with H1 title
	testFile := filepath.Join(tempDir, "test.md")
	originalContent := `# Test Document

This is the content.`
	if err := os.WriteFile(testFile, []byte(originalContent), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fm := &parser.Frontmatter{
		Title:       "Test Document",
		Description: "A test description",
		Tags:        []string{"test", "example"},
	}

	// First write
	if err := parser.WriteFrontmatter(testFile, fm, originalContent, false); err != nil {
		t.Fatalf("First WriteFrontmatter failed: %v", err)
	}

	content1, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after first write: %v", err)
	}

	// Extract body and write again (simulates running --force twice)
	_, body, err := parser.ExtractFrontmatter(testFile)
	if err != nil {
		t.Fatalf("ExtractFrontmatter failed: %v", err)
	}

	if err := parser.WriteFrontmatter(testFile, fm, body, false); err != nil {
		t.Fatalf("Second WriteFrontmatter failed: %v", err)
	}

	content2, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after second write: %v", err)
	}

	// Content should be identical after second write
	if !bytes.Equal(content1, content2) {
		t.Error("Content changed after second WriteFrontmatter - should be idempotent")
		t.Logf("After first write:\n%s", string(content1))
		t.Logf("After second write:\n%s", string(content2))
	}

	// Verify exactly one empty line between frontmatter and H1
	// The file should contain "---\n\n#" (closing delimiter, one empty line, H1)
	contentStr := string(content2)
	if !strings.Contains(contentStr, "---\n\n#") {
		t.Error("Expected exactly one empty line between frontmatter and H1")
		t.Logf("Content:\n%s", contentStr)
	}
	// And should NOT have two empty lines
	if strings.Contains(contentStr, "---\n\n\n") {
		t.Error("Found more than one empty line between frontmatter and content")
		t.Logf("Content:\n%s", contentStr)
	}
}

func TestFrontmatterSpacingWithExistingNewlines(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name         string
		body         string
		expectedBody string
	}{
		{
			name:         "no_leading_newlines",
			body:         "# Title\n\nContent",
			expectedBody: "# Title\n\nContent",
		},
		{
			name:         "one_leading_newline",
			body:         "\n# Title\n\nContent",
			expectedBody: "# Title\n\nContent",
		},
		{
			name:         "multiple_leading_newlines",
			body:         "\n\n\n# Title\n\nContent",
			expectedBody: "# Title\n\nContent",
		},
		{
			name:         "leading_spaces_and_newlines",
			body:         "  \n\n# Title\n\nContent",
			expectedBody: "# Title\n\nContent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, tc.name+".md")

			fm := &parser.Frontmatter{
				Title:       "Title",
				Description: "Description",
			}

			if err := parser.WriteFrontmatter(testFile, fm, tc.body, false); err != nil {
				t.Fatalf("WriteFrontmatter failed: %v", err)
			}

			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// Extract body from written content
			_, extractedBody, err := parser.ExtractFrontmatter(testFile)
			if err != nil {
				t.Fatalf("ExtractFrontmatter failed: %v", err)
			}

			// Body should match expected (trimmed)
			if extractedBody != tc.expectedBody {
				t.Errorf("Expected body:\n%q\nGot:\n%q", tc.expectedBody, extractedBody)
			}

			// Verify spacing in raw content
			contentStr := string(content)
			if !strings.Contains(contentStr, "---\n\n# Title") {
				t.Error("Expected exactly one empty line between frontmatter and H1")
				t.Logf("Content:\n%s", contentStr)
			}
		})
	}
}

func TestSummaryParsing(t *testing.T) {
	tempDir := t.TempDir()

	summaryContent := `# My Project

- [Introduction](intro.md)
- [Getting Started](start.md)
- Advanced
  - [Deep Dive](deep.md)
  - [Expert Mode](expert.md)
- [Conclusion](end.md)
`
	summaryPath := filepath.Join(tempDir, "SUMMARY.md")
	if err := os.WriteFile(summaryPath, []byte(summaryContent), 0o600); err != nil {
		t.Fatalf("Failed to write SUMMARY.md: %v", err)
	}

	entries, err := parser.ParseSummary(summaryPath)
	if err != nil {
		t.Fatalf("ParseSummary failed: %v", err)
	}

	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
	}

	// Check sections (Category field stores section from SUMMARY.md)
	advancedCount := 0
	for _, e := range entries {
		if e.Category == "Advanced" {
			advancedCount++
		}
	}
	if advancedCount != 2 {
		t.Errorf("Expected 2 Advanced section entries, got %d", advancedCount)
	}

	// Check paths
	for _, e := range entries {
		if e.Path == "" {
			t.Errorf("Entry '%s' has empty path", e.Title)
		}
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.Content != "docs/SUMMARY.md" {
		t.Errorf("Expected default content 'docs/SUMMARY.md', got '%s'", cfg.Content)
	}

	if cfg.Output.LLMsTxt != "llms.txt" {
		t.Errorf("Expected default llms_txt 'llms.txt', got '%s'", cfg.Output.LLMsTxt)
	}

	if cfg.AI.Enabled {
		t.Error("Expected AI to be disabled by default")
	}

	if !cfg.AI.GenerateSummaries {
		t.Error("Expected GenerateSummaries to be true by default")
	}
}

func TestFrontmatterPreservedInOutput(t *testing.T) {
	// Copy source to temp directory
	sourceDir := filepath.Join("..", "testdata", "blind", "source")
	tempDir := t.TempDir()

	if err := copyDir(sourceDir, tempDir); err != nil {
		t.Fatalf("Failed to copy source: %v", err)
	}

	configPath := filepath.Join(tempDir, ".aidocs.yaml")
	docs, cfg := runAidocs(t, configPath)

	// Read manifest
	manifestData, _ := os.ReadFile(cfg.Output.Manifest)
	var manifest map[string]any
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}

	manifestDocs, _ := manifest["documents"].([]any)

	// Find a document that should have frontmatter data
	foundEncoding := false
	for _, p := range manifestDocs {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		if pm["id"] == "blind-records-encoding" {
			foundEncoding = true

			// Should have description from frontmatter
			if pm["description"] != "Encoding format for blind records" {
				t.Errorf("Expected description from frontmatter, got: %v", pm["description"])
			}

			// Should have tags from frontmatter
			if tags, ok := pm["tags"].([]any); !ok || len(tags) != 3 {
				t.Errorf("Expected 3 tags from frontmatter, got: %v", pm["tags"])
			}

			// Should have estimatedTokens from frontmatter
			if tokens, ok := pm["estimatedTokens"].(float64); !ok || tokens != 450 {
				t.Errorf("Expected estimatedTokens 450, got: %v", pm["estimatedTokens"])
			}

			// Should have section from frontmatter
			if pm["section"] != "reference" {
				t.Errorf("Expected section 'reference', got: %v", pm["section"])
			}
		}
	}

	if !foundEncoding {
		t.Error("Document 'blind-records-encoding' not found in manifest")
	}

	// Verify documents include frontmatter data
	for _, doc := range docs {
		if strings.Contains(doc.Path, "record_encoding.md") {
			if doc.Frontmatter.Description == "" {
				t.Error("Document should have description from frontmatter")
			}
			if len(doc.Frontmatter.Tags) == 0 {
				t.Error("Document should have tags from frontmatter")
			}
		}
	}
}
