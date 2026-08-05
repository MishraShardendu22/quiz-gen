package controller

import (
	"database/sql"
	"strconv"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/worker"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Generate handles POST /generate requests with Idempotency Key support
func Generate(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.GenerateRequest
		if err := c.BodyParser(&req); err != nil {
			return util.ErrorResponse(c, 400, "Invalid request body", err)
		}

		idempotencyKey := c.Get("Idempotency-Key")
		if idempotencyKey == "" {
			idempotencyKey = req.IdempotencyKey
		}

		// Validate topic_id is provided
		if req.TopicID == "" {
			return util.ErrorResponse(c, 400, "topic_id is required", nil)
		}

		// Parse topic_id as UUID
		topicID, err := uuid.Parse(req.TopicID)
		if err != nil {
			return util.ErrorResponse(c, 400, "Invalid topic_id format", err)
		}

		// Validate requested_count
		if req.RequestedCount <= 0 {
			return util.ErrorResponse(c, 400, "requested_count must be greater than 0", nil)
		}

		if req.RequestedCount > worker.MaxRequestedCount {
			return util.ErrorResponse(c, 400, "requested_count exceeds maximum of "+strconv.Itoa(worker.MaxRequestedCount), nil)
		}

		// Validate token_budget
		if req.TokenBudget <= 0 {
			return util.ErrorResponse(c, 400, "token_budget must be greater than 0", nil)
		}

		// Verify topic exists in database
		var topicExists bool
		err = db.QueryRow("SELECT COUNT(*) > 0 FROM topics WHERE id = ?", topicID.String()).Scan(&topicExists)
		if err != nil {
			util.Error("failed to verify topic", "error", err.Error())
			return util.ErrorResponse(c, 500, "Database error", err)
		}

		if !topicExists {
			return util.ErrorResponse(c, 404, "Topic not found", nil)
		}

		// Atomically check idempotency key and create session inside a single transaction
		session, isExisting, err := worker.CreateSessionWithIdempotencyKey(db, topicID, req.RequestedCount, req.TokenBudget, idempotencyKey)
		if err != nil {
			util.Error("failed to create session", "error", err.Error())
			return util.ErrorResponse(c, 500, "Failed to create session", err)
		}

		if isExisting {
			util.Info("idempotency match found", "session_id", session.ID.String(), "idempotency_key", idempotencyKey)
			return util.JSONResponse(c, 200, "Idempotent request matched existing session", model.GenerateResponse{
				SessionID: session.ID.String(),
				Status:    string(session.Status),
			})
		}

		// Enqueue session for processing
		err = worker.Enqueue(session.ID)
		if err != nil {
			util.Error("failed to enqueue session", "session_id", session.ID.String(), "error", err.Error())
			// Mark session as failed so no orphan pending session remains in DB
			_ = worker.UpdateSessionError(db, session.ID, "Queue full")
			return util.ErrorResponse(c, 503, "Generation queue is full", err)
		}

		util.Info("generation request received", "session_id", session.ID.String(), "topic_id", topicID.String(), "requested_count", req.RequestedCount)

		// Return 202 Accepted immediately
		response := model.GenerateResponse{
			SessionID: session.ID.String(),
			Status:    string(session.Status),
		}

		return util.JSONResponse(c, 202, "Generation started", response)
	}
}
