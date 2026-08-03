package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Init initializes the database with proper configuration
func Init(dbPath string) (*sql.DB, error) {
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

	return db, nil
}

// Migrate creates all necessary tables
func Migrate(db *sql.DB) error {
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
			topic_id TEXT,
			status TEXT,
			requested_count INTEGER,
			generated_count INTEGER,
			token_budget INTEGER,
			tokens_used INTEGER,
			error TEXT,
			created_at DATETIME,
			completed_at DATETIME,
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
	return nil
}
