package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/judge"
	"github.com/MishraShardendu22/quiz-gen/service/llmretry"
	"github.com/MishraShardendu22/quiz-gen/service/openrouter"
	"github.com/MishraShardendu22/quiz-gen/service/prompt"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/google/uuid"
)

const MaxRegenerationAttempts = 5

// sessionQueue is a buffered channel (size 100) that holds session IDs waiting to be processed
var sessionQueue = make(chan uuid.UUID, 100)

// this is where worker starts
// begins the background worker goroutine that processes queued sessions
// processQueue runs in a goroutine and continuously processes queued sessions
func Start(db *sql.DB) {
	go func() {
		for sessionID := range sessionQueue {
			processSession(db, sessionID)
		}
	}()

	util.Info("session worker started")
}

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

	client := openrouter.GetClient()

	// number of questions = requested count
	// generation Loop until requested_count is reached or failure occurs
	for session.GeneratedCount < session.RequestedCount {
		// check budget before every OpenRouter request
		latestSession, err := GetSession(db, sessionID)
		if err == nil {
			session = latestSession
		}

		remainingBudget := session.TokenBudget - session.TokensUsed
		if remainingBudget <= 0 {
			failMsg := "Token budget exhausted."
			if session.GeneratedCount > 0 {
				failMsg = fmt.Sprintf("Token budget exhausted after generating %d of %d questions.", session.GeneratedCount, session.RequestedCount)
			}
			util.Warn("budget exhaustion", "session_id", sessionID.String(), "token_budget", session.TokenBudget, "tokens_used", session.TokensUsed, "generated_count", session.GeneratedCount)
			_ = UpdateSessionError(db, sessionID, failMsg)
			return
		}

		neededCount := session.RequestedCount - session.GeneratedCount

		// 4. Fetch existing topic questions for prompt under lock (released before network call)
		var existingTopicQuestions []model.Question
		func() {
			unlock := LockTopic(session.TopicID)
			defer unlock()
			existingTopicQuestions, _ = GetTopicQuestions(db, session.TopicID)
		}()

		// 5. Build prompt with existing questions
		promptStr, err := prompt.BuildPromptWithExisting(ctx, db, session.TopicID, neededCount, existingTopicQuestions)
		if err != nil {
			util.Error("failed to build prompt", "session_id", sessionID.String(), "error", err.Error())
			_ = UpdateSessionError(db, sessionID, "Prompt building failed: "+err.Error())
			return
		}

		// 6. Generate candidate questions (no lock held during network call)
		candidateQuestions, usage, err := llmretry.GenerateWithRetry(ctx, client, sessionID.String(), promptStr)

		// usage is basically the token usage and cost of the LLM call, we store it in DB for reporting
		if usage != nil {
			_ = StoreUsage(db, sessionID, usage, openrouter.DefaultModel)
		}
		if err != nil {
			util.Error("generation failed for session", "session_id", sessionID.String(), "error", err.Error())
			_ = UpdateSessionError(db, sessionID, err.Error())
			return
		}

		// 7. Judge & Regeneration Loop (up to MaxRegenerationAttempts = 5)
		acceptedQuestions := make([]model.LLMQuestion, 0)
		currentBatch := candidateQuestions

		var topicExhausted bool

		// judge duplicates and regenerate only the duplicates up to MaxRegenerationAttempts
		for regenAttempt := 1; regenAttempt <= MaxRegenerationAttempts; regenAttempt++ {
			// fetch fresh list of existing questions under lock for judging
			var currentTopicQuestions []model.Question
			func() {
				unlock := LockTopic(session.TopicID)
				defer unlock()
				currentTopicQuestions, _ = GetTopicQuestions(db, session.TopicID)
			}()

			// add questions accepted so far in this session to comparison set
			for _, aq := range acceptedQuestions {
				currentTopicQuestions = append(currentTopicQuestions, model.Question{
					Question: aq.Question,
				})
			}

			// judge duplicates (no lock held during network call)
			dupIndices, judgeUsage, judgeErr := judge.JudgeDuplicates(ctx, client, currentTopicQuestions, currentBatch)
			if judgeUsage != nil {
				_ = StoreUsage(db, sessionID, judgeUsage, openrouter.DefaultModel)
			}

			dupCount := len(dupIndices)
			util.Info("judge result", "session_id", sessionID.String(), "attempt", regenAttempt, "duplicate_count", dupCount, "judge_err", judgeErr)

			// map of duplicate indices
			isDuplicate := make(map[int]bool)
			for _, idx := range dupIndices {
				isDuplicate[idx] = true
			}

			// filter out unique questions from current batch
			nextBatchDuplicatesNeeded := 0
			for idx, q := range currentBatch {
				if isDuplicate[idx] {
					nextBatchDuplicatesNeeded++
				} else {
					acceptedQuestions = append(acceptedQuestions, q)
				}
			}

			if nextBatchDuplicatesNeeded == 0 {
				// all questions in current batch are unique!
				break
			}

			// if duplicate questions exist and maximum regeneration attempts reached
			if regenAttempt == MaxRegenerationAttempts {
				topicExhausted = true
				break
			}

			// check budget before regenerating duplicates
			currSess, _ := GetSession(db, sessionID)
			if currSess != nil && (currSess.TokenBudget-currSess.TokensUsed) <= 0 {
				util.Warn("budget exhaustion during regeneration", "session_id", sessionID.String())
				break
			}

			// regenerate ONLY the duplicate questions count
			regenPrompt, err := prompt.BuildPromptWithExisting(ctx, db, session.TopicID, nextBatchDuplicatesNeeded, currentTopicQuestions)
			if err != nil {
				util.Error("failed to build regen prompt", "session_id", sessionID.String(), "error", err.Error())
				break
			}

			// regenerate questions
			regenBatch, regenUsage, regenErr := llmretry.GenerateWithRetry(ctx, client, sessionID.String(), regenPrompt)
			if regenUsage != nil {
				_ = StoreUsage(db, sessionID, regenUsage, openrouter.DefaultModel)
			}
			if regenErr != nil {
				util.Error("regeneration LLM failed", "session_id", sessionID.String(), "error", regenErr.Error())
				break
			}

			currentBatch = regenBatch
			util.Info("regeneration attempts", "session_id", sessionID.String(), "attempt", regenAttempt, "regenerated_count", len(currentBatch))
		}

		// save accepted unique questions to DB under TopicLock
		if len(acceptedQuestions) > 0 {
			var saveErr error
			func() {
				unlock := LockTopic(session.TopicID)
				defer unlock()
				saveErr = SaveQuestions(ctx, db, sessionID, acceptedQuestions)
			}()

			if saveErr != nil {
				util.Error("failed to store questions", "session_id", sessionID.String(), "error", saveErr.Error())
				_ = UpdateSessionError(db, sessionID, "Question storage failed: "+saveErr.Error())
				return
			}
		}

		// reload session status after save
		latestSession, _ = GetSession(db, sessionID)
		if latestSession != nil {
			session = latestSession
		}

		if topicExhausted {
			failMsg := "Unable to generate additional unique questions after multiple attempts."
			util.Warn("topic exhaustion", "session_id", sessionID.String(), "generated_count", session.GeneratedCount, "requested_count", session.RequestedCount)
			_ = UpdateSessionError(db, sessionID, failMsg)
			return
		}
	}

	// 8. mark session completed if all requested questions were generated
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
		return ErrQueueFull
	}
}

// ErrQueueFull is returned when trying to enqueue a session but the queue is full
var ErrQueueFull = errors.New("session queue is full")
