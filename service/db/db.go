package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MishraShardendu22/quiz-gen/util"
	_ "github.com/mattn/go-sqlite3"
)

// Init initializes the database with proper configuration
func Init(dbPath string) (*sql.DB, error) {
	util.Info("initializing database", "path", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	util.Info("database initialized successfully")
	return db, nil
}

// Migrate creates all necessary tables
func Migrate(db *sql.DB) error {
	util.Info("running database migrations")

	statements := []string{
		`CREATE TABLE IF NOT EXISTS topics (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			parent_id TEXT NULL,
			status TEXT NOT NULL DEFAULT 'pending'
		)`,
		`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			topic_id TEXT NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			FOREIGN KEY (topic_id) REFERENCES topics(id),
			UNIQUE(topic_id, path)
		)`,
		`CREATE TABLE IF NOT EXISTS chunks (
			id TEXT PRIMARY KEY,
			topic_id TEXT NOT NULL,
			document_id TEXT NOT NULL,
			chunk_index INTEGER NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY (topic_id) REFERENCES topics(id),
			FOREIGN KEY (document_id) REFERENCES documents(id)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			topic_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			requested_count INTEGER NOT NULL DEFAULT 0,
			generated_count INTEGER NOT NULL DEFAULT 0,
			token_budget INTEGER NOT NULL DEFAULT 0,
			tokens_used INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			error TEXT,
			FOREIGN KEY (topic_id) REFERENCES topics(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_documents_topic_id ON documents(topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_topic_id ON chunks(topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON chunks(document_id)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("execute statement: %w", err)
		}
	}

	util.Info("database migrations completed")
	return nil
}
