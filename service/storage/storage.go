package storage

import (
	"database/sql"
	"fmt"

	"github.com/MishraShardendu22/quiz-gen/model"
	"github.com/MishraShardendu22/quiz-gen/service/loader"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// SyncTopicsDocumentsChunks synchronizes filesystem content with database incrementally
// Iterates over loaded topics, resolves/creates each topic's database UUID,
// then processes its documents with that UUID
func SyncTopicsDocumentsChunks(db *sql.DB, loadedTopics []model.LoadedTopic) error {
	// start tansaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// track which documents exist on filesystem
	// topicID|path -> true
	filesystemDocs := make(map[string]bool)

	// Process each topic and its documents
	for _, loadedTopic := range loadedTopics {

		// Look up or create topic
		var topicID string
		err := tx.QueryRow("SELECT id FROM topics WHERE name = ?", loadedTopic.Name).Scan(&topicID)

		// New topic: generate UUID and inse// topicID|path -> true
		if err == sql.ErrNoRows {
			topicID = uuid.Must(uuid.NewV7()).String()
			if _, err := tx.Exec(
				"INSERT INTO topics (id, name, status) VALUES (?, ?, ?)",
				topicID, loadedTopic.Name, "pending",
			); err != nil {
				return fmt.Errorf("insert topic %s: %w", loadedTopic.Name, err)
			}

			// Existing topic: use its UUID
		} else if err != nil {
			return fmt.Errorf("query topic %s: %w", loadedTopic.Name, err)
		}

		// Process this topic's documents with its resolved UUID
		for _, doc := range loadedTopic.Documents {
			key := topicID + "|" + doc.Path
			filesystemDocs[key] = true

			// Check if document already exists
			var existingID, existingHash string
			err := tx.QueryRow(
				"SELECT id, content_hash FROM documents WHERE topic_id = ? AND path = ?",
				topicID, doc.Path,
			).Scan(&existingID, &existingHash)

			// New document: insert it
			if err == sql.ErrNoRows {
				newID := uuid.Must(uuid.NewV7()).String()
				if _, err := tx.Exec(
					"INSERT INTO documents (id, topic_id, name, path, content_hash) VALUES (?, ?, ?, ?, ?)",
					newID, topicID, doc.Name, doc.Path, doc.Hash,
				); err != nil {
					return fmt.Errorf("insert document %s: %w", doc.Path, err)
				}

				// Chunk and store
				if err := insertChunks(tx, topicID, newID, doc.Content); err != nil {
					return err
				}

			// error handling for existing document
			} else if err != nil {
				return fmt.Errorf("query document %s: %w", doc.Path, err)

			// Document exists: check hash
			} else {
				if existingHash == doc.Hash {
					// Hash unchanged: skip completely
					continue
				}

				// Hash changed: delete old chunks, update document, re-chunk
				if _, err := tx.Exec("DELETE FROM chunks WHERE document_id = ?", existingID); err != nil {
					return fmt.Errorf("delete old chunks for %s: %w", doc.Path, err)
				}

				if _, err := tx.Exec(
					"UPDATE documents SET name = ?, content_hash = ? WHERE id = ?",
					doc.Name, doc.Hash, existingID,
				); err != nil {
					return fmt.Errorf("update document %s: %w", doc.Path, err)
				}

				if err := insertChunks(tx, topicID, existingID, doc.Content); err != nil {
					return err
				}
			}
		}
	}

	// Delete documents that exist in DB but not on filesystem
	allDBDocs, err := tx.Query("SELECT id, topic_id, path FROM documents")
	if err != nil {
		return fmt.Errorf("query all documents: %w", err)
	}
	defer allDBDocs.Close()

	for allDBDocs.Next() {
		var id, topicID, path string
		if err := allDBDocs.Scan(&id, &topicID, &path); err != nil {
			return err
		}

		key := topicID + "|" + path
		if !filesystemDocs[key] {
			// Document not on filesystem: delete its chunks and the document
			if _, err := tx.Exec("DELETE FROM chunks WHERE document_id = ?", id); err != nil {
				return fmt.Errorf("delete chunks for orphaned doc: %w", err)
			}
			if _, err := tx.Exec("DELETE FROM documents WHERE id = ?", id); err != nil {
				return fmt.Errorf("delete orphaned document: %w", err)
			}
		}
	}

	// Mark all topics as completed
	if _, err := tx.Exec("UPDATE topics SET status = ? WHERE status != ?", "completed", "completed"); err != nil {
		return fmt.Errorf("update topic status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// insertChunks chunks a document and inserts chunks into database
func insertChunks(tx *sql.Tx, topicID, documentID, content string) error {
	chunks := loader.SplitByHeading(content)
	for i, chunk := range chunks {
		if _, err := tx.Exec(
			"INSERT INTO chunks (id, topic_id, document_id, chunk_index, content) VALUES (?, ?, ?, ?, ?)",
			uuid.Must(uuid.NewV7()).String(),
			topicID,
			documentID,
			i,
			chunk,
		); err != nil {
			return fmt.Errorf("insert chunk: %w", err)
		}
	}
	return nil
}
