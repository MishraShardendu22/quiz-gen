package controller

import (
	"database/sql"
	"strconv"

	"github.com/MishraShardendu22/quiz-gen/service/worker"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GenerateRequest represents a quiz generation request
type GenerateRequest struct {
	TopicID        string `json:"topic_id"`
	RequestedCount int    `json:"requested_count"`
	TokenBudget    int    `json:"token_budget"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// GenerateResponse represents the response to a generation request
type GenerateResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

// Generate handles POST /generate requests with Idempotency Key support
func Generate(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req GenerateRequest
		if err := c.BodyParser(&req); err != nil {
			return util.ErrorResponse(c, 400, "Invalid request body", err)
		}

		idempotencyKey := c.Get("Idempotency-Key")
		if idempotencyKey == "" {
			idempotencyKey = req.IdempotencyKey
		}

		// Check idempotency key if provided
		if idempotencyKey != "" {
			existingSessionIDStr, err := worker.GetIdempotencyKey(db, idempotencyKey)
			if err == nil && existingSessionIDStr != "" {
				existingID, parseErr := uuid.Parse(existingSessionIDStr)
				if parseErr == nil {
					existingSession, fetchErr := worker.GetSession(db, existingID)
					if fetchErr == nil {
						util.Info("idempotency hits", "session_id", existingSession.ID.String(), "idempotency_key", idempotencyKey)
						return util.JSONResponse(c, 200, "Idempotent request matched existing session", GenerateResponse{
							SessionID: existingSession.ID.String(),
							Status:    string(existingSession.Status),
						})
					}
				}
			}
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

		// Create session with pending status
		session, err := worker.CreateSession(db, topicID, req.RequestedCount, req.TokenBudget)
		if err != nil {
			util.Error("failed to create session", "error", err.Error())
			return util.ErrorResponse(c, 500, "Failed to create session", err)
		}

		// Store idempotency key mapping if provided
		if idempotencyKey != "" {
			_ = worker.CreateIdempotencyKey(db, idempotencyKey, session.ID)
		}

		// Enqueue session for processing
		err = worker.Enqueue(session.ID)
		if err != nil {
			util.Error("failed to enqueue session", "error", err.Error())
			return util.ErrorResponse(c, 503, "Generation queue is full", err)
		}

		util.Info("generation request received", "session_id", session.ID.String(), "topic_id", topicID.String(), "requested_count", req.RequestedCount)

		// Return 202 Accepted immediately
		response := GenerateResponse{
			SessionID: session.ID.String(),
			Status:    string(session.Status),
		}

		return util.JSONResponse(c, 202, "Generation started", response)
	}
}
