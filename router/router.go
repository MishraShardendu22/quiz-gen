package router

import (
	"database/sql"

	"github.com/MishraShardendu22/quiz-gen/controller"
	"github.com/gofiber/fiber/v2"
)

// Setup registers all routes
func Setup(app *fiber.App, db *sql.DB) {
	app.Get("/topics", controller.GetTopics(db))
	app.Post("/generate", controller.Generate(db))
	app.Get("/sessions", controller.GetSessions(db))
	app.Get("/sessions/:id", controller.GetSession(db))
}
