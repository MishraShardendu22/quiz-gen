package worker

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/google/uuid"
)

// StoreUsage stores usage information for a session
func StoreUsage(db *sql.DB, sessionID uuid.UUID, usage *openrouter.Usage, model string) error {
	if usage == nil {
		return nil
	}

	now := time.Now().Unix()
	usageID := uuid.Must(uuid.NewV7()).String()
	estimatedCost := openrouter.EstimateCost(model, usage.PromptTokens, usage.CompletionTokens)

	_, err := db.Exec(`
		INSERT INTO usage (id, session_id, prompt_tokens, completion_tokens, total_tokens, estimated_cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, usageID, sessionID.String(), usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, estimatedCost, now)

	if err != nil {
		return fmt.Errorf("insert usage: %w", err)
	}

	// Update session tokens_used
	_, err = db.Exec(`
		UPDATE sessions
		SET tokens_used = tokens_used + ?, updated_at = ?
		WHERE id = ?
	`, usage.TotalTokens, now, sessionID.String())

	if err != nil {
		return fmt.Errorf("update tokens_used: %w", err)
	}

	return nil
}

// CheckTokenBudget checks if a session has enough tokens remaining
// Returns remaining tokens or error if budget exhausted
func CheckTokenBudget(db *sql.DB, sessionID uuid.UUID) (int, error) {
	session, err := GetSession(db, sessionID)
	if err != nil {
		return 0, err
	}

	remaining := session.TokenBudget - session.TokensUsed

	if remaining <= 0 {
		return 0, fmt.Errorf("token budget exhausted: budget=%d, used=%d", session.TokenBudget, session.TokensUsed)
	}

	return remaining, nil
}

// EnforceBudgetAfterGeneration checks if the session exceeded its budget after generation
// Returns error if budget exceeded
func EnforceBudgetAfterGeneration(db *sql.DB, sessionID uuid.UUID) error {
	session, err := GetSession(db, sessionID)
	if err != nil {
		return err
	}

	if session.TokensUsed > session.TokenBudget {
		return fmt.Errorf("token budget exceeded: budget=%d, used=%d", session.TokenBudget, session.TokensUsed)
	}

	return nil
}

// GetTotalUsage retrieves aggregated usage statistics
func GetTotalUsage(db *sql.DB) (totalPrompt, totalCompletion, totalTokens int, totalCost float64, err error) {
	err = db.QueryRow(`
		SELECT 
			COALESCE(SUM(prompt_tokens), 0),
			COALESCE(SUM(completion_tokens), 0),
			COALESCE(SUM(total_tokens), 0),
			COALESCE(SUM(estimated_cost), 0.0)
		FROM usage
	`).Scan(&totalPrompt, &totalCompletion, &totalTokens, &totalCost)

	if err != nil {
		return 0, 0, 0, 0.0, fmt.Errorf("query total usage: %w", err)
	}

	return totalPrompt, totalCompletion, totalTokens, totalCost, nil
}

// GetSessionUsage retrieves usage information for a specific session
func GetSessionUsage(db *sql.DB, sessionID uuid.UUID) (*model.SessionUsage, error) {
	var usage model.SessionUsage
	usage.SessionID = sessionID.String()

	err := db.QueryRow(`
		SELECT 
			COALESCE(SUM(prompt_tokens), 0),
			COALESCE(SUM(completion_tokens), 0),
			COALESCE(SUM(total_tokens), 0),
			COALESCE(SUM(estimated_cost), 0.0)
		FROM usage
		WHERE session_id = ?
	`, sessionID.String()).Scan(&usage.PromptTokens, &usage.CompletionTokens, &usage.TotalTokens, &usage.EstimatedCost)

	if err != nil {
		return nil, fmt.Errorf("query session usage: %w", err)
	}

	return &usage, nil
}
