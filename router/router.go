package router

import (
	"database/sql"

	"github.com/MishraShardendu22/quiz-gen/controller"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

/*
	Setup registers all routes and middlewares

	Allow all origins and headers for CORS - "*"
	Allow GET, POST, OPTIONS methods for CORS

	AllowHeaders - Idempotency-Key is included to support idempotent header.
	(Rest will be explained in their files and services)

	Simple Routes for the endpoints
*/

func Setup(app *fiber.App, db *sql.DB) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Idempotency-Key",
		AllowMethods: "GET, POST, OPTIONS",
	}))

	app.Get("/topics", controller.GetTopics(db))
	app.Post("/generate", controller.Generate(db))
	app.Get("/sessions", controller.GetSessions(db))
	app.Get("/sessions/:id", controller.GetSession(db))
	app.Post("/sessions/:id/retry", controller.RetrySession(db))
	app.Get("/usage", controller.GetUsage(db))
}
