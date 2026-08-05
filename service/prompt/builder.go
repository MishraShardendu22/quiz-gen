package prompt

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/google/uuid"
)

// BuildPrompt loads chunks from SQLite for the given topic and constructs an LLM prompt
func BuildPrompt(ctx context.Context, db *sql.DB, topicID uuid.UUID, requestedCount int) (string, error) {
	return BuildPromptWithExisting(ctx, db, topicID, requestedCount, nil)
}

// BuildPromptWithExisting constructs prompt including existing questions to avoid duplicates
func BuildPromptWithExisting(ctx context.Context, db *sql.DB, topicID uuid.UUID, requestedCount int, existingQuestions []model.Question) (string, error) {
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

	var existingStr string
	if len(existingQuestions) > 0 {
		var qList strings.Builder
		for i, q := range existingQuestions {
			qList.WriteString(fmt.Sprintf("%d. %s\n", i+1, q.Question))
		}
		existingStr = fmt.Sprintf(AvoidDuplicatesInstruction, qList.String())
	} else {
		existingStr = DefaultAvoidInstruction
	}

	prompt := fmt.Sprintf(GeneratorPromptTemplate, requestedCount, combinedContent, existingStr)

	return prompt, nil
}
