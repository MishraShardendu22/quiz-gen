package main

import (
	"github.com/MishraShardendu22/quiz-gen/router"
	"github.com/MishraShardendu22/quiz-gen/service/db"
	"github.com/MishraShardendu22/quiz-gen/service/loader"
	"github.com/MishraShardendu22/quiz-gen/service/storage"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
)

func main() {
	util.Info("starting application")

	// Initialize database
	sqlDB, err := db.Init("./quiz.db")
	if err != nil {
		util.Error("database initialization failed", "error", err.Error())
		panic(err)
	}
	defer sqlDB.Close()

	// Run migrations
	if err := db.Migrate(sqlDB); err != nil {
		util.Error("migration failed", "error", err.Error())
		panic(err)
	}

	// Load content from filesystem
	loadedTopics, err := loader.LoadContent("./content-pack")
	if err != nil {
		util.Error("content discovery failed", "error", err.Error())
		panic(err)
	}

	// Ingest data into database
	if err := storage.SyncTopicsDocumentsChunks(sqlDB, loadedTopics); err != nil {
		util.Error("synchronization failed", "error", err.Error())
		panic(err)
	}

	// Start HTTP server
	app := fiber.New()
	router.Setup(app, sqlDB)

	util.Info("starting http server", "address", ":9000")
	if err := app.Listen(":9000"); err != nil {
		util.Error("server failed", "error", err.Error())
		panic(err)
	}
}
