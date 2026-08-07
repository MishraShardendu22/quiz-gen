package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/google/uuid"
)

// MaxRequestedCount is the maximum number of questions that can be requested per session
const MaxRequestedCount = 100

// CreateSession inserts a new session with pending status into the database
func CreateSession(db *sql.DB, topicID uuid.UUID, requestedCount, tokenBudget int) (*model.Session, error) {
	session, _, err := CreateSessionWithIdempotencyKey(db, topicID, requestedCount, tokenBudget, "")
	return session, err
}

// CreateSessionWithIdempotencyKey atomically checks idempotency key and creates a session inside a transaction
func CreateSessionWithIdempotencyKey(db *sql.DB, topicID uuid.UUID, requestedCount, tokenBudget int, idempotencyKey string) (*model.Session, bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, false, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check idempotency key if supplied
	if idempotencyKey != "" {
		var existingSessionID string
		err := tx.QueryRow(`
			SELECT session_id FROM idempotency_keys WHERE idempotency_key = ?
		`, idempotencyKey).Scan(&existingSessionID)

		if err == nil && existingSessionID != "" {
			existingUUID, parseErr := uuid.Parse(existingSessionID)
			if parseErr == nil {
				existingSession, getErr := GetSession(db, existingUUID)
				if getErr == nil {
					_ = tx.Commit()
					return existingSession, true, nil
				}
			}
		}
	}

	now := time.Now().Unix()
	sessionID := uuid.Must(uuid.NewV7()).String()

	_, err = tx.Exec(`
		INSERT INTO sessions (id, topic_id, status, requested_count, token_budget, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sessionID, topicID.String(), model.SessionPending, requestedCount, tokenBudget, now, now)

	if err != nil {
		return nil, false, fmt.Errorf("insert session: %w", err)
	}

	if idempotencyKey != "" {
		_, err = tx.Exec(`
			INSERT INTO idempotency_keys (idempotency_key, session_id, created_at)
			VALUES (?, ?, ?)
		`, idempotencyKey, sessionID, now)
		if err != nil {
			return nil, false, fmt.Errorf("insert idempotency key: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit session transaction: %w", err)
	}

	newSession := &model.Session{
		ID:             uuid.MustParse(sessionID),
		TopicID:        topicID,
		Status:         model.SessionPending,
		RequestedCount: requestedCount,
		GeneratedCount: 0,
		TokenBudget:    tokenBudget,
		TokensUsed:     0,
		CreatedAt:      now,
		UpdatedAt:      now,
		Error:          nil,
	}

	return newSession, false, nil
}

// retrieves a session by ID from the database
func GetSession(db *sql.DB, sessionID uuid.UUID) (*model.Session, error) {
	var id, topicID, status string
	var requestedCount, generatedCount, tokenBudget, tokensUsed int
	var createdAt, updatedAt int64
	var errMsg *string

	err := db.QueryRow(`
		SELECT id, topic_id, status, requested_count, generated_count, token_budget, tokens_used, created_at, updated_at, error
		FROM sessions
		WHERE id = ?
	`, sessionID.String()).Scan(&id, &topicID, &status, &requestedCount, &generatedCount, &tokenBudget, &tokensUsed, &createdAt, &updatedAt, &errMsg)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}

	questions, _ := GetSessionQuestions(db, sessionID)

	return &model.Session{
		ID:             uuid.MustParse(id),
		TopicID:        uuid.MustParse(topicID),
		Status:         model.SessionStatus(status),
		RequestedCount: requestedCount,
		GeneratedCount: generatedCount,
		TokenBudget:    tokenBudget,
		TokensUsed:     tokensUsed,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		Error:          errMsg,
		Questions:      questions,
	}, nil
}

// retrieves all generated questions for a session
func GetSessionQuestions(db *sql.DB, sessionID uuid.UUID) ([]model.Question, error) {
	rows, err := db.Query(`
		SELECT id, session_id, question, option_1, option_2, option_3, option_4, correct_answer, explanation, created_at
		FROM questions
		WHERE session_id = ?
		ORDER BY created_at ASC
	`, sessionID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []model.Question
	for rows.Next() {
		var q model.Question
		var qID, sID string
		if err := rows.Scan(&qID, &sID, &q.Question, &q.Option1, &q.Option2, &q.Option3, &q.Option4, &q.CorrectAnswer, &q.Explanation, &q.CreatedAt); err != nil {
			return nil, err
		}
		q.ID = uuid.MustParse(qID)
		q.SessionID = uuid.MustParse(sID)
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

// retrieves all generated questions across all sessions for a specific topic
func GetTopicQuestions(db *sql.DB, topicID uuid.UUID) ([]model.Question, error) {
	rows, err := db.Query(`
		SELECT q.id, q.session_id, q.question, q.option_1, q.option_2, q.option_3, q.option_4, q.correct_answer, q.explanation, q.created_at
		FROM questions q
		JOIN sessions s ON q.session_id = s.id
		WHERE s.topic_id = ?
		ORDER BY q.created_at ASC
	`, topicID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []model.Question
	for rows.Next() {
		var q model.Question
		var qID, sID string
		if err := rows.Scan(&qID, &sID, &q.Question, &q.Option1, &q.Option2, &q.Option3, &q.Option4, &q.CorrectAnswer, &q.Explanation, &q.CreatedAt); err != nil {
			return nil, err
		}
		q.ID = uuid.MustParse(qID)
		q.SessionID = uuid.MustParse(sID)
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

// GetAllSessions retrieves all sessions ordered by newest first
func GetAllSessions(db *sql.DB) ([]*model.Session, error) {
	rows, err := db.Query(`
		SELECT id, topic_id, status, requested_count, generated_count, token_budget, tokens_used, created_at, updated_at, error
		FROM sessions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var id, topicID, status string
		var requestedCount, generatedCount, tokenBudget, tokensUsed int
		var createdAt, updatedAt int64
		var errMsg *string

		err := rows.Scan(&id, &topicID, &status, &requestedCount, &generatedCount, &tokenBudget, &tokensUsed, &createdAt, &updatedAt, &errMsg)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}

		sessions = append(sessions, &model.Session{
			ID:             uuid.MustParse(id),
			TopicID:        uuid.MustParse(topicID),
			Status:         model.SessionStatus(status),
			RequestedCount: requestedCount,
			GeneratedCount: generatedCount,
			TokenBudget:    tokenBudget,
			TokensUsed:     tokensUsed,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			Error:          errMsg,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return sessions, nil
}

// updates only the status and updated_at timestamp
func UpdateSessionStatus(db *sql.DB, sessionID uuid.UUID, newStatus model.SessionStatus) error {
	now := time.Now().Unix()
	result, err := db.Exec(`
		UPDATE sessions
		SET status = ?, updated_at = ?
		WHERE id = ?
	`, string(newStatus), now, sessionID.String())

	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("session not found: %s", sessionID.String())
	}

	return nil
}

// updates the status to failed and records the error message
func UpdateSessionError(db *sql.DB, sessionID uuid.UUID, errMsg string) error {
	now := time.Now().Unix()
	result, err := db.Exec(`
		UPDATE sessions
		SET status = ?, error = ?, updated_at = ?
		WHERE id = ?
	`, string(model.SessionFailed), errMsg, now, sessionID.String())

	if err != nil {
		return fmt.Errorf("update session error: %w", err)
	}

	// used to verify that the UPDATE actually modified a row matching the WHERE clause
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("session not found: %s", sessionID.String())
	}

	return nil
}

