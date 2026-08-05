package openrouter

import (
	"regexp"
	"strings"
)

// CleanJSON removes markdown code fences from a JSON response
func CleanJSON(raw string) string {
	raw = strings.TrimSpace(raw)

	// Remove markdown code fences (```json ... ``` or just ``` ... ```)
	re := regexp.MustCompile("(?s)^```(?:json)?\\s*(.*?)\\s*```$")
	if matches := re.FindStringSubmatch(raw); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Fallback: If JSON is embedded inside backticks with surrounding text
	if idx := strings.Index(raw, "```"); idx != -1 {
		endIdx := strings.LastIndex(raw, "```")
		if endIdx > idx {
			content := raw[idx:endIdx]
			if newlineIdx := strings.Index(content, "\n"); newlineIdx != -1 {
				content = content[newlineIdx+1:]
			}
			return strings.TrimSpace(content)
		}
	}

	return raw
}
