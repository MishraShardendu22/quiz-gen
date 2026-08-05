package controller

import (
	"database/sql"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/worker"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetSession handles GET /sessions/:id requests
func GetSession(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("id")

		// Parse session ID as UUID
		sid, err := uuid.Parse(sessionID)
		if err != nil {
			return util.ErrorResponse(c, 400, "Invalid session_id format", err)
		}

		// Retrieve session from database
		session, err := worker.GetSession(db, sid)
		if err != nil {
			util.Error("failed to get session", "session_id", sessionID, "error", err.Error())
			return util.ErrorResponse(c, 404, "Session not found", err)
		}

		util.Info("session retrieved", "session_id", sessionID)

		return util.JSONResponse(c, 200, "Session fetched successfully", session)
	}
}

// GetSessions handles GET /sessions requests
func GetSessions(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve all sessions from database
		sessions, err := worker.GetAllSessions(db)
		if err != nil {
			util.Error("failed to get sessions", "error", err.Error())
			return util.ErrorResponse(c, 500, "Failed to fetch sessions", err)
		}

		util.Info("sessions retrieved", "count", len(sessions))

		return util.JSONResponse(c, 200, "Sessions fetched successfully", sessions)
	}
}

// RetrySession handles POST /sessions/:id/retry requests
func RetrySession(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("id")
		oldID, err := uuid.Parse(sessionID)
		if err != nil {
			return util.ErrorResponse(c, 400, "Invalid session_id format", err)
		}

		var req model.RetrySessionRequest
		if err := c.BodyParser(&req); err != nil {
			return util.ErrorResponse(c, 400, "Invalid request body", err)
		}

		if req.TokenBudget <= 0 {
			return util.ErrorResponse(c, 400, "token_budget must be greater than 0", nil)
		}

		// Retrieve old session
		oldSession, err := worker.GetSession(db, oldID)
		if err != nil {
			return util.ErrorResponse(c, 404, "Session not found", err)
		}

		// Ensure ONLY failed sessions can be retried
		if oldSession.Status != model.SessionFailed {
			return util.ErrorResponse(c, 400, "Only failed sessions can be retried", nil)
		}

		// Create a completely new session copying topic_id, requested_count with new token_budget
		newSession, err := worker.CreateSession(db, oldSession.TopicID, oldSession.RequestedCount, req.TokenBudget)
		if err != nil {
			util.Error("failed to create retry session", "old_session_id", sessionID, "error", err.Error())
			return util.ErrorResponse(c, 500, "Failed to create retry session", err)
		}

		util.Info("retry session creation", "old_session_id", oldSession.ID.String(), "new_session_id", newSession.ID.String(), "token_budget", req.TokenBudget)

		// Enqueue new session
		if err := worker.Enqueue(newSession.ID); err != nil {
			util.Error("failed to enqueue retry session", "new_session_id", newSession.ID.String(), "error", err.Error())
			_ = worker.UpdateSessionError(db, newSession.ID, "Queue full")
			return util.ErrorResponse(c, 503, "Generation queue is full", err)
		}

		return util.JSONResponse(c, 202, "Retry session created", model.GenerateResponse{
			SessionID: newSession.ID.String(),
			Status:    string(newSession.Status),
		})
	}
}
