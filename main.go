package main

import (
	"fmt"
	"log"

	"github.com/MishraShardendu22/quiz-gen/router"
	"github.com/MishraShardendu22/quiz-gen/service/db"
	"github.com/MishraShardendu22/quiz-gen/service/loader"
	"github.com/MishraShardendu22/quiz-gen/service/storage"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize database
	sqlDB, err := db.Init("./quiz.db")
	if err != nil {
		log.Fatalf("Database init failed: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations
	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Load content from filesystem
	loadedTopics, err := loader.LoadContent("./content-pack")
	if err != nil {
		log.Fatalf("Load content failed: %v", err)
	}

	// Count total documents for logging
	totalDocs := 0
	for _, lt := range loadedTopics {
		totalDocs += len(lt.Documents)
	}

	// Ingest data into database
	if err := storage.SyncTopicsDocumentsChunks(sqlDB, loadedTopics); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}

	fmt.Printf("Ingestion complete: %d topics, %d documents\n", len(loadedTopics), totalDocs)

	// Start HTTP server
	app := fiber.New()
	router.Setup(app, sqlDB)

	fmt.Println("Server starting on :9000")
	if err := app.Listen(":9000"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
