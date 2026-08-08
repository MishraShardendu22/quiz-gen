package worker

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/judge"
	"github.com/MishraShardendu22/quiz-gen/service/llmretry"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/MishraShardendu22/quiz-gen/service/prompt"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/google/uuid"
)

const MaxRegenerationAttempts = 1

// sessionQueue is a buffered channel (size 100) that holds session IDs waiting to be processed
var sessionQueue = make(chan uuid.UUID, 100)

// Start begins the background worker pool that processes queued sessions
func Start(db *sql.DB) {
	workerCount := util.Config.WorkerCount
	if workerCount <= 0 {
		workerCount = 1
	}

	for i := 0; i < workerCount; i++ {
		go func() {
			for sessionID := range sessionQueue {
				processSession(db, sessionID)
			}
		}()
	}

	util.Info("session worker pool started", "workers", workerCount)

	// Recover pending or interrupted processing sessions from database on boot
	recoverPendingSessions(db)
}

func recoverPendingSessions(db *sql.DB) {
	rows, err := db.Query(`
		SELECT id, status FROM sessions
		WHERE status IN ('pending', 'processing')
		ORDER BY created_at ASC
	`)
	if err != nil {
		util.Error("failed to query pending sessions for recovery", "error", err.Error())
		return
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var sidStr, statusStr string
		if err := rows.Scan(&sidStr, &statusStr); err != nil {
			continue
		}
		sid, err := uuid.Parse(sidStr)
		if err != nil {
			continue
		}

		if model.SessionStatus(statusStr) == model.SessionProcessing {
			_ = UpdateSessionStatus(db, sid, model.SessionPending)
		}

		if err := Enqueue(sid); err != nil {
			util.Warn("failed to enqueue recovered session", "session_id", sidStr, "error", err.Error())
		} else {
			count++
		}
	}
	if count > 0 {
		util.Info("recovered pending sessions on startup", "count", count)
	}
}


