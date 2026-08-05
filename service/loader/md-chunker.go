package loader

import "strings"

// splitByHeading splits content by Markdown "##" headings.
// each chunk includes the "##" heading that precedes it.
// headings deeper than level 2 (e.g. ###, ####) remain part of the current chunk.
func SplitByHeading(content string) []string {
	lines := strings.Split(content, "\n")
	chunks := []string{}
	currentChunk := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Split only on level-2 headings ("## "), not deeper headings like "###".
		isH1 := strings.HasPrefix(trimmed, "# ") &&
			(len(trimmed) == 2 || trimmed[2] != '#')

		isH2 := strings.HasPrefix(trimmed, "## ") &&
			(len(trimmed) == 3 || trimmed[3] != '#')

		if (isH1 || isH2) && len(currentChunk) > 0 {
			chunks = append(chunks, strings.Join(currentChunk, "\n"))
			currentChunk = []string{line}
		} else {
			currentChunk = append(currentChunk, line)
		}
	}

	// Append the last chunk if it's not empty.
	if len(currentChunk) > 0 {
		chunk := strings.Join(currentChunk, "\n")
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// it considered ### to be a seperate chunk
// func SplitByHeading(content string) []string {
// 	lines := strings.Split(content, "\n")
// 	chunks := []string{}
// 	currentChunk := []string{}

// 	for _, line := range lines {
// 		// Check if this line is a "##" heading
// 		if strings.HasPrefix(strings.TrimSpace(line), "##") && len(currentChunk) > 0 {
// 			// Save the current chunk and start a new one
// 			chunks = append(chunks, strings.Join(currentChunk, "\n"))
// 			currentChunk = []string{line}
// 		} else {
// 			currentChunk = append(currentChunk, line)
// 		}
// 	}

// 	// Append the last chunk if it's not empty
// 	// same logic as before append last chunk if it has content
// 	if len(currentChunk) > 0 {
// 		chunk := strings.Join(currentChunk, "\n")
// 		if strings.TrimSpace(chunk) != "" {
// 			chunks = append(chunks, chunk)
// 		}
// 	}

// 	return chunks
// }
