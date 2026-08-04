package worker

import (
	"database/sql"
	"time"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/google/uuid"
)

// sessionQueue is a buffered channel that holds session IDs waiting to be processed
var sessionQueue = make(chan uuid.UUID, 100)

// Start begins the background worker goroutine that processes queued sessions
// The worker runs for the lifetime of the application
func Start(db *sql.DB) {
	go processQueue(db)
	util.Info("session worker started")
}

// Enqueue adds a session ID to the processing queue
// Returns an error if the queue is full (back pressure)
func Enqueue(sessionID uuid.UUID) error {
	select {
	case sessionQueue <- sessionID:
		util.Info("session queued", "session_id", sessionID.String())
		return nil
	default:
		return ErrQueueFull
	}
}

// processQueue runs in a goroutine and continuously processes queued sessions
func processQueue(db *sql.DB) {
	for sessionID := range sessionQueue {
		processSession(db, sessionID)
	}
}

// processSession handles a single session from start to finish
func processSession(db *sql.DB, sessionID uuid.UUID) {
	util.Info("processing session", "session_id", sessionID.String())

	// Load session from database
	session, err := GetSession(db, sessionID)
	if err != nil {
		util.Error("failed to load session", "session_id", sessionID.String(), "error", err.Error())
		return
	}

	// Update status to processing
	err = UpdateSessionStatus(db, sessionID, model.SessionProcessing)
	if err != nil {
		util.Error("failed to mark session as processing", "session_id", sessionID.String(), "error", err.Error())
		return
	}

	util.Info("session processing started", "session_id", sessionID.String(), "topic_id", session.TopicID.String())

	// Load topic to verify it exists
	topic, err := GetTopic(db, session.TopicID)
	if err != nil {
		util.Error("failed to load topic for session", "session_id", sessionID.String(), "topic_id", session.TopicID.String(), "error", err.Error())
		_ = UpdateSessionError(db, sessionID, "Topic not found")
		return
	}

	// Execute generation (stubbed - simulates work)
	err = executeGeneration(topic)
	if err != nil {
		util.Error("generation failed for session", "session_id", sessionID.String(), "topic_id", session.TopicID.String(), "error", err.Error())
		_ = UpdateSessionError(db, sessionID, err.Error())
		return
	}

	// Mark session as completed
	err = UpdateSessionStatus(db, sessionID, model.SessionCompleted)
	if err != nil {
		util.Error("failed to mark session as completed", "session_id", sessionID.String(), "error", err.Error())
		return
	}

	util.Info("session completed", "session_id", sessionID.String(), "topic_id", session.TopicID.String())
}

// executeGeneration simulates quiz generation work
// In the next milestone, this will be replaced with actual OpenRouter integration
func executeGeneration(topic *model.Topic) error {
	// Simulate generation work with a sleep
	time.Sleep(1 * time.Second)
	return nil
}
