package loader

import "strings"

// split closest to 500 lines '\n'
// SplitClosestTo500Lines splits content approximately every 500 lines
// Never splits in the middle of a line. If a chunk exceeds 500 lines, starts a new chunk
func SplitClosestTo500Lines(content string) []string {
	lines := strings.Split(content, "\n")
	chunks := []string{}
	currentChunk := []string{}

	for _, line := range lines {
		currentChunk = append(currentChunk, line)

		// If we've reached or exceeded 500 lines, save the chunk and start a new one
		if len(currentChunk) >= 500 {
			chunks = append(chunks, strings.Join(currentChunk, "\n"))
			currentChunk = []string{}
		}
	}

	// Append the last chunk if it's not empty
	// 499 lines in last chunk so it wont be added above
	// this basically end ka chunk append ho sakt hai naa hua ho
	if len(currentChunk) > 0 {
		chunk := strings.Join(currentChunk, "\n")
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}
