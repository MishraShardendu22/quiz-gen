package worker

import (
	"context"
	"database/sql"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/llmretry"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/MishraShardendu22/quiz-gen/service/prompt"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/google/uuid"
)

// sessionQueue is a buffered channel that holds session IDs waiting to be processed
var sessionQueue = make(chan uuid.UUID, 100)

// Start begins the background worker goroutine that processes queued sessions
func Start(db *sql.DB) {
	go processQueue(db)
	util.Info("session worker started")
}

// Enqueue adds a session ID to the processing queue
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
// Pipeline: Load Session -> Load Topic -> Load Chunks & Build Prompt -> Generate -> Clean -> Parse -> Validate -> Store Questions & Update generated_count -> Completed (or Session Failed on error)
func processSession(db *sql.DB, sessionID uuid.UUID) {
	ctx := context.Background()
	util.Info("processing session", "session_id", sessionID.String())

	// 1. Load Session
	session, err := GetSession(db, sessionID)
	if err != nil {
		util.Error("failed to load session", "session_id", sessionID.String(), "error", err.Error())
		return
	}

	// 2. Update status to processing
	err = UpdateSessionStatus(db, sessionID, model.SessionProcessing)
	if err != nil {
		util.Error("failed to mark session as processing", "session_id", sessionID.String(), "error", err.Error())
		return
	}

	util.Info("session processing started", "session_id", sessionID.String(), "topic_id", session.TopicID.String())

	// 3. Load Topic to verify existence
	_, err = GetTopic(db, session.TopicID)
	if err != nil {
		util.Error("failed to load topic for session", "session_id", sessionID.String(), "topic_id", session.TopicID.String(), "error", err.Error())
		_ = UpdateSessionError(db, sessionID, "Topic not found: "+err.Error())
		return
	}

	// 4. Load Chunks & Build Prompt
	promptStr, err := prompt.BuildPrompt(ctx, db, session.TopicID, session.RequestedCount)
	if err != nil {
		util.Error("failed to build prompt for session", "session_id", sessionID.String(), "error", err.Error())
		_ = UpdateSessionError(db, sessionID, "Prompt building failed: "+err.Error())
		return
	}

	// 5. Generate, Clean JSON, Parse, Validate (with LLM retries)
	client := openrouter.GetClient()
	questions, _, err := llmretry.GenerateWithRetry(ctx, client, sessionID.String(), promptStr)
	if err != nil {
		util.Error("generation failed for session", "session_id", sessionID.String(), "error", err.Error())
		_ = UpdateSessionError(db, sessionID, err.Error())
		return
	}

	// 6. Store Questions & Update generated_count in ONE SQL transaction
	err = SaveQuestions(ctx, db, sessionID, questions)
	if err != nil {
		util.Error("failed to store questions for session", "session_id", sessionID.String(), "error", err.Error())
		_ = UpdateSessionError(db, sessionID, "Question storage failed: "+err.Error())
		return
	}

	// 7. Mark session as completed
	err = UpdateSessionStatus(db, sessionID, model.SessionCompleted)
	if err != nil {
		util.Error("failed to mark session as completed", "session_id", sessionID.String(), "error", err.Error())
		return
	}

	util.Info("session completed", "session_id", sessionID.String(), "topic_id", session.TopicID.String(), "generated_count", len(questions))
}
