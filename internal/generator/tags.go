package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// maxTopTags is the maximum number of top tags to include in the output.
const maxTopTags = 10

// TagsFile represents the aggregated tags output.
type TagsFile struct {
	Tags      map[string]*TagInfo `json:"tags"`
	TotalTags int                 `json:"totalTags"`
	TopTags   []string            `json:"topTags"`
}

// TagInfo contains information about a single tag.
type TagInfo struct {
	Count     int      `json:"count"`
	Documents []string `json:"documents"`
}

// GenerateTags creates an aggregated tags structure from documents.
func GenerateTags(docs []*Document) *TagsFile {
	tags := make(map[string]*TagInfo)

	// Collect all tags and their documents
	for _, doc := range docs {
		docID := titleToID(doc.Frontmatter.Title)
		for _, tag := range doc.Frontmatter.Tags {
			if tags[tag] == nil {
				tags[tag] = &TagInfo{
					Documents: []string{},
				}
			}
			tags[tag].Count++
			tags[tag].Documents = append(tags[tag].Documents, docID)
		}
	}

	// Sort documents within each tag
	for _, info := range tags {
		sort.Strings(info.Documents)
	}

	// Get top tags (sorted by count, then alphabetically)
	topTags := getTopTags(tags, maxTopTags)

	return &TagsFile{
		Tags:      tags,
		TotalTags: len(tags),
		TopTags:   topTags,
	}
}

// getTopTags returns the top N tags sorted by count (descending), then alphabetically.
func getTopTags(tags map[string]*TagInfo, n int) []string {
	type tagCount struct {
		tag   string
		count int
	}

	// Collect and sort
	counts := make([]tagCount, 0, len(tags))
	for tag, info := range tags {
		counts = append(counts, tagCount{tag: tag, count: info.Count})
	}

	sort.Slice(counts, func(i, j int) bool {
		if counts[i].count != counts[j].count {
			return counts[i].count > counts[j].count
		}
		return counts[i].tag < counts[j].tag
	})

	// Take top N
	result := make([]string, 0, n)
	for i := 0; i < len(counts) && i < n; i++ {
		result = append(result, counts[i].tag)
	}

	return result
}

// WriteTags writes the tags file to disk.
func WriteTags(path string, tags *TagsFile, dryRun bool) error {
	if dryRun {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
