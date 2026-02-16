package ai

// Test exports for internal functions.

// ParseMetaJSONForTest exposes parseMetaJSON for testing.
func ParseMetaJSONForTest(text string) (*GeneratedMeta, error) {
	return parseMetaJSON(text)
}

// ExtractJSONForTest exposes extractJSON for testing.
func ExtractJSONForTest(text string) string {
	return extractJSON(text)
}

// TruncateContentForTest exposes truncateContent for testing.
func TruncateContentForTest(content string, maxLen int) string {
	return truncateContent(content, maxLen)
}

// ParseProjectInfoJSONForTest exposes parseProjectInfoJSON for testing.
func ParseProjectInfoJSONForTest(text string) (*GeneratedProjectInfo, error) {
	return parseProjectInfoJSON(text)
}