// retrieves a topic by ID from the database
func GetTopic(db *sql.DB, topicID uuid.UUID) (*model.Topic, error) {
	var id, name, status string
	var parentID *string

	err := db.QueryRow(`
		SELECT id, name, status, parent_id
		FROM topics
		WHERE id = ?
	`, topicID.String()).Scan(&id, &name, &status, &parentID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("topic not found: %s", topicID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("query topic: %w", err)
	}

	topic := &model.Topic{
		ID:     uuid.MustParse(id),
		Name:   name,
		Status: status,
	}

	if parentID != nil {
		parsedID := uuid.MustParse(*parentID)
		topic.ParentID = &parsedID
	}

	return topic, nil
}

// inserts all generated questions inside a single transaction and updates generated_count
func SaveQuestions(ctx context.Context, db *sql.DB, sessionID uuid.UUID, questions []model.LLMQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	
	// a prepared statement is a SQL statement that the database parses and compiles once
	// then can be executed multiple times with different parameters. 
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO questions (id, session_id, question, option_1, option_2, option_3, option_4, correct_answer, explanation, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert question statement: %w", err)
	}
	defer stmt.Close()

	// execute the prepared statement for each question
	for _, q := range questions {
		qID := uuid.Must(uuid.NewV7()).String()
		_, err := stmt.ExecContext(ctx,
			qID,
			sessionID.String(),
			q.Question,
			q.Options[0],
			q.Options[1],
			q.Options[2],
			q.Options[3],
			q.CorrectAnswer,
			q.Explanation,
			now,
		)
		if err != nil {
			return fmt.Errorf("insert question: %w", err)
		}
	}

	// update generated_count atomically in the same transaction
	_, err = tx.ExecContext(ctx, `
		UPDATE sessions
		SET generated_count = generated_count + ?, updated_at = ?
		WHERE id = ?
	`, len(questions), now, sessionID.String())
	if err != nil {
		return fmt.Errorf("update generated_count: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit questions transaction: %w", err)
	}

	return nil
}

// looks up an idempotency key and returns the mapped session_id if found
func GetIdempotencyKey(db *sql.DB, key string) (string, error) {
	var sessionID string
	err := db.QueryRow(`
		SELECT session_id FROM idempotency_keys WHERE idempotency_key = ?
	`, key).Scan(&sessionID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return sessionID, nil
}

// CreateIdempotencyKey stores an idempotency_key mapping
func CreateIdempotencyKey(db *sql.DB, key string, sessionID uuid.UUID) error {
	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO idempotency_keys (idempotency_key, session_id, created_at)
		VALUES (?, ?, ?)
	`, key, sessionID.String(), now)
	return err
}