/*
processSession()
    │
    ├── Load session
    ├── Mark session PROCESSING
    ├── Verify topic exists
    ├── Acquire topic lock
    │
    └── while generated < requested
          │
          ├── Reload session from DB
          ├── Check token budget
          ├── Load existing topic questions
          ├── Build generation prompt
          ├── Call LLM
          ├── Record token usage
          │
          ├── Duplicate/judge loop
          │     ├── Call judge LLM
          │     ├── Identify duplicates
          │     ├── Keep unique questions
          │     └── Regenerate duplicates
          │
          ├── Save accepted questions
          ├── Reload session
          └── Check budget / exhaustion
    │
    └── Mark session COMPLETED
*/ 
// processSession handles a single session from start to finish
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

	// Acquire per-topic lock for the duration of session processing to prevent concurrent same-topic sessions from racing
	unlock := LockTopic(session.TopicID)
	defer unlock()

	client := openrouter.GetClient()

	// Generation Loop until requested_count is reached or failure occurs
	for session.GeneratedCount < session.RequestedCount {
		// BEFORE EVERY OpenRouter REQUEST: load latest session from DB
		latestSession, err := GetSession(db, sessionID)
		if err == nil {
			session = latestSession
		}

		// Check actual recorded usage against budget before issuing another LLM request
		if session.TokensUsed >= session.TokenBudget {
			failMsg := "Token budget exhausted before next LLM request."
			if session.GeneratedCount > 0 {
				failMsg = fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
			}
			util.Warn("budget exhausted before request", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed, "generated_count", session.GeneratedCount)
			_ = UpdateSessionError(db, sessionID, failMsg)
			return
		}

		neededCount := session.RequestedCount - session.GeneratedCount

		// Fetch existing topic questions for prompt under topic lock
		existingTopicQuestions, _ := GetTopicQuestions(db, session.TopicID)

		// build prompt with existing questions
		promptStr, err := prompt.BuildPromptWithExisting(ctx, db, session.TopicID, neededCount, existingTopicQuestions)
		if err != nil {
			util.Error("failed to build prompt", "session_id", sessionID.String(), "error", err.Error())
			_ = UpdateSessionError(db, sessionID, "Prompt building failed: "+err.Error())
			return
		}

		// issue generation LLM request
		candidateQuestions, usage, err := llmretry.GenerateWithRetry(ctx, client, sessionID.String(), promptStr)

		// after EVERY OpenRouter REQUEST: record actual usage and update session.tokens_used
		if usage != nil {
			_ = StoreUsage(db, sessionID, usage, client.GetModel())
		}

		// Reload session to inspect updated tokens_used
		latestSession, _ = GetSession(db, sessionID)
		if latestSession != nil {
			session = latestSession
		}

		// If generation failed, log and mark session error
		if err != nil {
			util.Error("generation failed for session", "session_id", sessionID.String(), "error", err.Error())
			_ = UpdateSessionError(db, sessionID, err.Error())
			return
		}

		// post-request budget check: if actual reported tokens_used exceeded budget
		if session.TokensUsed > session.TokenBudget {
			failMsg := fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
			util.Warn("budget exhausted post-request", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed, "generated_count", session.GeneratedCount)
			_ = UpdateSessionError(db, sessionID, failMsg)
			return
		}

		// judge & regeneration Loop (up to MaxRegenerationAttempts = 5)
		acceptedQuestions := make([]model.LLMQuestion, 0)
		currentBatch := candidateQuestions

		var topicExhausted bool

		for regenAttempt := 1; regenAttempt <= MaxRegenerationAttempts; regenAttempt++ {
			// fetch fresh list of existing questions under topic lock for judging
			currentTopicQuestions, _ := GetTopicQuestions(db, session.TopicID)

			for _, aq := range acceptedQuestions {
				currentTopicQuestions = append(currentTopicQuestions, model.Question{
					Question: aq.Question,
				})
			}

			// before every openRouter request (LLM Judge): check budget using actual recorded usage
			currSess, _ := GetSession(db, sessionID)
			if currSess != nil {
				session = currSess
			}

			if session.TokensUsed >= session.TokenBudget {
				failMsg := fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
				util.Warn("budget exhausted before judge request", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed)
				_ = UpdateSessionError(db, sessionID, failMsg)
				return
			}

			dupIndices, judgeUsage, judgeErr := judge.JudgeDuplicates(ctx, client, currentTopicQuestions, currentBatch)
			if judgeUsage != nil {
				_ = StoreUsage(db, sessionID, judgeUsage, client.GetModel())
			}

			// Reload session to check tokens_used post judge call
			currSess, _ = GetSession(db, sessionID)
			if currSess != nil {
				session = currSess
			}

			if session.TokensUsed > session.TokenBudget {
				failMsg := fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
				util.Warn("budget exhausted post-judge request", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed)
				_ = UpdateSessionError(db, sessionID, failMsg)
				return
			}

			if judgeErr != nil {
				util.Error("judge duplicate detection failed", "session_id", sessionID.String(), "error", judgeErr.Error())
				_ = UpdateSessionError(db, sessionID, "Duplicate judge failed: "+judgeErr.Error())
				return
			}

			dupCount := len(dupIndices)
			util.Info("judge result", "session_id", sessionID.String(), "attempt", regenAttempt, "duplicate_count", dupCount)

			isDuplicate := make(map[int]bool)
			for _, idx := range dupIndices {
				isDuplicate[idx] = true
			}

			nextBatchDuplicatesNeeded := 0
			for idx, q := range currentBatch {
				if isDuplicate[idx] {
					nextBatchDuplicatesNeeded++
				} else {
					acceptedQuestions = append(acceptedQuestions, q)
				}
			}

			if nextBatchDuplicatesNeeded == 0 {
				break
			}

			if regenAttempt == MaxRegenerationAttempts {
				topicExhausted = true
				break
			}

			// BEFORE EVERY OpenRouter REQUEST (Regeneration): check budget using actual recorded usage
			if session.TokensUsed >= session.TokenBudget {
				failMsg := fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
				util.Warn("budget exhausted before regen request", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed)
				_ = UpdateSessionError(db, sessionID, failMsg)
				return
			}

			regenPrompt, err := prompt.BuildPromptWithExisting(ctx, db, session.TopicID, nextBatchDuplicatesNeeded, currentTopicQuestions)
			if err != nil {
				util.Error("failed to build regen prompt", "session_id", sessionID.String(), "error", err.Error())
				break
			}

			regenBatch, regenUsage, regenErr := llmretry.GenerateWithRetry(ctx, client, sessionID.String(), regenPrompt)
			if regenUsage != nil {
				_ = StoreUsage(db, sessionID, regenUsage, client.GetModel())
			}

			// Reload session to check tokens_used post regen call
			currSess, _ = GetSession(db, sessionID)
			if currSess != nil {
				session = currSess
			}

			if session.TokensUsed > session.TokenBudget {
				failMsg := fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
				util.Warn("budget exhaustion post-regen request", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed)
				_ = UpdateSessionError(db, sessionID, failMsg)
				return
			}

			if regenErr != nil {
				util.Error("regeneration LLM failed", "session_id", sessionID.String(), "error", regenErr.Error())
				break
			}

			currentBatch = regenBatch
			util.Info("regeneration attempts", "session_id", sessionID.String(), "attempt", regenAttempt, "regenerated_count", len(currentBatch))
		}

		// QUESTION STORAGE - Trim accepted questions before storing so generated_count never exceeds requested_count
		if len(acceptedQuestions) > neededCount {
			acceptedQuestions = acceptedQuestions[:neededCount]
		}

		// Save accepted unique questions to DB under TopicLock
		if len(acceptedQuestions) > 0 {
			saveErr := SaveQuestions(ctx, db, sessionID, acceptedQuestions)
			if saveErr != nil {
				util.Error("failed to store questions", "session_id", sessionID.String(), "error", saveErr.Error())
				_ = UpdateSessionError(db, sessionID, "Question storage failed: "+saveErr.Error())
				return
			}
		}

		// Reload session status after save to reflect updated tokens_used and generated_count
		latestSession, _ = GetSession(db, sessionID)
		if latestSession != nil {
			session = latestSession
		}

		// POST REQUEST VALIDATION - Compare actual tokens_used against token_budget
		if session.TokensUsed > session.TokenBudget {
			failMsg := fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
			util.Warn("budget exhaustion post-request", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed, "generated_count", session.GeneratedCount)
			_ = UpdateSessionError(db, sessionID, failMsg)
			return
		}

		if topicExhausted {
			failMsg := "Unable to generate additional unique questions after multiple attempts."
			util.Warn("topic exhaustion", "session_id", sessionID.String(), "generated_count", session.GeneratedCount, "requested_count", session.RequestedCount)
			_ = UpdateSessionError(db, sessionID, failMsg)
			return
		}
	}

	// Mark session completed if all requested questions were generated
	err = UpdateSessionStatus(db, sessionID, model.SessionCompleted)
	if err != nil {
		util.Error("failed to mark session as completed", "session_id", sessionID.String(), "error", err.Error())
		return
	}

	util.Info("session completed", "session_id", sessionID.String(), "topic_id", session.TopicID.String(), "generated_count", session.GeneratedCount)
}

// Enqueue adds a session ID to the processing queue
func Enqueue(sessionID uuid.UUID) error {
	select {
	case sessionQueue <- sessionID:
		util.Info("session queued", "session_id", sessionID.String())
		return nil
	default:
		return fmt.Errorf("session queue is full")
	}
}
