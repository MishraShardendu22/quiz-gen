package worker

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/google/uuid"
)

// CreateSession inserts a new session with pending status into the database
func CreateSession(db *sql.DB, topicID uuid.UUID) (*model.Session, error) {
	now := time.Now().Unix()
	sessionID := uuid.Must(uuid.NewV7()).String()

	_, err := db.Exec(`
		INSERT INTO sessions (id, topic_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, topicID.String(), model.SessionPending, now, now)

	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	return &model.Session{
		ID:        uuid.MustParse(sessionID),
		TopicID:   topicID,
		Status:    model.SessionPending,
		CreatedAt: now,
		UpdatedAt: now,
		Error:     nil,
	}, nil
}

// GetSession retrieves a session by ID from the database
func GetSession(db *sql.DB, sessionID uuid.UUID) (*model.Session, error) {
	var id, topicID, status string
	var createdAt, updatedAt int64
	var errMsg *string

	err := db.QueryRow(`
		SELECT id, topic_id, status, created_at, updated_at, error
		FROM sessions
		WHERE id = ?
	`, sessionID.String()).Scan(&id, &topicID, &status, &createdAt, &updatedAt, &errMsg)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}

	return &model.Session{
		ID:        uuid.MustParse(id),
		TopicID:   uuid.MustParse(topicID),
		Status:    model.SessionStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Error:     errMsg,
	}, nil
}

// UpdateSessionStatus updates only the status and updated_at timestamp
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

// UpdateSessionError updates the status to failed and records the error message
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

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("session not found: %s", sessionID.String())
	}

	return nil
}

// GetTopic retrieves a topic by ID from the database
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
		ID:   uuid.MustParse(id),
		Name: name,
		Status: status,
	}

	if parentID != nil {
		parsedID := uuid.MustParse(*parentID)
		topic.ParentID = &parsedID
	}

	return topic, nil
}
