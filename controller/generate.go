package controller

import (
	"database/sql"

	"github.com/MishraShardendu22/quiz-gen/service/worker"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GenerateRequest represents a quiz generation request
type GenerateRequest struct {
	TopicID string `json:"topic_id"`
}

// GenerateResponse represents the response to a generation request
type GenerateResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

// Generate handles POST /generate requests
// Validates the request, creates a session, enqueues it, and returns immediately
func Generate(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req GenerateRequest
		if err := c.BodyParser(&req); err != nil {
			return util.ErrorResponse(c, 400, "Invalid request body", err)
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
		session, err := worker.CreateSession(db, topicID)
		if err != nil {
			util.Error("failed to create session", "error", err.Error())
			return util.ErrorResponse(c, 500, "Failed to create session", err)
		}

		// Enqueue session for processing
		err = worker.Enqueue(session.ID)
		if err != nil {
			util.Error("failed to enqueue session", "error", err.Error())
			return util.ErrorResponse(c, 503, "Generation queue is full", err)
		}

		util.Info("generation request received", "session_id", session.ID.String(), "topic_id", topicID.String())

		// Return 202 Accepted immediately
		response := GenerateResponse{
			SessionID: session.ID.String(),
			Status:    string(session.Status),
		}

		return util.JSONResponse(c, 202, "Generation started", response)
	}
}
