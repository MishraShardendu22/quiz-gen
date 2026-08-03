package loader

import "strings"

// SplitByHeading splits content by Markdown "##" headings
// Each chunk includes the heading that precedes it
func SplitByHeading(content string) []string {
	lines := strings.Split(content, "\n")
	chunks := []string{}
	currentChunk := []string{}

	for _, line := range lines {
		// Check if this line is a "##" heading
		if strings.HasPrefix(strings.TrimSpace(line), "##") && len(currentChunk) > 0 {
			// Save the current chunk and start a new one
			chunks = append(chunks, strings.Join(currentChunk, "\n"))
			currentChunk = []string{line}
		} else {
			currentChunk = append(currentChunk, line)
		}
	}

	// Append the last chunk if it's not empty
	// same logic as before append last chunk if it has content
	if len(currentChunk) > 0 {
		chunk := strings.Join(currentChunk, "\n")
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}
