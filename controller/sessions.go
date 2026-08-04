package controller

import (
	"database/sql"

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
