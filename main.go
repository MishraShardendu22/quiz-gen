package main

import (
	"github.com/MishraShardendu22/quiz-gen/router"
	"github.com/MishraShardendu22/quiz-gen/service/db"
	"github.com/MishraShardendu22/quiz-gen/service/loader"
	"github.com/MishraShardendu22/quiz-gen/service/storage"
	"github.com/MishraShardendu22/quiz-gen/service/worker"
	"github.com/MishraShardendu22/quiz-gen/util"
	"github.com/gofiber/fiber/v2"
)

func main() {
	util.Info("starting application", "model", util.Config.ModelName)

	// Step - 1: Initialize the database
	sqlDB, err := db.Init(util.Config.DatabasePath)
	if err != nil {
		util.Error("database initialization failed", "error", err.Error())
		panic(err)
	}

	// close the database connection when the application exits
	defer sqlDB.Close()

	// run migrations
	if err := db.Migrate(sqlDB); err != nil {
		util.Error("migration failed", "error", err.Error())
		panic(err)
	}

	// Step - 2: Load content from filesystem
	loadedTopics, err := loader.LoadContent(util.Config.ContentPackDir)
	if err != nil {
		util.Error("content discovery failed", "error", err.Error())
		panic(err)
	}

	// Step - 3: Ingest data into database
	if err := storage.SyncTopicsDocumentsChunks(sqlDB, loadedTopics); err != nil {
		util.Error("synchronization failed", "error", err.Error())
		panic(err)
	}

	// Step - 4: Start HTTP server with timeouts configured from util.Config
	app := fiber.New(fiber.Config{
		ReadTimeout:  util.Config.ReadTimeout,
		WriteTimeout: util.Config.WriteTimeout,
		IdleTimeout:  util.Config.IdleTimeout,
	})
	router.Setup(app, sqlDB)

	// Step - 5: Start background worker for session processing
	worker.Start(sqlDB)

	util.Info("starting http server", "address", util.Config.ServerPort)
	if err := app.Listen(util.Config.ServerPort); err != nil {
		util.Error("server failed", "error", err.Error())
		panic(err)
	}
}
