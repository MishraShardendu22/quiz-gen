package worker

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/google/uuid"
)

// StoreUsage stores usage information for a session and updates tokens_used
func StoreUsage(db *sql.DB, sessionID uuid.UUID, usage *openrouter.Usage, modelName string) error {
	if usage == nil {
		return nil
	}

	now := time.Now().Unix()
	usageID := uuid.Must(uuid.NewV7()).String()
	estimatedCost := openrouter.EstimateCost(modelName, usage.PromptTokens, usage.CompletionTokens)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin usage tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO usage (id, session_id, prompt_tokens, completion_tokens, total_tokens, estimated_cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, usageID, sessionID.String(), usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, estimatedCost, now)
	if err != nil {
		return fmt.Errorf("insert usage: %w", err)
	}

	// Update session tokens_used
	_, err = tx.Exec(`
		UPDATE sessions
		SET tokens_used = tokens_used + ?, updated_at = ?
		WHERE id = ?
	`, usage.TotalTokens, now, sessionID.String())
	if err != nil {
		return fmt.Errorf("update tokens_used: %w", err)
	}

	return tx.Commit()
}

// GetUsageReport fetches aggregated usage and per-session breakdown
func GetUsageReport(db *sql.DB) (*model.UsageResponse, error) {
	var resp model.UsageResponse

	err := db.QueryRow(`
		SELECT 
			COALESCE(SUM(prompt_tokens), 0),
			COALESCE(SUM(completion_tokens), 0),
			COALESCE(SUM(total_tokens), 0),
			COALESCE(SUM(estimated_cost), 0.0)
		FROM usage
	`).Scan(&resp.TotalPromptTokens, &resp.TotalCompletionTokens, &resp.TotalTokens, &resp.EstimatedCost)
	if err != nil {
		return nil, fmt.Errorf("query total usage: %w", err)
	}

	rows, err := db.Query(`
		SELECT 
			session_id,
			COALESCE(SUM(prompt_tokens), 0),
			COALESCE(SUM(completion_tokens), 0),
			COALESCE(SUM(total_tokens), 0),
			COALESCE(SUM(estimated_cost), 0.0)
		FROM usage
		GROUP BY session_id
		ORDER BY session_id
	`)
	if err != nil {
		return nil, fmt.Errorf("query per-session usage: %w", err)
	}
	defer rows.Close()

	resp.Sessions = make([]model.UsageBreakdown, 0)
	for rows.Next() {
		var item model.UsageBreakdown
		if err := rows.Scan(&item.SessionID, &item.PromptTokens, &item.CompletionTokens, &item.TotalTokens, &item.EstimatedCost); err != nil {
			return nil, fmt.Errorf("scan usage breakdown: %w", err)
		}
		resp.Sessions = append(resp.Sessions, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("usage rows error: %w", err)
	}

	return &resp, nil
}
