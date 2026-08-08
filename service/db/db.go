package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MishraShardendu22/quiz-gen/util"
	_ "github.com/mattn/go-sqlite3"
)

// initializes the database with proper configuration
// treated as singleton here simply because it has one *sql.DB per database.
func Init(dbPath string) (*sql.DB, error) {
	util.Info("initializing database", "path", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// set connection pool settings for SQLite
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// using, write ahead logging for all operations that modify the database,
	// busy_timeout is basically, if there is a lock for some variable wait for it to open for 5 seconds before throwing an error.
	// foreign_keys enables enforcement of foreign key constraints in SQLite for the current database connection
	// synchronous =
	// OFF - SQLite rarely calls fsync()
	// FULL - SQLite calls fsync() after every write operation
	// NORMAL - SQLite calls fsync() less frequently than FULL but more frequently than OFF.
	// fsync() is a system call that flushes data to disk, ensuring that it is physically written and not just cached in memory.
	// The synchronous setting determines how often SQLite calls fsync() to ensure data durability and integrity.
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			util.Warn("failed to set SQLite pragma", "pragma", p, "error", err.Error())
		}
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	util.Info("database initialized successfully")
	return db, nil
}

// creates all necessary tables and indexes
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
			FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
			UNIQUE(topic_id, path)
		)`,
		`CREATE TABLE IF NOT EXISTS chunks (
			id TEXT PRIMARY KEY,
			topic_id TEXT NOT NULL,
			document_id TEXT NOT NULL,
			chunk_index INTEGER NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
			FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			topic_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			requested_count INTEGER NOT NULL DEFAULT 0 CHECK(requested_count > 0),
			generated_count INTEGER NOT NULL DEFAULT 0 CHECK(generated_count >= 0),
			token_budget INTEGER NOT NULL DEFAULT 0 CHECK(token_budget > 0),
			tokens_used INTEGER NOT NULL DEFAULT 0 CHECK(tokens_used >= 0),
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			error TEXT,
			FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS questions (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			question TEXT NOT NULL,
			option_1 TEXT NOT NULL,
			option_2 TEXT NOT NULL,
			option_3 TEXT NOT NULL,
			option_4 TEXT NOT NULL,
			correct_answer INTEGER NOT NULL CHECK(correct_answer BETWEEN 0 AND 3),
			explanation TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS usage (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			prompt_tokens INTEGER NOT NULL CHECK(prompt_tokens >= 0),
			completion_tokens INTEGER NOT NULL CHECK(completion_tokens >= 0),
			total_tokens INTEGER NOT NULL CHECK(total_tokens >= 0),
			estimated_cost REAL NOT NULL CHECK(estimated_cost >= 0.0),
			created_at INTEGER NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			idempotency_key TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)`,

		// Primary keys are indexed automatically

		`CREATE INDEX IF NOT EXISTS idx_chunks_topic_id ON chunks(topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_topic_id ON sessions(topic_id)`,
		`CREATE INDEX IF NOT EXISTS idx_documents_topic_id ON documents(topic_id)`,

		`CREATE INDEX IF NOT EXISTS idx_usage_session_id ON usage(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_questions_session_id ON questions(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_idempotency_keys_session_id ON idempotency_keys(session_id)`,

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
