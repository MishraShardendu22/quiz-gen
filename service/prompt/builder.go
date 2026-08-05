package prompt

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// BuildPrompt loads chunks from SQLite for the given topic and constructs an LLM prompt
func BuildPrompt(ctx context.Context, db *sql.DB, topicID uuid.UUID, requestedCount int) (string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT content
		FROM chunks
		WHERE topic_id = ?
		ORDER BY chunk_index ASC
	`, topicID.String())
	if err != nil {
		return "", fmt.Errorf("query chunks for topic %s: %w", topicID.String(), err)
	}
	defer rows.Close()

	var chunks []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return "", fmt.Errorf("scan chunk content: %w", err)
		}
		chunks = append(chunks, content)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate chunk rows: %w", err)
	}

	if len(chunks) == 0 {
		return "", fmt.Errorf("no chunks found for topic: %s", topicID.String())
	}

	combinedContent := strings.Join(chunks, "\n\n---\n\n")

	prompt := fmt.Sprintf("You are a quiz question generator. Generate exactly %d multiple-choice questions based ONLY on the provided context below.\n\nContext:\n%s\n\nRequirements for response:\nReturn ONLY JSON.\nNo markdown.\nNo ```json fences.\nNo explanations outside JSON.\n\nThe response must be a JSON object with the following schema:\n{\n  \"questions\": [\n    {\n      \"question\": \"question text\",\n      \"options\": [\"option 1\", \"option 2\", \"option 3\", \"option 4\"],\n      \"correct_answer\": 0,\n      \"explanation\": \"explanation text\"\n    }\n  ]\n}", requestedCount, combinedContent)

	return prompt, nil
}
